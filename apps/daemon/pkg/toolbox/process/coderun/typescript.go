// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package coderun

import (
	"encoding/base64"
	"fmt"
	"strings"
)

type typescriptToolbox struct{}

func (t *typescriptToolbox) GetRunCommand(code string, argv []string) string {
	encodedCode := base64.StdEncoding.EncodeToString([]byte("process.argv.splice(1, 1);\n" + code))

	parts := []string{
		`_f=/tmp/dtn_$$.ts`,
		fmt.Sprintf(`printf '%%s' '%s' | base64 -d > "$_f"`, encodedCode),
		fmt.Sprintf(`npm_config_loglevel=error npx ts-node -T --ignore-diagnostics 5107 -O '{"module":"CommonJS"}' "$_f" %s`, formatArgv(argv)),
		`_dtn_ec=$?`,
		`rm -f "$_f"`,
		`exit $_dtn_ec`,
	}

	return strings.Join(parts, "; ")
}
