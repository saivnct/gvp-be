package dao

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"log"
	"os"
	"strconv"
	"sync"
)

type Cache struct {
	RedisClient redis.Client
	RedSync     redsync.Redsync
}

var singletonCache *Cache
var onceCache sync.Once

func GetCache() *Cache {
	onceCache.Do(func() {
		fmt.Println("Init Cache...")

		redisDb, err := strconv.Atoi(os.Getenv("REDIS_DB"))
		if err != nil {
			log.Fatalf("Failed to read redis env: %v", err)
		}
		redisClient := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%v:%v", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
			Password: os.Getenv("REDIS_PASSWD"),
			DB:       redisDb,
		})

		// Create a pool with go-redis (or redigo) which is the pool redisync will
		// use while communicating with Redis. This can also be any pool that
		// implements the `redis.Pool` interface.
		pool := goredis.NewPool(redisClient)

		// Create an instance of redisync to be used to obtain a mutual exclusion
		// lock.
		rs := redsync.New(pool)

		singletonCache = &Cache{
			RedisClient: *redisClient,
			RedSync:     *rs,
		}
	})
	return singletonCache
}
