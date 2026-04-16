// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package coderun

import "fmt"

type CodeToolbox interface {
	GetRunCommand(code string, argv []string) string
}

func GetToolbox(language string) (CodeToolbox, error) {
	switch language {
	case "python":
		return &pythonToolbox{}, nil
	case "javascript":
		return &javascriptToolbox{}, nil
	case "typescript":
		return &typescriptToolbox{}, nil
	default:
		return nil, fmt.Errorf("unsupported language: %s", language)
	}
}
