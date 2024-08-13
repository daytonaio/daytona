# ProjectConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BuildConfig** | Pointer to [**ProjectBuildConfig**](ProjectBuildConfig.md) |  | [optional] 
**Default** | **bool** |  | 
**EnvVars** | **map[string]string** |  | 
**Image** | **string** |  | 
**Name** | **string** |  | 
**Repository** | [**GitRepository**](GitRepository.md) |  | 
**User** | **string** |  | 

## Methods

### NewProjectConfig

`func NewProjectConfig(default_ bool, envVars map[string]string, image string, name string, repository GitRepository, user string, ) *ProjectConfig`

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

`func (o *ProjectConfig) GetBuildConfig() ProjectBuildConfig`

GetBuildConfig returns the BuildConfig field if non-nil, zero value otherwise.

### GetBuildConfigOk

`func (o *ProjectConfig) GetBuildConfigOk() (*ProjectBuildConfig, bool)`

GetBuildConfigOk returns a tuple with the BuildConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBuildConfig

`func (o *ProjectConfig) SetBuildConfig(v ProjectBuildConfig)`

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


### GetRepository

`func (o *ProjectConfig) GetRepository() GitRepository`

GetRepository returns the Repository field if non-nil, zero value otherwise.

### GetRepositoryOk

`func (o *ProjectConfig) GetRepositoryOk() (*GitRepository, bool)`

GetRepositoryOk returns a tuple with the Repository field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRepository

`func (o *ProjectConfig) SetRepository(v GitRepository)`

SetRepository sets Repository field to given value.


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


