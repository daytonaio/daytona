# Target

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** |  | 
**Name** | **string** |  | 
**TargetConfig** | **string** |  | 
**Workspaces** | [**[]Workspace**](Workspace.md) |  | 

## Methods

### NewTarget

`func NewTarget(id string, name string, targetConfig string, workspaces []Workspace, ) *Target`

NewTarget instantiates a new Target object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTargetWithDefaults

`func NewTargetWithDefaults() *Target`

NewTargetWithDefaults instantiates a new Target object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *Target) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Target) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Target) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *Target) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Target) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Target) SetName(v string)`

SetName sets Name field to given value.


### GetTargetConfig

`func (o *Target) GetTargetConfig() string`

GetTargetConfig returns the TargetConfig field if non-nil, zero value otherwise.

### GetTargetConfigOk

`func (o *Target) GetTargetConfigOk() (*string, bool)`

GetTargetConfigOk returns a tuple with the TargetConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTargetConfig

`func (o *Target) SetTargetConfig(v string)`

SetTargetConfig sets TargetConfig field to given value.


### GetWorkspaces

`func (o *Target) GetWorkspaces() []Workspace`

GetWorkspaces returns the Workspaces field if non-nil, zero value otherwise.

### GetWorkspacesOk

`func (o *Target) GetWorkspacesOk() (*[]Workspace, bool)`

GetWorkspacesOk returns a tuple with the Workspaces field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkspaces

`func (o *Target) SetWorkspaces(v []Workspace)`

SetWorkspaces sets Workspaces field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


