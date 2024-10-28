# WorkspaceState

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**GitStatus** | [**GitStatus**](GitStatus.md) |  | 
**UpdatedAt** | **string** |  | 
**Uptime** | **int32** |  | 

## Methods

### NewWorkspaceState

`func NewWorkspaceState(gitStatus GitStatus, updatedAt string, uptime int32, ) *WorkspaceState`

NewWorkspaceState instantiates a new WorkspaceState object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspaceStateWithDefaults

`func NewWorkspaceStateWithDefaults() *WorkspaceState`

NewWorkspaceStateWithDefaults instantiates a new WorkspaceState object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGitStatus

`func (o *WorkspaceState) GetGitStatus() GitStatus`

GetGitStatus returns the GitStatus field if non-nil, zero value otherwise.

### GetGitStatusOk

`func (o *WorkspaceState) GetGitStatusOk() (*GitStatus, bool)`

GetGitStatusOk returns a tuple with the GitStatus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGitStatus

`func (o *WorkspaceState) SetGitStatus(v GitStatus)`

SetGitStatus sets GitStatus field to given value.


### GetUpdatedAt

`func (o *WorkspaceState) GetUpdatedAt() string`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *WorkspaceState) GetUpdatedAtOk() (*string, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *WorkspaceState) SetUpdatedAt(v string)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetUptime

`func (o *WorkspaceState) GetUptime() int32`

GetUptime returns the Uptime field if non-nil, zero value otherwise.

### GetUptimeOk

`func (o *WorkspaceState) GetUptimeOk() (*int32, bool)`

GetUptimeOk returns a tuple with the Uptime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUptime

`func (o *WorkspaceState) SetUptime(v int32)`

SetUptime sets Uptime field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


