// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"net/http"

	"github.com/daytonaio/daytona/pkg/sample"
)

func (s *Server) FetchSamples() ([]sample.Sample, *http.Response, error) {
	if s.config.SamplesIndexUrl == "" {
		return []sample.Sample{}, nil, nil
	}

	return sample.FetchSamples(s.config.SamplesIndexUrl)
}
