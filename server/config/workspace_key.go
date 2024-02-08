// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path"

	"golang.org/x/crypto/ssh"
)

func GetWorkspaceKey() (*ssh.Signer, error) {
	workspaceKeyPath, err := getWorkspaceKeyPath()
	if err != nil {
		return nil, err
	}

	workspaceKeyContent, err := os.ReadFile(workspaceKeyPath)
	if err != nil {
		return nil, err
	}

	workspaceKey, err := ssh.ParsePrivateKey([]byte(workspaceKeyContent))
	if err != nil {
		return nil, err
	}

	return &workspaceKey, err
}

func GetWorkspacePublicKey() (string, error) {
	workspaceKey, err := GetWorkspaceKey()
	if err != nil {
		return "", err
	}

	workspacePublicKey := string(ssh.MarshalAuthorizedKey((*workspaceKey).PublicKey()))

	return workspacePublicKey, nil
}

func GenerateWorkspaceKey() error {
	workspaceKeyPath, err := getWorkspaceKeyPath()
	if err != nil {
		return err
	}

	_, err = os.Stat(workspaceKeyPath)
	if os.IsNotExist(err) {
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

		err = os.WriteFile(workspaceKeyPath, privateKeyPem, 0600)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

func getWorkspaceKeyPath() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	dir := path.Join(userConfigDir, "daytona", "ssh")

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return "", err
	}

	workspaceKeyPath := path.Join(dir, workspaceKeyFileName)

	return workspaceKeyPath, nil
}
