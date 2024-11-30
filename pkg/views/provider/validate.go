// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

type DockerTargetOptions struct {
	SockPath         string `json:"Sock Path"`
	Name             string `json:"Name"`
	RemoteHostname   string `json:"Remote Hostname"`
	RemotePassword   string `json:"Remote Password"`
	RemotePort       int    `json:"Remote Port"`
	RemoteUser       string `json:"Remote User"`
	WorkspaceDataDir string `json:"Workspace Data Dir"`
}

type DigitalOceanTargetOptions struct {
	AuthToken string `json:"Auth Token"`
	DiskSize  int    `json:"Disk Size"`
	Image     string `json:"Image"`
	Region    string `json:"Region"`
	Size      string `json:"Size"`
}

type AWSTargetOptions struct {
	ListOfAvailableInstanceTypes string `json:"List of available instance types"`
	Region                       string `json:"Region"`
	SecretAccessKey              string `json:"Secret Access Key"`
	VolumeSize                   int    `json:"Volume Size"`
	VolumeType                   string `json:"Volume Type"`
}

type AzureTargetOptions struct {
	ClientID       string `json:"Client ID"`
	ClientSecret   string `json:"Client Secret"`
	DiskSize       int    `json:"Disk Size"`
	DiskType       string `json:"Disk Type"`
	ImageURN       string `json:"Image URN"`
	Region         string `json:"Region"`
	ResourceGroup  string `json:"Resource Group"`
	SubscriptionID string `json:"Subscription ID"`
	TenantID       string `json:"Tenant ID"`
	VmSize         string `json:"VM Size"`
}

type FlyTargetOptions struct {
	DiskSize int    `json:"Disk Size"`
	OrgSlug  string `json:"Org Slug"`
	Region   string `json:"Region"`
	Size     string `json:"Size"`
}

type GCPTargetOptions struct {
	DiskSize       int    `json:"Disk Size"`
	DiskType       string `json:"Disk Type"`
	MachineType    string `json:"Machine Type"`
	ProjectID      string `json:"Project ID"`
	VMImage        string `json:"VM Image"`
	Zone           string `json:"Zone"`
	CredentialFile string `json:"Credential File"`
}

type HetznerTargetOptions struct {
	APIToken           string `json:"API Token"`
	DiskImage          string `json:"Disk Image"`
	DiskSize           int    `json:"Disk Size"`
	Location           string `json:"Location"`
	LocationServerType string `json:"Location Server Type"`
}

func ParseJSON(data []byte, v interface{}) error {
	if err := json.Unmarshal(data, v); err == nil {
		return nil
	}
	return errors.New("input is not a valid JSON")
}

func ValidateJSONAgainstStruct(jsonStr string, expectedStruct interface{}) error {
	var jsonMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &jsonMap); err != nil {
		return fmt.Errorf("invalid json key format: %w", err)
	}
	//extract expected struct fields
	expectedType := reflect.TypeOf(expectedStruct)
	expectedKeys := make(map[string]bool)

	for i := 0; i < expectedType.NumField(); i++ {
		field := expectedType.Field(i)
		jsonKey := field.Tag.Get("json")
		expectedKeys[jsonKey] = true
		expectedType := field.Type.Kind()

		// Check if the key exists in the JSON map
		value, exists := jsonMap[jsonKey]
		if !exists {
			// Find the invalid key
			for inputKey := range jsonMap {
				if !expectedKeys[inputKey] {
					return fmt.Errorf("invalid key name: '%s'", inputKey)
				}
			}
			continue
		}

		//validate the value for each key type
		switch expectedType {
		case reflect.String:
			if _, ok := value.(string); !ok {
				return fmt.Errorf("field '%s' must be a string", jsonKey)
			}
		case reflect.Int:
			{
				if _, ok := value.(float64); !ok {
					return fmt.Errorf("field '%s' must be an integer", jsonKey)
				}
			}
		default:
			return fmt.Errorf("unsupported type for field '%s'", jsonKey)
		}

	}
	// Check for unexpected keys in the JSON map to enforce strict validation
	for key := range jsonMap {
		if _, ok := FindStructFieldByJSONTag(expectedType, key); !ok {
			return fmt.Errorf("unexpected field: '%s'", key)
		}
	}
	return nil
}

func FindStructFieldByJSONTag(t reflect.Type, jsonTag string) (reflect.StructField, bool) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Tag.Get("json") == jsonTag {
			return field, true
		}
	}
	return reflect.StructField{}, false
}

func ValidateDockerTarget(options string) error {
	return ValidateJSONAgainstStruct(options, DockerTargetOptions{})
}

func ValidateAWSTarget(options string) error {
	return ValidateJSONAgainstStruct(options, AWSTargetOptions{})
}

func ValidateAzureTarget(options string) error {
	return ValidateJSONAgainstStruct(options, AzureTargetOptions{})
}

func ValidateDigitalOceanTarget(options string) error {
	return ValidateJSONAgainstStruct(options, DigitalOceanTargetOptions{})
}

func ValidateGCPTarget(options string) error {
	return ValidateJSONAgainstStruct(options, GCPTargetOptions{})
}

func ValidateHetznerTarget(options string) error {
	return ValidateJSONAgainstStruct(options, HetznerTargetOptions{})
}

func ValidateFlyTarget(options string) error {
	return ValidateJSONAgainstStruct(options, FlyTargetOptions{})
}

func ValidateProviderTarget(providerName, options string) error {
	switch providerName {
	case "docker-provider":
		return ValidateDockerTarget(options)

	case "digitalocean-provider":
		return ValidateDigitalOceanTarget(options)

	case "aws-provider":
		return ValidateAWSTarget(options)

	case "azure-provider":
		return ValidateAzureTarget(options)

	case "gcp-provider":
		return ValidateGCPTarget(options)

	case "hetzner-provider":
		return ValidateHetznerTarget(options)

	case "fly-provider":
		return ValidateFlyTarget(options)
	default:
		return fmt.Errorf("unsupported provider: %s", providerName)
	}
}
