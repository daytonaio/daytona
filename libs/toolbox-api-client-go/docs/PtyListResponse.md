# PtyListResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Sessions** | Pointer to [**[]PtySessionInfo**](PtySessionInfo.md) |  | [optional] 

## Methods

### NewPtyListResponse

`func NewPtyListResponse() *PtyListResponse`

NewPtyListResponse instantiates a new PtyListResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPtyListResponseWithDefaults

`func NewPtyListResponseWithDefaults() *PtyListResponse`

NewPtyListResponseWithDefaults instantiates a new PtyListResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSessions

`func (o *PtyListResponse) GetSessions() []PtySessionInfo`

GetSessions returns the Sessions field if non-nil, zero value otherwise.

### GetSessionsOk

`func (o *PtyListResponse) GetSessionsOk() (*[]PtySessionInfo, bool)`

GetSessionsOk returns a tuple with the Sessions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSessions

`func (o *PtyListResponse) SetSessions(v []PtySessionInfo)`

SetSessions sets Sessions field to given value.

### HasSessions

`func (o *PtyListResponse) HasSessions() bool`

HasSessions returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


