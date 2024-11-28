# TargetConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Deleted** | **bool** |  | 
**Id** | **string** |  | 
**Name** | **string** |  | 
**Options** | **string** | JSON encoded map of options | 
**ProviderInfo** | [**TargetProviderInfo**](TargetProviderInfo.md) |  | 

## Methods

### NewTargetConfig

`func NewTargetConfig(deleted bool, id string, name string, options string, providerInfo TargetProviderInfo, ) *TargetConfig`

NewTargetConfig instantiates a new TargetConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTargetConfigWithDefaults

`func NewTargetConfigWithDefaults() *TargetConfig`

NewTargetConfigWithDefaults instantiates a new TargetConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDeleted

`func (o *TargetConfig) GetDeleted() bool`

GetDeleted returns the Deleted field if non-nil, zero value otherwise.

### GetDeletedOk

`func (o *TargetConfig) GetDeletedOk() (*bool, bool)`

GetDeletedOk returns a tuple with the Deleted field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeleted

`func (o *TargetConfig) SetDeleted(v bool)`

SetDeleted sets Deleted field to given value.


### GetId

`func (o *TargetConfig) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *TargetConfig) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *TargetConfig) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *TargetConfig) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *TargetConfig) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *TargetConfig) SetName(v string)`

SetName sets Name field to given value.


### GetOptions

`func (o *TargetConfig) GetOptions() string`

GetOptions returns the Options field if non-nil, zero value otherwise.

### GetOptionsOk

`func (o *TargetConfig) GetOptionsOk() (*string, bool)`

GetOptionsOk returns a tuple with the Options field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOptions

`func (o *TargetConfig) SetOptions(v string)`

SetOptions sets Options field to given value.


### GetProviderInfo

`func (o *TargetConfig) GetProviderInfo() TargetProviderInfo`

GetProviderInfo returns the ProviderInfo field if non-nil, zero value otherwise.

### GetProviderInfoOk

`func (o *TargetConfig) GetProviderInfoOk() (*TargetProviderInfo, bool)`

GetProviderInfoOk returns a tuple with the ProviderInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProviderInfo

`func (o *TargetConfig) SetProviderInfo(v TargetProviderInfo)`

SetProviderInfo sets ProviderInfo field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


