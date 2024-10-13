// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"os/exec"
)

type ProviderRequirementLogs interface {
	CheckDockerRequirements() ([]RequirementStatus, error)
}

const (
	ProviderDocker       string = "docker"
	ProviderAWS          string = "aws"
	ProviderDigitalOcean string = "digitalocean"
)

type RequirementStatus struct {
	Name   string
	Met    bool
	Reason string
}

type LogProvider struct{}

func (l *LogProvider) CheckDockerRequirements() ([]RequirementStatus, error) {
	var results []RequirementStatus

	//check if docker is installed
	_, err := exec.LookPath("docker")
	if err != nil {
		results = append(results, RequirementStatus{
			Name:   "Docker installed",
			Met:    false,
			Reason: "Docker is not installed",
		})
	} else {
		results = append(results, RequirementStatus{
			Name:   "Docker installed",
			Met:    true,
			Reason: "Docker is installed",
		})
	}
	//check if docker is running
	cmd := exec.Command("docker", "info")
	err = cmd.Run()
	if err != nil {
		results = append(results, RequirementStatus{
			Name:   "Docker running",
			Met:    false,
			Reason: "Docker is not running",
		})
	} else {
		results = append(results, RequirementStatus{
			Name:   "Docker running",
			Met:    true,
			Reason: "Docker is running",
		})
	}

	return results, nil
}
