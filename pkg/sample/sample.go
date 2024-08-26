// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package sample

import (
	"encoding/json"
	"io"
	"net/http"
)

type Sample struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
	GitUrl      string `json:"gitUrl" validate:"required"`
} // @name Sample

func FetchSamples(indexUrl string) ([]Sample, *http.Response, error) {
	var samples []Sample

	res, err := http.Get(indexUrl)
	if err != nil {
		return nil, res, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, res, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, res, err
	}

	err = json.Unmarshal(body, &samples)
	return samples, res, err
}
