# GitPushRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Password** | Pointer to **string** |  | [optional] 
**Path** | **string** |  | 
**Username** | Pointer to **string** |  | [optional] 

## Methods

### NewGitPushRequest

`func NewGitPushRequest(path string, ) *GitPushRequest`

NewGitPushRequest instantiates a new GitPushRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGitPushRequestWithDefaults

`func NewGitPushRequestWithDefaults() *GitPushRequest`

NewGitPushRequestWithDefaults instantiates a new GitPushRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPassword

`func (o *GitPushRequest) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *GitPushRequest) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *GitPushRequest) SetPassword(v string)`

SetPassword sets Password field to given value.

### HasPassword

`func (o *GitPushRequest) HasPassword() bool`

HasPassword returns a boolean if a field has been set.

### GetPath

`func (o *GitPushRequest) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *GitPushRequest) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *GitPushRequest) SetPath(v string)`

SetPath sets Path field to given value.


### GetUsername

`func (o *GitPushRequest) GetUsername() string`

GetUsername returns the Username field if non-nil, zero value otherwise.

### GetUsernameOk

`func (o *GitPushRequest) GetUsernameOk() (*string, bool)`

GetUsernameOk returns a tuple with the Username field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsername

`func (o *GitPushRequest) SetUsername(v string)`

SetUsername sets Username field to given value.

### HasUsername

`func (o *GitPushRequest) HasUsername() bool`

HasUsername returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


