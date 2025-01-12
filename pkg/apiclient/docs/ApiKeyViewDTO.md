# ApiKeyViewDTO

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Current** | **bool** |  | 
**Name** | **string** |  | 
**Type** | [**ModelsApiKeyType**](ModelsApiKeyType.md) |  | 

## Methods

### NewApiKeyViewDTO

`func NewApiKeyViewDTO(current bool, name string, type_ ModelsApiKeyType, ) *ApiKeyViewDTO`

NewApiKeyViewDTO instantiates a new ApiKeyViewDTO object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewApiKeyViewDTOWithDefaults

`func NewApiKeyViewDTOWithDefaults() *ApiKeyViewDTO`

NewApiKeyViewDTOWithDefaults instantiates a new ApiKeyViewDTO object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCurrent

`func (o *ApiKeyViewDTO) GetCurrent() bool`

GetCurrent returns the Current field if non-nil, zero value otherwise.

### GetCurrentOk

`func (o *ApiKeyViewDTO) GetCurrentOk() (*bool, bool)`

GetCurrentOk returns a tuple with the Current field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrent

`func (o *ApiKeyViewDTO) SetCurrent(v bool)`

SetCurrent sets Current field to given value.


### GetName

`func (o *ApiKeyViewDTO) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ApiKeyViewDTO) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ApiKeyViewDTO) SetName(v string)`

SetName sets Name field to given value.


### GetType

`func (o *ApiKeyViewDTO) GetType() ModelsApiKeyType`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *ApiKeyViewDTO) GetTypeOk() (*ModelsApiKeyType, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *ApiKeyViewDTO) SetType(v ModelsApiKeyType)`

SetType sets Type field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


