# TargetDTO

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** |  | 
**Info** | Pointer to [**TargetInfo**](TargetInfo.md) |  | [optional] 
**Name** | **string** |  | 
**TargetConfig** | **string** |  | 
**Workspaces** | [**[]Workspace**](Workspace.md) |  | 

## Methods

### NewTargetDTO

`func NewTargetDTO(id string, name string, targetConfig string, workspaces []Workspace, ) *TargetDTO`

NewTargetDTO instantiates a new TargetDTO object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTargetDTOWithDefaults

`func NewTargetDTOWithDefaults() *TargetDTO`

NewTargetDTOWithDefaults instantiates a new TargetDTO object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *TargetDTO) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *TargetDTO) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *TargetDTO) SetId(v string)`

SetId sets Id field to given value.


### GetInfo

`func (o *TargetDTO) GetInfo() TargetInfo`

GetInfo returns the Info field if non-nil, zero value otherwise.

### GetInfoOk

`func (o *TargetDTO) GetInfoOk() (*TargetInfo, bool)`

GetInfoOk returns a tuple with the Info field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInfo

`func (o *TargetDTO) SetInfo(v TargetInfo)`

SetInfo sets Info field to given value.

### HasInfo

`func (o *TargetDTO) HasInfo() bool`

HasInfo returns a boolean if a field has been set.

### GetName

`func (o *TargetDTO) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *TargetDTO) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *TargetDTO) SetName(v string)`

SetName sets Name field to given value.


### GetTargetConfig

`func (o *TargetDTO) GetTargetConfig() string`

GetTargetConfig returns the TargetConfig field if non-nil, zero value otherwise.

### GetTargetConfigOk

`func (o *TargetDTO) GetTargetConfigOk() (*string, bool)`

GetTargetConfigOk returns a tuple with the TargetConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTargetConfig

`func (o *TargetDTO) SetTargetConfig(v string)`

SetTargetConfig sets TargetConfig field to given value.


### GetWorkspaces

`func (o *TargetDTO) GetWorkspaces() []Workspace`

GetWorkspaces returns the Workspaces field if non-nil, zero value otherwise.

### GetWorkspacesOk

`func (o *TargetDTO) GetWorkspacesOk() (*[]Workspace, bool)`

GetWorkspacesOk returns a tuple with the Workspaces field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkspaces

`func (o *TargetDTO) SetWorkspaces(v []Workspace)`

SetWorkspaces sets Workspaces field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


