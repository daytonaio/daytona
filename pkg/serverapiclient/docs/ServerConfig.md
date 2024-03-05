# ServerConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ApiPort** | Pointer to **int32** |  | [optional] 
**Frps** | Pointer to [**FRPSConfig**](FRPSConfig.md) |  | [optional] 
**GitProviders** | Pointer to [**[]GitProvider**](GitProvider.md) |  | [optional] 
**HeadscalePort** | Pointer to **int32** |  | [optional] 
**Id** | Pointer to **string** |  | [optional] 
**ProvidersDir** | Pointer to **string** |  | [optional] 
**RegistryUrl** | Pointer to **string** |  | [optional] 
**ServerDownloadUrl** | Pointer to **string** |  | [optional] 
**TargetsFilePath** | Pointer to **string** |  | [optional] 

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

### GetTargetsFilePath

`func (o *ServerConfig) GetTargetsFilePath() string`

GetTargetsFilePath returns the TargetsFilePath field if non-nil, zero value otherwise.

### GetTargetsFilePathOk

`func (o *ServerConfig) GetTargetsFilePathOk() (*string, bool)`

GetTargetsFilePathOk returns a tuple with the TargetsFilePath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTargetsFilePath

`func (o *ServerConfig) SetTargetsFilePath(v string)`

SetTargetsFilePath sets TargetsFilePath field to given value.

### HasTargetsFilePath

`func (o *ServerConfig) HasTargetsFilePath() bool`

HasTargetsFilePath returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


