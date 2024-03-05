package targets

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/server/config"
)

func GetTargets() (map[string]provider.ProviderTarget, error) {
	c, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(c.TargetsFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]provider.ProviderTarget{}, nil
		}
		return nil, err
	}
	defer file.Close()

	targets := map[string]provider.ProviderTarget{}
	err = json.NewDecoder(file).Decode(&targets)
	if err != nil {
		return nil, err
	}

	return targets, nil
}

func GetTarget(targetName string) (*provider.ProviderTarget, error) {
	targets, err := GetTargets()
	if err != nil {
		return nil, err
	}

	target, ok := targets[targetName]
	if !ok {
		return nil, errors.New("target not found")
	}

	return &target, nil
}

func SetTarget(target provider.ProviderTarget) error {
	targets, err := GetTargets()
	if err != nil {
		return err
	}

	targets[target.Name] = target

	return saveTargets(targets)
}

func RemoveTarget(targetName string) error {
	targets, err := GetTargets()
	if err != nil {
		return err
	}

	delete(targets, targetName)

	return saveTargets(targets)
}

func saveTargets(targets map[string]provider.ProviderTarget) error {
	c, err := config.GetConfig()
	if err != nil {
		return err
	}

	file, err := os.Create(c.TargetsFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	content, err := json.MarshalIndent(targets, "", "  ")
	if err != nil {
		return err
	}

	_, err = file.Write(content)
	return err
}
