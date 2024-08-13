# ApiKey

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**KeyHash** | **string** |  | 
**Name** | **string** | Project or client name | 
**Type** | [**ApikeyApiKeyType**](ApikeyApiKeyType.md) |  | 

## Methods

### NewApiKey

`func NewApiKey(keyHash string, name string, type_ ApikeyApiKeyType, ) *ApiKey`

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



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


