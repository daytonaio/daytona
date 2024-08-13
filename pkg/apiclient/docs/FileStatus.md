# FileStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Extra** | **string** |  | 
**Name** | **string** |  | 
**Staging** | [**Status**](Status.md) |  | 
**Worktree** | [**Status**](Status.md) |  | 

## Methods

### NewFileStatus

`func NewFileStatus(extra string, name string, staging Status, worktree Status, ) *FileStatus`

NewFileStatus instantiates a new FileStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFileStatusWithDefaults

`func NewFileStatusWithDefaults() *FileStatus`

NewFileStatusWithDefaults instantiates a new FileStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetExtra

`func (o *FileStatus) GetExtra() string`

GetExtra returns the Extra field if non-nil, zero value otherwise.

### GetExtraOk

`func (o *FileStatus) GetExtraOk() (*string, bool)`

GetExtraOk returns a tuple with the Extra field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExtra

`func (o *FileStatus) SetExtra(v string)`

SetExtra sets Extra field to given value.


### GetName

`func (o *FileStatus) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *FileStatus) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *FileStatus) SetName(v string)`

SetName sets Name field to given value.


### GetStaging

`func (o *FileStatus) GetStaging() Status`

GetStaging returns the Staging field if non-nil, zero value otherwise.

### GetStagingOk

`func (o *FileStatus) GetStagingOk() (*Status, bool)`

GetStagingOk returns a tuple with the Staging field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStaging

`func (o *FileStatus) SetStaging(v Status)`

SetStaging sets Staging field to given value.


### GetWorktree

`func (o *FileStatus) GetWorktree() Status`

GetWorktree returns the Worktree field if non-nil, zero value otherwise.

### GetWorktreeOk

`func (o *FileStatus) GetWorktreeOk() (*Status, bool)`

GetWorktreeOk returns a tuple with the Worktree field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorktree

`func (o *FileStatus) SetWorktree(v Status)`

SetWorktree sets Worktree field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


