# ProjectConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BuildConfig** | Pointer to [**BuildConfig**](BuildConfig.md) |  | [optional] 
**Default** | **bool** |  | 
**EnvVars** | **map[string]string** |  | 
**GitProviderConfigId** | Pointer to **string** |  | [optional] 
**Image** | **string** |  | 
**Name** | **string** |  | 
**Prebuilds** | Pointer to [**[]PrebuildConfig**](PrebuildConfig.md) |  | [optional] 
**RepositoryUrl** | **string** |  | 
**User** | **string** |  | 

## Methods

### NewProjectConfig

`func NewProjectConfig(default_ bool, envVars map[string]string, image string, name string, repositoryUrl string, user string, ) *ProjectConfig`

NewProjectConfig instantiates a new ProjectConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectConfigWithDefaults

`func NewProjectConfigWithDefaults() *ProjectConfig`

NewProjectConfigWithDefaults instantiates a new ProjectConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBuildConfig

`func (o *ProjectConfig) GetBuildConfig() BuildConfig`

GetBuildConfig returns the BuildConfig field if non-nil, zero value otherwise.

### GetBuildConfigOk

`func (o *ProjectConfig) GetBuildConfigOk() (*BuildConfig, bool)`

GetBuildConfigOk returns a tuple with the BuildConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBuildConfig

`func (o *ProjectConfig) SetBuildConfig(v BuildConfig)`

SetBuildConfig sets BuildConfig field to given value.

### HasBuildConfig

`func (o *ProjectConfig) HasBuildConfig() bool`

HasBuildConfig returns a boolean if a field has been set.

### GetDefault

`func (o *ProjectConfig) GetDefault() bool`

GetDefault returns the Default field if non-nil, zero value otherwise.

### GetDefaultOk

`func (o *ProjectConfig) GetDefaultOk() (*bool, bool)`

GetDefaultOk returns a tuple with the Default field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefault

`func (o *ProjectConfig) SetDefault(v bool)`

SetDefault sets Default field to given value.


### GetEnvVars

`func (o *ProjectConfig) GetEnvVars() map[string]string`

GetEnvVars returns the EnvVars field if non-nil, zero value otherwise.

### GetEnvVarsOk

`func (o *ProjectConfig) GetEnvVarsOk() (*map[string]string, bool)`

GetEnvVarsOk returns a tuple with the EnvVars field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvVars

`func (o *ProjectConfig) SetEnvVars(v map[string]string)`

SetEnvVars sets EnvVars field to given value.


### GetGitProviderConfigId

`func (o *ProjectConfig) GetGitProviderConfigId() string`

GetGitProviderConfigId returns the GitProviderConfigId field if non-nil, zero value otherwise.

### GetGitProviderConfigIdOk

`func (o *ProjectConfig) GetGitProviderConfigIdOk() (*string, bool)`

GetGitProviderConfigIdOk returns a tuple with the GitProviderConfigId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGitProviderConfigId

`func (o *ProjectConfig) SetGitProviderConfigId(v string)`

SetGitProviderConfigId sets GitProviderConfigId field to given value.

### HasGitProviderConfigId

`func (o *ProjectConfig) HasGitProviderConfigId() bool`

HasGitProviderConfigId returns a boolean if a field has been set.

### GetImage

`func (o *ProjectConfig) GetImage() string`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *ProjectConfig) GetImageOk() (*string, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *ProjectConfig) SetImage(v string)`

SetImage sets Image field to given value.


### GetName

`func (o *ProjectConfig) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ProjectConfig) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ProjectConfig) SetName(v string)`

SetName sets Name field to given value.


### GetPrebuilds

`func (o *ProjectConfig) GetPrebuilds() []PrebuildConfig`

GetPrebuilds returns the Prebuilds field if non-nil, zero value otherwise.

### GetPrebuildsOk

`func (o *ProjectConfig) GetPrebuildsOk() (*[]PrebuildConfig, bool)`

GetPrebuildsOk returns a tuple with the Prebuilds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrebuilds

`func (o *ProjectConfig) SetPrebuilds(v []PrebuildConfig)`

SetPrebuilds sets Prebuilds field to given value.

### HasPrebuilds

`func (o *ProjectConfig) HasPrebuilds() bool`

HasPrebuilds returns a boolean if a field has been set.

### GetRepositoryUrl

`func (o *ProjectConfig) GetRepositoryUrl() string`

GetRepositoryUrl returns the RepositoryUrl field if non-nil, zero value otherwise.

### GetRepositoryUrlOk

`func (o *ProjectConfig) GetRepositoryUrlOk() (*string, bool)`

GetRepositoryUrlOk returns a tuple with the RepositoryUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRepositoryUrl

`func (o *ProjectConfig) SetRepositoryUrl(v string)`

SetRepositoryUrl sets RepositoryUrl field to given value.


### GetUser

`func (o *ProjectConfig) GetUser() string`

GetUser returns the User field if non-nil, zero value otherwise.

### GetUserOk

`func (o *ProjectConfig) GetUserOk() (*string, bool)`

GetUserOk returns a tuple with the User field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUser

`func (o *ProjectConfig) SetUser(v string)`

SetUser sets User field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


