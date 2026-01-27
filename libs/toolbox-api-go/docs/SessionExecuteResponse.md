# SessionExecuteResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CmdId** | Pointer to **string** |  | [optional] 
**ExitCode** | Pointer to **int32** |  | [optional] 
**Output** | Pointer to **string** |  | [optional] 
**Stderr** | Pointer to **string** |  | [optional] 
**Stdout** | Pointer to **string** |  | [optional] 

## Methods

### NewSessionExecuteResponse

`func NewSessionExecuteResponse() *SessionExecuteResponse`

NewSessionExecuteResponse instantiates a new SessionExecuteResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSessionExecuteResponseWithDefaults

`func NewSessionExecuteResponseWithDefaults() *SessionExecuteResponse`

NewSessionExecuteResponseWithDefaults instantiates a new SessionExecuteResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCmdId

`func (o *SessionExecuteResponse) GetCmdId() string`

GetCmdId returns the CmdId field if non-nil, zero value otherwise.

### GetCmdIdOk

`func (o *SessionExecuteResponse) GetCmdIdOk() (*string, bool)`

GetCmdIdOk returns a tuple with the CmdId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCmdId

`func (o *SessionExecuteResponse) SetCmdId(v string)`

SetCmdId sets CmdId field to given value.

### HasCmdId

`func (o *SessionExecuteResponse) HasCmdId() bool`

HasCmdId returns a boolean if a field has been set.

### GetExitCode

`func (o *SessionExecuteResponse) GetExitCode() int32`

GetExitCode returns the ExitCode field if non-nil, zero value otherwise.

### GetExitCodeOk

`func (o *SessionExecuteResponse) GetExitCodeOk() (*int32, bool)`

GetExitCodeOk returns a tuple with the ExitCode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExitCode

`func (o *SessionExecuteResponse) SetExitCode(v int32)`

SetExitCode sets ExitCode field to given value.

### HasExitCode

`func (o *SessionExecuteResponse) HasExitCode() bool`

HasExitCode returns a boolean if a field has been set.

### GetOutput

`func (o *SessionExecuteResponse) GetOutput() string`

GetOutput returns the Output field if non-nil, zero value otherwise.

### GetOutputOk

`func (o *SessionExecuteResponse) GetOutputOk() (*string, bool)`

GetOutputOk returns a tuple with the Output field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOutput

`func (o *SessionExecuteResponse) SetOutput(v string)`

SetOutput sets Output field to given value.

### HasOutput

`func (o *SessionExecuteResponse) HasOutput() bool`

HasOutput returns a boolean if a field has been set.

### GetStderr

`func (o *SessionExecuteResponse) GetStderr() string`

GetStderr returns the Stderr field if non-nil, zero value otherwise.

### GetStderrOk

`func (o *SessionExecuteResponse) GetStderrOk() (*string, bool)`

GetStderrOk returns a tuple with the Stderr field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStderr

`func (o *SessionExecuteResponse) SetStderr(v string)`

SetStderr sets Stderr field to given value.

### HasStderr

`func (o *SessionExecuteResponse) HasStderr() bool`

HasStderr returns a boolean if a field has been set.

### GetStdout

`func (o *SessionExecuteResponse) GetStdout() string`

GetStdout returns the Stdout field if non-nil, zero value otherwise.

### GetStdoutOk

`func (o *SessionExecuteResponse) GetStdoutOk() (*string, bool)`

GetStdoutOk returns a tuple with the Stdout field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStdout

`func (o *SessionExecuteResponse) SetStdout(v string)`

SetStdout sets Stdout field to given value.

### HasStdout

`func (o *SessionExecuteResponse) HasStdout() bool`

HasStdout returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


