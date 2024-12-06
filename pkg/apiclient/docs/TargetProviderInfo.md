# TargetProviderInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AgentlessTarget** | Pointer to **bool** |  | [optional] 
**Label** | Pointer to **string** |  | [optional] 
**Name** | **string** |  | 
**Version** | **string** |  | 

## Methods

### NewTargetProviderInfo

`func NewTargetProviderInfo(name string, version string, ) *TargetProviderInfo`

NewTargetProviderInfo instantiates a new TargetProviderInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTargetProviderInfoWithDefaults

`func NewTargetProviderInfoWithDefaults() *TargetProviderInfo`

NewTargetProviderInfoWithDefaults instantiates a new TargetProviderInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAgentlessTarget

`func (o *TargetProviderInfo) GetAgentlessTarget() bool`

GetAgentlessTarget returns the AgentlessTarget field if non-nil, zero value otherwise.

### GetAgentlessTargetOk

`func (o *TargetProviderInfo) GetAgentlessTargetOk() (*bool, bool)`

GetAgentlessTargetOk returns a tuple with the AgentlessTarget field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAgentlessTarget

`func (o *TargetProviderInfo) SetAgentlessTarget(v bool)`

SetAgentlessTarget sets AgentlessTarget field to given value.

### HasAgentlessTarget

`func (o *TargetProviderInfo) HasAgentlessTarget() bool`

HasAgentlessTarget returns a boolean if a field has been set.

### GetLabel

`func (o *TargetProviderInfo) GetLabel() string`

GetLabel returns the Label field if non-nil, zero value otherwise.

### GetLabelOk

`func (o *TargetProviderInfo) GetLabelOk() (*string, bool)`

GetLabelOk returns a tuple with the Label field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabel

`func (o *TargetProviderInfo) SetLabel(v string)`

SetLabel sets Label field to given value.

### HasLabel

`func (o *TargetProviderInfo) HasLabel() bool`

HasLabel returns a boolean if a field has been set.

### GetName

`func (o *TargetProviderInfo) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *TargetProviderInfo) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *TargetProviderInfo) SetName(v string)`

SetName sets Name field to given value.


### GetVersion

`func (o *TargetProviderInfo) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *TargetProviderInfo) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *TargetProviderInfo) SetVersion(v string)`

SetVersion sets Version field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


