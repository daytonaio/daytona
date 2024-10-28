# WorkspaceInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Created** | **string** |  | 
**IsRunning** | **bool** |  | 
**Name** | **string** |  | 
**ProviderMetadata** | Pointer to **string** |  | [optional] 
**TargetId** | **string** |  | 

## Methods

### NewWorkspaceInfo

`func NewWorkspaceInfo(created string, isRunning bool, name string, targetId string, ) *WorkspaceInfo`

NewWorkspaceInfo instantiates a new WorkspaceInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspaceInfoWithDefaults

`func NewWorkspaceInfoWithDefaults() *WorkspaceInfo`

NewWorkspaceInfoWithDefaults instantiates a new WorkspaceInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCreated

`func (o *WorkspaceInfo) GetCreated() string`

GetCreated returns the Created field if non-nil, zero value otherwise.

### GetCreatedOk

`func (o *WorkspaceInfo) GetCreatedOk() (*string, bool)`

GetCreatedOk returns a tuple with the Created field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreated

`func (o *WorkspaceInfo) SetCreated(v string)`

SetCreated sets Created field to given value.


### GetIsRunning

`func (o *WorkspaceInfo) GetIsRunning() bool`

GetIsRunning returns the IsRunning field if non-nil, zero value otherwise.

### GetIsRunningOk

`func (o *WorkspaceInfo) GetIsRunningOk() (*bool, bool)`

GetIsRunningOk returns a tuple with the IsRunning field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsRunning

`func (o *WorkspaceInfo) SetIsRunning(v bool)`

SetIsRunning sets IsRunning field to given value.


### GetName

`func (o *WorkspaceInfo) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkspaceInfo) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkspaceInfo) SetName(v string)`

SetName sets Name field to given value.


### GetProviderMetadata

`func (o *WorkspaceInfo) GetProviderMetadata() string`

GetProviderMetadata returns the ProviderMetadata field if non-nil, zero value otherwise.

### GetProviderMetadataOk

`func (o *WorkspaceInfo) GetProviderMetadataOk() (*string, bool)`

GetProviderMetadataOk returns a tuple with the ProviderMetadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProviderMetadata

`func (o *WorkspaceInfo) SetProviderMetadata(v string)`

SetProviderMetadata sets ProviderMetadata field to given value.

### HasProviderMetadata

`func (o *WorkspaceInfo) HasProviderMetadata() bool`

HasProviderMetadata returns a boolean if a field has been set.

### GetTargetId

`func (o *WorkspaceInfo) GetTargetId() string`

GetTargetId returns the TargetId field if non-nil, zero value otherwise.

### GetTargetIdOk

`func (o *WorkspaceInfo) GetTargetIdOk() (*string, bool)`

GetTargetIdOk returns a tuple with the TargetId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTargetId

`func (o *WorkspaceInfo) SetTargetId(v string)`

SetTargetId sets TargetId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


