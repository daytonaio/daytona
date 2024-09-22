# Build

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BuildConfig** | Pointer to [**BuildConfig**](BuildConfig.md) |  | [optional] 
**ContainerConfig** | [**ContainerConfig**](ContainerConfig.md) |  | 
**CreatedAt** | **string** |  | 
**EnvVars** | **map[string]string** |  | 
**Id** | **string** |  | 
**Image** | Pointer to **string** |  | [optional] 
**PrebuildId** | **string** |  | 
**Repository** | [**GitRepository**](GitRepository.md) |  | 
**State** | [**BuildBuildState**](BuildBuildState.md) |  | 
**UpdatedAt** | **string** |  | 
**User** | Pointer to **string** |  | [optional] 

## Methods

### NewBuild

`func NewBuild(containerConfig ContainerConfig, createdAt string, envVars map[string]string, id string, prebuildId string, repository GitRepository, state BuildBuildState, updatedAt string, ) *Build`

NewBuild instantiates a new Build object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBuildWithDefaults

`func NewBuildWithDefaults() *Build`

NewBuildWithDefaults instantiates a new Build object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBuildConfig

`func (o *Build) GetBuildConfig() BuildConfig`

GetBuildConfig returns the BuildConfig field if non-nil, zero value otherwise.

### GetBuildConfigOk

`func (o *Build) GetBuildConfigOk() (*BuildConfig, bool)`

GetBuildConfigOk returns a tuple with the BuildConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBuildConfig

`func (o *Build) SetBuildConfig(v BuildConfig)`

SetBuildConfig sets BuildConfig field to given value.

### HasBuildConfig

`func (o *Build) HasBuildConfig() bool`

HasBuildConfig returns a boolean if a field has been set.

### GetContainerConfig

`func (o *Build) GetContainerConfig() ContainerConfig`

GetContainerConfig returns the ContainerConfig field if non-nil, zero value otherwise.

### GetContainerConfigOk

`func (o *Build) GetContainerConfigOk() (*ContainerConfig, bool)`

GetContainerConfigOk returns a tuple with the ContainerConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContainerConfig

`func (o *Build) SetContainerConfig(v ContainerConfig)`

SetContainerConfig sets ContainerConfig field to given value.


### GetCreatedAt

`func (o *Build) GetCreatedAt() string`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Build) GetCreatedAtOk() (*string, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Build) SetCreatedAt(v string)`

SetCreatedAt sets CreatedAt field to given value.


### GetEnvVars

`func (o *Build) GetEnvVars() map[string]string`

GetEnvVars returns the EnvVars field if non-nil, zero value otherwise.

### GetEnvVarsOk

`func (o *Build) GetEnvVarsOk() (*map[string]string, bool)`

GetEnvVarsOk returns a tuple with the EnvVars field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvVars

`func (o *Build) SetEnvVars(v map[string]string)`

SetEnvVars sets EnvVars field to given value.


### GetId

`func (o *Build) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Build) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Build) SetId(v string)`

SetId sets Id field to given value.


### GetImage

`func (o *Build) GetImage() string`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *Build) GetImageOk() (*string, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *Build) SetImage(v string)`

SetImage sets Image field to given value.

### HasImage

`func (o *Build) HasImage() bool`

HasImage returns a boolean if a field has been set.

### GetPrebuildId

`func (o *Build) GetPrebuildId() string`

GetPrebuildId returns the PrebuildId field if non-nil, zero value otherwise.

### GetPrebuildIdOk

`func (o *Build) GetPrebuildIdOk() (*string, bool)`

GetPrebuildIdOk returns a tuple with the PrebuildId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrebuildId

`func (o *Build) SetPrebuildId(v string)`

SetPrebuildId sets PrebuildId field to given value.


### GetRepository

`func (o *Build) GetRepository() GitRepository`

GetRepository returns the Repository field if non-nil, zero value otherwise.

### GetRepositoryOk

`func (o *Build) GetRepositoryOk() (*GitRepository, bool)`

GetRepositoryOk returns a tuple with the Repository field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRepository

`func (o *Build) SetRepository(v GitRepository)`

SetRepository sets Repository field to given value.


### GetState

`func (o *Build) GetState() BuildBuildState`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *Build) GetStateOk() (*BuildBuildState, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *Build) SetState(v BuildBuildState)`

SetState sets State field to given value.


### GetUpdatedAt

`func (o *Build) GetUpdatedAt() string`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *Build) GetUpdatedAtOk() (*string, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *Build) SetUpdatedAt(v string)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetUser

`func (o *Build) GetUser() string`

GetUser returns the User field if non-nil, zero value otherwise.

### GetUserOk

`func (o *Build) GetUserOk() (*string, bool)`

GetUserOk returns a tuple with the User field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUser

`func (o *Build) SetUser(v string)`

SetUser sets User field to given value.

### HasUser

`func (o *Build) HasUser() bool`

HasUser returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


