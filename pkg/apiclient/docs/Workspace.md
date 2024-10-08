# Workspace

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** |  | 
**Name** | **string** |  | 
**Projects** | [**[]Project**](Project.md) |  | 
**Target** | **string** |  | 

## Methods

### NewWorkspace

`func NewWorkspace(id string, name string, projects []Project, target string, ) *Workspace`

NewWorkspace instantiates a new Workspace object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspaceWithDefaults

`func NewWorkspaceWithDefaults() *Workspace`

NewWorkspaceWithDefaults instantiates a new Workspace object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *Workspace) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Workspace) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Workspace) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *Workspace) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Workspace) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Workspace) SetName(v string)`

SetName sets Name field to given value.


### GetProjects

`func (o *Workspace) GetProjects() []Project`

GetProjects returns the Projects field if non-nil, zero value otherwise.

### GetProjectsOk

`func (o *Workspace) GetProjectsOk() (*[]Project, bool)`

GetProjectsOk returns a tuple with the Projects field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjects

`func (o *Workspace) SetProjects(v []Project)`

SetProjects sets Projects field to given value.


### GetTarget

`func (o *Workspace) GetTarget() string`

GetTarget returns the Target field if non-nil, zero value otherwise.

### GetTargetOk

`func (o *Workspace) GetTargetOk() (*string, bool)`

GetTargetOk returns a tuple with the Target field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTarget

`func (o *Workspace) SetTarget(v string)`

SetTarget sets Target field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


