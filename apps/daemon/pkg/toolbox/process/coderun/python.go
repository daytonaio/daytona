// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package coderun

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
)

//go:embed matplotlib_wrapper.py
var matplotlibWrapper string

type pythonToolbox struct{}

var matplotlibImportPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?m)^[^#]*import\s+matplotlib`),
	regexp.MustCompile(`(?m)^[^#]*from\s+matplotlib`),
	regexp.MustCompile(`(?m)^[^#]*__import__\s*\(\s*['"]matplotlib['"]`),
	regexp.MustCompile(`(?m)^[^#]*importlib\.import_module\s*\(\s*['"]matplotlib['"]`),
	regexp.MustCompile(`(?m)^[^#]*loader\.load_module\s*\(\s*['"]matplotlib['"]`),
	regexp.MustCompile(`(?m)^[^#]*sys\.modules\[['"]matplotlib['"]\]`),
}

func (t *pythonToolbox) GetRunCommand(code string, argv []string) string {
	encodedCode := base64.StdEncoding.EncodeToString([]byte(code))

	if isMatplotlibImported(code) {
		wrappedCode := strings.Replace(matplotlibWrapper, "{encoded_code}", encodedCode, 1)
		encodedCode = base64.StdEncoding.EncodeToString([]byte(wrappedCode))
	}

	return fmt.Sprintf("printf '%%s' '%s' | base64 -d | python3 -u - %s", encodedCode, formatArgv(argv))
}

func isMatplotlibImported(code string) bool {
	for _, pattern := range matplotlibImportPatterns {
		if pattern.MatchString(code) {
			return true
		}
	}

	return false
}
