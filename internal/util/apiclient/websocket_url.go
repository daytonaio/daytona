package apiclient

import (
	"errors"
	"fmt"
	"regexp"
)

func GetWebSocketUrl(apiUrl string) (string, error) {
	hostRegex := regexp.MustCompile(`(https*)://(.*)`)

	matches := hostRegex.FindStringSubmatch(apiUrl)

	if len(matches) != 3 {
		return "", errors.New("invalid API URL")
	}

	switch matches[1] {
	case "http":
		return fmt.Sprintf("ws://%s", matches[2]), nil
	case "https":
		return fmt.Sprintf("wss://%s", matches[2]), nil
	}

	return "", errors.New("invalid API URL protocol")
}
