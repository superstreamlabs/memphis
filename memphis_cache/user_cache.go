package memphis_cache

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/memphisdev/memphis/conf"
	"github.com/memphisdev/memphis/db"
	"github.com/memphisdev/memphis/models"

	"github.com/allegro/bigcache/v3"
)

var UCache UserCache
var configuration = conf.GetConfig()
var logger func(string, ...interface{})

type UserCache struct {
	Cache *MemphisCache
}

func InitializeUserCache(logger_func func(string, ...interface{})) error {
	logger = logger_func

	cache, err := New(context.Background(), configuration.USER_CACHE_LIFE_MINUTES, configuration.USER_CACHE_CLEAN_MINUTES, configuration.USER_CACHE_MAX_SIZE_MB)
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
		cache.Set(fmt.Sprintf("%v:%v", user.Username, user.TenantName), data)
	}

	UCache = UserCache{Cache: cache}
	return nil
}

func GetUser(username, tenentName string, forceGetFromDb bool) (bool, models.User, error) {
	var user models.User
	if forceGetFromDb {
		exist, userFromDB, db_err := db.GetUserByUsername(username, tenentName)
		if db_err != nil {
			return exist, models.User{}, db_err
		}
		SetUser(userFromDB)
		return exist, userFromDB, nil
	}
	data, err := UCache.Cache.Get(fmt.Sprintf("%v:%v", username, tenentName))
	if err != nil {
		exist, userFromDB, db_err := db.GetUserByUsername(username, tenentName)
		if db_err != nil {
			return exist, models.User{}, db_err
		}
		if err == bigcache.ErrEntryNotFound {
			SetUser(userFromDB)
			return exist, userFromDB, nil
		}
		logger("[tenant: %v][user: %v]error while using cache, error: %v", tenentName, username, err)
		return exist, userFromDB, nil
	}

	err = json.Unmarshal(data, &user)
	if err != nil {
		exist, userFromDB, db_err := db.GetUserByUsername(username, tenentName)
		if db_err != nil {
			return exist, models.User{}, db_err
		}
		logger("[tenant: %v][user: %v]error while using unmarshal in the cache, error: %v", tenentName, username, err)
		return exist, userFromDB, nil
	}

	return true, user, nil
}

func SetUser(user models.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	err = UCache.Cache.Set(fmt.Sprintf("%v:%v", user.Username, user.TenantName), data)
	return err
}

func DeleteUser(tenantName string, users []string) error {
	for _, user := range users {
		err := UCache.Cache.Delete(fmt.Sprintf("%v:%v", user, tenantName))
		if err == bigcache.ErrEntryNotFound {
			return nil
		} else if err != nil {
			return err
		}
	}
	return nil
}
