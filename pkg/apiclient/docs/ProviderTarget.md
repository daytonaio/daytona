# ProviderTarget

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** |  | [optional] 
**Options** | Pointer to **string** | JSON encoded map of options | [optional] 
**ProviderInfo** | Pointer to [**ProviderProviderInfo**](ProviderProviderInfo.md) |  | [optional] 

## Methods

### NewProviderTarget

`func NewProviderTarget() *ProviderTarget`

NewProviderTarget instantiates a new ProviderTarget object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProviderTargetWithDefaults

`func NewProviderTargetWithDefaults() *ProviderTarget`

NewProviderTargetWithDefaults instantiates a new ProviderTarget object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *ProviderTarget) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ProviderTarget) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ProviderTarget) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *ProviderTarget) HasName() bool`

HasName returns a boolean if a field has been set.

### GetOptions

`func (o *ProviderTarget) GetOptions() string`

GetOptions returns the Options field if non-nil, zero value otherwise.

### GetOptionsOk

`func (o *ProviderTarget) GetOptionsOk() (*string, bool)`

GetOptionsOk returns a tuple with the Options field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOptions

`func (o *ProviderTarget) SetOptions(v string)`

SetOptions sets Options field to given value.

### HasOptions

`func (o *ProviderTarget) HasOptions() bool`

HasOptions returns a boolean if a field has been set.

### GetProviderInfo

`func (o *ProviderTarget) GetProviderInfo() ProviderProviderInfo`

GetProviderInfo returns the ProviderInfo field if non-nil, zero value otherwise.

### GetProviderInfoOk

`func (o *ProviderTarget) GetProviderInfoOk() (*ProviderProviderInfo, bool)`

GetProviderInfoOk returns a tuple with the ProviderInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProviderInfo

`func (o *ProviderTarget) SetProviderInfo(v ProviderProviderInfo)`

SetProviderInfo sets ProviderInfo field to given value.

### HasProviderInfo

`func (o *ProviderTarget) HasProviderInfo() bool`

HasProviderInfo returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


