// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
	"time"
)

var timeLayout = "2006-01-02T15:04:05.999999999Z"

func FormatCreatedTime(input string) string {
	t, err := time.Parse(timeLayout, input)
	if err != nil {
		return "/"
	}

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

func FormatStatusTime(input string) string {
	t, err := time.Parse(timeLayout, input)
	if err != nil {
		return "stopped"
	}

	duration := time.Since(t)

	if duration < time.Minute {
		return "up < 1 minute"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "up 1 minute"
		}
		return fmt.Sprintf("up %d minutes", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "up 1 hour"
		}
		return fmt.Sprintf("up %d hours", hours)
	} else {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "up 1 day"
		}
		return fmt.Sprintf("up %d days", days)
	}
}
