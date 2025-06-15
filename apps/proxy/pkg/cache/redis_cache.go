// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/daytonaio/proxy/cmd/proxy/config"
	"github.com/redis/go-redis/v9"
)

type RedisCache[T any] struct {
	redis     *redis.Client
	keyPrefix string
}

type ValueObject[T any] struct {
	Value T `json:"value"`
}

var client *redis.Client

func (c *RedisCache[T]) Set(ctx context.Context, key string, value T, expiration time.Duration) error {
	jsonValue, err := json.Marshal(ValueObject[T]{Value: value})
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, c.keyPrefix+key, string(jsonValue), expiration).Err()
}

func (c *RedisCache[T]) Has(ctx context.Context, key string) (bool, error) {
	err := c.redis.Get(ctx, c.keyPrefix+key).Err()
	if err == nil {
		return true, nil
	}

	if err == redis.Nil {
		return false, nil
	}

	return false, err
}

func (c *RedisCache[T]) Get(ctx context.Context, key string) (*T, error) {
	value, err := c.redis.Get(ctx, c.keyPrefix+key).Result()
	if err != nil {
		return nil, err
	}
	var result ValueObject[T]
	err = json.Unmarshal([]byte(value), &result)
	if err != nil {
		return nil, err
	}
	return &result.Value, nil
}

func (c *RedisCache[T]) Delete(ctx context.Context, key string) error {
	return c.redis.Del(ctx, c.keyPrefix+key).Err()
}

func NewRedisCache[T any](config *config.RedisConfig, keyPrefix string) (*RedisCache[T], error) {
	if config.Host == nil || config.Port == nil {
		return nil, errors.New("host and port are required")
	}

	password := ""
	if config.Password != nil {
		password = *config.Password
	}

	if client == nil {
		client = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", *config.Host, *config.Port),
			Password: password,
		})
	}

	return &RedisCache[T]{
		redis:     client,
		keyPrefix: keyPrefix,
	}, nil
}
