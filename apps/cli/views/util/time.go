// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package util

import (
	"fmt"
	"time"
)

var timeLayout = "2006-01-02T15:04:05.999999999Z07:00"

func GetTimeSinceLabelFromString(input string) string {
	t, err := time.Parse(timeLayout, input)
	if err != nil {
		return "/"
	}

	return GetTimeSinceLabel(t)
}

func GetTimeSinceLabel(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "< 1 minute ago"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}
