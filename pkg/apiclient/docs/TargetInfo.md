# TargetInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**ProviderMetadata** | Pointer to **string** |  | [optional] 

## Methods

### NewTargetInfo

`func NewTargetInfo(name string, ) *TargetInfo`

NewTargetInfo instantiates a new TargetInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTargetInfoWithDefaults

`func NewTargetInfoWithDefaults() *TargetInfo`

NewTargetInfoWithDefaults instantiates a new TargetInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *TargetInfo) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *TargetInfo) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *TargetInfo) SetName(v string)`

SetName sets Name field to given value.


### GetProviderMetadata

`func (o *TargetInfo) GetProviderMetadata() string`

GetProviderMetadata returns the ProviderMetadata field if non-nil, zero value otherwise.

### GetProviderMetadataOk

`func (o *TargetInfo) GetProviderMetadataOk() (*string, bool)`

GetProviderMetadataOk returns a tuple with the ProviderMetadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProviderMetadata

`func (o *TargetInfo) SetProviderMetadata(v string)`

SetProviderMetadata sets ProviderMetadata field to given value.

### HasProviderMetadata

`func (o *TargetInfo) HasProviderMetadata() bool`

HasProviderMetadata returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


