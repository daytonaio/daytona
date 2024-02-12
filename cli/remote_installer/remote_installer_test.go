package remote_installer

import (
	"testing"

	"github.com/daytonaio/daytona/common/os"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/ssh"
)

type MockSession struct {
	mock.Mock
}

func (m *MockSession) Output(cmd string) ([]byte, error) {
	args := m.Called(cmd)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockSession) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockClient struct {
	mock.Mock
}

func (m *MockClient) NewSession() (*ssh.Session, error) {
	args := m.Called()
	session := args.Get(0).(ssh.Session)
	return &session, args.Error(1)
}

func TestDetectOs(t *testing.T) {
	expectedOutput := "Linux test 4.15.0-106-generic #107-Ubuntu SMP Thu Jun 4 11:27:52 UTC 2020 x86_64 x86_64 x86_64 GNU/Linux"

	mockSession := new(MockSession)
	mockSession.On("Output", "uname -a").Return([]byte(expectedOutput), nil)
	mockSession.On("Close").Return(nil)

	mockClient := new(MockClient)
	mockClient.On("NewSession").Return(mockSession, nil)

	installer := &RemoteInstaller{Client: mockClient}
	remoteOs, err := installer.DetectOs()

	mockSession.AssertExpectations(t)
	mockClient.AssertExpectations(t)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if *remoteOs != os.Linux_64_86 {
		t.Errorf("Expected os.Linux_64_86, but got %v", remoteOs)
	}
}

// func TestServerExists(t *testing.T) {
// 	mockSession := new(MockSession)
// 	mockSession.On("Output", "test -f /usr/local/bin/daytona").Return([]byte(""), nil)
// 	mockSession.On("Close").Return(nil)

// 	mockClient := new(MockClient)
// 	mockClient.On("NewSession").Return(mockSession, nil)

// 	installer := &RemoteInstaller{Client: mockClient}
// 	exists, err := installer.ServerExists(os.Linux_64_86)

// 	mockSession.AssertExpectations(t)
// 	mockClient.AssertExpectations(t)

// 	if err != nil {
// 		t.Errorf("Unexpected error: %v", err)
// 	}

// 	if !*exists {
// 		t.Errorf("Expected server to exist, but it does not")
// 	}
// }

func TestInstall(t *testing.T) {
	mockSession := new(MockSession)
	mockSession.On("Output", "curl -o /tmp/daytona_install.tar.gz https://example.com/linux_64_86_binary | tar -xz -C /tmp -f /tmp/daytona_install.tar.gz && mv /tmp/daytona /usr/local/bin").Return([]byte(""), nil)
	mockSession.On("Output", "chmod +x /usr/local/bin/daytona").Return([]byte(""), nil)
	mockSession.On("Close").Return(nil)

	mockClient := new(MockClient)
	mockClient.On("NewSession").Return(mockSession, nil)

	BinaryUrl_linux_64_86 := "https://example.com/linux_64_86_binary"
	BinaryUrl_linux_arm64 := "https://example.com/linux_arm64_binary"

	installer := &RemoteInstaller{
		Client: mockClient,
		BinaryUrl: map[os.OperatingSystem]string{
			os.Linux_64_86: BinaryUrl_linux_64_86,
			os.Linux_arm64: BinaryUrl_linux_arm64,
		},
	}
	err := installer.InstallBinary(os.Linux_64_86)

	mockSession.AssertExpectations(t)
	mockClient.AssertExpectations(t)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
