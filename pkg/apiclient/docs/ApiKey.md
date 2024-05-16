# ApiKey

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**KeyHash** | Pointer to **string** |  | [optional] 
**Name** | Pointer to **string** | Project or client name | [optional] 
**Type** | Pointer to [**ApikeyApiKeyType**](ApikeyApiKeyType.md) |  | [optional] 

## Methods

### NewApiKey

`func NewApiKey() *ApiKey`

NewApiKey instantiates a new ApiKey object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewApiKeyWithDefaults

`func NewApiKeyWithDefaults() *ApiKey`

NewApiKeyWithDefaults instantiates a new ApiKey object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetKeyHash

`func (o *ApiKey) GetKeyHash() string`

GetKeyHash returns the KeyHash field if non-nil, zero value otherwise.

### GetKeyHashOk

`func (o *ApiKey) GetKeyHashOk() (*string, bool)`

GetKeyHashOk returns a tuple with the KeyHash field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKeyHash

`func (o *ApiKey) SetKeyHash(v string)`

SetKeyHash sets KeyHash field to given value.

### HasKeyHash

`func (o *ApiKey) HasKeyHash() bool`

HasKeyHash returns a boolean if a field has been set.

### GetName

`func (o *ApiKey) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ApiKey) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ApiKey) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *ApiKey) HasName() bool`

HasName returns a boolean if a field has been set.

### GetType

`func (o *ApiKey) GetType() ApikeyApiKeyType`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *ApiKey) GetTypeOk() (*ApikeyApiKeyType, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *ApiKey) SetType(v ApikeyApiKeyType)`

SetType sets Type field to given value.

### HasType

`func (o *ApiKey) HasType() bool`

HasType returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


