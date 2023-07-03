package cache

import (
	"fmt"

	"github.com/allegro/bigcache"
)

func InitializeUserCache() (*bigcache.BigCache, error) {
	cache, err := InitializeCache()
	if err != nil {
		return nil, fmt.Errorf("error occured at InitializeUserCache while creating new chace error: %v", err)
	}
}
