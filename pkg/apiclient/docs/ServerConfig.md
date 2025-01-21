# ServerConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ApiPort** | **int32** |  | 
**BinariesPath** | **string** |  | 
**BuildImageNamespace** | Pointer to **string** |  | [optional] 
**BuilderImage** | **string** |  | 
**BuilderRegistryServer** | **string** |  | 
**DefaultWorkspaceImage** | **string** |  | 
**DefaultWorkspaceUser** | **string** |  | 
**Frps** | Pointer to [**FRPSConfig**](FRPSConfig.md) |  | [optional] 
**HeadscalePort** | **int32** |  | 
**Id** | **string** |  | 
**LocalBuilderRegistryImage** | **string** |  | 
**LocalBuilderRegistryPort** | **int32** |  | 
**LocalRunnerDisabled** | Pointer to **bool** |  | [optional] 
**LogFile** | [**LogFileConfig**](LogFileConfig.md) |  | 
**RegistryUrl** | **string** |  | 
**SamplesIndexUrl** | Pointer to **string** |  | [optional] 
**ServerDownloadUrl** | **string** |  | 

## Methods

### NewServerConfig

`func NewServerConfig(apiPort int32, binariesPath string, builderImage string, builderRegistryServer string, defaultWorkspaceImage string, defaultWorkspaceUser string, headscalePort int32, id string, localBuilderRegistryImage string, localBuilderRegistryPort int32, logFile LogFileConfig, registryUrl string, serverDownloadUrl string, ) *ServerConfig`

NewServerConfig instantiates a new ServerConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewServerConfigWithDefaults

`func NewServerConfigWithDefaults() *ServerConfig`

NewServerConfigWithDefaults instantiates a new ServerConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetApiPort

`func (o *ServerConfig) GetApiPort() int32`

GetApiPort returns the ApiPort field if non-nil, zero value otherwise.

### GetApiPortOk

`func (o *ServerConfig) GetApiPortOk() (*int32, bool)`

GetApiPortOk returns a tuple with the ApiPort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetApiPort

`func (o *ServerConfig) SetApiPort(v int32)`

SetApiPort sets ApiPort field to given value.


### GetBinariesPath

`func (o *ServerConfig) GetBinariesPath() string`

GetBinariesPath returns the BinariesPath field if non-nil, zero value otherwise.

### GetBinariesPathOk

`func (o *ServerConfig) GetBinariesPathOk() (*string, bool)`

GetBinariesPathOk returns a tuple with the BinariesPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBinariesPath

`func (o *ServerConfig) SetBinariesPath(v string)`

SetBinariesPath sets BinariesPath field to given value.


### GetBuildImageNamespace

`func (o *ServerConfig) GetBuildImageNamespace() string`

GetBuildImageNamespace returns the BuildImageNamespace field if non-nil, zero value otherwise.

### GetBuildImageNamespaceOk

`func (o *ServerConfig) GetBuildImageNamespaceOk() (*string, bool)`

GetBuildImageNamespaceOk returns a tuple with the BuildImageNamespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBuildImageNamespace

`func (o *ServerConfig) SetBuildImageNamespace(v string)`

SetBuildImageNamespace sets BuildImageNamespace field to given value.

### HasBuildImageNamespace

`func (o *ServerConfig) HasBuildImageNamespace() bool`

HasBuildImageNamespace returns a boolean if a field has been set.

### GetBuilderImage

`func (o *ServerConfig) GetBuilderImage() string`

GetBuilderImage returns the BuilderImage field if non-nil, zero value otherwise.

### GetBuilderImageOk

`func (o *ServerConfig) GetBuilderImageOk() (*string, bool)`

GetBuilderImageOk returns a tuple with the BuilderImage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBuilderImage

`func (o *ServerConfig) SetBuilderImage(v string)`

SetBuilderImage sets BuilderImage field to given value.


### GetBuilderRegistryServer

`func (o *ServerConfig) GetBuilderRegistryServer() string`

GetBuilderRegistryServer returns the BuilderRegistryServer field if non-nil, zero value otherwise.

### GetBuilderRegistryServerOk

`func (o *ServerConfig) GetBuilderRegistryServerOk() (*string, bool)`

GetBuilderRegistryServerOk returns a tuple with the BuilderRegistryServer field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBuilderRegistryServer

`func (o *ServerConfig) SetBuilderRegistryServer(v string)`

SetBuilderRegistryServer sets BuilderRegistryServer field to given value.


### GetDefaultWorkspaceImage

`func (o *ServerConfig) GetDefaultWorkspaceImage() string`

GetDefaultWorkspaceImage returns the DefaultWorkspaceImage field if non-nil, zero value otherwise.

### GetDefaultWorkspaceImageOk

`func (o *ServerConfig) GetDefaultWorkspaceImageOk() (*string, bool)`

GetDefaultWorkspaceImageOk returns a tuple with the DefaultWorkspaceImage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultWorkspaceImage

`func (o *ServerConfig) SetDefaultWorkspaceImage(v string)`

SetDefaultWorkspaceImage sets DefaultWorkspaceImage field to given value.


### GetDefaultWorkspaceUser

`func (o *ServerConfig) GetDefaultWorkspaceUser() string`

GetDefaultWorkspaceUser returns the DefaultWorkspaceUser field if non-nil, zero value otherwise.

### GetDefaultWorkspaceUserOk

`func (o *ServerConfig) GetDefaultWorkspaceUserOk() (*string, bool)`

GetDefaultWorkspaceUserOk returns a tuple with the DefaultWorkspaceUser field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultWorkspaceUser

`func (o *ServerConfig) SetDefaultWorkspaceUser(v string)`

SetDefaultWorkspaceUser sets DefaultWorkspaceUser field to given value.


### GetFrps

`func (o *ServerConfig) GetFrps() FRPSConfig`

GetFrps returns the Frps field if non-nil, zero value otherwise.

### GetFrpsOk

`func (o *ServerConfig) GetFrpsOk() (*FRPSConfig, bool)`

GetFrpsOk returns a tuple with the Frps field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFrps

`func (o *ServerConfig) SetFrps(v FRPSConfig)`

SetFrps sets Frps field to given value.

### HasFrps

`func (o *ServerConfig) HasFrps() bool`

HasFrps returns a boolean if a field has been set.

### GetHeadscalePort

`func (o *ServerConfig) GetHeadscalePort() int32`

GetHeadscalePort returns the HeadscalePort field if non-nil, zero value otherwise.

### GetHeadscalePortOk

`func (o *ServerConfig) GetHeadscalePortOk() (*int32, bool)`

GetHeadscalePortOk returns a tuple with the HeadscalePort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHeadscalePort

`func (o *ServerConfig) SetHeadscalePort(v int32)`

SetHeadscalePort sets HeadscalePort field to given value.


### GetId

`func (o *ServerConfig) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ServerConfig) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ServerConfig) SetId(v string)`

SetId sets Id field to given value.


### GetLocalBuilderRegistryImage

`func (o *ServerConfig) GetLocalBuilderRegistryImage() string`

GetLocalBuilderRegistryImage returns the LocalBuilderRegistryImage field if non-nil, zero value otherwise.

### GetLocalBuilderRegistryImageOk

`func (o *ServerConfig) GetLocalBuilderRegistryImageOk() (*string, bool)`

GetLocalBuilderRegistryImageOk returns a tuple with the LocalBuilderRegistryImage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLocalBuilderRegistryImage

`func (o *ServerConfig) SetLocalBuilderRegistryImage(v string)`

SetLocalBuilderRegistryImage sets LocalBuilderRegistryImage field to given value.


### GetLocalBuilderRegistryPort

`func (o *ServerConfig) GetLocalBuilderRegistryPort() int32`

GetLocalBuilderRegistryPort returns the LocalBuilderRegistryPort field if non-nil, zero value otherwise.

### GetLocalBuilderRegistryPortOk

`func (o *ServerConfig) GetLocalBuilderRegistryPortOk() (*int32, bool)`

GetLocalBuilderRegistryPortOk returns a tuple with the LocalBuilderRegistryPort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLocalBuilderRegistryPort

`func (o *ServerConfig) SetLocalBuilderRegistryPort(v int32)`

SetLocalBuilderRegistryPort sets LocalBuilderRegistryPort field to given value.


### GetLocalRunnerDisabled

`func (o *ServerConfig) GetLocalRunnerDisabled() bool`

GetLocalRunnerDisabled returns the LocalRunnerDisabled field if non-nil, zero value otherwise.

### GetLocalRunnerDisabledOk

`func (o *ServerConfig) GetLocalRunnerDisabledOk() (*bool, bool)`

GetLocalRunnerDisabledOk returns a tuple with the LocalRunnerDisabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLocalRunnerDisabled

`func (o *ServerConfig) SetLocalRunnerDisabled(v bool)`

SetLocalRunnerDisabled sets LocalRunnerDisabled field to given value.

### HasLocalRunnerDisabled

`func (o *ServerConfig) HasLocalRunnerDisabled() bool`

HasLocalRunnerDisabled returns a boolean if a field has been set.

### GetLogFile

`func (o *ServerConfig) GetLogFile() LogFileConfig`

GetLogFile returns the LogFile field if non-nil, zero value otherwise.

### GetLogFileOk

`func (o *ServerConfig) GetLogFileOk() (*LogFileConfig, bool)`

GetLogFileOk returns a tuple with the LogFile field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLogFile

`func (o *ServerConfig) SetLogFile(v LogFileConfig)`

SetLogFile sets LogFile field to given value.


### GetRegistryUrl

`func (o *ServerConfig) GetRegistryUrl() string`

GetRegistryUrl returns the RegistryUrl field if non-nil, zero value otherwise.

### GetRegistryUrlOk

`func (o *ServerConfig) GetRegistryUrlOk() (*string, bool)`

GetRegistryUrlOk returns a tuple with the RegistryUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRegistryUrl

`func (o *ServerConfig) SetRegistryUrl(v string)`

SetRegistryUrl sets RegistryUrl field to given value.


### GetSamplesIndexUrl

`func (o *ServerConfig) GetSamplesIndexUrl() string`

GetSamplesIndexUrl returns the SamplesIndexUrl field if non-nil, zero value otherwise.

### GetSamplesIndexUrlOk

`func (o *ServerConfig) GetSamplesIndexUrlOk() (*string, bool)`

GetSamplesIndexUrlOk returns a tuple with the SamplesIndexUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSamplesIndexUrl

`func (o *ServerConfig) SetSamplesIndexUrl(v string)`

SetSamplesIndexUrl sets SamplesIndexUrl field to given value.

### HasSamplesIndexUrl

`func (o *ServerConfig) HasSamplesIndexUrl() bool`

HasSamplesIndexUrl returns a boolean if a field has been set.

### GetServerDownloadUrl

`func (o *ServerConfig) GetServerDownloadUrl() string`

GetServerDownloadUrl returns the ServerDownloadUrl field if non-nil, zero value otherwise.

### GetServerDownloadUrlOk

`func (o *ServerConfig) GetServerDownloadUrlOk() (*string, bool)`

GetServerDownloadUrlOk returns a tuple with the ServerDownloadUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServerDownloadUrl

`func (o *ServerConfig) SetServerDownloadUrl(v string)`

SetServerDownloadUrl sets ServerDownloadUrl field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


