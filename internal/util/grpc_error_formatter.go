package util

import (
	"fmt"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

type GRPCErrorFormatter struct {
}

func (f *GRPCErrorFormatter) Format(entry *log.Entry) ([]byte, error) {
	if strings.Contains(entry.Message, "rpc error") {
		re := regexp.MustCompile(`rpc error: code = (.*?) desc = (.*)`)
		matches := re.FindStringSubmatch(entry.Message)
		if len(matches) == 3 {
			code := matches[1]
			description := matches[2]
			switch code {
			case "Unavailable":
				return []byte("Daytona Agent is not running. Please run `daytona agent` first\n"), nil
			case "Unknown":
				return []byte(formatUnkownDescription(description)), nil
			default:
				return []byte(fmt.Sprintf("%s: %s\n", code, description)), nil
			}
		}
	}

	textFormatter := &log.TextFormatter{}

	return textFormatter.Format(entry)
}

func formatUnkownDescription(description string) string {
	if strings.Contains(description, "You cannot remove a running container") {
		return "You cannot remove a running workspace. Please stop the workspace first or force remove with `-f`\n"
	}

	if strings.Contains(description, "record not found") {
		return "Resource not found\n"
	}

	return strings.ToUpper(description[:1]) + description[1:] + "\n"
}
