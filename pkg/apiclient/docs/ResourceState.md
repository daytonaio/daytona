# ResourceState

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Error** | Pointer to **string** |  | [optional] 
**Name** | [**ModelsResourceStateName**](ModelsResourceStateName.md) |  | 
**UpdatedAt** | **string** |  | 

## Methods

### NewResourceState

`func NewResourceState(name ModelsResourceStateName, updatedAt string, ) *ResourceState`

NewResourceState instantiates a new ResourceState object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResourceStateWithDefaults

`func NewResourceStateWithDefaults() *ResourceState`

NewResourceStateWithDefaults instantiates a new ResourceState object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetError

`func (o *ResourceState) GetError() string`

GetError returns the Error field if non-nil, zero value otherwise.

### GetErrorOk

`func (o *ResourceState) GetErrorOk() (*string, bool)`

GetErrorOk returns a tuple with the Error field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetError

`func (o *ResourceState) SetError(v string)`

SetError sets Error field to given value.

### HasError

`func (o *ResourceState) HasError() bool`

HasError returns a boolean if a field has been set.

### GetName

`func (o *ResourceState) GetName() ModelsResourceStateName`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ResourceState) GetNameOk() (*ModelsResourceStateName, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ResourceState) SetName(v ModelsResourceStateName)`

SetName sets Name field to given value.


### GetUpdatedAt

`func (o *ResourceState) GetUpdatedAt() string`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *ResourceState) GetUpdatedAtOk() (*string, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *ResourceState) SetUpdatedAt(v string)`

SetUpdatedAt sets UpdatedAt field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


