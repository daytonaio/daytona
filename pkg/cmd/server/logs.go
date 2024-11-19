// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var followFlag bool
var fileFlag string
var localFlag bool

func init() {
	logsCmd.Flags().BoolVarP(&followFlag, "follow", "f", false, "Follow logs")
	logsCmd.Flags().StringVar(&fileFlag, "file", "", "Read specific log file from server log files")
	logsCmd.Flags().BoolVarP(&localFlag, "local", "l", false, "Read logs from local server log files")
}

var logsCmd = &cobra.Command{
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
		query += fmt.Sprintf("file=%s", fileFlag)
	}

	ws, res, err := apiclient_util.GetWebsocketConn(context.Background(), "/log/server", &activeProfile, &query)

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
	views.RenderBorderedMessage("Reading from local server log file...")

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

		selectedFile := selection.GetLogFileFromPrompt(logFiles)
		if selectedFile == nil {
			return nil
		}

		logFile = *selectedFile
	}

	var reader io.Reader
	if strings.HasSuffix(logFile, ".zip") || strings.HasSuffix(logFile, ".gz") {
		reader, err = readCompressedFile(logFile)
	} else {
		reader, err = os.Open(logFile)
	}
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	msgChan := make(chan []byte)
	errChan := make(chan error)

	go util.ReadLog(context.Background(), reader, followFlag, msgChan, errChan)

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
		if file.Name() == "daytona.log" || strings.HasPrefix(file.Name(), "daytona-") && (strings.HasSuffix(file.Name(), ".log") || strings.HasSuffix(file.Name(), ".zip") || strings.HasSuffix(file.Name(), ".gz")) {
			logFiles = append(logFiles, filepath.Join(dir, file.Name()))
		}
	}

	return logFiles, nil
}

func readCompressedFile(filePath string) (io.Reader, error) {
	zipFile, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, err
	}

	if len(zipFile.File) == 0 {
		return nil, fmt.Errorf("empty zip file")
	}

	return zipFile.File[0].Open()
}
