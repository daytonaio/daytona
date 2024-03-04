package provider

type ProviderProfile struct {
	Name   string
	Config interface{}
}

type ProviderInfo struct {
	Name    string
	Version string
}

type InitializeProviderRequest struct {
	BasePath          string
	ServerDownloadUrl string
	ServerVersion     string
	ServerUrl         string
	ServerApiUrl      string
}

type ProviderTarget struct {
	Name string
	// JSON encoded map of options
	Options string
}

type ProviderTargetManifest map[string]ProviderTargetProperty // @name ProviderTargetManifest

type ProviderTargetPropertyType string

const (
	ProviderTargetPropertyTypeString  ProviderTargetPropertyType = "string"
	ProviderTargetPropertyTypeOption  ProviderTargetPropertyType = "option"
	ProviderTargetPropertyTypeBoolean ProviderTargetPropertyType = "boolean"
	ProviderTargetPropertyTypeInt     ProviderTargetPropertyType = "int"
	ProviderTargetPropertyTypeFloat   ProviderTargetPropertyType = "float"
)

type ProviderTargetProperty struct {
	Type ProviderTargetPropertyType
	// DefaultValue is converted into the appropriate type based on the Type
	DefaultValue string
	// Options is only used if the Type is ProviderTargetPropertyTypeOption
	Options []string
}
