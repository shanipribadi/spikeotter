package spikeotter

import (
	"context"
	"time"

	"github.com/maypok86/otter/v2"
)

type Cache struct {
	cache  *otter.Cache[string, *Model]
	source *Source
}

func NewCache() *Cache {
	source := NewSource()
	cache := otter.Must(
		&otter.Options[string, *Model]{
			MaximumSize:      1000000,
			ExpiryCalculator: otter.ExpiryCreating[string, *Model](time.Minute),
		})

	return &Cache{
		cache:  cache,
		source: source,
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
