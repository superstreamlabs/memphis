package memphis_cache

import (
	"context"
	"encoding/json"
	"fmt"
	"memphis/conf"
	"memphis/db"
	"memphis/models"

	"github.com/allegro/bigcache/v3"
)

var UCache UserCache
var configuration = conf.GetConfig()

type UserCache struct {
	Cache *MemphisCache
}

func InitializeUserCache() error {
	cache, err := New(context.Background(), configuration.USER_CACHE_LIFE_MINUTES, configuration.USER_CACHE_CLEAN_MINUTES, configuration.USER_CACHE_MAX_SIZE)
	if err != nil {
		UCache = UserCache{Cache: cache}
		return err
	}

	exists, users, err := db.GetAllUsersInDB()
	if err != nil {
		UCache = UserCache{Cache: cache}
		return err
	} else if !exists {
		UCache = UserCache{Cache: cache}
	}

	for _, user := range users {
		data, err := json.Marshal(user)
		if err != nil {
			UCache = UserCache{Cache: cache}
			return err
		}
		cache.Set(fmt.Sprint("%v:%v", user.Username, user.TenantName), data)
	}

	UCache = UserCache{Cache: cache}
	return nil

}

func GetUser(username, tenentName string) (models.User, error) {
	var user models.User
	data, err := UCache.Cache.Get(fmt.Sprint("%v:%v", username, tenentName))
	if err == bigcache.ErrEntryNotFound {
		_, userFromDB, err := db.GetUserByUsername(username, tenentName)
		if err != nil {
			return models.User{}, err
		}
		SetUser(userFromDB)
		return userFromDB, nil
	}

	err = json.Unmarshal(data, &user)
	if err != nil {
		return models.User{}, err
	}

	return user, nil

}

func SetUser(user models.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	err = UCache.Cache.Set(fmt.Sprint("%v:%v", user.Username, user.TenantName), data)
	return err
}

func DeleteUser(tenantName string, users []string) error {
	for _, user := range users {
		err := UCache.Cache.Delete(fmt.Sprint("%v:%v", user, tenantName))
		if err != nil {
			return err
		}
	}
	return nil
}
