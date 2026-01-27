# ProcessRestartResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Message** | Pointer to **string** |  | [optional] 
**ProcessName** | Pointer to **string** |  | [optional] 

## Methods

### NewProcessRestartResponse

`func NewProcessRestartResponse() *ProcessRestartResponse`

NewProcessRestartResponse instantiates a new ProcessRestartResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProcessRestartResponseWithDefaults

`func NewProcessRestartResponseWithDefaults() *ProcessRestartResponse`

NewProcessRestartResponseWithDefaults instantiates a new ProcessRestartResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMessage

`func (o *ProcessRestartResponse) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *ProcessRestartResponse) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *ProcessRestartResponse) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *ProcessRestartResponse) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetProcessName

`func (o *ProcessRestartResponse) GetProcessName() string`

GetProcessName returns the ProcessName field if non-nil, zero value otherwise.

### GetProcessNameOk

`func (o *ProcessRestartResponse) GetProcessNameOk() (*string, bool)`

GetProcessNameOk returns a tuple with the ProcessName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProcessName

`func (o *ProcessRestartResponse) SetProcessName(v string)`

SetProcessName sets ProcessName field to given value.

### HasProcessName

`func (o *ProcessRestartResponse) HasProcessName() bool`

HasProcessName returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


