# CreateWorkspaceRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] 
**Name** | Pointer to **string** |  | [optional] 
**Projects** | [**[]CreateWorkspaceRequestProject**](CreateWorkspaceRequestProject.md) |  | 
**Target** | Pointer to **string** |  | [optional] 

## Methods

### NewCreateWorkspaceRequest

`func NewCreateWorkspaceRequest(projects []CreateWorkspaceRequestProject, ) *CreateWorkspaceRequest`

NewCreateWorkspaceRequest instantiates a new CreateWorkspaceRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateWorkspaceRequestWithDefaults

`func NewCreateWorkspaceRequestWithDefaults() *CreateWorkspaceRequest`

NewCreateWorkspaceRequestWithDefaults instantiates a new CreateWorkspaceRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *CreateWorkspaceRequest) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *CreateWorkspaceRequest) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *CreateWorkspaceRequest) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *CreateWorkspaceRequest) HasId() bool`

HasId returns a boolean if a field has been set.

### GetName

`func (o *CreateWorkspaceRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *CreateWorkspaceRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *CreateWorkspaceRequest) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *CreateWorkspaceRequest) HasName() bool`

HasName returns a boolean if a field has been set.

### GetProjects

`func (o *CreateWorkspaceRequest) GetProjects() []CreateWorkspaceRequestProject`

GetProjects returns the Projects field if non-nil, zero value otherwise.

### GetProjectsOk

`func (o *CreateWorkspaceRequest) GetProjectsOk() (*[]CreateWorkspaceRequestProject, bool)`

GetProjectsOk returns a tuple with the Projects field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjects

`func (o *CreateWorkspaceRequest) SetProjects(v []CreateWorkspaceRequestProject)`

SetProjects sets Projects field to given value.


### GetTarget

`func (o *CreateWorkspaceRequest) GetTarget() string`

GetTarget returns the Target field if non-nil, zero value otherwise.

### GetTargetOk

`func (o *CreateWorkspaceRequest) GetTargetOk() (*string, bool)`

GetTargetOk returns a tuple with the Target field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTarget

`func (o *CreateWorkspaceRequest) SetTarget(v string)`

SetTarget sets Target field to given value.

### HasTarget

`func (o *CreateWorkspaceRequest) HasTarget() bool`

HasTarget returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


