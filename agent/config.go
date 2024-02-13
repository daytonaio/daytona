// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"

	log "github.com/sirupsen/logrus"
)

type IConfig struct {
	ReverseProxy struct {
		Hostname string `envconfig:"DAYTONA_PROXY_HOSTNAME" validate:"required"`
		Port     int    `envconfig:"DAYTONA_PROXY_PORT" validate:"required"`
		AuthKey  string `envconfig:"DAYTONA_PROXY_AUTH_KEY" validate:"required"`
	}
}

var Config IConfig

func init() {
	err := envconfig.Process("", &Config)
	if err != nil {
		log.Error(err)
		os.Exit(2)
	}

	var validate = validator.New()
	err = validate.Struct(&Config)
	if err != nil {
		log.Error(err)
		os.Exit(2)
	}
}
