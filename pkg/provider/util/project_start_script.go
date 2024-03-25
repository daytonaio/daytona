package util

import "fmt"

func GetProjectStartScript(serverDownloadUrl string, apiKey string) string {
	return fmt.Sprintf(`curl -sfL -H "Authorization: Bearer %s" %s | sudo -E bash && daytona agent`, apiKey, serverDownloadUrl)
}
