package cache

import (
	"fmt"
	"memphis/conf"
	"time"

	"github.com/allegro/bigcache"
)

var configuration = conf.GetConfig()

func InitializeCache() (*bigcache.BigCache, error) {
	cache, err := bigcache.NewBigCache(bigcache.DefaultConfig(10 * time.Minute))
	if err != nil {
		return nil, fmt.Errorf("error occured at InitializeCache while creating new chace error: %v", err)
	}

	return cache, nil

}

func (bc *bigcache.BigCache) Set(key string, value []byte) error {
	return bc.Set(key, value)
}

func (bc *bigcache.BigCache) Get(key string) ([]byte, error) {
	return bc.Get(key)
}

func (bc *bigcache.BigCache) Delete(key string) error {
	return bc.Delete(key)
}
