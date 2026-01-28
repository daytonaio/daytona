# SessionExecuteRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Async** | Pointer to **bool** |  | [optional] 
**Command** | **string** |  | 
**RunAsync** | Pointer to **bool** |  | [optional] 

## Methods

### NewSessionExecuteRequest

`func NewSessionExecuteRequest(command string, ) *SessionExecuteRequest`

NewSessionExecuteRequest instantiates a new SessionExecuteRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSessionExecuteRequestWithDefaults

`func NewSessionExecuteRequestWithDefaults() *SessionExecuteRequest`

NewSessionExecuteRequestWithDefaults instantiates a new SessionExecuteRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAsync

`func (o *SessionExecuteRequest) GetAsync() bool`

GetAsync returns the Async field if non-nil, zero value otherwise.

### GetAsyncOk

`func (o *SessionExecuteRequest) GetAsyncOk() (*bool, bool)`

GetAsyncOk returns a tuple with the Async field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAsync

`func (o *SessionExecuteRequest) SetAsync(v bool)`

SetAsync sets Async field to given value.

### HasAsync

`func (o *SessionExecuteRequest) HasAsync() bool`

HasAsync returns a boolean if a field has been set.

### GetCommand

`func (o *SessionExecuteRequest) GetCommand() string`

GetCommand returns the Command field if non-nil, zero value otherwise.

### GetCommandOk

`func (o *SessionExecuteRequest) GetCommandOk() (*string, bool)`

GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommand

`func (o *SessionExecuteRequest) SetCommand(v string)`

SetCommand sets Command field to given value.


### GetRunAsync

`func (o *SessionExecuteRequest) GetRunAsync() bool`

GetRunAsync returns the RunAsync field if non-nil, zero value otherwise.

### GetRunAsyncOk

`func (o *SessionExecuteRequest) GetRunAsyncOk() (*bool, bool)`

GetRunAsyncOk returns a tuple with the RunAsync field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsync

`func (o *SessionExecuteRequest) SetRunAsync(v bool)`

SetRunAsync sets RunAsync field to given value.

### HasRunAsync

`func (o *SessionExecuteRequest) HasRunAsync() bool`

HasRunAsync returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


