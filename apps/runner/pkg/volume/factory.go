package volume

import (
	"fmt"
	"sync"
)

// providerFactories holds registered provider factories
var (
	providerFactories = make(map[string]func() Provider)
	factoriesMutex    sync.RWMutex
)

// RegisterProviderFactory registers a factory function for a provider type
func RegisterProviderFactory(providerType string, factory func() Provider) {
	factoriesMutex.Lock()
	defer factoriesMutex.Unlock()

	providerFactories[providerType] = factory
}

// CreateProvider creates a provider instance based on the type
func CreateProvider(providerType string) (Provider, error) {
	factoriesMutex.RLock()
	defer factoriesMutex.RUnlock()

	factory, exists := providerFactories[providerType]
	if !exists {
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}

	return factory(), nil
}
