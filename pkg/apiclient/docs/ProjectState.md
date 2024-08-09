# ProjectState

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**GitStatus** | [**GitStatus**](GitStatus.md) |  | 
**UpdatedAt** | **string** |  | 
**Uptime** | **int32** |  | 

## Methods

### NewProjectState

`func NewProjectState(gitStatus GitStatus, updatedAt string, uptime int32, ) *ProjectState`

NewProjectState instantiates a new ProjectState object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectStateWithDefaults

`func NewProjectStateWithDefaults() *ProjectState`

NewProjectStateWithDefaults instantiates a new ProjectState object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGitStatus

`func (o *ProjectState) GetGitStatus() GitStatus`

GetGitStatus returns the GitStatus field if non-nil, zero value otherwise.

### GetGitStatusOk

`func (o *ProjectState) GetGitStatusOk() (*GitStatus, bool)`

GetGitStatusOk returns a tuple with the GitStatus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGitStatus

`func (o *ProjectState) SetGitStatus(v GitStatus)`

SetGitStatus sets GitStatus field to given value.


### GetUpdatedAt

`func (o *ProjectState) GetUpdatedAt() string`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *ProjectState) GetUpdatedAtOk() (*string, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *ProjectState) SetUpdatedAt(v string)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetUptime

`func (o *ProjectState) GetUptime() int32`

GetUptime returns the Uptime field if non-nil, zero value otherwise.

### GetUptimeOk

`func (o *ProjectState) GetUptimeOk() (*int32, bool)`

GetUptimeOk returns a tuple with the Uptime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUptime

`func (o *ProjectState) SetUptime(v int32)`

SetUptime sets Uptime field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


