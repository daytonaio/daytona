# ProviderInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AgentlessTarget** | Pointer to **bool** |  | [optional] 
**Label** | Pointer to **string** |  | [optional] 
**Name** | **string** |  | 
**RunnerId** | **string** |  | 
**RunnerName** | **string** |  | 
**TargetConfigManifest** | [**map[string]TargetConfigProperty**](TargetConfigProperty.md) |  | 
**Version** | **string** |  | 

## Methods

### NewProviderInfo

`func NewProviderInfo(name string, runnerId string, runnerName string, targetConfigManifest map[string]TargetConfigProperty, version string, ) *ProviderInfo`

NewProviderInfo instantiates a new ProviderInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProviderInfoWithDefaults

`func NewProviderInfoWithDefaults() *ProviderInfo`

NewProviderInfoWithDefaults instantiates a new ProviderInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAgentlessTarget

`func (o *ProviderInfo) GetAgentlessTarget() bool`

GetAgentlessTarget returns the AgentlessTarget field if non-nil, zero value otherwise.

### GetAgentlessTargetOk

`func (o *ProviderInfo) GetAgentlessTargetOk() (*bool, bool)`

GetAgentlessTargetOk returns a tuple with the AgentlessTarget field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAgentlessTarget

`func (o *ProviderInfo) SetAgentlessTarget(v bool)`

SetAgentlessTarget sets AgentlessTarget field to given value.

### HasAgentlessTarget

`func (o *ProviderInfo) HasAgentlessTarget() bool`

HasAgentlessTarget returns a boolean if a field has been set.

### GetLabel

`func (o *ProviderInfo) GetLabel() string`

GetLabel returns the Label field if non-nil, zero value otherwise.

### GetLabelOk

`func (o *ProviderInfo) GetLabelOk() (*string, bool)`

GetLabelOk returns a tuple with the Label field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabel

`func (o *ProviderInfo) SetLabel(v string)`

SetLabel sets Label field to given value.

### HasLabel

`func (o *ProviderInfo) HasLabel() bool`

HasLabel returns a boolean if a field has been set.

### GetName

`func (o *ProviderInfo) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ProviderInfo) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ProviderInfo) SetName(v string)`

SetName sets Name field to given value.


### GetRunnerId

`func (o *ProviderInfo) GetRunnerId() string`

GetRunnerId returns the RunnerId field if non-nil, zero value otherwise.

### GetRunnerIdOk

`func (o *ProviderInfo) GetRunnerIdOk() (*string, bool)`

GetRunnerIdOk returns a tuple with the RunnerId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunnerId

`func (o *ProviderInfo) SetRunnerId(v string)`

SetRunnerId sets RunnerId field to given value.


### GetRunnerName

`func (o *ProviderInfo) GetRunnerName() string`

GetRunnerName returns the RunnerName field if non-nil, zero value otherwise.

### GetRunnerNameOk

`func (o *ProviderInfo) GetRunnerNameOk() (*string, bool)`

GetRunnerNameOk returns a tuple with the RunnerName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunnerName

`func (o *ProviderInfo) SetRunnerName(v string)`

SetRunnerName sets RunnerName field to given value.


### GetTargetConfigManifest

`func (o *ProviderInfo) GetTargetConfigManifest() map[string]TargetConfigProperty`

GetTargetConfigManifest returns the TargetConfigManifest field if non-nil, zero value otherwise.

### GetTargetConfigManifestOk

`func (o *ProviderInfo) GetTargetConfigManifestOk() (*map[string]TargetConfigProperty, bool)`

GetTargetConfigManifestOk returns a tuple with the TargetConfigManifest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTargetConfigManifest

`func (o *ProviderInfo) SetTargetConfigManifest(v map[string]TargetConfigProperty)`

SetTargetConfigManifest sets TargetConfigManifest field to given value.


### GetVersion

`func (o *ProviderInfo) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *ProviderInfo) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *ProviderInfo) SetVersion(v string)`

SetVersion sets Version field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


