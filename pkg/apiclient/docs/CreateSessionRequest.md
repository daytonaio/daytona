# CreateSessionRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Alias** | Pointer to **string** |  | [optional] 
**SessionId** | **string** |  | 

## Methods

### NewCreateSessionRequest

`func NewCreateSessionRequest(sessionId string, ) *CreateSessionRequest`

NewCreateSessionRequest instantiates a new CreateSessionRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateSessionRequestWithDefaults

`func NewCreateSessionRequestWithDefaults() *CreateSessionRequest`

NewCreateSessionRequestWithDefaults instantiates a new CreateSessionRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAlias

`func (o *CreateSessionRequest) GetAlias() string`

GetAlias returns the Alias field if non-nil, zero value otherwise.

### GetAliasOk

`func (o *CreateSessionRequest) GetAliasOk() (*string, bool)`

GetAliasOk returns a tuple with the Alias field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAlias

`func (o *CreateSessionRequest) SetAlias(v string)`

SetAlias sets Alias field to given value.

### HasAlias

`func (o *CreateSessionRequest) HasAlias() bool`

HasAlias returns a boolean if a field has been set.

### GetSessionId

`func (o *CreateSessionRequest) GetSessionId() string`

GetSessionId returns the SessionId field if non-nil, zero value otherwise.

### GetSessionIdOk

`func (o *CreateSessionRequest) GetSessionIdOk() (*string, bool)`

GetSessionIdOk returns a tuple with the SessionId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSessionId

`func (o *CreateSessionRequest) SetSessionId(v string)`

SetSessionId sets SessionId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


