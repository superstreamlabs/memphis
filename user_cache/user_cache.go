package user_cache

import (
	"context"
	"encoding/json"
	"fmt"
	"memphis/db"
	"memphis/models"
	"time"

	"github.com/allegro/bigcache/v3"
)

var UCache UserCache

type UserCache struct {
	Cache *bigcache.BigCache
}

func InitializeUserCache() error {
	conf := bigcache.DefaultConfig(10 * time.Second)
	conf.HardMaxCacheSize = 1
	conf.CleanWindow = time.Second * 1
	cache, err := bigcache.New(context.Background(), conf)
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
		fmt.Printf("new cache entity with key - %v:%v and size - %v \r\n", user.Username, user.TenantName, cap(data))

		cache.Set(fmt.Sprint("%v:%v", user.Username, user.TenantName), data)
	}

	UCache = UserCache{Cache: cache}
	return nil

}

func Get(username, tenentName string) (models.User, error) {
	var user models.User
	fmt.Printf("Get entity with key - %v:%v \r\n", username, tenentName)
	data, err := UCache.Cache.Get(fmt.Sprint("%v:%v", username, tenentName))
	if err == bigcache.ErrEntryNotFound {
		fmt.Printf("used the DB : %v \r\n", err)
		_, userFromDB, err := db.GetUserByUsername(username, tenentName)
		if err != nil {
			return models.User{}, err
		}
		Set(userFromDB)
		return userFromDB, nil
	}

	err = json.Unmarshal(data, &user)
	if err != nil {
		return models.User{}, err
	}

	return user, nil

}

func Set(user models.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	err = UCache.Cache.Set(fmt.Sprint("%v:%v", user.Username, user.TenantName), data)
	return err
}
