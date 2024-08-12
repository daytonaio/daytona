//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package constants

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDaytonaScript(t *testing.T) {
	BaseUrl := "https://download.daytona.io/daytona"
	ExpectedString := "https://download.daytona.io/daytona"
	t.Run("Test_Get_Daytona_Script", func(t *testing.T) {
		script := GetDaytonaScript(BaseUrl)
		assert.Contains(t, script, ExpectedString, "the script should contain the correct base URL")
	})

}
