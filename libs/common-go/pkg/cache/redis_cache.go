// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Host         *string `envconfig:"HOST" mapstructure:"host"`
	Port         *int    `envconfig:"PORT" mapstructure:"port"`
	Username     *string `envconfig:"USERNAME" mapstructure:"username"`
	Password     *string `envconfig:"PASSWORD" mapstructure:"password"`
	TLS          *bool   `envconfig:"TLS" mapstructure:"tls"`
	Mode         *string `envconfig:"MODE" mapstructure:"mode"`
	ClusterNodes *string `envconfig:"CLUSTER_NODES" mapstructure:"cluster_nodes"`
}

type RedisCache[T any] struct {
	redis     redis.Cmdable
	keyPrefix string
}

type ValueObject[T any] struct {
	Value T `json:"value"`
}

var (
	redisClient redis.Cmdable
	redisOnce   sync.Once
	redisErr    error
)

func parseClusterNodes(nodesStr string) []string {
	parts := strings.Split(nodesStr, ",")
	addrs := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		if !strings.Contains(trimmed, ":") {
			trimmed = trimmed + ":6379"
		}
		addrs = append(addrs, trimmed)
	}
	return addrs
}

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

func NewRedisCache[T any](config *RedisConfig, keyPrefix string) (*RedisCache[T], error) {
	redisOnce.Do(func() {
		redisClient, redisErr = createClient(config)
	})
	if redisErr != nil {
		return nil, redisErr
	}

	return &RedisCache[T]{
		redis:     redisClient,
		keyPrefix: keyPrefix,
	}, nil
}

func createClient(config *RedisConfig) (redis.Cmdable, error) {
	mode := "single"
	if config.Mode != nil {
		mode = strings.ToLower(strings.TrimSpace(*config.Mode))
	}

	switch mode {
	case "single", "cluster":
	default:
		return nil, fmt.Errorf("unsupported redis mode %q: must be \"single\" or \"cluster\"", mode)
	}

	username := ""
	if config.Username != nil {
		username = *config.Username
	}

	password := ""
	if config.Password != nil {
		password = *config.Password
	}

	var tlsConfig *tls.Config
	if config.TLS != nil && *config.TLS {
		tlsConfig = &tls.Config{}
	}

	if mode == "cluster" {
		if config.ClusterNodes == nil || *config.ClusterNodes == "" {
			return nil, errors.New("cluster nodes are required when mode is cluster")
		}
		addrs := parseClusterNodes(*config.ClusterNodes)
		if len(addrs) == 0 {
			return nil, errors.New("cluster nodes are required when mode is cluster")
		}
		return redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:     addrs,
			Username:  username,
			Password:  password,
			TLSConfig: tlsConfig,
		}), nil
	}

	if config.Host == nil || config.Port == nil {
		return nil, errors.New("host and port are required")
	}

	return redis.NewClient(&redis.Options{
		Addr:      fmt.Sprintf("%s:%d", *config.Host, *config.Port),
		Username:  username,
		Password:  password,
		TLSConfig: tlsConfig,
	}), nil
}

// resetClient resets the global Redis client singleton. For testing only.
func resetClient() {
	redisClient = nil
	redisErr = nil
	redisOnce = sync.Once{}
}
