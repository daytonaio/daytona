// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package logs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/constants"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/server/selection"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var followFlag bool
var fileFlag string
var localFlag bool

func init() {
	LogsCmd.AddCommand(listCmd)

	LogsCmd.Flags().BoolVarP(&followFlag, "follow", "f", false, "Follow logs")
	LogsCmd.Flags().StringVar(&fileFlag, "file", "", "Read specific log file")
	LogsCmd.Flags().BoolVarP(&localFlag, "local", "l", false, "Read local server log files")
}

var LogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Output Daytona Server logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			return err
		}

		query := ""
		if followFlag {
			query += "follow=true"
		}

		switch {
		case localFlag:
			return readLocalServerLogFile()
		default:
			return readRemoteServerLogFile(ctx, activeProfile, query)
		}
	},
}

func readRemoteServerLogFile(ctx context.Context, activeProfile config.Profile, query string) error {
	apiclient, err := apiclient_util.GetApiClient(&activeProfile)
	if err != nil {
		return err
	}

	logFiles, res, err := apiclient.ServerAPI.GetServerLogFiles(ctx).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	if len(logFiles) == 0 {
		return promptReadingLocalServerLogs("No server log files found.")
	}

	if fileFlag == "" && len(logFiles) == 1 {
		fileFlag = logFiles[0]
	}

	if fileFlag == "" {
		selectedFile := selection.GetLogFileFromPrompt(logFiles)
		if selectedFile == nil {
			return nil
		}

		fileFlag = *selectedFile
	}

	if fileFlag != "" {
		if query != "" {
			query += "&"
		}
		query += fmt.Sprintf("file=%s", filepath.Base(fileFlag))
	}

	ws, res, err := util.GetWebsocketConn(context.Background(), "/log/server", activeProfile.Api.Url, activeProfile.Api.Key, &query)
	if res.StatusCode == http.StatusNotFound {
		return apiclient_util.HandleErrorResponse(res, err)
	}
	if err != nil {
		log.Error(apiclient_util.HandleErrorResponse(res, err))

		if activeProfile.Id != "default" {
			return nil
		}

		return promptReadingLocalServerLogs("An error occurred while connecting to the server.")
	}

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			return nil
		}

		fmt.Println(string(msg))
	}
}

func promptReadingLocalServerLogs(info string) error {
	readLocalLogsFile := true
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().Description(info + " Would you like to read from local log files instead?").
				Value(&readLocalLogsFile),
		),
	).WithTheme(views.GetCustomTheme())
	formErr := form.Run()
	if formErr != nil {
		return formErr
	}

	if readLocalLogsFile {
		return readLocalServerLogFile()
	}

	return nil
}

func readLocalServerLogFile() error {
	cfg, err := server.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get server config: %w", err)
	}

	logFile := fmt.Sprintf("%s/%s", filepath.Dir(cfg.LogFile.Path), fileFlag)
	if fileFlag == "" {
		logDir := filepath.Dir(cfg.LogFile.Path)
		logFiles, err := getLocalServerLogFiles(logDir)
		if err != nil {
			return fmt.Errorf("failed to get log files: %w", err)
		}

		if len(logFiles) == 0 {
			return fmt.Errorf("no log files found in %s", logDir)
		}

		selectedFile := &logFiles[0]

		if len(logFiles) > 1 {
			selectedFile = selection.GetLogFileFromPrompt(logFiles)
			if selectedFile == nil {
				return nil
			}
		}

		logFile = *selectedFile
	}

	var reader io.Reader
	if regexp.MustCompile(constants.ZIP_LOG_FILE_NAME_SUFFIX_PATTERN).MatchString(logFile) {
		reader, err = logs.ReadCompressedFile(logFile)
	} else {
		reader, err = os.Open(logFile)
	}
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	msgChan := make(chan []byte)
	errChan := make(chan error)

	go logs.ReadLog(context.Background(), reader, followFlag, msgChan, errChan)

	for {
		select {
		case <-context.Background().Done():
			return nil
		case err := <-errChan:
			if err != nil {
				if err != io.EOF {
					return err
				}
				return nil
			}
		case msg := <-msgChan:
			fmt.Println(string(msg))
		}
	}
}

func getLocalServerLogFiles(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var logFiles []string
	for _, file := range files {
		if regexp.MustCompile(constants.LOG_FILE_NAME_PATTERN).MatchString(file.Name()) {
			logFiles = append(logFiles, filepath.Join(dir, file.Name()))
		}
	}

	return logFiles, nil
}
