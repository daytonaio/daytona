//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0
package apikeys

import (
	"encoding/base64"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGenerateRandomKey(t *testing.T) {
	key := GenerateRandomKey()
	//check if the key is a valid base64 string
	_, err := base64.RawStdEncoding.DecodeString(key)
	assert.NoError(t, err, "generated key is not a valid base64 string")

	//check if the key is a valid uuid
	uuidStr, err := base64.RawStdEncoding.DecodeString(key)
	assert.NoError(t, err)
	_, err = uuid.Parse(string(uuidStr))
	assert.NoError(t, err, "generated key is not a valid UUID")

}
