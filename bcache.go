package spikeotter

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"maps"
	"math"
	"time"

	"github.com/allegro/bigcache"
	"github.com/maypok86/otter/v2/stats"
)

type BCache struct {
	cache   *bigcache.BigCache
	source  *Source
	counter *stats.Counter
}

func NewBCache(uniques int, loadFactor int, maxsize int, expiry time.Duration, refresh time.Duration) *BCache {
	source := NewSource(uniques, loadFactor)
	counter := stats.NewCounter()
	opts := bigcache.DefaultConfig(expiry)
	// average serialize *Model is ~212 bytes
	avgSize := 212
	opts.MaxEntrySize = avgSize * 3
	opts.MaxEntriesInWindow = maxsize                                              // controls initial allocation (how many unique entries are present within LifeWindow)
	opts.HardMaxCacheSize = int(math.Ceil(float64(maxsize*avgSize) / 1024 / 1024)) // hard limit on memory size in MiB
	opts.OnRemove = func(key string, entry []byte) {
		counter.RecordEviction(uint32(len(entry)))
	}
	if expiry > 0 { // LifeWindow is how long before items are marked as expired
		opts.LifeWindow = expiry
	}
	if refresh > 0 { // CleanWindow is interval of cache cleanups (evicting expired entries)
		opts.CleanWindow = refresh
	}
	cache, err := bigcache.NewBigCache(opts)
	if err != nil {
		panic(err.Error())
	}

	return &BCache{
		cache:   cache,
		source:  source,
		counter: counter,
	}
}

func (c *BCache) StatsLoop(ctx context.Context) error {
	t := time.NewTicker(time.Second)
	defer t.Stop()
	var old stats.Stats
	var new stats.Stats

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			old = new
			new = c.counter.Snapshot()
			diff := new.Minus(old)
			slog.LogAttrs(ctx, slog.LevelInfo, "bcache",
				slog.Float64("hitRatio", diff.HitRatio()),
				slog.Uint64("loads", diff.Loads()),
				slog.Uint64("evictions", diff.Evictions),
				slog.Uint64("evictionWeight", diff.EvictionWeight),
				slog.Int("len", c.cache.Len()),
				slog.Int("capacity", c.cache.Capacity()),
			)
		}
	}
}

func (c *BCache) Load(ctx context.Context, id string) (*Model, error) {
	start := time.Now()
	model, err := c.source.Get(ctx, id)
	if err != nil {
		c.counter.RecordLoadFailure(time.Since(start))
		return nil, err
	}
	buf, err := json.Marshal(model)
	if err != nil {
		c.counter.RecordLoadFailure(time.Since(start))
		return nil, err
	}
	// there can be data race on concurrent Load here (value set might not come from latest Get).
	err = c.cache.Set(id, buf)
	if err != nil {
		c.counter.RecordLoadFailure(time.Since(start))
		return nil, err
	}
	c.counter.RecordLoadSuccess(time.Since(start))
	return model, nil
}

func (c *BCache) BulkLoad(ctx context.Context, ids []string) (map[string]*Model, error) {
	start := time.Now()
	models, err := c.source.BulkGet(ctx, ids)
	if err != nil {
		c.counter.RecordLoadFailure(time.Since(start))
		return nil, err
	}
	for id, model := range models {
		buf, err := json.Marshal(model)
		if err != nil {
			c.counter.RecordLoadFailure(time.Since(start))
			return nil, err
		}
		// there can be data race on concurrent Load here (value set might not come from latest Get).
		err = c.cache.Set(id, buf)
		if err != nil {
			c.counter.RecordLoadFailure(time.Since(start))
			return nil, err
		}
	}
	c.counter.RecordLoadSuccess(time.Since(start))
	return models, nil
}

func (c *BCache) Get(ctx context.Context, id string) (*Model, error) {
	model := &Model{}
	buf, err := c.cache.Get(id)
	if err == nil {
		c.counter.RecordHits(1)
		err := json.Unmarshal(buf, model)
		return model, err
	}
	if errors.Is(err, bigcache.ErrEntryNotFound) {
		c.counter.RecordMisses(1)
		return c.Load(ctx, id)
	}
	return nil, err
}

func (c *BCache) BulkGet(ctx context.Context, ids []string) (map[string]*Model, error) {
	models := make(map[string]*Model)
	misses := []string{}
	for _, id := range ids {
		model := &Model{}
		buf, err := c.cache.Get(id)
		if err == nil {
			c.counter.RecordHits(1)
			err := json.Unmarshal(buf, model)
			if err != nil {
				return nil, err
			}
			models[id] = model
			continue
		}
		if errors.Is(err, bigcache.ErrEntryNotFound) {
			c.counter.RecordMisses(1)
			misses = append(misses, id)
			continue
		}
		return nil, err
	}
	if len(misses) >= 0 {
		loads, err := c.BulkLoad(ctx, misses)
		if err != nil {
			return nil, err
		}
		maps.Copy(models, loads)
	}
	return models, nil
}

func (c *BCache) GenIDs() []string {
	return c.source.GenIDs()
}

func (c *BCache) Source() *Source {
	return c.source
}
