# WorkspaceDTO

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** |  | 
**Info** | Pointer to [**WorkspaceInfo**](WorkspaceInfo.md) |  | [optional] 
**Name** | **string** |  | 
**Projects** | [**[]Project**](Project.md) |  | 
**Target** | **string** |  | 

## Methods

### NewWorkspaceDTO

`func NewWorkspaceDTO(id string, name string, projects []Project, target string, ) *WorkspaceDTO`

NewWorkspaceDTO instantiates a new WorkspaceDTO object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspaceDTOWithDefaults

`func NewWorkspaceDTOWithDefaults() *WorkspaceDTO`

NewWorkspaceDTOWithDefaults instantiates a new WorkspaceDTO object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

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


### GetProjects

`func (o *WorkspaceDTO) GetProjects() []Project`

GetProjects returns the Projects field if non-nil, zero value otherwise.

### GetProjectsOk

`func (o *WorkspaceDTO) GetProjectsOk() (*[]Project, bool)`

GetProjectsOk returns a tuple with the Projects field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjects

`func (o *WorkspaceDTO) SetProjects(v []Project)`

SetProjects sets Projects field to given value.


### GetTarget

`func (o *WorkspaceDTO) GetTarget() string`

GetTarget returns the Target field if non-nil, zero value otherwise.

### GetTargetOk

`func (o *WorkspaceDTO) GetTargetOk() (*string, bool)`

GetTargetOk returns a tuple with the Target field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTarget

`func (o *WorkspaceDTO) SetTarget(v string)`

SetTarget sets Target field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


