//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package constants

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDaytonaScript(t *testing.T) {
	baseUrl := "http://localhost:8080/daytona"
	expectedString := "http://localhost:8080/daytona"
	t.Run("Test_Get_Daytona_Script", func(t *testing.T) {
		script := GetDaytonaScript(baseUrl)
		assert.Contains(t, script, expectedString, "the script should contain the correct base URL")
	})

}
