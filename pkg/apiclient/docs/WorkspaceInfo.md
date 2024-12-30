# WorkspaceInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**Projects** | [**[]ProjectInfo**](ProjectInfo.md) |  | 
**ProviderMetadata** | Pointer to **string** |  | [optional] 

## Methods

### NewWorkspaceInfo

`func NewWorkspaceInfo(name string, projects []ProjectInfo, ) *WorkspaceInfo`

NewWorkspaceInfo instantiates a new WorkspaceInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspaceInfoWithDefaults

`func NewWorkspaceInfoWithDefaults() *WorkspaceInfo`

NewWorkspaceInfoWithDefaults instantiates a new WorkspaceInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

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


### GetProjects

`func (o *WorkspaceInfo) GetProjects() []ProjectInfo`

GetProjects returns the Projects field if non-nil, zero value otherwise.

### GetProjectsOk

`func (o *WorkspaceInfo) GetProjectsOk() (*[]ProjectInfo, bool)`

GetProjectsOk returns a tuple with the Projects field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjects

`func (o *WorkspaceInfo) SetProjects(v []ProjectInfo)`

SetProjects sets Projects field to given value.


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


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


