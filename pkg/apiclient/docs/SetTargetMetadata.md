# SetTargetMetadata

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ProviderMetadata** | Pointer to **string** |  | [optional] 
**Uptime** | **int32** |  | 

## Methods

### NewSetTargetMetadata

`func NewSetTargetMetadata(uptime int32, ) *SetTargetMetadata`

NewSetTargetMetadata instantiates a new SetTargetMetadata object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSetTargetMetadataWithDefaults

`func NewSetTargetMetadataWithDefaults() *SetTargetMetadata`

NewSetTargetMetadataWithDefaults instantiates a new SetTargetMetadata object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetProviderMetadata

`func (o *SetTargetMetadata) GetProviderMetadata() string`

GetProviderMetadata returns the ProviderMetadata field if non-nil, zero value otherwise.

### GetProviderMetadataOk

`func (o *SetTargetMetadata) GetProviderMetadataOk() (*string, bool)`

GetProviderMetadataOk returns a tuple with the ProviderMetadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProviderMetadata

`func (o *SetTargetMetadata) SetProviderMetadata(v string)`

SetProviderMetadata sets ProviderMetadata field to given value.

### HasProviderMetadata

`func (o *SetTargetMetadata) HasProviderMetadata() bool`

HasProviderMetadata returns a boolean if a field has been set.

### GetUptime

`func (o *SetTargetMetadata) GetUptime() int32`

GetUptime returns the Uptime field if non-nil, zero value otherwise.

### GetUptimeOk

`func (o *SetTargetMetadata) GetUptimeOk() (*int32, bool)`

GetUptimeOk returns a tuple with the Uptime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUptime

`func (o *SetTargetMetadata) SetUptime(v int32)`

SetUptime sets Uptime field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


