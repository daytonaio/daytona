# ProcessStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AutoRestart** | Pointer to **bool** |  | [optional] 
**Pid** | Pointer to **int32** |  | [optional] 
**Priority** | Pointer to **int32** |  | [optional] 
**Running** | Pointer to **bool** |  | [optional] 

## Methods

### NewProcessStatus

`func NewProcessStatus() *ProcessStatus`

NewProcessStatus instantiates a new ProcessStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProcessStatusWithDefaults

`func NewProcessStatusWithDefaults() *ProcessStatus`

NewProcessStatusWithDefaults instantiates a new ProcessStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAutoRestart

`func (o *ProcessStatus) GetAutoRestart() bool`

GetAutoRestart returns the AutoRestart field if non-nil, zero value otherwise.

### GetAutoRestartOk

`func (o *ProcessStatus) GetAutoRestartOk() (*bool, bool)`

GetAutoRestartOk returns a tuple with the AutoRestart field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoRestart

`func (o *ProcessStatus) SetAutoRestart(v bool)`

SetAutoRestart sets AutoRestart field to given value.

### HasAutoRestart

`func (o *ProcessStatus) HasAutoRestart() bool`

HasAutoRestart returns a boolean if a field has been set.

### GetPid

`func (o *ProcessStatus) GetPid() int32`

GetPid returns the Pid field if non-nil, zero value otherwise.

### GetPidOk

`func (o *ProcessStatus) GetPidOk() (*int32, bool)`

GetPidOk returns a tuple with the Pid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPid

`func (o *ProcessStatus) SetPid(v int32)`

SetPid sets Pid field to given value.

### HasPid

`func (o *ProcessStatus) HasPid() bool`

HasPid returns a boolean if a field has been set.

### GetPriority

`func (o *ProcessStatus) GetPriority() int32`

GetPriority returns the Priority field if non-nil, zero value otherwise.

### GetPriorityOk

`func (o *ProcessStatus) GetPriorityOk() (*int32, bool)`

GetPriorityOk returns a tuple with the Priority field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriority

`func (o *ProcessStatus) SetPriority(v int32)`

SetPriority sets Priority field to given value.

### HasPriority

`func (o *ProcessStatus) HasPriority() bool`

HasPriority returns a boolean if a field has been set.

### GetRunning

`func (o *ProcessStatus) GetRunning() bool`

GetRunning returns the Running field if non-nil, zero value otherwise.

### GetRunningOk

`func (o *ProcessStatus) GetRunningOk() (*bool, bool)`

GetRunningOk returns a tuple with the Running field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunning

`func (o *ProcessStatus) SetRunning(v bool)`

SetRunning sets Running field to given value.

### HasRunning

`func (o *ProcessStatus) HasRunning() bool`

HasRunning returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


