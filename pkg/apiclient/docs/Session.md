# Session

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Alias** | Pointer to **string** |  | [optional] 
**Commands** | [**[]Command**](Command.md) |  | 
**SessionId** | **string** |  | 

## Methods

### NewSession

`func NewSession(commands []Command, sessionId string, ) *Session`

NewSession instantiates a new Session object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSessionWithDefaults

`func NewSessionWithDefaults() *Session`

NewSessionWithDefaults instantiates a new Session object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAlias

`func (o *Session) GetAlias() string`

GetAlias returns the Alias field if non-nil, zero value otherwise.

### GetAliasOk

`func (o *Session) GetAliasOk() (*string, bool)`

GetAliasOk returns a tuple with the Alias field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAlias

`func (o *Session) SetAlias(v string)`

SetAlias sets Alias field to given value.

### HasAlias

`func (o *Session) HasAlias() bool`

HasAlias returns a boolean if a field has been set.

### GetCommands

`func (o *Session) GetCommands() []Command`

GetCommands returns the Commands field if non-nil, zero value otherwise.

### GetCommandsOk

`func (o *Session) GetCommandsOk() (*[]Command, bool)`

GetCommandsOk returns a tuple with the Commands field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommands

`func (o *Session) SetCommands(v []Command)`

SetCommands sets Commands field to given value.


### GetSessionId

`func (o *Session) GetSessionId() string`

GetSessionId returns the SessionId field if non-nil, zero value otherwise.

### GetSessionIdOk

`func (o *Session) GetSessionIdOk() (*string, bool)`

GetSessionIdOk returns a tuple with the SessionId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSessionId

`func (o *Session) SetSessionId(v string)`

SetSessionId sets SessionId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


