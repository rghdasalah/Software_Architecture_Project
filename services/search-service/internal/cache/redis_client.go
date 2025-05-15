package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pb "search-service/proto"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

func NewRedisClient(addr string) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisClient{Client: rdb}
}

func (r *RedisClient) GetCachedSearch(ctx context.Context, key string) ([]*pb.Ride, error) {
	val, err := r.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // cache miss
	} else if err != nil {
		return nil, err
	}

	var rides []*pb.Ride
	err = json.Unmarshal([]byte(val), &rides)
	if err != nil {
		return nil, err
	}
	return rides, nil
}

func (r *RedisClient) CacheSearchResult(ctx context.Context, key string, rides []*pb.Ride, ttl time.Duration) error {
	data, err := json.Marshal(rides)
	if err != nil {
		return err
	}

	return r.Client.Set(ctx, key, data, ttl).Err()
}

func GenerateCacheKey(startHash, endHash string) string {
	return fmt.Sprintf("search:%s:%s", startHash, endHash)
}
