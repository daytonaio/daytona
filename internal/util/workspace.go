package util

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/daytonaio/daytona/common/grpc/proto"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func GetFirstWorkspaceProjectName(conn *grpc.ClientConn, workspaceId string, projectName string) (string, error) {
	ctx := context.Background()

	client := proto.NewWorkspaceServiceClient(conn)
	workspaceInfoRequest := &proto.WorkspaceInfoRequest{
		Id: workspaceId,
	}

	response, err := client.Info(ctx, workspaceInfoRequest)
	if err != nil {
		log.Fatal(err)
	}

	if projectName == "" {
		if len(response.Projects) == 0 {
			return "", errors.New("no projects found in workspace")
		}

		return response.Projects[0].Name, nil
	}

	for _, project := range response.Projects {
		if project.Name == projectName {
			return project.Name, nil
		}
	}

	return "", errors.New("project not found in workspace")
}

func GetValidatedWorkspaceName(input string) (string, error) {
	input = strings.ToLower(input)

	input = strings.ReplaceAll(input, " ", "-")

	// Regular expression that catches letters, numbers, and dashes
	pattern := "^[a-zA-Z0-9-]+$"

	matched, err := regexp.MatchString(pattern, input)
	if err != nil {
		return "", err
	}

	if !matched {
		return "", fmt.Errorf("only letters, numbers, and dashes are allowed")
	}

	return input, nil
}

func GetValidatedUrl(input string) (string, error) {
	// Check if the input starts with a scheme (e.g., http:// or https://)
	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		// If it doesn't, add "http://" as the default scheme
		input = "https://" + input
	}

	// Try to parse the input as a URL
	parsedURL, err := url.Parse(input)
	if err != nil {
		return "", fmt.Errorf("input is not a valid URL")
	}

	// Validate the URL's host (domain) has a proper extension
	host := parsedURL.Host
	if !isValidTLD(host) {
		return "", fmt.Errorf("the URL does not have a valid TLD")
	}

	// If parsing was successful, return the fixed URL
	return parsedURL.String(), nil
}

func isValidTLD(host string) bool {
	// Regular expression to match common domain extensions like .com, .org, etc.
	extensionPattern := `\.([a-zA-Z]{2,6})$`
	regex := regexp.MustCompile(extensionPattern)

	// Check if the host (domain) matches the extension pattern
	return regex.MatchString(host)
}
