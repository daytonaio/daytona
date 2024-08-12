package apikeys

import (
	"encoding/base64"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestHashKeys(t *testing.T) {
	tests := []struct {
		Name         string
		Key          string
		ExpectedHash string
	}{

		{
			Name:         "testcase-1",
			Key:          "apikey1",
			ExpectedHash: "\xd4\xf7\x9b1?\x81\x06\xf5\xaf\x10\x8a\xd9o\xf5\x16\"-\xbfՠ\xabR\xf40\x8eK\x1a\xd1\xd7@\xde`",
		},
		{
			Name:         "testcase-2",
			Key:          "apikey2",
			ExpectedHash: "\x15\xfa\xc8\xfa\x1c\x99\x02%h\xb0\b\xb9\xdf\a\xb0KE5J\xc5\xcaG@\x04\x1d\x90L\xd3\xcf+9\xe3",
		},
		{
			Name:         "testcase-3",
			Key:          "apikey",
			ExpectedHash: "ly6\x95\x17\x1ey=}\x00\x80\xadw\x00\xa2\xbcP%i\x12\xce\xf2I, \x1e\x8e\xccT\xb2J\xb5",
		},
		{
			Name:         "testcase-4",
			Key:          "",
			ExpectedHash: "\xe3\xb0\xc4B\x98\xfc\x1c\x14\x9a\xfb\xf4șo\xb9$'\xaeA\xe4d\x9b\x93L\xa4\x95\x99\x1bxR\xb8U",
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			hashKey := HashKey(tc.Key)
			assert.NotNil(t, hashKey)
			assert.Equal(t, tc.ExpectedHash, hashKey)
		})
	}
}

func TestGenerateRandomKey(t *testing.T) {
	key := GenerateRandomKey()
	//check if the key is a valid base64 string
	_, err := base64.RawStdEncoding.DecodeString(key)
	assert.NoError(t, err, "generatd key is not a valid base64 string")

	//check if the key is a valid uuid
	uuidStr, err := base64.RawStdEncoding.DecodeString(key)
	assert.NoError(t, err)
	_, err = uuid.Parse(string(uuidStr))
	assert.NoError(t, err, "Generated key is not a valid UUID")

}

func TestEqualsKeyHashFromApi(t *testing.T) {
	tests := []struct {
		Name               string
		Key                string
		ExpectedKeyHash    string
		ExpectedComparison bool
	}{
		{
			Name:               "Case 1 - Valid key and incorrect hash",
			Key:                "apikey1",
			ExpectedKeyHash:    "invalidhash",
			ExpectedComparison: false,
		},
		{
			Name:               "Case 2 - Empty key and non-empty hash",
			Key:                "",
			ExpectedKeyHash:    "nonemptyhash",
			ExpectedComparison: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			comparisonResult := EqualsKeyHashFromApi(tc.Key, tc.ExpectedKeyHash)
			assert.Equal(t, tc.ExpectedComparison, comparisonResult, "expected comparism should match the actual comparism")
			assert.NotNil(t, comparisonResult, "actual comparison shouldn't be nil")
		})
	}
}
