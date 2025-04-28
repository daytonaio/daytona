// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package ports

import (
	"fmt"
	"math/rand/v2"
	"net"
	"net/http"
	"strconv"
	"time"

	cmap "github.com/orcaman/concurrent-map/v2"
	log "github.com/sirupsen/logrus"
)

var portMap = cmap.New[bool]()

func GetAvailableEphemeralPort() (uint16, error) {
	var ephemeralPort uint16
	for {
		randomPort := 50000 + rand.IntN(60000-50000)
		ephemeralPort = uint16(randomPort)

		if IsPortAvailable(ephemeralPort) {
			log.Debug("EPHEMERAL PORT: " + strconv.FormatUint(uint64(ephemeralPort), 10))
			portMap.Set(strconv.FormatUint(uint64(ephemeralPort), 10), true)
			return ephemeralPort, nil
		}
	}
}

func IsPortAvailable(port uint16) bool {
	if portMap.Has(strconv.FormatUint(uint64(port), 10)) {
		return false
	}

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
