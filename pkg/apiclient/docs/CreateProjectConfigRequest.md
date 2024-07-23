# CreateProjectConfigRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Build** | Pointer to [**ProjectBuild**](ProjectBuild.md) |  | [optional] 
**EnvVars** | Pointer to **map[string]string** |  | [optional] 
**Image** | Pointer to **string** |  | [optional] 
**Name** | **string** |  | 
**RepositoryUrl** | Pointer to **string** |  | [optional] 
**User** | Pointer to **string** |  | [optional] 

## Methods

### NewCreateProjectConfigRequest

`func NewCreateProjectConfigRequest(name string, ) *CreateProjectConfigRequest`

NewCreateProjectConfigRequest instantiates a new CreateProjectConfigRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateProjectConfigRequestWithDefaults

`func NewCreateProjectConfigRequestWithDefaults() *CreateProjectConfigRequest`

NewCreateProjectConfigRequestWithDefaults instantiates a new CreateProjectConfigRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBuild

`func (o *CreateProjectConfigRequest) GetBuild() ProjectBuild`

GetBuild returns the Build field if non-nil, zero value otherwise.

### GetBuildOk

`func (o *CreateProjectConfigRequest) GetBuildOk() (*ProjectBuild, bool)`

GetBuildOk returns a tuple with the Build field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBuild

`func (o *CreateProjectConfigRequest) SetBuild(v ProjectBuild)`

SetBuild sets Build field to given value.

### HasBuild

`func (o *CreateProjectConfigRequest) HasBuild() bool`

HasBuild returns a boolean if a field has been set.

### GetEnvVars

`func (o *CreateProjectConfigRequest) GetEnvVars() map[string]string`

GetEnvVars returns the EnvVars field if non-nil, zero value otherwise.

### GetEnvVarsOk

`func (o *CreateProjectConfigRequest) GetEnvVarsOk() (*map[string]string, bool)`

GetEnvVarsOk returns a tuple with the EnvVars field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvVars

`func (o *CreateProjectConfigRequest) SetEnvVars(v map[string]string)`

SetEnvVars sets EnvVars field to given value.

### HasEnvVars

`func (o *CreateProjectConfigRequest) HasEnvVars() bool`

HasEnvVars returns a boolean if a field has been set.

### GetImage

`func (o *CreateProjectConfigRequest) GetImage() string`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *CreateProjectConfigRequest) GetImageOk() (*string, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *CreateProjectConfigRequest) SetImage(v string)`

SetImage sets Image field to given value.

### HasImage

`func (o *CreateProjectConfigRequest) HasImage() bool`

HasImage returns a boolean if a field has been set.

### GetName

`func (o *CreateProjectConfigRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *CreateProjectConfigRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *CreateProjectConfigRequest) SetName(v string)`

SetName sets Name field to given value.


### GetRepositoryUrl

`func (o *CreateProjectConfigRequest) GetRepositoryUrl() string`

GetRepositoryUrl returns the RepositoryUrl field if non-nil, zero value otherwise.

### GetRepositoryUrlOk

`func (o *CreateProjectConfigRequest) GetRepositoryUrlOk() (*string, bool)`

GetRepositoryUrlOk returns a tuple with the RepositoryUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRepositoryUrl

`func (o *CreateProjectConfigRequest) SetRepositoryUrl(v string)`

SetRepositoryUrl sets RepositoryUrl field to given value.

### HasRepositoryUrl

`func (o *CreateProjectConfigRequest) HasRepositoryUrl() bool`

HasRepositoryUrl returns a boolean if a field has been set.

### GetUser

`func (o *CreateProjectConfigRequest) GetUser() string`

GetUser returns the User field if non-nil, zero value otherwise.

### GetUserOk

`func (o *CreateProjectConfigRequest) GetUserOk() (*string, bool)`

GetUserOk returns a tuple with the User field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUser

`func (o *CreateProjectConfigRequest) SetUser(v string)`

SetUser sets User field to given value.

### HasUser

`func (o *CreateProjectConfigRequest) HasUser() bool`

HasUser returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


