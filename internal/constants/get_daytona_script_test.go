package constants

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDaytonaScript(t *testing.T) {
	tests := []struct {
		Name           string
		BaseUrl        string
		ExpectedString string
	}{
		{
			Name:           "replace default base URL with custom URL",
			BaseUrl:        "https://custom.url/daytona",
			ExpectedString: "https://custom.url/daytona",
		},
		{
			Name:           "replace with localhost URL",
			BaseUrl:        "http://localhost:8080/daytona",
			ExpectedString: "http://localhost:8080/daytona",
		},
		{
			Name:           "malformed URL",
			BaseUrl:        "htp:/bad-url",
			ExpectedString: "htp:/bad-url",
		},
		{
			Name:           "no substitution needed when default URL is used",
			BaseUrl:        "https://download.daytona.io/daytona",
			ExpectedString: "https://download.daytona.io/daytona",
		},
		{
			Name:           "trailing slash in base URL",
			BaseUrl:        "https://example.com/daytona/",
			ExpectedString: "https://example.com/daytona/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			script := GetDaytonaScript(tt.BaseUrl)
			assert.Contains(t, script, tt.ExpectedString, "the script should contain the correct base URL")
			assert.NotNil(t, script, "the script should not be nil")
			assert.NotEmpty(t, script, "the script should not be empty")
			// Length check
			assert.Greater(t, len(script), 100, "The script should have a reasonable length")
		})
	}
}
