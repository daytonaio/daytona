# ProjectConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Build** | Pointer to [**ProjectBuildConfig**](ProjectBuildConfig.md) |  | [optional] 
**Default** | Pointer to **bool** |  | [optional] 
**Image** | Pointer to **string** |  | [optional] 
**Name** | Pointer to **string** |  | [optional] 
**Repository** | Pointer to [**GitRepository**](GitRepository.md) |  | [optional] 
**User** | Pointer to **string** |  | [optional] 

## Methods

### NewProjectConfig

`func NewProjectConfig() *ProjectConfig`

NewProjectConfig instantiates a new ProjectConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectConfigWithDefaults

`func NewProjectConfigWithDefaults() *ProjectConfig`

NewProjectConfigWithDefaults instantiates a new ProjectConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBuild

`func (o *ProjectConfig) GetBuild() ProjectBuildConfig`

GetBuild returns the Build field if non-nil, zero value otherwise.

### GetBuildOk

`func (o *ProjectConfig) GetBuildOk() (*ProjectBuildConfig, bool)`

GetBuildOk returns a tuple with the Build field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBuild

`func (o *ProjectConfig) SetBuild(v ProjectBuildConfig)`

SetBuild sets Build field to given value.

### HasBuild

`func (o *ProjectConfig) HasBuild() bool`

HasBuild returns a boolean if a field has been set.

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

### HasDefault

`func (o *ProjectConfig) HasDefault() bool`

HasDefault returns a boolean if a field has been set.

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

### HasImage

`func (o *ProjectConfig) HasImage() bool`

HasImage returns a boolean if a field has been set.

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

### HasName

`func (o *ProjectConfig) HasName() bool`

HasName returns a boolean if a field has been set.

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

### HasRepository

`func (o *ProjectConfig) HasRepository() bool`

HasRepository returns a boolean if a field has been set.

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

### HasUser

`func (o *ProjectConfig) HasUser() bool`

HasUser returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


