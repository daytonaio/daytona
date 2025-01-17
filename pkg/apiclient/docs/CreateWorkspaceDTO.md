# CreateWorkspaceDTO

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BuildConfig** | Pointer to [**BuildConfig**](BuildConfig.md) |  | [optional] 
**EnvVars** | **map[string]string** |  | 
**GitProviderConfigId** | Pointer to **string** |  | [optional] 
**Id** | **string** |  | 
**Image** | Pointer to **string** |  | [optional] 
**Labels** | **map[string]string** |  | 
**Name** | **string** |  | 
**Source** | [**CreateWorkspaceSourceDTO**](CreateWorkspaceSourceDTO.md) |  | 
**TargetId** | **string** |  | 
**User** | Pointer to **string** |  | [optional] 

## Methods

### NewCreateWorkspaceDTO

`func NewCreateWorkspaceDTO(envVars map[string]string, id string, labels map[string]string, name string, source CreateWorkspaceSourceDTO, targetId string, ) *CreateWorkspaceDTO`

NewCreateWorkspaceDTO instantiates a new CreateWorkspaceDTO object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateWorkspaceDTOWithDefaults

`func NewCreateWorkspaceDTOWithDefaults() *CreateWorkspaceDTO`

NewCreateWorkspaceDTOWithDefaults instantiates a new CreateWorkspaceDTO object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBuildConfig

`func (o *CreateWorkspaceDTO) GetBuildConfig() BuildConfig`

GetBuildConfig returns the BuildConfig field if non-nil, zero value otherwise.

### GetBuildConfigOk

`func (o *CreateWorkspaceDTO) GetBuildConfigOk() (*BuildConfig, bool)`

GetBuildConfigOk returns a tuple with the BuildConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBuildConfig

`func (o *CreateWorkspaceDTO) SetBuildConfig(v BuildConfig)`

SetBuildConfig sets BuildConfig field to given value.

### HasBuildConfig

`func (o *CreateWorkspaceDTO) HasBuildConfig() bool`

HasBuildConfig returns a boolean if a field has been set.

### GetEnvVars

`func (o *CreateWorkspaceDTO) GetEnvVars() map[string]string`

GetEnvVars returns the EnvVars field if non-nil, zero value otherwise.

### GetEnvVarsOk

`func (o *CreateWorkspaceDTO) GetEnvVarsOk() (*map[string]string, bool)`

GetEnvVarsOk returns a tuple with the EnvVars field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvVars

`func (o *CreateWorkspaceDTO) SetEnvVars(v map[string]string)`

SetEnvVars sets EnvVars field to given value.


### GetGitProviderConfigId

`func (o *CreateWorkspaceDTO) GetGitProviderConfigId() string`

GetGitProviderConfigId returns the GitProviderConfigId field if non-nil, zero value otherwise.

### GetGitProviderConfigIdOk

`func (o *CreateWorkspaceDTO) GetGitProviderConfigIdOk() (*string, bool)`

GetGitProviderConfigIdOk returns a tuple with the GitProviderConfigId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGitProviderConfigId

`func (o *CreateWorkspaceDTO) SetGitProviderConfigId(v string)`

SetGitProviderConfigId sets GitProviderConfigId field to given value.

### HasGitProviderConfigId

`func (o *CreateWorkspaceDTO) HasGitProviderConfigId() bool`

HasGitProviderConfigId returns a boolean if a field has been set.

### GetId

`func (o *CreateWorkspaceDTO) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *CreateWorkspaceDTO) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *CreateWorkspaceDTO) SetId(v string)`

SetId sets Id field to given value.


### GetImage

`func (o *CreateWorkspaceDTO) GetImage() string`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *CreateWorkspaceDTO) GetImageOk() (*string, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *CreateWorkspaceDTO) SetImage(v string)`

SetImage sets Image field to given value.

### HasImage

`func (o *CreateWorkspaceDTO) HasImage() bool`

HasImage returns a boolean if a field has been set.

### GetLabels

`func (o *CreateWorkspaceDTO) GetLabels() map[string]string`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *CreateWorkspaceDTO) GetLabelsOk() (*map[string]string, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *CreateWorkspaceDTO) SetLabels(v map[string]string)`

SetLabels sets Labels field to given value.


### GetName

`func (o *CreateWorkspaceDTO) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *CreateWorkspaceDTO) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *CreateWorkspaceDTO) SetName(v string)`

SetName sets Name field to given value.


### GetSource

`func (o *CreateWorkspaceDTO) GetSource() CreateWorkspaceSourceDTO`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *CreateWorkspaceDTO) GetSourceOk() (*CreateWorkspaceSourceDTO, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *CreateWorkspaceDTO) SetSource(v CreateWorkspaceSourceDTO)`

SetSource sets Source field to given value.


### GetTargetId

`func (o *CreateWorkspaceDTO) GetTargetId() string`

GetTargetId returns the TargetId field if non-nil, zero value otherwise.

### GetTargetIdOk

`func (o *CreateWorkspaceDTO) GetTargetIdOk() (*string, bool)`

GetTargetIdOk returns a tuple with the TargetId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTargetId

`func (o *CreateWorkspaceDTO) SetTargetId(v string)`

SetTargetId sets TargetId field to given value.


### GetUser

`func (o *CreateWorkspaceDTO) GetUser() string`

GetUser returns the User field if non-nil, zero value otherwise.

### GetUserOk

`func (o *CreateWorkspaceDTO) GetUserOk() (*string, bool)`

GetUserOk returns a tuple with the User field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUser

`func (o *CreateWorkspaceDTO) SetUser(v string)`

SetUser sets User field to given value.

### HasUser

`func (o *CreateWorkspaceDTO) HasUser() bool`

HasUser returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


