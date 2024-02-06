// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"os"
	"strings"
)

var scriptContent string

func createTemporaryScriptFile() (*os.File, error) {
	// Create a temporary file in the current directory
	file, err := os.CreateTemp(".", "script-*.sh")
	if err != nil {
		return nil, err
	}

	SetScriptContent()

	// Write the script content to the file
	_, err = file.WriteString(scriptContent)
	if err != nil {
		file.Close() // Close the file if there's an error
		return nil, err
	}

	return file, nil
}

func SetScriptContent() {
	var result string

	tempScriptContent := `
	#!/bin/bash

	# Define the service file path
	USER_SERVICE_DIR="$HOME/.config/systemd/user"
	USER_SERVICE_PATH="$HOME/.config/systemd/user/daytona-agent.service"
	LOG_FILE="$HOME/daytona-agent-install.log"
	
	# Create the service file content
	SERVICE_CONTENT="[Unit]
	Description=Daytona Agent
	
	[Service]
	Type=simple
	ExecStart=/usr/local/bin/daytona agent
	Restart=always
	StandardOutput=syslog
	StandardError=syslog
	SyslogIdentifier=dagentservice
	
	[Install]
	WantedBy=default.target"
	
	# Check if the service file already exists for the user
	if [ -f "$USER_SERVICE_PATH" ]; then
		echo "Service file '$USER_SERVICE_PATH' already exists. Aborting."
		exit 1
	fi
	
	# Create the directory if it doesn't exist
	if [ ! -d "$USER_SERVICE_DIR" ]; then
		mkdir -p "$USER_SERVICE_DIR"
	fi
	
	# Create the service file for user
	echo "$SERVICE_CONTENT" > "$USER_SERVICE_PATH"
	
	# Reload systemd daemon for the root user
	systemctl daemon-reload
	
	# Enable and start the service for root
	systemctl --user enable daytona-agent.service
	systemctl --user start daytona-agent.service
	
	# Check if the service is running
	if systemctl --user is-active daytona-agent.service >/dev/null 2>&1; then
		echo "Daytona Agent service is installed and running." | tee -a "$LOG_FILE"
	else
		echo "Daytona Agent service failed to start. Check the logs for details." | tee -a "$LOG_FILE"
	fi
	`
	result = strings.Replace(tempScriptContent, "<NEEDLE_STRING>", "", 1)
	scriptContent = result
}
