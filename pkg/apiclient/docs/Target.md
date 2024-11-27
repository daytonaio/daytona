# Target

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Default** | **bool** |  | 
**EnvVars** | **map[string]string** |  | 
**Id** | **string** |  | 
**LastJob** | Pointer to [**Job**](Job.md) |  | [optional] 
**Metadata** | Pointer to [**TargetMetadata**](TargetMetadata.md) |  | [optional] 
**Name** | **string** |  | 
**Options** | **string** | JSON encoded map of options | 
**ProviderInfo** | [**TargetProviderInfo**](TargetProviderInfo.md) |  | 
**Workspaces** | Pointer to [**[]Workspace**](Workspace.md) |  | [optional] 

## Methods

### NewTarget

`func NewTarget(default_ bool, envVars map[string]string, id string, name string, options string, providerInfo TargetProviderInfo, ) *Target`

NewTarget instantiates a new Target object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTargetWithDefaults

`func NewTargetWithDefaults() *Target`

NewTargetWithDefaults instantiates a new Target object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDefault

`func (o *Target) GetDefault() bool`

GetDefault returns the Default field if non-nil, zero value otherwise.

### GetDefaultOk

`func (o *Target) GetDefaultOk() (*bool, bool)`

GetDefaultOk returns a tuple with the Default field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefault

`func (o *Target) SetDefault(v bool)`

SetDefault sets Default field to given value.


### GetEnvVars

`func (o *Target) GetEnvVars() map[string]string`

GetEnvVars returns the EnvVars field if non-nil, zero value otherwise.

### GetEnvVarsOk

`func (o *Target) GetEnvVarsOk() (*map[string]string, bool)`

GetEnvVarsOk returns a tuple with the EnvVars field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvVars

`func (o *Target) SetEnvVars(v map[string]string)`

SetEnvVars sets EnvVars field to given value.


### GetId

`func (o *Target) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Target) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Target) SetId(v string)`

SetId sets Id field to given value.


### GetLastJob

`func (o *Target) GetLastJob() Job`

GetLastJob returns the LastJob field if non-nil, zero value otherwise.

### GetLastJobOk

`func (o *Target) GetLastJobOk() (*Job, bool)`

GetLastJobOk returns a tuple with the LastJob field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastJob

`func (o *Target) SetLastJob(v Job)`

SetLastJob sets LastJob field to given value.

### HasLastJob

`func (o *Target) HasLastJob() bool`

HasLastJob returns a boolean if a field has been set.

### GetMetadata

`func (o *Target) GetMetadata() TargetMetadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *Target) GetMetadataOk() (*TargetMetadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *Target) SetMetadata(v TargetMetadata)`

SetMetadata sets Metadata field to given value.

### HasMetadata

`func (o *Target) HasMetadata() bool`

HasMetadata returns a boolean if a field has been set.

### GetName

`func (o *Target) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Target) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Target) SetName(v string)`

SetName sets Name field to given value.


### GetOptions

`func (o *Target) GetOptions() string`

GetOptions returns the Options field if non-nil, zero value otherwise.

### GetOptionsOk

`func (o *Target) GetOptionsOk() (*string, bool)`

GetOptionsOk returns a tuple with the Options field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOptions

`func (o *Target) SetOptions(v string)`

SetOptions sets Options field to given value.


### GetProviderInfo

`func (o *Target) GetProviderInfo() TargetProviderInfo`

GetProviderInfo returns the ProviderInfo field if non-nil, zero value otherwise.

### GetProviderInfoOk

`func (o *Target) GetProviderInfoOk() (*TargetProviderInfo, bool)`

GetProviderInfoOk returns a tuple with the ProviderInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProviderInfo

`func (o *Target) SetProviderInfo(v TargetProviderInfo)`

SetProviderInfo sets ProviderInfo field to given value.


### GetWorkspaces

`func (o *Target) GetWorkspaces() []Workspace`

GetWorkspaces returns the Workspaces field if non-nil, zero value otherwise.

### GetWorkspacesOk

`func (o *Target) GetWorkspacesOk() (*[]Workspace, bool)`

GetWorkspacesOk returns a tuple with the Workspaces field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkspaces

`func (o *Target) SetWorkspaces(v []Workspace)`

SetWorkspaces sets Workspaces field to given value.

### HasWorkspaces

`func (o *Target) HasWorkspaces() bool`

HasWorkspaces returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


