package spikeotter

import (
	"context"
	"log/slog"
	"time"

	"github.com/maypok86/otter/v2"
	"github.com/maypok86/otter/v2/stats"
)

type Cache struct {
	cache   *otter.Cache[string, *Model]
	source  *Source
	counter *stats.Counter
}

func NewCache(uniques int, loadFactor int, maxsize int, expiry time.Duration, refresh time.Duration) *Cache {
	source := NewSource(uniques, loadFactor)
	counter := stats.NewCounter()
	opts := &otter.Options[string, *Model]{
		MaximumSize:   maxsize,
		StatsRecorder: counter,
	}
	if expiry > 0 {
		opts.ExpiryCalculator = otter.ExpiryAccessing[string, *Model](expiry)
	}
	if refresh > 0 {
		opts.RefreshCalculator = otter.RefreshWriting[string, *Model](refresh)
	}
	cache := otter.Must(opts)

	return &Cache{
		cache:   cache,
		source:  source,
		counter: counter,
	}
}

func (c *Cache) StatsLoop(ctx context.Context) error {
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
			slog.LogAttrs(ctx, slog.LevelInfo, "ticker",
				slog.Float64("hitRatio", diff.HitRatio()),
				slog.Uint64("loads", diff.Loads()),
				slog.Uint64("eviction", diff.Evictions),
				slog.Int("estimatedSize", c.cache.EstimatedSize()),
			)
		}
	}
}

func (c *Cache) Get(ctx context.Context, id string) (*Model, error) {
	model, err := c.cache.Get(ctx, id, otter.LoaderFunc[string, *Model](c.source.Get))
	if err != nil {
		return nil, err
	}
	return model, nil
}

func (c *Cache) BulkGet(ctx context.Context, ids []string) (map[string]*Model, error) {
	models, err := c.cache.BulkGet(ctx, ids, otter.BulkLoaderFunc[string, *Model](c.source.BulkGet))
	if err != nil {
		return nil, err
	}
	return models, nil
}

func (c *Cache) GenIDs() []string {
	return c.source.GenIDs()
}

func (c *Cache) Source() *Source {
	return c.source
}
