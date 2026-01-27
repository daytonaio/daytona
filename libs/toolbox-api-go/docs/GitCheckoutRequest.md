# GitCheckoutRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Branch** | **string** |  | 
**Path** | **string** |  | 

## Methods

### NewGitCheckoutRequest

`func NewGitCheckoutRequest(branch string, path string, ) *GitCheckoutRequest`

NewGitCheckoutRequest instantiates a new GitCheckoutRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGitCheckoutRequestWithDefaults

`func NewGitCheckoutRequestWithDefaults() *GitCheckoutRequest`

NewGitCheckoutRequestWithDefaults instantiates a new GitCheckoutRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBranch

`func (o *GitCheckoutRequest) GetBranch() string`

GetBranch returns the Branch field if non-nil, zero value otherwise.

### GetBranchOk

`func (o *GitCheckoutRequest) GetBranchOk() (*string, bool)`

GetBranchOk returns a tuple with the Branch field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBranch

`func (o *GitCheckoutRequest) SetBranch(v string)`

SetBranch sets Branch field to given value.


### GetPath

`func (o *GitCheckoutRequest) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *GitCheckoutRequest) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *GitCheckoutRequest) SetPath(v string)`

SetPath sets Path field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


