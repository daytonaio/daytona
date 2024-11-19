# WorkspaceMetadata

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**GitStatus** | [**GitStatus**](GitStatus.md) |  | 
**UpdatedAt** | **string** |  | 
**Uptime** | **int32** |  | 
**WorkspaceId** | **string** |  | 

## Methods

### NewWorkspaceMetadata

`func NewWorkspaceMetadata(gitStatus GitStatus, updatedAt string, uptime int32, workspaceId string, ) *WorkspaceMetadata`

NewWorkspaceMetadata instantiates a new WorkspaceMetadata object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspaceMetadataWithDefaults

`func NewWorkspaceMetadataWithDefaults() *WorkspaceMetadata`

NewWorkspaceMetadataWithDefaults instantiates a new WorkspaceMetadata object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGitStatus

`func (o *WorkspaceMetadata) GetGitStatus() GitStatus`

GetGitStatus returns the GitStatus field if non-nil, zero value otherwise.

### GetGitStatusOk

`func (o *WorkspaceMetadata) GetGitStatusOk() (*GitStatus, bool)`

GetGitStatusOk returns a tuple with the GitStatus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGitStatus

`func (o *WorkspaceMetadata) SetGitStatus(v GitStatus)`

SetGitStatus sets GitStatus field to given value.


### GetUpdatedAt

`func (o *WorkspaceMetadata) GetUpdatedAt() string`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *WorkspaceMetadata) GetUpdatedAtOk() (*string, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *WorkspaceMetadata) SetUpdatedAt(v string)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetUptime

`func (o *WorkspaceMetadata) GetUptime() int32`

GetUptime returns the Uptime field if non-nil, zero value otherwise.

### GetUptimeOk

`func (o *WorkspaceMetadata) GetUptimeOk() (*int32, bool)`

GetUptimeOk returns a tuple with the Uptime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUptime

`func (o *WorkspaceMetadata) SetUptime(v int32)`

SetUptime sets Uptime field to given value.


### GetWorkspaceId

`func (o *WorkspaceMetadata) GetWorkspaceId() string`

GetWorkspaceId returns the WorkspaceId field if non-nil, zero value otherwise.

### GetWorkspaceIdOk

`func (o *WorkspaceMetadata) GetWorkspaceIdOk() (*string, bool)`

GetWorkspaceIdOk returns a tuple with the WorkspaceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkspaceId

`func (o *WorkspaceMetadata) SetWorkspaceId(v string)`

SetWorkspaceId sets WorkspaceId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


