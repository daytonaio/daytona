package util

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// GetSshPrivateKeyPath returns the path to the private key and the password if it's encrypted
func GetSshPrivateKeyPath(privateKeyPath string) (string, *string, error) {
	keyContent, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return "", nil, err
	}

	_, err = ssh.ParsePrivateKey(keyContent)
	if err == nil {
		return privateKeyPath, nil, err
	}

	if err.Error() == (&ssh.PassphraseMissingError{}).Error() {
		fmt.Print("Enter password for key: ")
		password, err := term.ReadPassword(0)
		fmt.Println()
		if err != nil {
			return "", nil, err
		}

		stringPassword := string(password)

		return privateKeyPath, &stringPassword, nil
	}

	return "", nil, err
}
