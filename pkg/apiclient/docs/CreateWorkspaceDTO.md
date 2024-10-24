# CreateWorkspaceDTO

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** |  | 
**Name** | **string** |  | 
**Projects** | [**[]CreateProjectDTO**](CreateProjectDTO.md) |  | 
**TargetConfig** | **string** |  | 

## Methods

### NewCreateWorkspaceDTO

`func NewCreateWorkspaceDTO(id string, name string, projects []CreateProjectDTO, targetConfig string, ) *CreateWorkspaceDTO`

NewCreateWorkspaceDTO instantiates a new CreateWorkspaceDTO object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateWorkspaceDTOWithDefaults

`func NewCreateWorkspaceDTOWithDefaults() *CreateWorkspaceDTO`

NewCreateWorkspaceDTOWithDefaults instantiates a new CreateWorkspaceDTO object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

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


### GetProjects

`func (o *CreateWorkspaceDTO) GetProjects() []CreateProjectDTO`

GetProjects returns the Projects field if non-nil, zero value otherwise.

### GetProjectsOk

`func (o *CreateWorkspaceDTO) GetProjectsOk() (*[]CreateProjectDTO, bool)`

GetProjectsOk returns a tuple with the Projects field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjects

`func (o *CreateWorkspaceDTO) SetProjects(v []CreateProjectDTO)`

SetProjects sets Projects field to given value.


### GetTargetConfig

`func (o *CreateWorkspaceDTO) GetTargetConfig() string`

GetTargetConfig returns the TargetConfig field if non-nil, zero value otherwise.

### GetTargetConfigOk

`func (o *CreateWorkspaceDTO) GetTargetConfigOk() (*string, bool)`

GetTargetConfigOk returns a tuple with the TargetConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTargetConfig

`func (o *CreateWorkspaceDTO) SetTargetConfig(v string)`

SetTargetConfig sets TargetConfig field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


