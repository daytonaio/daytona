# ProcessStatusResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ProcessName** | Pointer to **string** |  | [optional] 
**Running** | Pointer to **bool** |  | [optional] 

## Methods

### NewProcessStatusResponse

`func NewProcessStatusResponse() *ProcessStatusResponse`

NewProcessStatusResponse instantiates a new ProcessStatusResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProcessStatusResponseWithDefaults

`func NewProcessStatusResponseWithDefaults() *ProcessStatusResponse`

NewProcessStatusResponseWithDefaults instantiates a new ProcessStatusResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetProcessName

`func (o *ProcessStatusResponse) GetProcessName() string`

GetProcessName returns the ProcessName field if non-nil, zero value otherwise.

### GetProcessNameOk

`func (o *ProcessStatusResponse) GetProcessNameOk() (*string, bool)`

GetProcessNameOk returns a tuple with the ProcessName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProcessName

`func (o *ProcessStatusResponse) SetProcessName(v string)`

SetProcessName sets ProcessName field to given value.

### HasProcessName

`func (o *ProcessStatusResponse) HasProcessName() bool`

HasProcessName returns a boolean if a field has been set.

### GetRunning

`func (o *ProcessStatusResponse) GetRunning() bool`

GetRunning returns the Running field if non-nil, zero value otherwise.

### GetRunningOk

`func (o *ProcessStatusResponse) GetRunningOk() (*bool, bool)`

GetRunningOk returns a tuple with the Running field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunning

`func (o *ProcessStatusResponse) SetRunning(v bool)`

SetRunning sets Running field to given value.

### HasRunning

`func (o *ProcessStatusResponse) HasRunning() bool`

HasRunning returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


