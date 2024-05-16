# GitStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CurrentBranch** | Pointer to **string** |  | [optional] 
**FileStatus** | Pointer to [**[]FileStatus**](FileStatus.md) |  | [optional] 

## Methods

### NewGitStatus

`func NewGitStatus() *GitStatus`

NewGitStatus instantiates a new GitStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGitStatusWithDefaults

`func NewGitStatusWithDefaults() *GitStatus`

NewGitStatusWithDefaults instantiates a new GitStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

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

### HasCurrentBranch

`func (o *GitStatus) HasCurrentBranch() bool`

HasCurrentBranch returns a boolean if a field has been set.

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

### HasFileStatus

`func (o *GitStatus) HasFileStatus() bool`

HasFileStatus returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


