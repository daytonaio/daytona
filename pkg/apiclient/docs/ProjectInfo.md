# ProjectInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Created** | **string** |  | 
**IsRunning** | **bool** |  | 
**Name** | **string** |  | 
**ProviderMetadata** | Pointer to **string** |  | [optional] 
**WorkspaceId** | **string** |  | 

## Methods

### NewProjectInfo

`func NewProjectInfo(created string, isRunning bool, name string, workspaceId string, ) *ProjectInfo`

NewProjectInfo instantiates a new ProjectInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectInfoWithDefaults

`func NewProjectInfoWithDefaults() *ProjectInfo`

NewProjectInfoWithDefaults instantiates a new ProjectInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCreated

`func (o *ProjectInfo) GetCreated() string`

GetCreated returns the Created field if non-nil, zero value otherwise.

### GetCreatedOk

`func (o *ProjectInfo) GetCreatedOk() (*string, bool)`

GetCreatedOk returns a tuple with the Created field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreated

`func (o *ProjectInfo) SetCreated(v string)`

SetCreated sets Created field to given value.


### GetIsRunning

`func (o *ProjectInfo) GetIsRunning() bool`

GetIsRunning returns the IsRunning field if non-nil, zero value otherwise.

### GetIsRunningOk

`func (o *ProjectInfo) GetIsRunningOk() (*bool, bool)`

GetIsRunningOk returns a tuple with the IsRunning field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsRunning

`func (o *ProjectInfo) SetIsRunning(v bool)`

SetIsRunning sets IsRunning field to given value.


### GetName

`func (o *ProjectInfo) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ProjectInfo) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ProjectInfo) SetName(v string)`

SetName sets Name field to given value.


### GetProviderMetadata

`func (o *ProjectInfo) GetProviderMetadata() string`

GetProviderMetadata returns the ProviderMetadata field if non-nil, zero value otherwise.

### GetProviderMetadataOk

`func (o *ProjectInfo) GetProviderMetadataOk() (*string, bool)`

GetProviderMetadataOk returns a tuple with the ProviderMetadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProviderMetadata

`func (o *ProjectInfo) SetProviderMetadata(v string)`

SetProviderMetadata sets ProviderMetadata field to given value.

### HasProviderMetadata

`func (o *ProjectInfo) HasProviderMetadata() bool`

HasProviderMetadata returns a boolean if a field has been set.

### GetWorkspaceId

`func (o *ProjectInfo) GetWorkspaceId() string`

GetWorkspaceId returns the WorkspaceId field if non-nil, zero value otherwise.

### GetWorkspaceIdOk

`func (o *ProjectInfo) GetWorkspaceIdOk() (*string, bool)`

GetWorkspaceIdOk returns a tuple with the WorkspaceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkspaceId

`func (o *ProjectInfo) SetWorkspaceId(v string)`

SetWorkspaceId sets WorkspaceId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


