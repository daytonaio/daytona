# GithubWebHookPayload

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**After** | Pointer to **string** |  | [optional] 
**Before** | Pointer to **string** |  | [optional] 
**Commits** | Pointer to [**[]GithubWebHookCommit**](GithubWebHookCommit.md) |  | [optional] 
**Compare** | Pointer to **string** |  | [optional] 
**Created** | Pointer to **bool** |  | [optional] 
**Deleted** | Pointer to **bool** |  | [optional] 
**Forced** | Pointer to **bool** |  | [optional] 
**HeadCommit** | Pointer to [**GithubWebHookCommit**](GithubWebHookCommit.md) |  | [optional] 
**Pusher** | Pointer to [**GithubUser**](GithubUser.md) |  | [optional] 
**Ref** | Pointer to **string** |  | [optional] 
**Repository** | Pointer to [**GithubRepository**](GithubRepository.md) |  | [optional] 
**Sender** | Pointer to [**GithubUser**](GithubUser.md) |  | [optional] 

## Methods

### NewGithubWebHookPayload

`func NewGithubWebHookPayload() *GithubWebHookPayload`

NewGithubWebHookPayload instantiates a new GithubWebHookPayload object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGithubWebHookPayloadWithDefaults

`func NewGithubWebHookPayloadWithDefaults() *GithubWebHookPayload`

NewGithubWebHookPayloadWithDefaults instantiates a new GithubWebHookPayload object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAfter

`func (o *GithubWebHookPayload) GetAfter() string`

GetAfter returns the After field if non-nil, zero value otherwise.

### GetAfterOk

`func (o *GithubWebHookPayload) GetAfterOk() (*string, bool)`

GetAfterOk returns a tuple with the After field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAfter

`func (o *GithubWebHookPayload) SetAfter(v string)`

SetAfter sets After field to given value.

### HasAfter

`func (o *GithubWebHookPayload) HasAfter() bool`

HasAfter returns a boolean if a field has been set.

### GetBefore

`func (o *GithubWebHookPayload) GetBefore() string`

GetBefore returns the Before field if non-nil, zero value otherwise.

### GetBeforeOk

`func (o *GithubWebHookPayload) GetBeforeOk() (*string, bool)`

GetBeforeOk returns a tuple with the Before field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBefore

`func (o *GithubWebHookPayload) SetBefore(v string)`

SetBefore sets Before field to given value.

### HasBefore

`func (o *GithubWebHookPayload) HasBefore() bool`

HasBefore returns a boolean if a field has been set.

### GetCommits

`func (o *GithubWebHookPayload) GetCommits() []GithubWebHookCommit`

GetCommits returns the Commits field if non-nil, zero value otherwise.

### GetCommitsOk

`func (o *GithubWebHookPayload) GetCommitsOk() (*[]GithubWebHookCommit, bool)`

GetCommitsOk returns a tuple with the Commits field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommits

`func (o *GithubWebHookPayload) SetCommits(v []GithubWebHookCommit)`

SetCommits sets Commits field to given value.

### HasCommits

`func (o *GithubWebHookPayload) HasCommits() bool`

HasCommits returns a boolean if a field has been set.

### GetCompare

`func (o *GithubWebHookPayload) GetCompare() string`

GetCompare returns the Compare field if non-nil, zero value otherwise.

### GetCompareOk

`func (o *GithubWebHookPayload) GetCompareOk() (*string, bool)`

GetCompareOk returns a tuple with the Compare field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompare

`func (o *GithubWebHookPayload) SetCompare(v string)`

SetCompare sets Compare field to given value.

### HasCompare

`func (o *GithubWebHookPayload) HasCompare() bool`

HasCompare returns a boolean if a field has been set.

### GetCreated

`func (o *GithubWebHookPayload) GetCreated() bool`

GetCreated returns the Created field if non-nil, zero value otherwise.

### GetCreatedOk

`func (o *GithubWebHookPayload) GetCreatedOk() (*bool, bool)`

GetCreatedOk returns a tuple with the Created field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreated

`func (o *GithubWebHookPayload) SetCreated(v bool)`

SetCreated sets Created field to given value.

### HasCreated

`func (o *GithubWebHookPayload) HasCreated() bool`

HasCreated returns a boolean if a field has been set.

### GetDeleted

`func (o *GithubWebHookPayload) GetDeleted() bool`

GetDeleted returns the Deleted field if non-nil, zero value otherwise.

### GetDeletedOk

`func (o *GithubWebHookPayload) GetDeletedOk() (*bool, bool)`

GetDeletedOk returns a tuple with the Deleted field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeleted

`func (o *GithubWebHookPayload) SetDeleted(v bool)`

SetDeleted sets Deleted field to given value.

### HasDeleted

`func (o *GithubWebHookPayload) HasDeleted() bool`

HasDeleted returns a boolean if a field has been set.

### GetForced

`func (o *GithubWebHookPayload) GetForced() bool`

GetForced returns the Forced field if non-nil, zero value otherwise.

### GetForcedOk

`func (o *GithubWebHookPayload) GetForcedOk() (*bool, bool)`

GetForcedOk returns a tuple with the Forced field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetForced

`func (o *GithubWebHookPayload) SetForced(v bool)`

SetForced sets Forced field to given value.

### HasForced

`func (o *GithubWebHookPayload) HasForced() bool`

HasForced returns a boolean if a field has been set.

### GetHeadCommit

`func (o *GithubWebHookPayload) GetHeadCommit() GithubWebHookCommit`

GetHeadCommit returns the HeadCommit field if non-nil, zero value otherwise.

### GetHeadCommitOk

`func (o *GithubWebHookPayload) GetHeadCommitOk() (*GithubWebHookCommit, bool)`

GetHeadCommitOk returns a tuple with the HeadCommit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHeadCommit

`func (o *GithubWebHookPayload) SetHeadCommit(v GithubWebHookCommit)`

SetHeadCommit sets HeadCommit field to given value.

### HasHeadCommit

`func (o *GithubWebHookPayload) HasHeadCommit() bool`

HasHeadCommit returns a boolean if a field has been set.

### GetPusher

`func (o *GithubWebHookPayload) GetPusher() GithubUser`

GetPusher returns the Pusher field if non-nil, zero value otherwise.

### GetPusherOk

`func (o *GithubWebHookPayload) GetPusherOk() (*GithubUser, bool)`

GetPusherOk returns a tuple with the Pusher field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPusher

`func (o *GithubWebHookPayload) SetPusher(v GithubUser)`

SetPusher sets Pusher field to given value.

### HasPusher

`func (o *GithubWebHookPayload) HasPusher() bool`

HasPusher returns a boolean if a field has been set.

### GetRef

`func (o *GithubWebHookPayload) GetRef() string`

GetRef returns the Ref field if non-nil, zero value otherwise.

### GetRefOk

`func (o *GithubWebHookPayload) GetRefOk() (*string, bool)`

GetRefOk returns a tuple with the Ref field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRef

`func (o *GithubWebHookPayload) SetRef(v string)`

SetRef sets Ref field to given value.

### HasRef

`func (o *GithubWebHookPayload) HasRef() bool`

HasRef returns a boolean if a field has been set.

### GetRepository

`func (o *GithubWebHookPayload) GetRepository() GithubRepository`

GetRepository returns the Repository field if non-nil, zero value otherwise.

### GetRepositoryOk

`func (o *GithubWebHookPayload) GetRepositoryOk() (*GithubRepository, bool)`

GetRepositoryOk returns a tuple with the Repository field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRepository

`func (o *GithubWebHookPayload) SetRepository(v GithubRepository)`

SetRepository sets Repository field to given value.

### HasRepository

`func (o *GithubWebHookPayload) HasRepository() bool`

HasRepository returns a boolean if a field has been set.

### GetSender

`func (o *GithubWebHookPayload) GetSender() GithubUser`

GetSender returns the Sender field if non-nil, zero value otherwise.

### GetSenderOk

`func (o *GithubWebHookPayload) GetSenderOk() (*GithubUser, bool)`

GetSenderOk returns a tuple with the Sender field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSender

`func (o *GithubWebHookPayload) SetSender(v GithubUser)`

SetSender sets Sender field to given value.

### HasSender

`func (o *GithubWebHookPayload) HasSender() bool`

HasSender returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


