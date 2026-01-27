# ExecuteRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Command** | **string** |  | 
**Cwd** | Pointer to **string** | Current working directory | [optional] 
**Timeout** | Pointer to **int32** | Timeout in seconds, defaults to 10 seconds | [optional] 

## Methods

### NewExecuteRequest

`func NewExecuteRequest(command string, ) *ExecuteRequest`

NewExecuteRequest instantiates a new ExecuteRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewExecuteRequestWithDefaults

`func NewExecuteRequestWithDefaults() *ExecuteRequest`

NewExecuteRequestWithDefaults instantiates a new ExecuteRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCommand

`func (o *ExecuteRequest) GetCommand() string`

GetCommand returns the Command field if non-nil, zero value otherwise.

### GetCommandOk

`func (o *ExecuteRequest) GetCommandOk() (*string, bool)`

GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommand

`func (o *ExecuteRequest) SetCommand(v string)`

SetCommand sets Command field to given value.


### GetCwd

`func (o *ExecuteRequest) GetCwd() string`

GetCwd returns the Cwd field if non-nil, zero value otherwise.

### GetCwdOk

`func (o *ExecuteRequest) GetCwdOk() (*string, bool)`

GetCwdOk returns a tuple with the Cwd field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCwd

`func (o *ExecuteRequest) SetCwd(v string)`

SetCwd sets Cwd field to given value.

### HasCwd

`func (o *ExecuteRequest) HasCwd() bool`

HasCwd returns a boolean if a field has been set.

### GetTimeout

`func (o *ExecuteRequest) GetTimeout() int32`

GetTimeout returns the Timeout field if non-nil, zero value otherwise.

### GetTimeoutOk

`func (o *ExecuteRequest) GetTimeoutOk() (*int32, bool)`

GetTimeoutOk returns a tuple with the Timeout field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeout

`func (o *ExecuteRequest) SetTimeout(v int32)`

SetTimeout sets Timeout field to given value.

### HasTimeout

`func (o *ExecuteRequest) HasTimeout() bool`

HasTimeout returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


