# GitStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Ahead** | Pointer to **int32** |  | [optional] 
**Behind** | Pointer to **int32** |  | [optional] 
**BranchPublished** | Pointer to **bool** |  | [optional] 
**CurrentBranch** | **string** |  | 
**FileStatus** | [**[]FileStatus**](FileStatus.md) |  | 

## Methods

### NewGitStatus

`func NewGitStatus(currentBranch string, fileStatus []FileStatus, ) *GitStatus`

NewGitStatus instantiates a new GitStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGitStatusWithDefaults

`func NewGitStatusWithDefaults() *GitStatus`

NewGitStatusWithDefaults instantiates a new GitStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAhead

`func (o *GitStatus) GetAhead() int32`

GetAhead returns the Ahead field if non-nil, zero value otherwise.

### GetAheadOk

`func (o *GitStatus) GetAheadOk() (*int32, bool)`

GetAheadOk returns a tuple with the Ahead field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAhead

`func (o *GitStatus) SetAhead(v int32)`

SetAhead sets Ahead field to given value.

### HasAhead

`func (o *GitStatus) HasAhead() bool`

HasAhead returns a boolean if a field has been set.

### GetBehind

`func (o *GitStatus) GetBehind() int32`

GetBehind returns the Behind field if non-nil, zero value otherwise.

### GetBehindOk

`func (o *GitStatus) GetBehindOk() (*int32, bool)`

GetBehindOk returns a tuple with the Behind field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBehind

`func (o *GitStatus) SetBehind(v int32)`

SetBehind sets Behind field to given value.

### HasBehind

`func (o *GitStatus) HasBehind() bool`

HasBehind returns a boolean if a field has been set.

### GetBranchPublished

`func (o *GitStatus) GetBranchPublished() bool`

GetBranchPublished returns the BranchPublished field if non-nil, zero value otherwise.

### GetBranchPublishedOk

`func (o *GitStatus) GetBranchPublishedOk() (*bool, bool)`

GetBranchPublishedOk returns a tuple with the BranchPublished field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBranchPublished

`func (o *GitStatus) SetBranchPublished(v bool)`

SetBranchPublished sets BranchPublished field to given value.

### HasBranchPublished

`func (o *GitStatus) HasBranchPublished() bool`

HasBranchPublished returns a boolean if a field has been set.

### GetCurrentBranch

`func (o *GitStatus) GetCurrentBranch() string`

GetCurrentBranch returns the CurrentBranch field if non-nil, zero value otherwise.

### GetCurrentBranchOk

`func (o *GitStatus) GetCurrentBranchOk() (*string, bool)`

GetCurrentBranchOk returns a tuple with the CurrentBranch field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrentBranch

`func (o *GitStatus) SetCurrentBranch(v string)`

SetCurrentBranch sets CurrentBranch field to given value.


### GetFileStatus

`func (o *GitStatus) GetFileStatus() []FileStatus`

GetFileStatus returns the FileStatus field if non-nil, zero value otherwise.

### GetFileStatusOk

`func (o *GitStatus) GetFileStatusOk() (*[]FileStatus, bool)`

GetFileStatusOk returns a tuple with the FileStatus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFileStatus

`func (o *GitStatus) SetFileStatus(v []FileStatus)`

SetFileStatus sets FileStatus field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


