# SessionDTO

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Alias** | Pointer to **string** |  | [optional] 
**SessionId** | **string** |  | 

## Methods

### NewSessionDTO

`func NewSessionDTO(sessionId string, ) *SessionDTO`

NewSessionDTO instantiates a new SessionDTO object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSessionDTOWithDefaults

`func NewSessionDTOWithDefaults() *SessionDTO`

NewSessionDTOWithDefaults instantiates a new SessionDTO object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAlias

`func (o *SessionDTO) GetAlias() string`

GetAlias returns the Alias field if non-nil, zero value otherwise.

### GetAliasOk

`func (o *SessionDTO) GetAliasOk() (*string, bool)`

GetAliasOk returns a tuple with the Alias field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAlias

`func (o *SessionDTO) SetAlias(v string)`

SetAlias sets Alias field to given value.

### HasAlias

`func (o *SessionDTO) HasAlias() bool`

HasAlias returns a boolean if a field has been set.

### GetSessionId

`func (o *SessionDTO) GetSessionId() string`

GetSessionId returns the SessionId field if non-nil, zero value otherwise.

### GetSessionIdOk

`func (o *SessionDTO) GetSessionIdOk() (*string, bool)`

GetSessionIdOk returns a tuple with the SessionId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSessionId

`func (o *SessionDTO) SetSessionId(v string)`

SetSessionId sets SessionId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


