// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package devcontainer

import (
	"fmt"
	"strings"
)

func ConvertToArray(mergedCommands interface{}) ([]string, error) {
	switch mergedCommands := mergedCommands.(type) {
	case string:
		return []string{mergedCommands}, nil
	case []interface{}:
		var commandArray []string
		for _, arg := range mergedCommands {
			argString, ok := arg.(string)
			if !ok {
				return nil, fmt.Errorf("invalid command type: %v", arg)
			}
			commandArray = append(commandArray, argString)
		}
		return []string{strings.Join(commandArray, " ")}, nil
	case map[string]interface{}:
		var commandArray []string
		for _, command := range mergedCommands {
			switch command := command.(type) {
			case string:
				commandArray = append(commandArray, command)
			case []interface{}:
				var cmd []string
				for _, arg := range command {
					argString, ok := arg.(string)
					if !ok {
						return nil, fmt.Errorf("invalid command type: %v", command)
					}
					cmd = append(cmd, argString)
				}
				commandArray = append(commandArray, strings.Join(cmd, " "))
			}
		}
		return commandArray, nil
	}

	return nil, fmt.Errorf("invalid command type: %v", mergedCommands)
}
