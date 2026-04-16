// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package coderun

import (
	"encoding/base64"
	"fmt"
)

type javascriptToolbox struct{}

func (t *javascriptToolbox) GetRunCommand(code string, argv []string) string {
	encodedCode := base64.StdEncoding.EncodeToString([]byte("process.argv.splice(1, 1);\n" + code))
	return fmt.Sprintf("printf '%%s' '%s' | base64 -d | node - %s", encodedCode, formatArgv(argv))
}
