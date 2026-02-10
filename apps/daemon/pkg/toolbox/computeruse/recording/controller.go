// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"github.com/daytonaio/daemon/pkg/recording"
)

type RecordingController struct {
	recordingService *recording.RecordingService
}

func NewRecordingController(recordingService *recording.RecordingService) *RecordingController {
	return &RecordingController{
		recordingService: recordingService,
	}
}
