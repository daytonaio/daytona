# GitCommitRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AllowEmpty** | Pointer to **bool** |  | [optional] 
**Author** | **string** |  | 
**Email** | **string** |  | 
**Message** | **string** |  | 
**Path** | **string** |  | 

## Methods

### NewGitCommitRequest

`func NewGitCommitRequest(author string, email string, message string, path string, ) *GitCommitRequest`

NewGitCommitRequest instantiates a new GitCommitRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGitCommitRequestWithDefaults

`func NewGitCommitRequestWithDefaults() *GitCommitRequest`

NewGitCommitRequestWithDefaults instantiates a new GitCommitRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAllowEmpty

`func (o *GitCommitRequest) GetAllowEmpty() bool`

GetAllowEmpty returns the AllowEmpty field if non-nil, zero value otherwise.

### GetAllowEmptyOk

`func (o *GitCommitRequest) GetAllowEmptyOk() (*bool, bool)`

GetAllowEmptyOk returns a tuple with the AllowEmpty field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowEmpty

`func (o *GitCommitRequest) SetAllowEmpty(v bool)`

SetAllowEmpty sets AllowEmpty field to given value.

### HasAllowEmpty

`func (o *GitCommitRequest) HasAllowEmpty() bool`

HasAllowEmpty returns a boolean if a field has been set.

### GetAuthor

`func (o *GitCommitRequest) GetAuthor() string`

GetAuthor returns the Author field if non-nil, zero value otherwise.

### GetAuthorOk

`func (o *GitCommitRequest) GetAuthorOk() (*string, bool)`

GetAuthorOk returns a tuple with the Author field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthor

`func (o *GitCommitRequest) SetAuthor(v string)`

SetAuthor sets Author field to given value.


### GetEmail

`func (o *GitCommitRequest) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *GitCommitRequest) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *GitCommitRequest) SetEmail(v string)`

SetEmail sets Email field to given value.


### GetMessage

`func (o *GitCommitRequest) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *GitCommitRequest) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *GitCommitRequest) SetMessage(v string)`

SetMessage sets Message field to given value.


### GetPath

`func (o *GitCommitRequest) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *GitCommitRequest) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *GitCommitRequest) SetPath(v string)`

SetPath sets Path field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


