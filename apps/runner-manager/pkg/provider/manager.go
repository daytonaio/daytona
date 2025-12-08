// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package provider

import (
	"errors"
	"sync"

	"github.com/daytonaio/runner-manager/pkg/provider/aws"
	"github.com/daytonaio/runner-manager/pkg/provider/k8s"
)

var (
	instance *ProviderManager
	once     sync.Once
)

// ProviderManager manages the runner provider instances
type ProviderManager struct {
	provider IRunnerProvider
}

// GetInstance returns the singleton instance of ProviderManager
func GetInstance() *ProviderManager {
	once.Do(func() {
		instance = &ProviderManager{}
	})
	return instance
}

// SetProvider sets the active provider
func (pm *ProviderManager) SetProvider(providerType string, namespace string, waitTimeout int, kubeconfig string) error {
	switch providerType {
	case "kubernetes", "k8s":
		config := k8s.K8sProviderConfig{
			Namespace:   namespace,
			WaitTimeout: waitTimeout,
			Kubeconfig:  kubeconfig,
		}
		provider, err := k8s.NewK8sProvider(config)
		if err != nil {
			return errors.New("failed to create k8s provider: " + err.Error())
		}
		pm.provider = provider
	case "aws":
		pm.provider = aws.NewAwsProvider()
	default:
		return errors.New("unsupported provider type: " + providerType)
	}
	return nil
}

// GetProvider returns the currently active provider
func (pm *ProviderManager) GetProvider() (IRunnerProvider, error) {
	if pm.provider == nil {
		return nil, errors.New("no provider configured")
	}
	return pm.provider, nil
}
