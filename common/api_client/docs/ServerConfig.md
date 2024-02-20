# ServerConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ApiPort** | Pointer to **int32** |  | [optional] 
**Frps** | Pointer to [**FRPSConfig**](FRPSConfig.md) |  | [optional] 
**GitProviders** | Pointer to [**[]GitProvider**](GitProvider.md) |  | [optional] 
**HeadscalePort** | Pointer to **int32** |  | [optional] 
**Id** | Pointer to **string** |  | [optional] 
**PluginRegistryUrl** | Pointer to **string** |  | [optional] 
**PluginsDir** | Pointer to **string** |  | [optional] 
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

### GetGitProviders

`func (o *ServerConfig) GetGitProviders() []GitProvider`

GetGitProviders returns the GitProviders field if non-nil, zero value otherwise.

### GetGitProvidersOk

`func (o *ServerConfig) GetGitProvidersOk() (*[]GitProvider, bool)`

GetGitProvidersOk returns a tuple with the GitProviders field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGitProviders

`func (o *ServerConfig) SetGitProviders(v []GitProvider)`

SetGitProviders sets GitProviders field to given value.

### HasGitProviders

`func (o *ServerConfig) HasGitProviders() bool`

HasGitProviders returns a boolean if a field has been set.

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

### GetPluginRegistryUrl

`func (o *ServerConfig) GetPluginRegistryUrl() string`

GetPluginRegistryUrl returns the PluginRegistryUrl field if non-nil, zero value otherwise.

### GetPluginRegistryUrlOk

`func (o *ServerConfig) GetPluginRegistryUrlOk() (*string, bool)`

GetPluginRegistryUrlOk returns a tuple with the PluginRegistryUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPluginRegistryUrl

`func (o *ServerConfig) SetPluginRegistryUrl(v string)`

SetPluginRegistryUrl sets PluginRegistryUrl field to given value.

### HasPluginRegistryUrl

`func (o *ServerConfig) HasPluginRegistryUrl() bool`

HasPluginRegistryUrl returns a boolean if a field has been set.

### GetPluginsDir

`func (o *ServerConfig) GetPluginsDir() string`

GetPluginsDir returns the PluginsDir field if non-nil, zero value otherwise.

### GetPluginsDirOk

`func (o *ServerConfig) GetPluginsDirOk() (*string, bool)`

GetPluginsDirOk returns a tuple with the PluginsDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPluginsDir

`func (o *ServerConfig) SetPluginsDir(v string)`

SetPluginsDir sets PluginsDir field to given value.

### HasPluginsDir

`func (o *ServerConfig) HasPluginsDir() bool`

HasPluginsDir returns a boolean if a field has been set.

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


