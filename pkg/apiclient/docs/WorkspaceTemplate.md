# WorkspaceTemplate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BuildConfig** | Pointer to [**BuildConfig**](BuildConfig.md) |  | [optional] 
**Default** | **bool** |  | 
**EnvVars** | **map[string]string** |  | 
**GitProviderConfigId** | Pointer to **string** |  | [optional] 
**Image** | **string** |  | 
**Labels** | **map[string]string** |  | 
**Name** | **string** |  | 
**Prebuilds** | Pointer to [**[]PrebuildConfig**](PrebuildConfig.md) |  | [optional] 
**RepositoryUrl** | **string** |  | 
**User** | **string** |  | 

## Methods

### NewWorkspaceTemplate

`func NewWorkspaceTemplate(default_ bool, envVars map[string]string, image string, labels map[string]string, name string, repositoryUrl string, user string, ) *WorkspaceTemplate`

NewWorkspaceTemplate instantiates a new WorkspaceTemplate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspaceTemplateWithDefaults

`func NewWorkspaceTemplateWithDefaults() *WorkspaceTemplate`

NewWorkspaceTemplateWithDefaults instantiates a new WorkspaceTemplate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBuildConfig

`func (o *WorkspaceTemplate) GetBuildConfig() BuildConfig`

GetBuildConfig returns the BuildConfig field if non-nil, zero value otherwise.

### GetBuildConfigOk

`func (o *WorkspaceTemplate) GetBuildConfigOk() (*BuildConfig, bool)`

GetBuildConfigOk returns a tuple with the BuildConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBuildConfig

`func (o *WorkspaceTemplate) SetBuildConfig(v BuildConfig)`

SetBuildConfig sets BuildConfig field to given value.

### HasBuildConfig

`func (o *WorkspaceTemplate) HasBuildConfig() bool`

HasBuildConfig returns a boolean if a field has been set.

### GetDefault

`func (o *WorkspaceTemplate) GetDefault() bool`

GetDefault returns the Default field if non-nil, zero value otherwise.

### GetDefaultOk

`func (o *WorkspaceTemplate) GetDefaultOk() (*bool, bool)`

GetDefaultOk returns a tuple with the Default field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefault

`func (o *WorkspaceTemplate) SetDefault(v bool)`

SetDefault sets Default field to given value.


### GetEnvVars

`func (o *WorkspaceTemplate) GetEnvVars() map[string]string`

GetEnvVars returns the EnvVars field if non-nil, zero value otherwise.

### GetEnvVarsOk

`func (o *WorkspaceTemplate) GetEnvVarsOk() (*map[string]string, bool)`

GetEnvVarsOk returns a tuple with the EnvVars field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvVars

`func (o *WorkspaceTemplate) SetEnvVars(v map[string]string)`

SetEnvVars sets EnvVars field to given value.


### GetGitProviderConfigId

`func (o *WorkspaceTemplate) GetGitProviderConfigId() string`

GetGitProviderConfigId returns the GitProviderConfigId field if non-nil, zero value otherwise.

### GetGitProviderConfigIdOk

`func (o *WorkspaceTemplate) GetGitProviderConfigIdOk() (*string, bool)`

GetGitProviderConfigIdOk returns a tuple with the GitProviderConfigId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGitProviderConfigId

`func (o *WorkspaceTemplate) SetGitProviderConfigId(v string)`

SetGitProviderConfigId sets GitProviderConfigId field to given value.

### HasGitProviderConfigId

`func (o *WorkspaceTemplate) HasGitProviderConfigId() bool`

HasGitProviderConfigId returns a boolean if a field has been set.

### GetImage

`func (o *WorkspaceTemplate) GetImage() string`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *WorkspaceTemplate) GetImageOk() (*string, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *WorkspaceTemplate) SetImage(v string)`

SetImage sets Image field to given value.


### GetLabels

`func (o *WorkspaceTemplate) GetLabels() map[string]string`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *WorkspaceTemplate) GetLabelsOk() (*map[string]string, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *WorkspaceTemplate) SetLabels(v map[string]string)`

SetLabels sets Labels field to given value.


### GetName

`func (o *WorkspaceTemplate) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkspaceTemplate) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkspaceTemplate) SetName(v string)`

SetName sets Name field to given value.


### GetPrebuilds

`func (o *WorkspaceTemplate) GetPrebuilds() []PrebuildConfig`

GetPrebuilds returns the Prebuilds field if non-nil, zero value otherwise.

### GetPrebuildsOk

`func (o *WorkspaceTemplate) GetPrebuildsOk() (*[]PrebuildConfig, bool)`

GetPrebuildsOk returns a tuple with the Prebuilds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrebuilds

`func (o *WorkspaceTemplate) SetPrebuilds(v []PrebuildConfig)`

SetPrebuilds sets Prebuilds field to given value.

### HasPrebuilds

`func (o *WorkspaceTemplate) HasPrebuilds() bool`

HasPrebuilds returns a boolean if a field has been set.

### GetRepositoryUrl

`func (o *WorkspaceTemplate) GetRepositoryUrl() string`

GetRepositoryUrl returns the RepositoryUrl field if non-nil, zero value otherwise.

### GetRepositoryUrlOk

`func (o *WorkspaceTemplate) GetRepositoryUrlOk() (*string, bool)`

GetRepositoryUrlOk returns a tuple with the RepositoryUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRepositoryUrl

`func (o *WorkspaceTemplate) SetRepositoryUrl(v string)`

SetRepositoryUrl sets RepositoryUrl field to given value.


### GetUser

`func (o *WorkspaceTemplate) GetUser() string`

GetUser returns the User field if non-nil, zero value otherwise.

### GetUserOk

`func (o *WorkspaceTemplate) GetUserOk() (*string, bool)`

GetUserOk returns a tuple with the User field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUser

`func (o *WorkspaceTemplate) SetUser(v string)`

SetUser sets User field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


