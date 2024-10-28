# CreateWorkspaceConfigDTO

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BuildConfig** | Pointer to [**BuildConfig**](BuildConfig.md) |  | [optional] 
**EnvVars** | **map[string]string** |  | 
**GitProviderConfigId** | Pointer to **string** |  | [optional] 
**Image** | Pointer to **string** |  | [optional] 
**Name** | **string** |  | 
**RepositoryUrl** | **string** |  | 
**User** | Pointer to **string** |  | [optional] 

## Methods

### NewCreateWorkspaceConfigDTO

`func NewCreateWorkspaceConfigDTO(envVars map[string]string, name string, repositoryUrl string, ) *CreateWorkspaceConfigDTO`

NewCreateWorkspaceConfigDTO instantiates a new CreateWorkspaceConfigDTO object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateWorkspaceConfigDTOWithDefaults

`func NewCreateWorkspaceConfigDTOWithDefaults() *CreateWorkspaceConfigDTO`

NewCreateWorkspaceConfigDTOWithDefaults instantiates a new CreateWorkspaceConfigDTO object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBuildConfig

`func (o *CreateWorkspaceConfigDTO) GetBuildConfig() BuildConfig`

GetBuildConfig returns the BuildConfig field if non-nil, zero value otherwise.

### GetBuildConfigOk

`func (o *CreateWorkspaceConfigDTO) GetBuildConfigOk() (*BuildConfig, bool)`

GetBuildConfigOk returns a tuple with the BuildConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBuildConfig

`func (o *CreateWorkspaceConfigDTO) SetBuildConfig(v BuildConfig)`

SetBuildConfig sets BuildConfig field to given value.

### HasBuildConfig

`func (o *CreateWorkspaceConfigDTO) HasBuildConfig() bool`

HasBuildConfig returns a boolean if a field has been set.

### GetEnvVars

`func (o *CreateWorkspaceConfigDTO) GetEnvVars() map[string]string`

GetEnvVars returns the EnvVars field if non-nil, zero value otherwise.

### GetEnvVarsOk

`func (o *CreateWorkspaceConfigDTO) GetEnvVarsOk() (*map[string]string, bool)`

GetEnvVarsOk returns a tuple with the EnvVars field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvVars

`func (o *CreateWorkspaceConfigDTO) SetEnvVars(v map[string]string)`

SetEnvVars sets EnvVars field to given value.


### GetGitProviderConfigId

`func (o *CreateWorkspaceConfigDTO) GetGitProviderConfigId() string`

GetGitProviderConfigId returns the GitProviderConfigId field if non-nil, zero value otherwise.

### GetGitProviderConfigIdOk

`func (o *CreateWorkspaceConfigDTO) GetGitProviderConfigIdOk() (*string, bool)`

GetGitProviderConfigIdOk returns a tuple with the GitProviderConfigId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGitProviderConfigId

`func (o *CreateWorkspaceConfigDTO) SetGitProviderConfigId(v string)`

SetGitProviderConfigId sets GitProviderConfigId field to given value.

### HasGitProviderConfigId

`func (o *CreateWorkspaceConfigDTO) HasGitProviderConfigId() bool`

HasGitProviderConfigId returns a boolean if a field has been set.

### GetImage

`func (o *CreateWorkspaceConfigDTO) GetImage() string`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *CreateWorkspaceConfigDTO) GetImageOk() (*string, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *CreateWorkspaceConfigDTO) SetImage(v string)`

SetImage sets Image field to given value.

### HasImage

`func (o *CreateWorkspaceConfigDTO) HasImage() bool`

HasImage returns a boolean if a field has been set.

### GetName

`func (o *CreateWorkspaceConfigDTO) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *CreateWorkspaceConfigDTO) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *CreateWorkspaceConfigDTO) SetName(v string)`

SetName sets Name field to given value.


### GetRepositoryUrl

`func (o *CreateWorkspaceConfigDTO) GetRepositoryUrl() string`

GetRepositoryUrl returns the RepositoryUrl field if non-nil, zero value otherwise.

### GetRepositoryUrlOk

`func (o *CreateWorkspaceConfigDTO) GetRepositoryUrlOk() (*string, bool)`

GetRepositoryUrlOk returns a tuple with the RepositoryUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRepositoryUrl

`func (o *CreateWorkspaceConfigDTO) SetRepositoryUrl(v string)`

SetRepositoryUrl sets RepositoryUrl field to given value.


### GetUser

`func (o *CreateWorkspaceConfigDTO) GetUser() string`

GetUser returns the User field if non-nil, zero value otherwise.

### GetUserOk

`func (o *CreateWorkspaceConfigDTO) GetUserOk() (*string, bool)`

GetUserOk returns a tuple with the User field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUser

`func (o *CreateWorkspaceConfigDTO) SetUser(v string)`

SetUser sets User field to given value.

### HasUser

`func (o *CreateWorkspaceConfigDTO) HasUser() bool`

HasUser returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


