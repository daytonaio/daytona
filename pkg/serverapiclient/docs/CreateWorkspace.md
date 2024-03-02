# CreateWorkspace

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** |  | [optional] 
**Repositories** | Pointer to [**[]Repository**](Repository.md) |  | [optional] 
**Target** | Pointer to **string** |  | [optional] 

## Methods

### NewCreateWorkspace

`func NewCreateWorkspace() *CreateWorkspace`

NewCreateWorkspace instantiates a new CreateWorkspace object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateWorkspaceWithDefaults

`func NewCreateWorkspaceWithDefaults() *CreateWorkspace`

NewCreateWorkspaceWithDefaults instantiates a new CreateWorkspace object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *CreateWorkspace) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *CreateWorkspace) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *CreateWorkspace) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *CreateWorkspace) HasName() bool`

HasName returns a boolean if a field has been set.

### GetRepositories

`func (o *CreateWorkspace) GetRepositories() []Repository`

GetRepositories returns the Repositories field if non-nil, zero value otherwise.

### GetRepositoriesOk

`func (o *CreateWorkspace) GetRepositoriesOk() (*[]Repository, bool)`

GetRepositoriesOk returns a tuple with the Repositories field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRepositories

`func (o *CreateWorkspace) SetRepositories(v []Repository)`

SetRepositories sets Repositories field to given value.

### HasRepositories

`func (o *CreateWorkspace) HasRepositories() bool`

HasRepositories returns a boolean if a field has been set.

### GetTarget

`func (o *CreateWorkspace) GetTarget() string`

GetTarget returns the Target field if non-nil, zero value otherwise.

### GetTargetOk

`func (o *CreateWorkspace) GetTargetOk() (*string, bool)`

GetTargetOk returns a tuple with the Target field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTarget

`func (o *CreateWorkspace) SetTarget(v string)`

SetTarget sets Target field to given value.

### HasTarget

`func (o *CreateWorkspace) HasTarget() bool`

HasTarget returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


