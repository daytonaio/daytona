package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type GitCredResponse struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}

var gitCredCmd = &cobra.Command{
	Use:     "git-cred get",
	Aliases: []string{"rev"},
	Args:    cobra.ExactArgs(1),
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "get":
			result, err := parseFromStdin()
			host := result["host"]
			if err != nil || host == "" {
				fmt.Println("error parsing 'host' from stdin")
				return
			}

			supervisorUrl := getSupervisorHostUrl()

			// grpc server..., get git url from env
			resp, err := http.Get(supervisorUrl + "/user/git-auth-token/" + host)
			if err != nil {
				log.Fatalln(err)
			}
			//We Read the response body on the line below.
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatalln(err)
			}

			var response GitCredResponse
			if err := json.Unmarshal(body, &response); err != nil { // Parse []byte to go struct pointer
				fmt.Println("Can not unmarshal JSON")
			}

			fmt.Println("username=" + response.Username)
			fmt.Println("password=" + response.Token)
		default:
			return
		}
	},
}

func getSupervisorHostUrl() string {
	val, ok := os.LookupEnv("WS_SUPERVISOR_HOST_URL")
	if !ok {
		return "http://172.17.0.1:63899"
	} else {
		return val
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
