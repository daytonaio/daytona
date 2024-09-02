// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apiclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal"
	"github.com/daytonaio/daytona/pkg/api/middlewares"
	log "github.com/sirupsen/logrus"
)

type ApiErrorResponse struct {
	Error string `json:"error"`
}

func HandleErrorResponse(res *http.Response, requestErr error) error {
	if res == nil {
		return requestErr
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	contentType := res.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "text/html") {
		errorLogsDir, err := config.GetErrorLogsDir()
		if err != nil {
			return errors.New(string(body))
		}

		fileName := fmt.Sprintf("%s/api-error-%s.html", errorLogsDir, time.Now().Format("2006-01-02-15-04-05"))
		err = os.WriteFile(fileName, body, 0644)
		if err != nil {
			return errors.New(string(body))
		}

		return fmt.Errorf("an error occurred and an html page was returned. You can check the page at %s", fileName)
	}

	var errResponse ApiErrorResponse

	err = json.Unmarshal(body, &errResponse)
	if err != nil {
		return errors.New(string(body))
	}

	if !IsHealthCheckFailed(errors.New(errResponse.Error)) {
		checkVersionsMismatch(res)
	}

	return errors.New(errResponse.Error)
}

func checkVersionsMismatch(res *http.Response) {
	serverVersion := res.Header.Get(middlewares.SERVER_VERSION_HEADER)
	if internal.Version != serverVersion {
		log.Warn(fmt.Sprintf("Version mismatch detected. CLI is on version %s, Daytona Server is on version %s. To ensure maximum compatibility, please make sure the versions are aligned.", internal.Version, serverVersion))
	}
}
