// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"fmt"
	"strings"
)

const (
	defaultMouseButton  = "left"
	middleMouseButton   = "middle"
	defaultScrollAmount = 1
	scrollDirectionUp   = "up"
	scrollDirectionDown = "down"
)

func normalizeMouseButton(button string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(button))
	if normalized == "" {
		return defaultMouseButton, nil
	}

	switch normalized {
	case defaultMouseButton, "right", middleMouseButton:
		return normalized, nil
	default:
		return "", fmt.Errorf("unsupported mouse button %q: expected one of left, right, middle", button)
	}
}

// robotgoMouseButton translates our canonical "middle" to robotgo's "center".
func robotgoMouseButton(canonical string) string {
	if canonical == middleMouseButton {
		return "center"
	}
	return canonical
}

func normalizeScrollDirection(direction string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(direction))

	switch normalized {
	case scrollDirectionUp, scrollDirectionDown:
		return normalized, nil
	default:
		return "", fmt.Errorf("unsupported scroll direction %q: expected up or down", direction)
	}
}

func normalizeScrollAmount(amount int) (int, error) {
	if amount < 0 {
		return 0, fmt.Errorf("scroll amount must be greater than or equal to 0")
	}

	if amount == 0 {
		return defaultScrollAmount, nil
	}

	return amount, nil
}
