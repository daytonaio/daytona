// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package devcontainer

import "fmt"

type Configuration struct {
	Name              string                    `json:"name"`
	DockerFile        string                    `json:"dockerFile"`
	RunArgs           []string                  `json:"runArgs"`
	InitializeCommand string                    `json:"initializeCommand"`
	PostCreateCommand string                    `json:"postCreateCommand"`
	RemoteUser        string                    `json:"remoteUser"`
	Features          map[string]interface{}    `json:"features"`
	ForwardPorts      map[string]PortAttributes `json:"forwardPorts"`
	Customizations    map[string]interface{}    `json:"customizations"`
	ConfigFilePath    ConfigFilePath            `json:"configFilePath"`
}

type PortAttributes struct {
	Label            *string
	OnAutoForward    *string
	Protocol         *string
	RequireLocalPort *bool
	ElevateIfNeeded  *bool
}

type ConfigFilePath struct {
	Mid    int    `json:"$mid"`
	FsPath string `json:"fsPath"`
	Path   string `json:"path"`
	Scheme string `json:"scheme"`
}

type Workspace struct {
	WorkspaceFolder string `json:"workspaceFolder"`
	WorkspaceMount  string `json:"workspaceMount"`
}

type FeatureRef struct {
	Id        string `json:"id"`
	Owner     string `json:"owner"`
	Namespace string `json:"namespace"`
	Registry  string `json:"registry"`
	Resource  string `json:"resource"`
	Path      string `json:"path"`
	Version   string `json:"version"`
	Tag       string `json:"tag"`
}

type Feature struct {
	Id               string                 `json:"id"`
	Version          string                 `json:"version"`
	Name             string                 `json:"name"`
	DocumentationURL string                 `json:"documentationURL"`
	Description      string                 `json:"description"`
	Options          map[string]interface{} `json:"options"`
	LicenceURL       string                 `json:"licenceURL"`
	Keywords         string                 `json:"keywords"`
	Entrypoint       string                 `json:"entrypoint"`
	Privileged       bool                   `json:"privileged"`
	ContainerEnv     map[string]string      `json:"containerEnv"`
	Customizations   map[string]interface{} `json:"customizations"`
	Mounts           []Mount                `json:"mounts"`
	InstallsAfter    []string               `json:"installsAfter"`
	Included         bool                   `json:"included"`
	Value            string                 `json:"value"`
	CachePath        string                 `json:"cachePath"`
	ConsecutiveId    string                 `json:"consecutiveId"`
	Init             bool                   `json:"init"`
	CapAdd           []string               `json:"capAdd"`
	SecurityOpt      []string               `json:"securityOpt"`
	LegacyIds        []string               `json:"legacyIds"`
	Deprecated       bool                   `json:"deprecated"`
}

type Mount struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"`
}

type Root struct {
	/*
		Configuration         Configuration         `json:"configuration"`
		Workspace             Workspace             `json:"workspace"`
		FeaturesConfiguration FeaturesConfiguration `json:"featuresConfiguration"`
	*/
	MergedConfiguration MergedConfiguration `json:"mergedConfiguration"`
}

type MergedConfiguration struct {
	Name            string                    `json:"name"`
	DockerFile      string                    `json:"dockerFile"`
	RunArgs         []string                  `json:"runArgs"`
	RemoteUser      string                    `json:"remoteUser"`
	Features        map[string]interface{}    `json:"features"`
	ForwardPorts    []int                     `json:"forwardPorts"`
	ConfigFilePath  ConfigFilePath            `json:"configFilePath"`
	Init            bool                      `json:"init"`
	Privileged      bool                      `json:"privileged"`
	Entrypoints     []string                  `json:"entrypoints"`
	Mounts          []Mount                   `json:"mounts"`
	RemoteEnv       map[string]string         `json:"remoteEnv"`
	ContainerEnv    map[string]string         `json:"containerEnv"`
	PortsAttributes map[string]PortAttributes `json:"portsAttributes"`

	// Commands
	InitializeCommand     []interface{} `json:"initializeCommand"`
	OnCreateCommands      []interface{} `json:"onCreateCommands"`
	UpdateContentCommands []interface{} `json:"updateContentCommands"`
	PostCreateCommands    []interface{} `json:"postCreateCommands"`
	PostStartCommands     []interface{} `json:"postStartCommands"`
	PostAttachCommands    []interface{} `json:"postAttachCommands"`
}

func ConvertCommands(mergedCommands []interface{}) ([]string, error) {
	// Convert the commands to a string array
	var commandArray []string
	for _, commands := range mergedCommands {
		switch commands := commands.(type) {
		case []interface{}:
			for _, command := range commands {
				commandString, ok := command.(string)
				if !ok {
					return nil, fmt.Errorf("invalid command type: %v", command)
				}
				commandArray = append(commandArray, commandString)
			}
		case map[string]interface{}:
			for _, command := range commands {
				commandString, ok := command.(string)
				if !ok {
					return nil, fmt.Errorf("invalid command type: %v", command)
				}
				commandArray = append(commandArray, commandString)
			}
		case string:
			commandArray = append(commandArray, commands)
		default:
			return nil, fmt.Errorf("invalid command type")
		}
	}

	return commandArray, nil
}
