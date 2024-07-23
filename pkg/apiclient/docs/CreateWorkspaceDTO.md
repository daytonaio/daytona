# CreateWorkspaceDTO

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] 
**Name** | Pointer to **string** |  | [optional] 
**Projects** | [**[]CreateProjectDTO**](CreateProjectDTO.md) |  | 
**Target** | Pointer to **string** |  | [optional] 

## Methods

### NewCreateWorkspaceDTO

`func NewCreateWorkspaceDTO(projects []CreateProjectDTO, ) *CreateWorkspaceDTO`

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

### HasId

`func (o *CreateWorkspaceDTO) HasId() bool`

HasId returns a boolean if a field has been set.

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

### HasName

`func (o *CreateWorkspaceDTO) HasName() bool`

HasName returns a boolean if a field has been set.

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


### GetTarget

`func (o *CreateWorkspaceDTO) GetTarget() string`

GetTarget returns the Target field if non-nil, zero value otherwise.

### GetTargetOk

`func (o *CreateWorkspaceDTO) GetTargetOk() (*string, bool)`

GetTargetOk returns a tuple with the Target field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTarget

`func (o *CreateWorkspaceDTO) SetTarget(v string)`

SetTarget sets Target field to given value.

### HasTarget

`func (o *CreateWorkspaceDTO) HasTarget() bool`

HasTarget returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


