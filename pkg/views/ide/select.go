// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ide

import (
	"github.com/daytonaio/daytona/cmd/daytona/config"
)

func GetIdeIdFromPrompt(ideList []config.Ide) string {
	chosenIdeId := ""
	choiceChan := make(chan string)

	go Render(ideList, choiceChan)

	chosenIdeId = <-choiceChan

	return chosenIdeId
}
