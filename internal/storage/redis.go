package storage

import (
    "context"
    "encoding/json"
    "time"

    "github.com/redis/go-redis/v9"

    "github.com/lehater/go-k8s-analyzer/internal/model"
)

type RedisStorage struct {
    client *redis.Client
}

func NewRedisStorage(addr string, dialTimeout, readTimeout, writeTimeout time.Duration) *RedisStorage {
    rdb := redis.NewClient(&redis.Options{
        Addr:         addr,
        DialTimeout:  dialTimeout,
        ReadTimeout:  readTimeout,
        WriteTimeout: writeTimeout,
    })
    return &RedisStorage{client: rdb}
}

func (s *RedisStorage) SaveMetric(ctx context.Context, key string, m model.Metric, ttl time.Duration) error {
    data, err := json.Marshal(m)
    if err != nil {
        return err
    }
    return s.client.Set(ctx, key, data, ttl).Err()
}
