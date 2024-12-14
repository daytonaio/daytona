# GitAddRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Files** | **[]string** | files to add (use . for all files) | 
**Path** | **string** |  | 

## Methods

### NewGitAddRequest

`func NewGitAddRequest(files []string, path string, ) *GitAddRequest`

NewGitAddRequest instantiates a new GitAddRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGitAddRequestWithDefaults

`func NewGitAddRequestWithDefaults() *GitAddRequest`

NewGitAddRequestWithDefaults instantiates a new GitAddRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFiles

`func (o *GitAddRequest) GetFiles() []string`

GetFiles returns the Files field if non-nil, zero value otherwise.

### GetFilesOk

`func (o *GitAddRequest) GetFilesOk() (*[]string, bool)`

GetFilesOk returns a tuple with the Files field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFiles

`func (o *GitAddRequest) SetFiles(v []string)`

SetFiles sets Files field to given value.


### GetPath

`func (o *GitAddRequest) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *GitAddRequest) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *GitAddRequest) SetPath(v string)`

SetPath sets Path field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


