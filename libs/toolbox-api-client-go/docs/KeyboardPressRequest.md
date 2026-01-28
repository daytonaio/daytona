# KeyboardPressRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Key** | Pointer to **string** |  | [optional] 
**Modifiers** | Pointer to **[]string** | ctrl, alt, shift, cmd | [optional] 

## Methods

### NewKeyboardPressRequest

`func NewKeyboardPressRequest() *KeyboardPressRequest`

NewKeyboardPressRequest instantiates a new KeyboardPressRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewKeyboardPressRequestWithDefaults

`func NewKeyboardPressRequestWithDefaults() *KeyboardPressRequest`

NewKeyboardPressRequestWithDefaults instantiates a new KeyboardPressRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetKey

`func (o *KeyboardPressRequest) GetKey() string`

GetKey returns the Key field if non-nil, zero value otherwise.

### GetKeyOk

`func (o *KeyboardPressRequest) GetKeyOk() (*string, bool)`

GetKeyOk returns a tuple with the Key field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKey

`func (o *KeyboardPressRequest) SetKey(v string)`

SetKey sets Key field to given value.

### HasKey

`func (o *KeyboardPressRequest) HasKey() bool`

HasKey returns a boolean if a field has been set.

### GetModifiers

`func (o *KeyboardPressRequest) GetModifiers() []string`

GetModifiers returns the Modifiers field if non-nil, zero value otherwise.

### GetModifiersOk

`func (o *KeyboardPressRequest) GetModifiersOk() (*[]string, bool)`

GetModifiersOk returns a tuple with the Modifiers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetModifiers

`func (o *KeyboardPressRequest) SetModifiers(v []string)`

SetModifiers sets Modifiers field to given value.

### HasModifiers

`func (o *KeyboardPressRequest) HasModifiers() bool`

HasModifiers returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


