# Target

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Default** | **bool** |  | 
**EnvVars** | **map[string]string** |  | 
**Id** | **string** |  | 
**LastJob** | Pointer to [**Job**](Job.md) |  | [optional] 
**LastJobId** | Pointer to **string** |  | [optional] 
**Metadata** | Pointer to [**TargetMetadata**](TargetMetadata.md) |  | [optional] 
**Name** | **string** |  | 
**ProviderMetadata** | Pointer to **string** |  | [optional] 
**TargetConfig** | [**TargetConfig**](TargetConfig.md) |  | 
**TargetConfigId** | **string** |  | 
**Workspaces** | [**[]Workspace**](Workspace.md) |  | 

## Methods

### NewTarget

`func NewTarget(default_ bool, envVars map[string]string, id string, name string, targetConfig TargetConfig, targetConfigId string, workspaces []Workspace, ) *Target`

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

### GetLastJobId

`func (o *Target) GetLastJobId() string`

GetLastJobId returns the LastJobId field if non-nil, zero value otherwise.

### GetLastJobIdOk

`func (o *Target) GetLastJobIdOk() (*string, bool)`

GetLastJobIdOk returns a tuple with the LastJobId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastJobId

`func (o *Target) SetLastJobId(v string)`

SetLastJobId sets LastJobId field to given value.

### HasLastJobId

`func (o *Target) HasLastJobId() bool`

HasLastJobId returns a boolean if a field has been set.

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


### GetProviderMetadata

`func (o *Target) GetProviderMetadata() string`

GetProviderMetadata returns the ProviderMetadata field if non-nil, zero value otherwise.

### GetProviderMetadataOk

`func (o *Target) GetProviderMetadataOk() (*string, bool)`

GetProviderMetadataOk returns a tuple with the ProviderMetadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProviderMetadata

`func (o *Target) SetProviderMetadata(v string)`

SetProviderMetadata sets ProviderMetadata field to given value.

### HasProviderMetadata

`func (o *Target) HasProviderMetadata() bool`

HasProviderMetadata returns a boolean if a field has been set.

### GetTargetConfig

`func (o *Target) GetTargetConfig() TargetConfig`

GetTargetConfig returns the TargetConfig field if non-nil, zero value otherwise.

### GetTargetConfigOk

`func (o *Target) GetTargetConfigOk() (*TargetConfig, bool)`

GetTargetConfigOk returns a tuple with the TargetConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTargetConfig

`func (o *Target) SetTargetConfig(v TargetConfig)`

SetTargetConfig sets TargetConfig field to given value.


### GetTargetConfigId

`func (o *Target) GetTargetConfigId() string`

GetTargetConfigId returns the TargetConfigId field if non-nil, zero value otherwise.

### GetTargetConfigIdOk

`func (o *Target) GetTargetConfigIdOk() (*string, bool)`

GetTargetConfigIdOk returns a tuple with the TargetConfigId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTargetConfigId

`func (o *Target) SetTargetConfigId(v string)`

SetTargetConfigId sets TargetConfigId field to given value.


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



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


