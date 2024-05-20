// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package devcontainer

type Configuration struct {
	Name              string         `json:"name"`
	DockerFile        string         `json:"dockerFile"`
	RunArgs           []string       `json:"runArgs"`
	InitializeCommand string         `json:"initializeCommand"`
	PostCreateCommand string         `json:"postCreateCommand"`
	RemoteUser        string         `json:"remoteUser"`
	Features          Features       `json:"features"`
	ForwardPorts      []int          `json:"forwardPorts"`
	Customizations    Customizations `json:"customizations"`
	ConfigFilePath    ConfigFilePath `json:"configFilePath"`
}

type Features struct {
	DockerInDocker string `json:"docker-in-docker"`
}

type Customizations struct {
	Vscode Vscode `json:"vscode"`
}

type Vscode struct {
	Extensions []string       `json:"extensions"`
	Settings   VscodeSettings `json:"settings"`
}

type VscodeSettings struct {
	JavaImportGradleEnabled bool   `json:"java.import.gradle.enabled"`
	JavaServerLaunchMode    string `json:"java.server.launchMode"`
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

type FeatureSets struct {
	SourceInformation SourceInformation `json:"sourceInformation"`
	Features          []Feature         `json:"features"`
	InternalVersion   string            `json:"internalVersion"`
	ComputedDigest    string            `json:"computedDigest"`
	DstFolder         string            `json:"dstFolder"`
}

type SourceInformation struct {
	Type                        string     `json:"type"`
	Manifest                    Manifest   `json:"manifest"`
	ManifestDigest              string     `json:"manifestDigest"`
	FeatureRef                  FeatureRef `json:"featureRef"`
	UserFeatureId               string     `json:"userFeatureId"`
	UserFeatureIdWithoutVersion string     `json:"userFeatureIdWithoutVersion"`
}

type Manifest struct {
	SchemaVersion int         `json:"schemaVersion"`
	MediaType     string      `json:"mediaType"`
	Config        Config      `json:"config"`
	Layers        []Layer     `json:"layers"`
	Annotations   Annotations `json:"annotations"`
}

type Config struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int    `json:"size"`
}

type Layer struct {
	MediaType   string           `json:"mediaType"`
	Digest      string           `json:"digest"`
	Size        int              `json:"size"`
	Annotations LayerAnnotations `json:"annotations"`
}

type Annotations struct {
	PackageType string `json:"com.github.package.type"`
}

type LayerAnnotations struct {
	Title string `json:"org.opencontainers.image.title"`
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
	Id               string               `json:"id"`
	Version          string               `json:"version"`
	Name             string               `json:"name"`
	DocumentationURL string               `json:"documentationURL"`
	Description      string               `json:"description"`
	Options          Options              `json:"options"`
	Entrypoint       string               `json:"entrypoint"`
	Privileged       bool                 `json:"privileged"`
	ContainerEnv     ContainerEnv         `json:"containerEnv"`
	Customizations   VscodeCustomizations `json:"customizations"`
	Mounts           []Mount              `json:"mounts"`
	InstallsAfter    []string             `json:"installsAfter"`
	Included         bool                 `json:"included"`
	Value            string               `json:"value"`
	CachePath        string               `json:"cachePath"`
	ConsecutiveId    string               `json:"consecutiveId"`
}

type Options struct {
	Version                  OptionDetails `json:"version"`
	Moby                     OptionDetails `json:"moby"`
	DockerDashComposeVersion OptionDetails `json:"dockerDashComposeVersion"`
	AzureDnsAutoDetection    OptionDetails `json:"azureDnsAutoDetection"`
	DockerDefaultAddressPool OptionDetails `json:"dockerDefaultAddressPool"`
}

type OptionDetails struct {
	Type        string   `json:"type"`
	Proposals   []string `json:"proposals,omitempty"`
	Default     string   `json:"default,omitempty"`
	Description string   `json:"description"`
}

type ContainerEnv struct {
	DockerBuildkit string `json:"DOCKER_BUILDKIT"`
}

type VscodeCustomizations struct {
	Vscode VscodeExtensions `json:"vscode"`
}

type VscodeExtensions struct {
	Extensions []string `json:"extensions"`
}

type Mount struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"`
}

type FeaturesConfiguration struct {
	FeatureSets []FeatureSets `json:"featureSets"`
	DstFolder   string        `json:"dstFolder"`
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
	Name                  string            `json:"name"`
	DockerFile            string            `json:"dockerFile"`
	RunArgs               []string          `json:"runArgs"`
	InitializeCommand     string            `json:"initializeCommand"`
	RemoteUser            string            `json:"remoteUser"`
	Features              Features          `json:"features"`
	ForwardPorts          []int             `json:"forwardPorts"`
	ConfigFilePath        ConfigFilePath    `json:"configFilePath"`
	Init                  bool              `json:"init"`
	Privileged            bool              `json:"privileged"`
	Entrypoints           []string          `json:"entrypoints"`
	Mounts                []Mount           `json:"mounts"`
	OnCreateCommands      []string          `json:"onCreateCommands"`
	UpdateContentCommands []string          `json:"updateContentCommands"`
	PostCreateCommands    []string          `json:"postCreateCommands"`
	PostStartCommands     []string          `json:"postStartCommands"`
	PostAttachCommands    []string          `json:"postAttachCommands"`
	RemoteEnv             map[string]string `json:"remoteEnv"`
	ContainerEnv          ContainerEnv      `json:"containerEnv"`
	PortsAttributes       PortsAttributes   `json:"portsAttributes"`
}

type PortsAttributes struct{}
