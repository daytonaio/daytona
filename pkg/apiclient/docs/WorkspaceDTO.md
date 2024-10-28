# WorkspaceDTO

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BuildConfig** | Pointer to [**BuildConfig**](BuildConfig.md) |  | [optional] 
**EnvVars** | **map[string]string** |  | 
**GitProviderConfigId** | Pointer to **string** |  | [optional] 
**Id** | **string** |  | 
**Image** | **string** |  | 
**Info** | Pointer to [**WorkspaceInfo**](WorkspaceInfo.md) |  | [optional] 
**Name** | **string** |  | 
**Repository** | [**GitRepository**](GitRepository.md) |  | 
**State** | Pointer to [**WorkspaceState**](WorkspaceState.md) |  | [optional] 
**TargetId** | **string** |  | 
**User** | **string** |  | 

## Methods

### NewWorkspaceDTO

`func NewWorkspaceDTO(envVars map[string]string, id string, image string, name string, repository GitRepository, targetId string, user string, ) *WorkspaceDTO`

NewWorkspaceDTO instantiates a new WorkspaceDTO object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspaceDTOWithDefaults

`func NewWorkspaceDTOWithDefaults() *WorkspaceDTO`

NewWorkspaceDTOWithDefaults instantiates a new WorkspaceDTO object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBuildConfig

`func (o *WorkspaceDTO) GetBuildConfig() BuildConfig`

GetBuildConfig returns the BuildConfig field if non-nil, zero value otherwise.

### GetBuildConfigOk

`func (o *WorkspaceDTO) GetBuildConfigOk() (*BuildConfig, bool)`

GetBuildConfigOk returns a tuple with the BuildConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBuildConfig

`func (o *WorkspaceDTO) SetBuildConfig(v BuildConfig)`

SetBuildConfig sets BuildConfig field to given value.

### HasBuildConfig

`func (o *WorkspaceDTO) HasBuildConfig() bool`

HasBuildConfig returns a boolean if a field has been set.

### GetEnvVars

`func (o *WorkspaceDTO) GetEnvVars() map[string]string`

GetEnvVars returns the EnvVars field if non-nil, zero value otherwise.

### GetEnvVarsOk

`func (o *WorkspaceDTO) GetEnvVarsOk() (*map[string]string, bool)`

GetEnvVarsOk returns a tuple with the EnvVars field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvVars

`func (o *WorkspaceDTO) SetEnvVars(v map[string]string)`

SetEnvVars sets EnvVars field to given value.


### GetGitProviderConfigId

`func (o *WorkspaceDTO) GetGitProviderConfigId() string`

GetGitProviderConfigId returns the GitProviderConfigId field if non-nil, zero value otherwise.

### GetGitProviderConfigIdOk

`func (o *WorkspaceDTO) GetGitProviderConfigIdOk() (*string, bool)`

GetGitProviderConfigIdOk returns a tuple with the GitProviderConfigId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGitProviderConfigId

`func (o *WorkspaceDTO) SetGitProviderConfigId(v string)`

SetGitProviderConfigId sets GitProviderConfigId field to given value.

### HasGitProviderConfigId

`func (o *WorkspaceDTO) HasGitProviderConfigId() bool`

HasGitProviderConfigId returns a boolean if a field has been set.

### GetId

`func (o *WorkspaceDTO) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *WorkspaceDTO) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *WorkspaceDTO) SetId(v string)`

SetId sets Id field to given value.


### GetImage

`func (o *WorkspaceDTO) GetImage() string`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *WorkspaceDTO) GetImageOk() (*string, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *WorkspaceDTO) SetImage(v string)`

SetImage sets Image field to given value.


### GetInfo

`func (o *WorkspaceDTO) GetInfo() WorkspaceInfo`

GetInfo returns the Info field if non-nil, zero value otherwise.

### GetInfoOk

`func (o *WorkspaceDTO) GetInfoOk() (*WorkspaceInfo, bool)`

GetInfoOk returns a tuple with the Info field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInfo

`func (o *WorkspaceDTO) SetInfo(v WorkspaceInfo)`

SetInfo sets Info field to given value.

### HasInfo

`func (o *WorkspaceDTO) HasInfo() bool`

HasInfo returns a boolean if a field has been set.

### GetName

`func (o *WorkspaceDTO) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkspaceDTO) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkspaceDTO) SetName(v string)`

SetName sets Name field to given value.


### GetRepository

`func (o *WorkspaceDTO) GetRepository() GitRepository`

GetRepository returns the Repository field if non-nil, zero value otherwise.

### GetRepositoryOk

`func (o *WorkspaceDTO) GetRepositoryOk() (*GitRepository, bool)`

GetRepositoryOk returns a tuple with the Repository field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRepository

`func (o *WorkspaceDTO) SetRepository(v GitRepository)`

SetRepository sets Repository field to given value.


### GetState

`func (o *WorkspaceDTO) GetState() WorkspaceState`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *WorkspaceDTO) GetStateOk() (*WorkspaceState, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *WorkspaceDTO) SetState(v WorkspaceState)`

SetState sets State field to given value.

### HasState

`func (o *WorkspaceDTO) HasState() bool`

HasState returns a boolean if a field has been set.

### GetTargetId

`func (o *WorkspaceDTO) GetTargetId() string`

GetTargetId returns the TargetId field if non-nil, zero value otherwise.

### GetTargetIdOk

`func (o *WorkspaceDTO) GetTargetIdOk() (*string, bool)`

GetTargetIdOk returns a tuple with the TargetId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTargetId

`func (o *WorkspaceDTO) SetTargetId(v string)`

SetTargetId sets TargetId field to given value.


### GetUser

`func (o *WorkspaceDTO) GetUser() string`

GetUser returns the User field if non-nil, zero value otherwise.

### GetUserOk

`func (o *WorkspaceDTO) GetUserOk() (*string, bool)`

GetUserOk returns a tuple with the User field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUser

`func (o *WorkspaceDTO) SetUser(v string)`

SetUser sets User field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


