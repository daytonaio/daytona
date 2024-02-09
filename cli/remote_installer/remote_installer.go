package remote_installer

import (
	"fmt"

	"github.com/daytonaio/daytona/common/os"
	"golang.org/x/crypto/ssh"
)

type SSHConfig struct {
	Hostname       string
	User           string
	Password       string
	PrivateKeyPath string
}

type SSHSession interface {
	Close() error
	Output(cmd string) ([]byte, error)
}

type SSHClient interface {
	NewSession() (*ssh.Session, error)
}

type RemoteInstaller struct {
	Client     SSHClient
	Password   string
	BinaryUrl  map[os.OperatingSystem]string
	Downloader DownloaderType
}

type DownloaderType int

const (
	DownloaderCurl DownloaderType = iota
	DownloaderWget
)

func (s *RemoteInstaller) InstallBinary(os os.OperatingSystem) error {
	session, err := s.Client.NewSession()
	if err != nil {
		return err
	}
	defer (*session).Close()

	url, ok := s.BinaryUrl[os]
	if !ok {
		return fmt.Errorf("url for os %d not found", os)
	}

	var cmd string

	// todo: separate into multiple cmd calls

	if s.Downloader == DownloaderCurl {
		cmd = fmt.Sprintf("curl -Lo daytona %s && (echo '%s' | sudo -S chmod +x daytona 2>/dev/null) && (echo '%s' | sudo -S mv daytona /usr/local/bin/ 2>/dev/null) && rm -f daytona", url, s.Password, s.Password)
	} else {
		cmd = fmt.Sprintf("wget -O daytona %s && (echo '%s' | sudo -S chmod +x daytona 2>/dev/null) && (echo '%s' | sudo -S mv daytona /usr/local/bin/ 2>/dev/null) && rm -f daytona", url, s.Password, s.Password)
	}

	_, err = (*session).Output(cmd)
	if err != nil {
		return err
	}

	return nil
}

func (s *RemoteInstaller) RegisterDaemon(remoteOs os.OperatingSystem) error {
	session, err := s.Client.NewSession()
	if err != nil {
		return err
	}
	defer (*session).Close()

	switch remoteOs {
	case os.Darwin_64_86:
		fallthrough
	case os.Darwin_arm64:
		fallthrough
	case os.Linux_64_86:
		fallthrough
	case os.Linux_arm64:
		output, err := (*session).Output("echo $(daytona server startup > /dev/null 2>&1; echo $?)")
		if err != nil {
			return err
		}

		if string(output) == "0\n" { // Exit status 0
			return nil
		} else {
			return fmt.Errorf("unexpected exit status: %s", string(output))
		}
	default:
		return fmt.Errorf("unexpected os: %d", remoteOs)
	}
}

func (s *RemoteInstaller) InstallDocker(remoteOS os.OperatingSystem) error {
	session, err := s.Client.NewSession()
	if err != nil {
		return err
	}
	defer (*session).Close()

	var cmd string

	switch remoteOS {
	case os.Darwin_64_86:
		fallthrough
	case os.Darwin_arm64:
		fallthrough
	case os.Linux_64_86:
		fallthrough
	case os.Linux_arm64:
		if s.Downloader == DownloaderCurl {
			cmd = "curl -fsSL https://get.docker.com | sh"
		} else {
			cmd = "wget -qO- https://get.docker.com | sh"
		}
	default:
		return fmt.Errorf("unexpected os: %d", remoteOS)
	}

	err = (*session).Run(cmd)
	if err != nil {
		return err
	}

	return nil
}

func (s *RemoteInstaller) AddUserToDockerGroup(user string) error {
	session, err := s.Client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	cmd := fmt.Sprintf("echo $(echo '%s' | sudo -S usermod -aG docker %s  > /dev/null 2>&1; echo $?)", s.Password, user)
	output, _ := (*session).CombinedOutput(cmd)

	if string(output) == "0\n" {
		return nil
	} else {
		return nil
	}
}

func (s *RemoteInstaller) DetectOs() (*os.OperatingSystem, error) {
	session, err := s.Client.NewSession()
	if err != nil {
		return nil, err
	}
	defer (*session).Close()

	output, err := (*session).Output("uname -a")
	if err != nil {
		return nil, err
	}

	return os.OSFromUnameA(string(output))
}

func (s *RemoteInstaller) ServerRegistered() (bool, error) {
	session, err := s.Client.NewSession()
	if err != nil {
		return false, err
	}
	defer session.Close()

	output, _ := (*session).CombinedOutput("echo $(systemctl --user is-active daytona-server.service > /dev/null 2>&1; echo $?)")

	if string(output) == "0\n" {
		return true, nil
	} else {
		return false, nil
	}
}

func (s *RemoteInstaller) SudoPasswordRequired() (bool, error) {
	session, err := s.Client.NewSession()
	if err != nil {
		return false, err
	}
	defer session.Close()

	output, _ := (*session).CombinedOutput("echo $(sudo -n true > /dev/null 2>&1; echo $?)")

	if string(output) == "0\n" {
		return false, nil
	} else {
		return true, nil
	}
}

func (s *RemoteInstaller) DockerInstalled() (bool, error) {
	session, err := s.Client.NewSession()
	if err != nil {
		return false, err
	}
	defer session.Close()

	output, _ := (*session).CombinedOutput("echo $(docker -v > /dev/null 2>&1; echo $?)")

	if string(output) == "0\n" {
		return true, nil
	} else {
		return false, nil
	}
}

func (s *RemoteInstaller) CurlInstalled() (bool, error) {
	session, err := s.Client.NewSession()
	if err != nil {
		return false, err
	}
	defer session.Close()

	output, _ := (*session).CombinedOutput("echo $(curl -V > /dev/null 2>&1; echo $?)")

	if string(output) == "0\n" {
		return true, nil
	} else {
		return false, nil
	}
}

func (s *RemoteInstaller) WgetInstalled() (bool, error) {
	session, err := s.Client.NewSession()
	if err != nil {
		return false, err
	}
	defer session.Close()

	output, _ := (*session).CombinedOutput("echo $(wget -V > /dev/null 2>&1; echo $?)")

	if string(output) == "0\n" {
		return true, nil
	} else {
		return false, nil
	}
}

func (s *RemoteInstaller) RemoveBinary(remoteOS os.OperatingSystem) error {
	session, err := s.Client.NewSession()
	if err != nil {
		return err
	}
	defer (*session).Close()

	switch remoteOS {
	case os.Darwin_64_86:
		fallthrough
	case os.Darwin_arm64:
		fallthrough
	case os.Linux_64_86:
		fallthrough
	case os.Linux_arm64:

		cmd := fmt.Sprintf("echo $(echo '%s' | sudo -S rm /usr/local/bin/daytona > /dev/null 2>&1; echo $?)", s.Password)
		output, err := (*session).CombinedOutput(cmd)
		if err != nil {
			return err
		}

		if string(output) == "0\n" { // Exit status 0
			return nil
		} else {
			return fmt.Errorf("unexpected exit status: %s", string(output))
		}
	default:
		return fmt.Errorf("unexpected os: %d", remoteOS)
	}
}

func (s *RemoteInstaller) RemoveDaemon(remoteOS os.OperatingSystem) error {
	session, err := s.Client.NewSession()
	if err != nil {
		return err
	}
	defer (*session).Close()

	switch remoteOS {
	case os.Darwin_64_86:
		fallthrough
	case os.Darwin_arm64:
		fallthrough
	case os.Linux_64_86:
		fallthrough
	case os.Linux_arm64:
		output, _ := (*session).CombinedOutput("echo $((systemctl --user stop daytona-server.service && systemctl --user disable daytona-server.service && rm $HOME/.config/systemd/user/daytona-server.service) > /dev/null 2>&1; echo $?)")
		if err != nil {
			return err
		}

		if string(output) == "0\n" { // Exit status 0
			return nil
		} else {
			return fmt.Errorf("unexpected exit status: %s", string(output))
		}
	default:
		return fmt.Errorf("unexpected os: %d", remoteOS)
	}
}
