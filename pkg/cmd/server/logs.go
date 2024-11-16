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
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/views"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var followFlag bool
var fileFlag bool

func init() {
	logsCmd.Flags().BoolVarP(&followFlag, "follow", "f", false, "Follow logs")
	logsCmd.Flags().BoolVar(&fileFlag, "file", false, "Read logs from local server log file")
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Output Daytona Server logs",
	RunE: func(cmd *cobra.Command, args []string) error {
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
		case fileFlag:
			return readServerLogFile()

		default:
			ws, res, err := apiclient.GetWebsocketConn(context.Background(), "/log/server", &activeProfile, &query)

			if err != nil {
				log.Error(apiclient.HandleErrorResponse(res, err))

				if activeProfile.Id != "default" {
					return nil
				}

				readLogsFile := true
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewConfirm().Description("An error occurred while connecting to the server. Would you like to read from local log file instead?").
							Value(&readLogsFile),
					),
				).WithTheme(views.GetCustomTheme())
				formErr := form.Run()
				if formErr != nil {
					return formErr
				}

				if readLogsFile {
					return readServerLogFile()
				}
				return nil

			}

			for {
				_, msg, err := ws.ReadMessage()
				if err != nil {
					return nil
				}

				fmt.Println(string(msg))
			}
		}
	},
}

func readServerLogFile() error {
	views.RenderBorderedMessage("Reading from server log file...")
	cfg, err := server.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get server config: %w", err)
	}

	logDir := filepath.Dir(cfg.LogFile.Path)
	logFiles, err := getLogFiles(logDir)
	if err != nil {
		return fmt.Errorf("failed to get log files: %w", err)
	}

	if len(logFiles) == 0 {
		return fmt.Errorf("no log files found in %s", logDir)
	}

	selectedFile, err := selectLogFile(logFiles)
	if err != nil {
		return fmt.Errorf("failed to select log file: %w", err)
	}

	var reader io.Reader
	if strings.HasSuffix(selectedFile, ".zip") {
		reader, err = readCompressedFile(selectedFile)
	} else {
		reader, err = os.Open(selectedFile)
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

func getLogFiles(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var logFiles []string
	for _, file := range files {
		if file.Name() == "daytona.log" || strings.HasPrefix(file.Name(), "daytona-") && (strings.HasSuffix(file.Name(), ".log") || strings.HasSuffix(file.Name(), ".zip")) {
			logFiles = append(logFiles, filepath.Join(dir, file.Name()))
		}
	}

	return logFiles, nil
}

func selectLogFile(files []string) (string, error) {
	var options []huh.Option[string]
	for _, file := range files {
		options = append(options, huh.Option[string]{Key: filepath.Base(file), Value: filepath.Base(file)})
	}

	var selected string
	prompt := huh.NewSelect[string]().
		Title("Select a log file to read").
		Options(options...).
		Value(&selected)

	err := prompt.Run()
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if filepath.Base(file) == selected {
			return file, nil
		}
	}

	return "", fmt.Errorf("selected file not found")
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
