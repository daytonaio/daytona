# Project

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** |  | [optional] 
**Repository** | Pointer to [**GitRepository**](GitRepository.md) |  | [optional] 
**State** | Pointer to [**ProjectState**](ProjectState.md) |  | [optional] 
**Target** | Pointer to **string** |  | [optional] 
**WorkspaceId** | Pointer to **string** |  | [optional] 

## Methods

### NewProject

`func NewProject() *Project`

NewProject instantiates a new Project object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectWithDefaults

`func NewProjectWithDefaults() *Project`

NewProjectWithDefaults instantiates a new Project object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *Project) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Project) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Project) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *Project) HasName() bool`

HasName returns a boolean if a field has been set.

### GetRepository

`func (o *Project) GetRepository() GitRepository`

GetRepository returns the Repository field if non-nil, zero value otherwise.

### GetRepositoryOk

`func (o *Project) GetRepositoryOk() (*GitRepository, bool)`

GetRepositoryOk returns a tuple with the Repository field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRepository

`func (o *Project) SetRepository(v GitRepository)`

SetRepository sets Repository field to given value.

### HasRepository

`func (o *Project) HasRepository() bool`

HasRepository returns a boolean if a field has been set.

### GetState

`func (o *Project) GetState() ProjectState`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *Project) GetStateOk() (*ProjectState, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *Project) SetState(v ProjectState)`

SetState sets State field to given value.

### HasState

`func (o *Project) HasState() bool`

HasState returns a boolean if a field has been set.

### GetTarget

`func (o *Project) GetTarget() string`

GetTarget returns the Target field if non-nil, zero value otherwise.

### GetTargetOk

`func (o *Project) GetTargetOk() (*string, bool)`

GetTargetOk returns a tuple with the Target field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTarget

`func (o *Project) SetTarget(v string)`

SetTarget sets Target field to given value.

### HasTarget

`func (o *Project) HasTarget() bool`

HasTarget returns a boolean if a field has been set.

### GetWorkspaceId

`func (o *Project) GetWorkspaceId() string`

GetWorkspaceId returns the WorkspaceId field if non-nil, zero value otherwise.

### GetWorkspaceIdOk

`func (o *Project) GetWorkspaceIdOk() (*string, bool)`

GetWorkspaceIdOk returns a tuple with the WorkspaceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkspaceId

`func (o *Project) SetWorkspaceId(v string)`

SetWorkspaceId sets WorkspaceId field to given value.

### HasWorkspaceId

`func (o *Project) HasWorkspaceId() bool`

HasWorkspaceId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


