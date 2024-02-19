package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/daytonaio/daytona/cli/connection"
	server_proto "github.com/daytonaio/daytona/common/grpc/proto"
	"github.com/daytonaio/daytona/common/grpc/proto/types"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/spf13/cobra"
)

type GitCredentials struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}

var gitCredCmd = &cobra.Command{
	Use:     "git-cred get",
	Aliases: []string{"rev"},
	Args:    cobra.ExactArgs(1),
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		if args[0] != "get" {
			return
		}

		ctx := context.Background()
		result, err := parseFromStdin()
		host := result["host"]

		if err != nil || host == "" {
			fmt.Println("error parsing 'host' from stdin")
			return
		}

		conn, err := connection.Get(nil)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		serverClient := server_proto.NewServerClient(conn)

		var gitCredentials GitCredentials
		serverConfig, err := serverClient.GetConfig(ctx, &empty.Empty{})
		if err != nil {
			log.Fatal(err)
		}

		projectRepositoryUrl := getProjectRepositoryUrl()
		gitProviderId := getGitProviderFromRepositoryUrl(projectRepositoryUrl)

		if gitProviderId == "" {
			fmt.Println("error: unable to determine git provider")
			return
		}

		var gitProvider *types.GitProvider
		for _, provider := range serverConfig.GitProviders {
			if provider.Id == gitProviderId {
				gitProvider = provider
				break
			}
		}

		gitCredentials = GitCredentials{
			Username: gitProvider.Username,
			Token:    gitProvider.Token,
		}

		fmt.Println("username=" + gitCredentials.Username)
		fmt.Println("password=" + gitCredentials.Token)
	},
}

func getProjectRepositoryUrl() string {
	val, ok := os.LookupEnv("DAYTONA_WS_PROJECT_REPOSITORY_URL")
	if ok {
		return val
	} else {
		return ""
	}
}

func getGitProviderFromRepositoryUrl(url string) string {
	if strings.Contains(url, "github.com/") {
		return "github"
	} else if strings.Contains(url, "gitlab.com/") {
		return "gitlab"
	} else if strings.Contains(url, "bitbucket.org/") {
		return "bitbucket"
	} else {
		return ""
	}
}

func parseFromStdin() (map[string]string, error) {
	result := make(map[string]string)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) > 0 {
			tuple := strings.Split(line, "=")
			if len(tuple) == 2 {
				result[tuple[0]] = strings.TrimSpace(tuple[1])
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
