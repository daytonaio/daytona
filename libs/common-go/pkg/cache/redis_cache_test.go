// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"reflect"
	"strings"
	"testing"

	"github.com/redis/go-redis/v9"
)

func stringPtr(v string) *string {
	return &v
}

func intPtr(v int) *int {
	return &v
}

func TestNewRedisCacheSingleModeDefault(t *testing.T) {
	resetClient()

	cache, err := NewRedisCache[string](&RedisConfig{
		Host: stringPtr("localhost"),
		Port: intPtr(6379),
	}, "default:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cache == nil {
		t.Fatal("expected cache to be non-nil")
	}
	if redisClient == nil {
		t.Fatal("expected redis client to be initialized")
	}
	if _, ok := redisClient.(*redis.Client); !ok {
		t.Fatal("expected single redis client")
	}
}

func TestNewRedisCacheSingleModeExplicit(t *testing.T) {
	resetClient()

	mode := "single"
	cache, err := NewRedisCache[string](&RedisConfig{
		Host: stringPtr("localhost"),
		Port: intPtr(6379),
		Mode: &mode,
	}, "explicit:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cache == nil {
		t.Fatal("expected cache to be non-nil")
	}
	if _, ok := redisClient.(*redis.Client); !ok {
		t.Fatal("expected explicit single mode to create single redis client")
	}
}

func TestNewRedisCacheClusterMode(t *testing.T) {
	resetClient()

	mode := "cluster"
	cache, err := NewRedisCache[string](&RedisConfig{
		Mode:         &mode,
		ClusterNodes: stringPtr("host1:7000,host2:7001"),
	}, "cluster:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cache == nil {
		t.Fatal("expected cache to be non-nil")
	}
	if redisClient == nil {
		t.Fatal("expected redis client to be initialized")
	}
	if _, ok := redisClient.(*redis.ClusterClient); !ok {
		t.Fatal("expected cluster redis client")
	}
}

func TestNewRedisCacheClusterMissingNodes(t *testing.T) {
	resetClient()

	mode := "cluster"
	cache, err := NewRedisCache[string](&RedisConfig{
		Mode: &mode,
	}, "cluster:")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "cluster nodes are required when mode is cluster" {
		t.Fatalf("unexpected error: %v", err)
	}
	if cache != nil {
		t.Fatal("expected cache to be nil on error")
	}
}

func TestNewRedisCacheSingleMissingHost(t *testing.T) {
	resetClient()

	cache, err := NewRedisCache[string](&RedisConfig{
		Port: intPtr(6379),
	}, "missing-host:")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "host and port are required" {
		t.Fatalf("unexpected error: %v", err)
	}
	if cache != nil {
		t.Fatal("expected cache to be nil on error")
	}
}

func TestNewRedisCacheSingleton(t *testing.T) {
	resetClient()

	first, err := NewRedisCache[string](&RedisConfig{
		Host: stringPtr("localhost"),
		Port: intPtr(6379),
	}, "first:")
	if err != nil {
		t.Fatalf("expected no error creating first cache, got %v", err)
	}

	second, err := NewRedisCache[string](&RedisConfig{
		Host: stringPtr("otherhost"),
		Port: intPtr(6380),
	}, "second:")
	if err != nil {
		t.Fatalf("expected no error creating second cache, got %v", err)
	}

	if first.redis != second.redis {
		t.Fatal("expected both caches to reuse the same redis client")
	}
}

func TestParseClusterNodes(t *testing.T) {
	resetClient()

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{name: "single node", input: "host1:7000", want: []string{"host1:7000"}},
		{name: "multiple nodes", input: "host1:7000,host2:7001,host3:7002", want: []string{"host1:7000", "host2:7001", "host3:7002"}},
		{name: "default port", input: "host1", want: []string{"host1:6379"}},
		{name: "trim whitespace", input: " host1:7000 , host2:7001 ", want: []string{"host1:7000", "host2:7001"}},
		{name: "empty", input: "", want: []string{}},
		{name: "commas only", input: ",,", want: []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseClusterNodes(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("parseClusterNodes(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestNewRedisCacheUnsupportedMode(t *testing.T) {
	resetClient()

	mode := "sentinel"
	_, err := NewRedisCache[string](&RedisConfig{
		Mode: &mode,
	}, "bad:")
	if err == nil {
		t.Fatal("expected error for unsupported mode")
	}
	if !strings.Contains(err.Error(), "unsupported redis mode") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewRedisCacheClusterCommaOnlyNodes(t *testing.T) {
	resetClient()

	mode := "cluster"
	_, err := NewRedisCache[string](&RedisConfig{
		Mode:         &mode,
		ClusterNodes: stringPtr(",,"),
	}, "bad:")
	if err == nil {
		t.Fatal("expected error for comma-only cluster nodes")
	}
	if !strings.Contains(err.Error(), "cluster nodes are required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewRedisCacheModeNormalization(t *testing.T) {
	resetClient()

	mode := " Cluster "
	cache, err := NewRedisCache[string](&RedisConfig{
		Mode:         &mode,
		ClusterNodes: stringPtr("host1:7000"),
	}, "norm:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cache == nil {
		t.Fatal("expected cache to be non-nil")
	}
	if _, ok := redisClient.(*redis.ClusterClient); !ok {
		t.Fatal("expected cluster client after mode normalization")
	}
}

func TestRedisCacheKeyPrefix(t *testing.T) {
	resetClient()

	cache, err := NewRedisCache[string](&RedisConfig{
		Host: stringPtr("localhost"),
		Port: intPtr(6379),
	}, "test:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cache == nil {
		t.Fatal("expected cache to be non-nil")
	}
	if cache.keyPrefix != "test:" {
		t.Fatalf("expected keyPrefix to be %q, got %q", "test:", cache.keyPrefix)
	}
}
