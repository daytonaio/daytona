# ServerConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ApiPort** | Pointer to **int32** |  | [optional] 
**BinariesPath** | Pointer to **string** |  | [optional] 
**BuildImageNamespace** | Pointer to **string** |  | [optional] 
**BuilderImage** | Pointer to **string** |  | [optional] 
**BuilderRegistryServer** | Pointer to **string** |  | [optional] 
**DefaultProjectImage** | Pointer to **string** |  | [optional] 
**DefaultProjectPostStartCommands** | Pointer to **[]string** |  | [optional] 
**DefaultProjectUser** | Pointer to **string** |  | [optional] 
**Frps** | Pointer to [**FRPSConfig**](FRPSConfig.md) |  | [optional] 
**HeadscalePort** | Pointer to **int32** |  | [optional] 
**Id** | Pointer to **string** |  | [optional] 
**IpWithProtocol** | Pointer to **string** |  | [optional] 
**LocalBuilderRegistryPort** | Pointer to **int32** |  | [optional] 
**LogFilePath** | Pointer to **string** |  | [optional] 
**ProvidersDir** | Pointer to **string** |  | [optional] 
**RegistryUrl** | Pointer to **string** |  | [optional] 
**ServerDownloadUrl** | Pointer to **string** |  | [optional] 

## Methods

### NewServerConfig

`func NewServerConfig() *ServerConfig`

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

### HasApiPort

`func (o *ServerConfig) HasApiPort() bool`

HasApiPort returns a boolean if a field has been set.

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

### HasBinariesPath

`func (o *ServerConfig) HasBinariesPath() bool`

HasBinariesPath returns a boolean if a field has been set.

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

### HasBuilderImage

`func (o *ServerConfig) HasBuilderImage() bool`

HasBuilderImage returns a boolean if a field has been set.

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

### HasBuilderRegistryServer

`func (o *ServerConfig) HasBuilderRegistryServer() bool`

HasBuilderRegistryServer returns a boolean if a field has been set.

### GetDefaultProjectImage

`func (o *ServerConfig) GetDefaultProjectImage() string`

GetDefaultProjectImage returns the DefaultProjectImage field if non-nil, zero value otherwise.

### GetDefaultProjectImageOk

`func (o *ServerConfig) GetDefaultProjectImageOk() (*string, bool)`

GetDefaultProjectImageOk returns a tuple with the DefaultProjectImage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultProjectImage

`func (o *ServerConfig) SetDefaultProjectImage(v string)`

SetDefaultProjectImage sets DefaultProjectImage field to given value.

### HasDefaultProjectImage

`func (o *ServerConfig) HasDefaultProjectImage() bool`

HasDefaultProjectImage returns a boolean if a field has been set.

### GetDefaultProjectPostStartCommands

`func (o *ServerConfig) GetDefaultProjectPostStartCommands() []string`

GetDefaultProjectPostStartCommands returns the DefaultProjectPostStartCommands field if non-nil, zero value otherwise.

### GetDefaultProjectPostStartCommandsOk

`func (o *ServerConfig) GetDefaultProjectPostStartCommandsOk() (*[]string, bool)`

GetDefaultProjectPostStartCommandsOk returns a tuple with the DefaultProjectPostStartCommands field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultProjectPostStartCommands

`func (o *ServerConfig) SetDefaultProjectPostStartCommands(v []string)`

SetDefaultProjectPostStartCommands sets DefaultProjectPostStartCommands field to given value.

### HasDefaultProjectPostStartCommands

`func (o *ServerConfig) HasDefaultProjectPostStartCommands() bool`

HasDefaultProjectPostStartCommands returns a boolean if a field has been set.

### GetDefaultProjectUser

`func (o *ServerConfig) GetDefaultProjectUser() string`

GetDefaultProjectUser returns the DefaultProjectUser field if non-nil, zero value otherwise.

### GetDefaultProjectUserOk

`func (o *ServerConfig) GetDefaultProjectUserOk() (*string, bool)`

GetDefaultProjectUserOk returns a tuple with the DefaultProjectUser field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultProjectUser

`func (o *ServerConfig) SetDefaultProjectUser(v string)`

SetDefaultProjectUser sets DefaultProjectUser field to given value.

### HasDefaultProjectUser

`func (o *ServerConfig) HasDefaultProjectUser() bool`

HasDefaultProjectUser returns a boolean if a field has been set.

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

### HasHeadscalePort

`func (o *ServerConfig) HasHeadscalePort() bool`

HasHeadscalePort returns a boolean if a field has been set.

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

### HasId

`func (o *ServerConfig) HasId() bool`

HasId returns a boolean if a field has been set.

### GetIpWithProtocol

`func (o *ServerConfig) GetIpWithProtocol() string`

GetIpWithProtocol returns the IpWithProtocol field if non-nil, zero value otherwise.

### GetIpWithProtocolOk

`func (o *ServerConfig) GetIpWithProtocolOk() (*string, bool)`

GetIpWithProtocolOk returns a tuple with the IpWithProtocol field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIpWithProtocol

`func (o *ServerConfig) SetIpWithProtocol(v string)`

SetIpWithProtocol sets IpWithProtocol field to given value.

### HasIpWithProtocol

`func (o *ServerConfig) HasIpWithProtocol() bool`

HasIpWithProtocol returns a boolean if a field has been set.

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

### HasLocalBuilderRegistryPort

`func (o *ServerConfig) HasLocalBuilderRegistryPort() bool`

HasLocalBuilderRegistryPort returns a boolean if a field has been set.

### GetLogFilePath

`func (o *ServerConfig) GetLogFilePath() string`

GetLogFilePath returns the LogFilePath field if non-nil, zero value otherwise.

### GetLogFilePathOk

`func (o *ServerConfig) GetLogFilePathOk() (*string, bool)`

GetLogFilePathOk returns a tuple with the LogFilePath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLogFilePath

`func (o *ServerConfig) SetLogFilePath(v string)`

SetLogFilePath sets LogFilePath field to given value.

### HasLogFilePath

`func (o *ServerConfig) HasLogFilePath() bool`

HasLogFilePath returns a boolean if a field has been set.

### GetProvidersDir

`func (o *ServerConfig) GetProvidersDir() string`

GetProvidersDir returns the ProvidersDir field if non-nil, zero value otherwise.

### GetProvidersDirOk

`func (o *ServerConfig) GetProvidersDirOk() (*string, bool)`

GetProvidersDirOk returns a tuple with the ProvidersDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProvidersDir

`func (o *ServerConfig) SetProvidersDir(v string)`

SetProvidersDir sets ProvidersDir field to given value.

### HasProvidersDir

`func (o *ServerConfig) HasProvidersDir() bool`

HasProvidersDir returns a boolean if a field has been set.

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

### HasRegistryUrl

`func (o *ServerConfig) HasRegistryUrl() bool`

HasRegistryUrl returns a boolean if a field has been set.

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

### HasServerDownloadUrl

`func (o *ServerConfig) HasServerDownloadUrl() bool`

HasServerDownloadUrl returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


