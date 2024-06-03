# CreateWorkspaceRequestProject

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Build** | Pointer to [**ProjectBuild**](ProjectBuild.md) |  | [optional] 
**EnvVars** | Pointer to **map[string]string** |  | [optional] 
**Image** | Pointer to **string** |  | [optional] 
**Name** | **string** |  | 
**PostStartCommands** | Pointer to **[]string** |  | [optional] 
**Source** | Pointer to [**CreateWorkspaceRequestProjectSource**](CreateWorkspaceRequestProjectSource.md) |  | [optional] 
**User** | Pointer to **string** |  | [optional] 

## Methods

### NewCreateWorkspaceRequestProject

`func NewCreateWorkspaceRequestProject(name string, ) *CreateWorkspaceRequestProject`

NewCreateWorkspaceRequestProject instantiates a new CreateWorkspaceRequestProject object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateWorkspaceRequestProjectWithDefaults

`func NewCreateWorkspaceRequestProjectWithDefaults() *CreateWorkspaceRequestProject`

NewCreateWorkspaceRequestProjectWithDefaults instantiates a new CreateWorkspaceRequestProject object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBuild

`func (o *CreateWorkspaceRequestProject) GetBuild() ProjectBuild`

GetBuild returns the Build field if non-nil, zero value otherwise.

### GetBuildOk

`func (o *CreateWorkspaceRequestProject) GetBuildOk() (*ProjectBuild, bool)`

GetBuildOk returns a tuple with the Build field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBuild

`func (o *CreateWorkspaceRequestProject) SetBuild(v ProjectBuild)`

SetBuild sets Build field to given value.

### HasBuild

`func (o *CreateWorkspaceRequestProject) HasBuild() bool`

HasBuild returns a boolean if a field has been set.

### GetEnvVars

`func (o *CreateWorkspaceRequestProject) GetEnvVars() map[string]string`

GetEnvVars returns the EnvVars field if non-nil, zero value otherwise.

### GetEnvVarsOk

`func (o *CreateWorkspaceRequestProject) GetEnvVarsOk() (*map[string]string, bool)`

GetEnvVarsOk returns a tuple with the EnvVars field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvVars

`func (o *CreateWorkspaceRequestProject) SetEnvVars(v map[string]string)`

SetEnvVars sets EnvVars field to given value.

### HasEnvVars

`func (o *CreateWorkspaceRequestProject) HasEnvVars() bool`

HasEnvVars returns a boolean if a field has been set.

### GetImage

`func (o *CreateWorkspaceRequestProject) GetImage() string`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *CreateWorkspaceRequestProject) GetImageOk() (*string, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *CreateWorkspaceRequestProject) SetImage(v string)`

SetImage sets Image field to given value.

### HasImage

`func (o *CreateWorkspaceRequestProject) HasImage() bool`

HasImage returns a boolean if a field has been set.

### GetName

`func (o *CreateWorkspaceRequestProject) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *CreateWorkspaceRequestProject) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *CreateWorkspaceRequestProject) SetName(v string)`

SetName sets Name field to given value.


### GetPostStartCommands

`func (o *CreateWorkspaceRequestProject) GetPostStartCommands() []string`

GetPostStartCommands returns the PostStartCommands field if non-nil, zero value otherwise.

### GetPostStartCommandsOk

`func (o *CreateWorkspaceRequestProject) GetPostStartCommandsOk() (*[]string, bool)`

GetPostStartCommandsOk returns a tuple with the PostStartCommands field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPostStartCommands

`func (o *CreateWorkspaceRequestProject) SetPostStartCommands(v []string)`

SetPostStartCommands sets PostStartCommands field to given value.

### HasPostStartCommands

`func (o *CreateWorkspaceRequestProject) HasPostStartCommands() bool`

HasPostStartCommands returns a boolean if a field has been set.

### GetSource

`func (o *CreateWorkspaceRequestProject) GetSource() CreateWorkspaceRequestProjectSource`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *CreateWorkspaceRequestProject) GetSourceOk() (*CreateWorkspaceRequestProjectSource, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *CreateWorkspaceRequestProject) SetSource(v CreateWorkspaceRequestProjectSource)`

SetSource sets Source field to given value.

### HasSource

`func (o *CreateWorkspaceRequestProject) HasSource() bool`

HasSource returns a boolean if a field has been set.

### GetUser

`func (o *CreateWorkspaceRequestProject) GetUser() string`

GetUser returns the User field if non-nil, zero value otherwise.

### GetUserOk

`func (o *CreateWorkspaceRequestProject) GetUserOk() (*string, bool)`

GetUserOk returns a tuple with the User field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUser

`func (o *CreateWorkspaceRequestProject) SetUser(v string)`

SetUser sets User field to given value.

### HasUser

`func (o *CreateWorkspaceRequestProject) HasUser() bool`

HasUser returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


