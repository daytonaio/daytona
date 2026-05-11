// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// Config is the env-driven configuration for daytona-session-daemon.
type Config struct {
	// Networking. Loopback by default — the daemon relies on the authenticated
	// 3-hop proxy chain (proxy -> runner -> daytona-daemon's /proxy/:port re-proxy
	// -> loopback) for external access. See the plan's "Auth on the daemon" section.
	BindAddr string `envconfig:"SESSION_DAEMON_BIND_ADDR" default:"127.0.0.1"`
	Port     int    `envconfig:"SESSION_DAEMON_PORT" default:"2281"`
	LogLevel string `envconfig:"SESSION_DAEMON_LOG_LEVEL" default:"info"`

	// Workspace + bundle paths.
	WorkspaceRoot      string `envconfig:"SESSION_DAEMON_USER_NODE_MODULES_ROOT" default:"/workspace"`
	NodeBundleRoot     string `envconfig:"SESSION_DAEMON_NODE_BUNDLE_ROOT" default:"/usr/lib/daytona/repl_host"`
	PythonInterpreter  string `envconfig:"SESSION_DAEMON_PYTHON" default:"python3"`
	NodeInterpreter    string `envconfig:"SESSION_DAEMON_NODE" default:"node"`
	WorkerScriptDir    string `envconfig:"SESSION_DAEMON_WORKER_SCRIPT_DIR" default:"/tmp"`
	HostScriptCacheDir string `envconfig:"SESSION_DAEMON_HOST_SCRIPT_CACHE_DIR" default:"/tmp"`

	// Per-engine concurrency caps. The 4x difference between TS and Python is intentional
	// and reflects the memory floor of each engine — see plan §7. Bash isolates are
	// virtual just-bash shells (no subprocess, no V8 heap), so they're the cheapest
	// of the three and carry a much higher default cap.
	TSMaxContexts   int `envconfig:"SESSION_DAEMON_TS_MAX_SESSIONS" default:"64"`
	PyMaxContexts   int `envconfig:"SESSION_DAEMON_PY_MAX_SESSIONS" default:"16"`
	BashMaxContexts int `envconfig:"SESSION_DAEMON_BASH_MAX_SESSIONS" default:"128"`

	// Default + ceiling for TS context memory limit (in MB).
	TSDefaultMemoryMB int `envconfig:"SESSION_DAEMON_TS_DEFAULT_MEMORY_MB" default:"128"`
	TSMaxMemoryMB     int `envconfig:"SESSION_DAEMON_TS_MAX_MEMORY_MB" default:"512"`

	// PyWarmPoolSize is the number of CPython subprocesses the daemon keeps
	// pre-spawned, ready to be claimed by the next CreateSession call. ~300ms
	// of `python3 + import` cost per process moves out of the request path —
	// see plan "transient-context perf" and the runtime benchmark in /tmp/timing.py.
	// 0 disables the pool (each context spawns its own python on-demand).
	PyWarmPoolSize int `envconfig:"SESSION_DAEMON_PY_WARM_POOL_SIZE" default:"2"`

	// Defense-in-depth: in-sandbox idle-context sweeper (intentionally 1.5x the API's
	// idle TTL). The API stamps its own idleTtlSeconds via ApiIdleTtlHintSeconds at
	// sandbox-create time so we can warn on inverted ratios at boot.
	ContextIdleTTLSeconds int `envconfig:"SESSION_DAEMON_IDLE_TTL_SECONDS" default:"5400"`
	ContextGCIntervalSec  int `envconfig:"SESSION_DAEMON_GC_INTERVAL_SECONDS" default:"60"`
	ApiIdleTtlHintSeconds int `envconfig:"SESSION_DAEMON_API_IDLE_TTL_SECONDS_HINT" default:"0"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to process env config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Port)
	}
	if c.TSMaxContexts < 1 {
		return fmt.Errorf("SESSION_DAEMON_TS_MAX_SESSIONS must be >= 1")
	}
	if c.PyMaxContexts < 1 {
		return fmt.Errorf("SESSION_DAEMON_PY_MAX_SESSIONS must be >= 1")
	}
	if c.BashMaxContexts < 1 {
		return fmt.Errorf("SESSION_DAEMON_BASH_MAX_SESSIONS must be >= 1")
	}
	if c.TSDefaultMemoryMB <= 0 || c.TSDefaultMemoryMB > c.TSMaxMemoryMB {
		return fmt.Errorf("SESSION_DAEMON_TS_DEFAULT_MEMORY_MB must be in (0, %d]", c.TSMaxMemoryMB)
	}
	return nil
}
