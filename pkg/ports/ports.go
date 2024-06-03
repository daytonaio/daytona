// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ports

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

func GetAvailableEphemeralPort() (uint16, error) {
	var ephemeralPort uint16
	for ephemeralPort = 50000; ephemeralPort < 60000; ephemeralPort++ {
		if IsPortAvailable(ephemeralPort) {
			log.Debug("EPHEMERAL PORT: " + strconv.FormatUint(uint64(ephemeralPort), 10))
			return ephemeralPort, nil
		}
	}
	return 0, errors.New("no more ephemeral ports available")
}

func IsPortAvailable(port uint16) bool {
	timeout := time.Millisecond * 50
	conn, err := net.DialTimeout("tcp", fmt.Sprintf(":%d", port), timeout)
	if err != nil {
		return true
	}
	if conn != nil {
		defer conn.Close()
	}
	return false
}

// This function checks if a service is ready on the specified port

func IsPortReady(port uint16) bool {
	client := &http.Client{Timeout: time.Second * 10}
	// Close the idle connections after the function returns
	defer client.CloseIdleConnections()

	response, err := client.Get(fmt.Sprintf("http://localhost:%d", port))
	if err != nil {
		return false
	}
	defer response.Body.Close()
	return response.StatusCode == http.StatusOK
}
