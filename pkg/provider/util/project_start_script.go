package util

import "fmt"

func GetProjectStartScript(serverDownloadUrl string) string {
	return fmt.Sprintf("curl -sfL %s | sudo -E bash && daytona agent", serverDownloadUrl)
}
