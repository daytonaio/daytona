package jetbrains

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetIdeas(t *testing.T) {
	ideas := GetIdes()

	tests := []struct {
		id             Id
		expectedIDE    Ide
		version        string
		arch           string
		expectedIDEURL string
	}{
		{
			id:             CLion,
			expectedIDE:    clion,
			version:        "2024.1",
			arch:           "Amd64",
			expectedIDEURL: "https://download.jetbrains.com/cpp/CLion-2024.1.tar.gz",
		},
		{
			id:             IntelliJ,
			expectedIDE:    intellij,
			version:        "2024.1",
			arch:           "Arm64",
			expectedIDEURL: "https://download.jetbrains.com/idea/ideaIU-2024.1-aarch64.tar.gz",
		},
		{
			id:             GoLand,
			expectedIDE:    goland,
			version:        "2024.1",
			arch:           "Amd64",
			expectedIDEURL: "https://download.jetbrains.com/go/goland-2024.1.tar.gz",
		},
		{
			id:             PyCharm,
			expectedIDE:    pycharm,
			version:        "2024.1",
			arch:           "Arm64",
			expectedIDEURL: "https://download.jetbrains.com/python/pycharm-professional-2024.1-aarch64.tar.gz",
		},
		{
			id:             PhpStorm,
			expectedIDE:    phpstorm,
			version:        "2024.1",
			arch:           "Amd64",
			expectedIDEURL: "https://download.jetbrains.com/webide/PhpStorm-2024.1.tar.gz",
		},
		{
			id:             WebStorm,
			expectedIDE:    webstorm,
			version:        "2024.1",
			arch:           "Arm64",
			expectedIDEURL: "https://download.jetbrains.com/webstorm/WebStorm-2024.1-aarch64.tar.gz",
		},
		{
			id:             Rider,
			expectedIDE:    rider,
			version:        "2024.1",
			arch:           "Amd64",
			expectedIDEURL: "https://download.jetbrains.com/rider/JetBrains.Rider-2024.1.tar.gz",
		},
		{
			id:             RubyMine,
			expectedIDE:    rubymine,
			version:        "2024.1",
			arch:           "Arm64",
			expectedIDEURL: "https://download.jetbrains.com/ruby/RubyMine-2024.1-aarch64.tar.gz",
		},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("Tetsing IDE %s", tc.expectedIDE.Name), func(t *testing.T) {
			ide, exists := ideas[tc.id]
			assert.True(t, exists, "IDE sould exist in map")
			assert.Equal(t, tc.expectedIDE, ide, "expedted IDE should match the returned IDE")
			assert.NotNil(t, tc.expectedIDE, "IDE should not be nil")
			//actual url assertions for system architecture
			var actualURL string
			switch tc.arch {
			case "Amd64":
				actualURL = fmt.Sprintf(ide.UrlTemplates.Amd64, tc.version)
			case "Arm64":
				actualURL = fmt.Sprintf(ide.UrlTemplates.Arm64, tc.version)
			default:
				t.Fatalf("unsupported architecture: %s", tc.arch)
			}
			assert.Equal(t, tc.expectedIDEURL, actualURL, "URL should match expected URL format")
			assert.NotEmpty(t, ide.ProductCode, "ProductCode should not be empty")
			assert.NotEmpty(t, ide.Name, "IDE name should not be empty")

		})
	}
}
