# GithubWebHookCommit

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Added** | Pointer to **[]string** |  | [optional] 
**Author** | Pointer to [**GithubWebHookAuthor**](GithubWebHookAuthor.md) |  | [optional] 
**Committer** | Pointer to [**GithubWebHookAuthor**](GithubWebHookAuthor.md) |  | [optional] 
**Distinct** | Pointer to **bool** |  | [optional] 
**Id** | Pointer to **string** |  | [optional] 
**Message** | Pointer to **string** |  | [optional] 
**Modified** | Pointer to **[]string** |  | [optional] 
**Removed** | Pointer to **[]string** |  | [optional] 
**Timestamp** | Pointer to **string** |  | [optional] 

## Methods

### NewGithubWebHookCommit

`func NewGithubWebHookCommit() *GithubWebHookCommit`

NewGithubWebHookCommit instantiates a new GithubWebHookCommit object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGithubWebHookCommitWithDefaults

`func NewGithubWebHookCommitWithDefaults() *GithubWebHookCommit`

NewGithubWebHookCommitWithDefaults instantiates a new GithubWebHookCommit object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAdded

`func (o *GithubWebHookCommit) GetAdded() []string`

GetAdded returns the Added field if non-nil, zero value otherwise.

### GetAddedOk

`func (o *GithubWebHookCommit) GetAddedOk() (*[]string, bool)`

GetAddedOk returns a tuple with the Added field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAdded

`func (o *GithubWebHookCommit) SetAdded(v []string)`

SetAdded sets Added field to given value.

### HasAdded

`func (o *GithubWebHookCommit) HasAdded() bool`

HasAdded returns a boolean if a field has been set.

### GetAuthor

`func (o *GithubWebHookCommit) GetAuthor() GithubWebHookAuthor`

GetAuthor returns the Author field if non-nil, zero value otherwise.

### GetAuthorOk

`func (o *GithubWebHookCommit) GetAuthorOk() (*GithubWebHookAuthor, bool)`

GetAuthorOk returns a tuple with the Author field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthor

`func (o *GithubWebHookCommit) SetAuthor(v GithubWebHookAuthor)`

SetAuthor sets Author field to given value.

### HasAuthor

`func (o *GithubWebHookCommit) HasAuthor() bool`

HasAuthor returns a boolean if a field has been set.

### GetCommitter

`func (o *GithubWebHookCommit) GetCommitter() GithubWebHookAuthor`

GetCommitter returns the Committer field if non-nil, zero value otherwise.

### GetCommitterOk

`func (o *GithubWebHookCommit) GetCommitterOk() (*GithubWebHookAuthor, bool)`

GetCommitterOk returns a tuple with the Committer field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommitter

`func (o *GithubWebHookCommit) SetCommitter(v GithubWebHookAuthor)`

SetCommitter sets Committer field to given value.

### HasCommitter

`func (o *GithubWebHookCommit) HasCommitter() bool`

HasCommitter returns a boolean if a field has been set.

### GetDistinct

`func (o *GithubWebHookCommit) GetDistinct() bool`

GetDistinct returns the Distinct field if non-nil, zero value otherwise.

### GetDistinctOk

`func (o *GithubWebHookCommit) GetDistinctOk() (*bool, bool)`

GetDistinctOk returns a tuple with the Distinct field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDistinct

`func (o *GithubWebHookCommit) SetDistinct(v bool)`

SetDistinct sets Distinct field to given value.

### HasDistinct

`func (o *GithubWebHookCommit) HasDistinct() bool`

HasDistinct returns a boolean if a field has been set.

### GetId

`func (o *GithubWebHookCommit) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *GithubWebHookCommit) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *GithubWebHookCommit) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *GithubWebHookCommit) HasId() bool`

HasId returns a boolean if a field has been set.

### GetMessage

`func (o *GithubWebHookCommit) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *GithubWebHookCommit) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *GithubWebHookCommit) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *GithubWebHookCommit) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetModified

`func (o *GithubWebHookCommit) GetModified() []string`

GetModified returns the Modified field if non-nil, zero value otherwise.

### GetModifiedOk

`func (o *GithubWebHookCommit) GetModifiedOk() (*[]string, bool)`

GetModifiedOk returns a tuple with the Modified field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetModified

`func (o *GithubWebHookCommit) SetModified(v []string)`

SetModified sets Modified field to given value.

### HasModified

`func (o *GithubWebHookCommit) HasModified() bool`

HasModified returns a boolean if a field has been set.

### GetRemoved

`func (o *GithubWebHookCommit) GetRemoved() []string`

GetRemoved returns the Removed field if non-nil, zero value otherwise.

### GetRemovedOk

`func (o *GithubWebHookCommit) GetRemovedOk() (*[]string, bool)`

GetRemovedOk returns a tuple with the Removed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRemoved

`func (o *GithubWebHookCommit) SetRemoved(v []string)`

SetRemoved sets Removed field to given value.

### HasRemoved

`func (o *GithubWebHookCommit) HasRemoved() bool`

HasRemoved returns a boolean if a field has been set.

### GetTimestamp

`func (o *GithubWebHookCommit) GetTimestamp() string`

GetTimestamp returns the Timestamp field if non-nil, zero value otherwise.

### GetTimestampOk

`func (o *GithubWebHookCommit) GetTimestampOk() (*string, bool)`

GetTimestampOk returns a tuple with the Timestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestamp

`func (o *GithubWebHookCommit) SetTimestamp(v string)`

SetTimestamp sets Timestamp field to given value.

### HasTimestamp

`func (o *GithubWebHookCommit) HasTimestamp() bool`

HasTimestamp returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


