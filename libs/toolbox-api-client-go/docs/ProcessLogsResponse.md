# ProcessLogsResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Logs** | Pointer to **string** |  | [optional] 
**ProcessName** | Pointer to **string** |  | [optional] 

## Methods

### NewProcessLogsResponse

`func NewProcessLogsResponse() *ProcessLogsResponse`

NewProcessLogsResponse instantiates a new ProcessLogsResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProcessLogsResponseWithDefaults

`func NewProcessLogsResponseWithDefaults() *ProcessLogsResponse`

NewProcessLogsResponseWithDefaults instantiates a new ProcessLogsResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLogs

`func (o *ProcessLogsResponse) GetLogs() string`

GetLogs returns the Logs field if non-nil, zero value otherwise.

### GetLogsOk

`func (o *ProcessLogsResponse) GetLogsOk() (*string, bool)`

GetLogsOk returns a tuple with the Logs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLogs

`func (o *ProcessLogsResponse) SetLogs(v string)`

SetLogs sets Logs field to given value.

### HasLogs

`func (o *ProcessLogsResponse) HasLogs() bool`

HasLogs returns a boolean if a field has been set.

### GetProcessName

`func (o *ProcessLogsResponse) GetProcessName() string`

GetProcessName returns the ProcessName field if non-nil, zero value otherwise.

### GetProcessNameOk

`func (o *ProcessLogsResponse) GetProcessNameOk() (*string, bool)`

GetProcessNameOk returns a tuple with the ProcessName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProcessName

`func (o *ProcessLogsResponse) SetProcessName(v string)`

SetProcessName sets ProcessName field to given value.

### HasProcessName

`func (o *ProcessLogsResponse) HasProcessName() bool`

HasProcessName returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


