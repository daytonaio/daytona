# TargetDTO

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Default** | **bool** |  | 
**EnvVars** | **map[string]string** |  | 
**Id** | **string** |  | 
**LastJob** | Pointer to [**Job**](Job.md) |  | [optional] 
**Metadata** | Pointer to [**TargetMetadata**](TargetMetadata.md) |  | [optional] 
**Name** | **string** |  | 
**ProviderMetadata** | Pointer to **string** |  | [optional] 
**State** | [**ResourceState**](ResourceState.md) |  | 
**TargetConfig** | [**TargetConfig**](TargetConfig.md) |  | 
**TargetConfigId** | **string** |  | 
**Workspaces** | [**[]Workspace**](Workspace.md) |  | 

## Methods

### NewTargetDTO

`func NewTargetDTO(default_ bool, envVars map[string]string, id string, name string, state ResourceState, targetConfig TargetConfig, targetConfigId string, workspaces []Workspace, ) *TargetDTO`

NewTargetDTO instantiates a new TargetDTO object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTargetDTOWithDefaults

`func NewTargetDTOWithDefaults() *TargetDTO`

NewTargetDTOWithDefaults instantiates a new TargetDTO object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDefault

`func (o *TargetDTO) GetDefault() bool`

GetDefault returns the Default field if non-nil, zero value otherwise.

### GetDefaultOk

`func (o *TargetDTO) GetDefaultOk() (*bool, bool)`

GetDefaultOk returns a tuple with the Default field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefault

`func (o *TargetDTO) SetDefault(v bool)`

SetDefault sets Default field to given value.


### GetEnvVars

`func (o *TargetDTO) GetEnvVars() map[string]string`

GetEnvVars returns the EnvVars field if non-nil, zero value otherwise.

### GetEnvVarsOk

`func (o *TargetDTO) GetEnvVarsOk() (*map[string]string, bool)`

GetEnvVarsOk returns a tuple with the EnvVars field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvVars

`func (o *TargetDTO) SetEnvVars(v map[string]string)`

SetEnvVars sets EnvVars field to given value.


### GetId

`func (o *TargetDTO) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *TargetDTO) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *TargetDTO) SetId(v string)`

SetId sets Id field to given value.


### GetLastJob

`func (o *TargetDTO) GetLastJob() Job`

GetLastJob returns the LastJob field if non-nil, zero value otherwise.

### GetLastJobOk

`func (o *TargetDTO) GetLastJobOk() (*Job, bool)`

GetLastJobOk returns a tuple with the LastJob field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastJob

`func (o *TargetDTO) SetLastJob(v Job)`

SetLastJob sets LastJob field to given value.

### HasLastJob

`func (o *TargetDTO) HasLastJob() bool`

HasLastJob returns a boolean if a field has been set.

### GetMetadata

`func (o *TargetDTO) GetMetadata() TargetMetadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *TargetDTO) GetMetadataOk() (*TargetMetadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *TargetDTO) SetMetadata(v TargetMetadata)`

SetMetadata sets Metadata field to given value.

### HasMetadata

`func (o *TargetDTO) HasMetadata() bool`

HasMetadata returns a boolean if a field has been set.

### GetName

`func (o *TargetDTO) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *TargetDTO) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *TargetDTO) SetName(v string)`

SetName sets Name field to given value.


### GetProviderMetadata

`func (o *TargetDTO) GetProviderMetadata() string`

GetProviderMetadata returns the ProviderMetadata field if non-nil, zero value otherwise.

### GetProviderMetadataOk

`func (o *TargetDTO) GetProviderMetadataOk() (*string, bool)`

GetProviderMetadataOk returns a tuple with the ProviderMetadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProviderMetadata

`func (o *TargetDTO) SetProviderMetadata(v string)`

SetProviderMetadata sets ProviderMetadata field to given value.

### HasProviderMetadata

`func (o *TargetDTO) HasProviderMetadata() bool`

HasProviderMetadata returns a boolean if a field has been set.

### GetState

`func (o *TargetDTO) GetState() ResourceState`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *TargetDTO) GetStateOk() (*ResourceState, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *TargetDTO) SetState(v ResourceState)`

SetState sets State field to given value.


### GetTargetConfig

`func (o *TargetDTO) GetTargetConfig() TargetConfig`

GetTargetConfig returns the TargetConfig field if non-nil, zero value otherwise.

### GetTargetConfigOk

`func (o *TargetDTO) GetTargetConfigOk() (*TargetConfig, bool)`

GetTargetConfigOk returns a tuple with the TargetConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTargetConfig

`func (o *TargetDTO) SetTargetConfig(v TargetConfig)`

SetTargetConfig sets TargetConfig field to given value.


### GetTargetConfigId

`func (o *TargetDTO) GetTargetConfigId() string`

GetTargetConfigId returns the TargetConfigId field if non-nil, zero value otherwise.

### GetTargetConfigIdOk

`func (o *TargetDTO) GetTargetConfigIdOk() (*string, bool)`

GetTargetConfigIdOk returns a tuple with the TargetConfigId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTargetConfigId

`func (o *TargetDTO) SetTargetConfigId(v string)`

SetTargetConfigId sets TargetConfigId field to given value.


### GetWorkspaces

`func (o *TargetDTO) GetWorkspaces() []Workspace`

GetWorkspaces returns the Workspaces field if non-nil, zero value otherwise.

### GetWorkspacesOk

`func (o *TargetDTO) GetWorkspacesOk() (*[]Workspace, bool)`

GetWorkspacesOk returns a tuple with the Workspaces field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkspaces

`func (o *TargetDTO) SetWorkspaces(v []Workspace)`

SetWorkspaces sets Workspaces field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


