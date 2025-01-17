// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"errors"
	"strconv"

	"github.com/daytonaio/daytona/pkg/apiclient"
)

func CreateServerPortValidator(config *apiclient.ServerConfig, portView *string, port *int32) func(string) error {
	return func(string) error {
		validatePort, err := strconv.Atoi(*portView)
		if err != nil {
			return errors.New("failed to parse port")
		}
		if validatePort < 0 || validatePort > 65535 {
			return errors.New("port out of range")
		}
		*port = int32(validatePort)

		if config.ApiPort == config.HeadscalePort {
			return errors.New("port conflict")
		}

		return nil
	}
}

func CreatePortValidator(portView *string, port *int32) func(string) error {
	return func(string) error {
		validatePort, err := strconv.Atoi(*portView)
		if err != nil {
			return errors.New("failed to parse port")
		}
		if validatePort < 0 || validatePort > 65535 {
			return errors.New("port out of range")
		}
		*port = int32(validatePort)

		return nil
	}
}

func CreateIntValidator(viewValue *string, value *int32) func(string) error {
	return func(string) error {
		validateInt, err := strconv.Atoi(*viewValue)
		if err != nil {
			return errors.New("failed to parse int")
		}

		if validateInt <= 0 {
			return errors.New("int out of range")
		}

		*value = int32(validateInt)

		return nil
	}
}
