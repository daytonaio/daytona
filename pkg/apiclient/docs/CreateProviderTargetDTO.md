# CreateProviderTargetDTO

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**Options** | **string** |  | 
**ProviderInfo** | [**ProviderProviderInfo**](ProviderProviderInfo.md) |  | 

## Methods

### NewCreateProviderTargetDTO

`func NewCreateProviderTargetDTO(name string, options string, providerInfo ProviderProviderInfo, ) *CreateProviderTargetDTO`

NewCreateProviderTargetDTO instantiates a new CreateProviderTargetDTO object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateProviderTargetDTOWithDefaults

`func NewCreateProviderTargetDTOWithDefaults() *CreateProviderTargetDTO`

NewCreateProviderTargetDTOWithDefaults instantiates a new CreateProviderTargetDTO object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *CreateProviderTargetDTO) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *CreateProviderTargetDTO) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *CreateProviderTargetDTO) SetName(v string)`

SetName sets Name field to given value.


### GetOptions

`func (o *CreateProviderTargetDTO) GetOptions() string`

GetOptions returns the Options field if non-nil, zero value otherwise.

### GetOptionsOk

`func (o *CreateProviderTargetDTO) GetOptionsOk() (*string, bool)`

GetOptionsOk returns a tuple with the Options field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOptions

`func (o *CreateProviderTargetDTO) SetOptions(v string)`

SetOptions sets Options field to given value.


### GetProviderInfo

`func (o *CreateProviderTargetDTO) GetProviderInfo() ProviderProviderInfo`

GetProviderInfo returns the ProviderInfo field if non-nil, zero value otherwise.

### GetProviderInfoOk

`func (o *CreateProviderTargetDTO) GetProviderInfoOk() (*ProviderProviderInfo, bool)`

GetProviderInfoOk returns a tuple with the ProviderInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProviderInfo

`func (o *CreateProviderTargetDTO) SetProviderInfo(v ProviderProviderInfo)`

SetProviderInfo sets ProviderInfo field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


