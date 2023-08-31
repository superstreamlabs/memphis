package memphis_cache

import (
	"context"
	"time"

	"github.com/allegro/bigcache/v3"
)

type MemphisCache struct {
	Cache *bigcache.BigCache
}

func New(ctx context.Context, life_window, clean_window, cache_size int) (*MemphisCache, error) {
	cache_conf := bigcache.DefaultConfig(time.Duration(time.Duration(life_window) * time.Minute))
	cache_conf.CleanWindow = time.Duration(time.Duration(clean_window) * time.Minute)
	cache_conf.HardMaxCacheSize = cache_size

	cache, err := bigcache.New(ctx, cache_conf)

	return &MemphisCache{Cache: cache}, err
}

func (mc *MemphisCache) Get(key string) ([]byte, error) {
	return mc.Cache.Get(key)
}

func (mc *MemphisCache) Set(key string, data []byte) error {
	return mc.Cache.Set(key, data)
}

func (mc *MemphisCache) Delete(key string) error {
	return mc.Cache.Delete(key)
}

func (mc *MemphisCache) Iterator() *bigcache.EntryInfoIterator {
	return mc.Cache.Iterator()
}
