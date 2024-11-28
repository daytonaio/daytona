# AddTargetConfigDTO

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**Options** | **string** |  | 
**ProviderInfo** | [**TargetProviderInfo**](TargetProviderInfo.md) |  | 

## Methods

### NewAddTargetConfigDTO

`func NewAddTargetConfigDTO(name string, options string, providerInfo TargetProviderInfo, ) *AddTargetConfigDTO`

NewAddTargetConfigDTO instantiates a new AddTargetConfigDTO object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAddTargetConfigDTOWithDefaults

`func NewAddTargetConfigDTOWithDefaults() *AddTargetConfigDTO`

NewAddTargetConfigDTOWithDefaults instantiates a new AddTargetConfigDTO object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *AddTargetConfigDTO) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *AddTargetConfigDTO) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *AddTargetConfigDTO) SetName(v string)`

SetName sets Name field to given value.


### GetOptions

`func (o *AddTargetConfigDTO) GetOptions() string`

GetOptions returns the Options field if non-nil, zero value otherwise.

### GetOptionsOk

`func (o *AddTargetConfigDTO) GetOptionsOk() (*string, bool)`

GetOptionsOk returns a tuple with the Options field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOptions

`func (o *AddTargetConfigDTO) SetOptions(v string)`

SetOptions sets Options field to given value.


### GetProviderInfo

`func (o *AddTargetConfigDTO) GetProviderInfo() TargetProviderInfo`

GetProviderInfo returns the ProviderInfo field if non-nil, zero value otherwise.

### GetProviderInfoOk

`func (o *AddTargetConfigDTO) GetProviderInfoOk() (*TargetProviderInfo, bool)`

GetProviderInfoOk returns a tuple with the ProviderInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProviderInfo

`func (o *AddTargetConfigDTO) SetProviderInfo(v TargetProviderInfo)`

SetProviderInfo sets ProviderInfo field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


