# SetWorkspaceMetadata

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**GitStatus** | Pointer to [**GitStatus**](GitStatus.md) |  | [optional] 
**Uptime** | **int32** |  | 

## Methods

### NewSetWorkspaceMetadata

`func NewSetWorkspaceMetadata(uptime int32, ) *SetWorkspaceMetadata`

NewSetWorkspaceMetadata instantiates a new SetWorkspaceMetadata object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSetWorkspaceMetadataWithDefaults

`func NewSetWorkspaceMetadataWithDefaults() *SetWorkspaceMetadata`

NewSetWorkspaceMetadataWithDefaults instantiates a new SetWorkspaceMetadata object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGitStatus

`func (o *SetWorkspaceMetadata) GetGitStatus() GitStatus`

GetGitStatus returns the GitStatus field if non-nil, zero value otherwise.

### GetGitStatusOk

`func (o *SetWorkspaceMetadata) GetGitStatusOk() (*GitStatus, bool)`

GetGitStatusOk returns a tuple with the GitStatus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGitStatus

`func (o *SetWorkspaceMetadata) SetGitStatus(v GitStatus)`

SetGitStatus sets GitStatus field to given value.

### HasGitStatus

`func (o *SetWorkspaceMetadata) HasGitStatus() bool`

HasGitStatus returns a boolean if a field has been set.

### GetUptime

`func (o *SetWorkspaceMetadata) GetUptime() int32`

GetUptime returns the Uptime field if non-nil, zero value otherwise.

### GetUptimeOk

`func (o *SetWorkspaceMetadata) GetUptimeOk() (*int32, bool)`

GetUptimeOk returns a tuple with the Uptime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUptime

`func (o *SetWorkspaceMetadata) SetUptime(v int32)`

SetUptime sets Uptime field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


