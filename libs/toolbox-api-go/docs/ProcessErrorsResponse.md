# ProcessErrorsResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Errors** | Pointer to **string** |  | [optional] 
**ProcessName** | Pointer to **string** |  | [optional] 

## Methods

### NewProcessErrorsResponse

`func NewProcessErrorsResponse() *ProcessErrorsResponse`

NewProcessErrorsResponse instantiates a new ProcessErrorsResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProcessErrorsResponseWithDefaults

`func NewProcessErrorsResponseWithDefaults() *ProcessErrorsResponse`

NewProcessErrorsResponseWithDefaults instantiates a new ProcessErrorsResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetErrors

`func (o *ProcessErrorsResponse) GetErrors() string`

GetErrors returns the Errors field if non-nil, zero value otherwise.

### GetErrorsOk

`func (o *ProcessErrorsResponse) GetErrorsOk() (*string, bool)`

GetErrorsOk returns a tuple with the Errors field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetErrors

`func (o *ProcessErrorsResponse) SetErrors(v string)`

SetErrors sets Errors field to given value.

### HasErrors

`func (o *ProcessErrorsResponse) HasErrors() bool`

HasErrors returns a boolean if a field has been set.

### GetProcessName

`func (o *ProcessErrorsResponse) GetProcessName() string`

GetProcessName returns the ProcessName field if non-nil, zero value otherwise.

### GetProcessNameOk

`func (o *ProcessErrorsResponse) GetProcessNameOk() (*string, bool)`

GetProcessNameOk returns a tuple with the ProcessName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProcessName

`func (o *ProcessErrorsResponse) SetProcessName(v string)`

SetProcessName sets ProcessName field to given value.

### HasProcessName

`func (o *ProcessErrorsResponse) HasProcessName() bool`

HasProcessName returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


