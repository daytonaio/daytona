# ProviderDTO

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Label** | Pointer to **string** |  | [optional] 
**Latest** | **bool** |  | 
**Name** | **string** |  | 
**Version** | **string** |  | 

## Methods

### NewProviderDTO

`func NewProviderDTO(latest bool, name string, version string, ) *ProviderDTO`

NewProviderDTO instantiates a new ProviderDTO object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProviderDTOWithDefaults

`func NewProviderDTOWithDefaults() *ProviderDTO`

NewProviderDTOWithDefaults instantiates a new ProviderDTO object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLabel

`func (o *ProviderDTO) GetLabel() string`

GetLabel returns the Label field if non-nil, zero value otherwise.

### GetLabelOk

`func (o *ProviderDTO) GetLabelOk() (*string, bool)`

GetLabelOk returns a tuple with the Label field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabel

`func (o *ProviderDTO) SetLabel(v string)`

SetLabel sets Label field to given value.

### HasLabel

`func (o *ProviderDTO) HasLabel() bool`

HasLabel returns a boolean if a field has been set.

### GetLatest

`func (o *ProviderDTO) GetLatest() bool`

GetLatest returns the Latest field if non-nil, zero value otherwise.

### GetLatestOk

`func (o *ProviderDTO) GetLatestOk() (*bool, bool)`

GetLatestOk returns a tuple with the Latest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLatest

`func (o *ProviderDTO) SetLatest(v bool)`

SetLatest sets Latest field to given value.


### GetName

`func (o *ProviderDTO) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ProviderDTO) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ProviderDTO) SetName(v string)`

SetName sets Name field to given value.


### GetVersion

`func (o *ProviderDTO) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *ProviderDTO) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *ProviderDTO) SetVersion(v string)`

SetVersion sets Version field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


