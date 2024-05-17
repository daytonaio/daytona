# GitBranch

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** |  | [optional] 
**Sha** | Pointer to **string** |  | [optional] 

## Methods

### NewGitBranch

`func NewGitBranch() *GitBranch`

NewGitBranch instantiates a new GitBranch object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGitBranchWithDefaults

`func NewGitBranchWithDefaults() *GitBranch`

NewGitBranchWithDefaults instantiates a new GitBranch object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *GitBranch) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *GitBranch) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *GitBranch) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *GitBranch) HasName() bool`

HasName returns a boolean if a field has been set.

### GetSha

`func (o *GitBranch) GetSha() string`

GetSha returns the Sha field if non-nil, zero value otherwise.

### GetShaOk

`func (o *GitBranch) GetShaOk() (*string, bool)`

GetShaOk returns a tuple with the Sha field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSha

`func (o *GitBranch) SetSha(v string)`

SetSha sets Sha field to given value.

### HasSha

`func (o *GitBranch) HasSha() bool`

HasSha returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


