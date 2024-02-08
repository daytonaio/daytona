package config_ssh_key

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path"

	"golang.org/x/crypto/ssh"
)

func GetPrivateKeyPath() string {
	return path.Join(os.Getenv("HOME"), ".ssh", "id_daytona")
}

func GeneratePrivateKey() error {
	privateKeyPath := GetPrivateKeyPath()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	privateKeyDer := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privateKeyDer,
	}

	privateKeyPem := pem.EncodeToMemory(&privateKeyBlock)
	if privateKeyPem == nil {
		return errors.New("failed to encode private key")
	}

	err = os.WriteFile(privateKeyPath, privateKeyPem, 0600)
	if err != nil {
		return err
	}

	return nil
}

func GetPublicKey() (string, error) {
	privateKeyPath := GetPrivateKeyPath()
	privateKeyContent, err := os.ReadFile(privateKeyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", errors.New("private key not found")
		}
		return "", err
	}

	privateKey, err := ssh.ParsePrivateKey([]byte(privateKeyContent))
	if err != nil {
		return "", err
	}

	return string(ssh.MarshalAuthorizedKey(privateKey.PublicKey())), nil
}

func SetPrivateKey(privateKey string) error {
	privateKeyPath := GetPrivateKeyPath()

	return os.WriteFile(privateKeyPath, []byte(privateKey), 0600)
}

func GetPrivateKey() (*ssh.Signer, error) {
	privateKeyPath := GetPrivateKeyPath()
	privateKeyContent, err := os.ReadFile(privateKeyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("private key not found")
		}
		return nil, err
	}

	privateKey, err := ssh.ParsePrivateKey([]byte(privateKeyContent))
	if err != nil {
		return nil, err
	}

	return &privateKey, nil
}

func DeletePrivateKey() error {
	privateKeyPath := GetPrivateKeyPath()

	return os.Remove(privateKeyPath)
}
