// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// MergeEnvVars merges multiple environment variables maps into one
// and resolves environment variables that are referenced in the values.
// If an environment variable is not found, a warning is logged.
// The order of the maps is important. Latter maps will override
// the values of the previous ones.
func MergeEnvVars(envVars ...map[string]string) map[string]string {
	vars := map[string]string{}

	for _, env := range envVars {
		for k, v := range env {
			vars[k] = v
		}
	}

	for k, v := range vars {
		if strings.HasPrefix(v, "$") {
			env, ok := os.LookupEnv(v[1:])
			if ok {
				vars[k] = env
			} else {
				log.Warnf("Environment variable %s not found", v[1:])
			}
		} else {
			vars[k] = v
		}
	}

	return vars
}

func GetEnvVarsFromShell() map[string]string {
	envMap := map[string]string{}

	for _, env := range os.Environ() {
		kv := strings.SplitN(env, "=", 2)
		if len(kv) == 2 {
			envMap[kv[0]] = kv[1]
		}
	}
	return envMap
}
