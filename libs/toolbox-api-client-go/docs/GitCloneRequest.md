# GitCloneRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Branch** | Pointer to **string** |  | [optional] 
**CommitId** | Pointer to **string** |  | [optional] 
**Password** | Pointer to **string** |  | [optional] 
**Path** | **string** |  | 
**Url** | **string** |  | 
**Username** | Pointer to **string** |  | [optional] 

## Methods

### NewGitCloneRequest

`func NewGitCloneRequest(path string, url string, ) *GitCloneRequest`

NewGitCloneRequest instantiates a new GitCloneRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGitCloneRequestWithDefaults

`func NewGitCloneRequestWithDefaults() *GitCloneRequest`

NewGitCloneRequestWithDefaults instantiates a new GitCloneRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBranch

`func (o *GitCloneRequest) GetBranch() string`

GetBranch returns the Branch field if non-nil, zero value otherwise.

### GetBranchOk

`func (o *GitCloneRequest) GetBranchOk() (*string, bool)`

GetBranchOk returns a tuple with the Branch field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBranch

`func (o *GitCloneRequest) SetBranch(v string)`

SetBranch sets Branch field to given value.

### HasBranch

`func (o *GitCloneRequest) HasBranch() bool`

HasBranch returns a boolean if a field has been set.

### GetCommitId

`func (o *GitCloneRequest) GetCommitId() string`

GetCommitId returns the CommitId field if non-nil, zero value otherwise.

### GetCommitIdOk

`func (o *GitCloneRequest) GetCommitIdOk() (*string, bool)`

GetCommitIdOk returns a tuple with the CommitId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommitId

`func (o *GitCloneRequest) SetCommitId(v string)`

SetCommitId sets CommitId field to given value.

### HasCommitId

`func (o *GitCloneRequest) HasCommitId() bool`

HasCommitId returns a boolean if a field has been set.

### GetPassword

`func (o *GitCloneRequest) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *GitCloneRequest) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *GitCloneRequest) SetPassword(v string)`

SetPassword sets Password field to given value.

### HasPassword

`func (o *GitCloneRequest) HasPassword() bool`

HasPassword returns a boolean if a field has been set.

### GetPath

`func (o *GitCloneRequest) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *GitCloneRequest) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *GitCloneRequest) SetPath(v string)`

SetPath sets Path field to given value.


### GetUrl

`func (o *GitCloneRequest) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *GitCloneRequest) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *GitCloneRequest) SetUrl(v string)`

SetUrl sets Url field to given value.


### GetUsername

`func (o *GitCloneRequest) GetUsername() string`

GetUsername returns the Username field if non-nil, zero value otherwise.

### GetUsernameOk

`func (o *GitCloneRequest) GetUsernameOk() (*string, bool)`

GetUsernameOk returns a tuple with the Username field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsername

`func (o *GitCloneRequest) SetUsername(v string)`

SetUsername sets Username field to given value.

### HasUsername

`func (o *GitCloneRequest) HasUsername() bool`

HasUsername returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


