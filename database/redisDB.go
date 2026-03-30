package database

import (
	"api/config"
	"api/types"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	rdb  *redis.Client
	rctx context.Context
)

func ConnectRedis() error {
	opt, err := redis.ParseURL(config.Config("REDIS_URL"))
	if err != nil {
		fmt.Println(err)
	}
	rdb = redis.NewClient(opt)
	rctx = context.Background()
	res := rdb.Ping(rctx)
	if res.Err() != nil {
		return res.Err()
	}
	fmt.Println("Connected to Redis")
	return nil
}

func StoreMapping(link *types.LinkDTO) error {
	var result *redis.StatusCmd
	if link.Expiration == 0 {
		result = rdb.Set(rctx, link.ShortURL, link.LongURL, 0)
	} else {
		result = rdb.Set(rctx, link.ShortURL, link.LongURL, time.Duration(link.Expiration)*time.Second)
	}
	if result.Err() != nil {
		fmt.Println(result.Err())
		return result.Err()
	}
	return nil
}

func GetLongURL(shortURL string) (string, error) {
	result := rdb.Get(rctx, shortURL)
	if result.Err() != nil {
		return "", result.Err()
	}
	return result.Val(), nil
}

func IncrementClickCount(shortURL string) error {
    result := rdb.Incr(rctx, "clicks:"+shortURL)
    if result.Err() != nil {
        fmt.Println(result.Err())
        return result.Err()
    }
    return nil
}

func GetClickCount(shortURL string) (int64, error) {
    result := rdb.Get(rctx, "clicks:"+shortURL)
    if result.Err() == redis.Nil {
        return 0, nil // key doesn't exist yet = 0 clicks
    }
    if result.Err() != nil {
        return 0, result.Err()
    }
    count, err := result.Int64()
    if err != nil {
        return 0, err
    }
    return count, nil
}