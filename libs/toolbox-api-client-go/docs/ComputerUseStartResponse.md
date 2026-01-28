# ComputerUseStartResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Message** | Pointer to **string** |  | [optional] 
**Status** | Pointer to [**map[string]ProcessStatus**](ProcessStatus.md) |  | [optional] 

## Methods

### NewComputerUseStartResponse

`func NewComputerUseStartResponse() *ComputerUseStartResponse`

NewComputerUseStartResponse instantiates a new ComputerUseStartResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewComputerUseStartResponseWithDefaults

`func NewComputerUseStartResponseWithDefaults() *ComputerUseStartResponse`

NewComputerUseStartResponseWithDefaults instantiates a new ComputerUseStartResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMessage

`func (o *ComputerUseStartResponse) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *ComputerUseStartResponse) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *ComputerUseStartResponse) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *ComputerUseStartResponse) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetStatus

`func (o *ComputerUseStartResponse) GetStatus() map[string]ProcessStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *ComputerUseStartResponse) GetStatusOk() (*map[string]ProcessStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *ComputerUseStartResponse) SetStatus(v map[string]ProcessStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *ComputerUseStartResponse) HasStatus() bool`

HasStatus returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


