# UpdateJobState

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ErrorMessage** | Pointer to **string** |  | [optional] 
**State** | [**JobState**](JobState.md) |  | 

## Methods

### NewUpdateJobState

`func NewUpdateJobState(state JobState, ) *UpdateJobState`

NewUpdateJobState instantiates a new UpdateJobState object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateJobStateWithDefaults

`func NewUpdateJobStateWithDefaults() *UpdateJobState`

NewUpdateJobStateWithDefaults instantiates a new UpdateJobState object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetErrorMessage

`func (o *UpdateJobState) GetErrorMessage() string`

GetErrorMessage returns the ErrorMessage field if non-nil, zero value otherwise.

### GetErrorMessageOk

`func (o *UpdateJobState) GetErrorMessageOk() (*string, bool)`

GetErrorMessageOk returns a tuple with the ErrorMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetErrorMessage

`func (o *UpdateJobState) SetErrorMessage(v string)`

SetErrorMessage sets ErrorMessage field to given value.

### HasErrorMessage

`func (o *UpdateJobState) HasErrorMessage() bool`

HasErrorMessage returns a boolean if a field has been set.

### GetState

`func (o *UpdateJobState) GetState() JobState`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *UpdateJobState) GetStateOk() (*JobState, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *UpdateJobState) SetState(v JobState)`

SetState sets State field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


