# KeyboardTypeRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Delay** | Pointer to **int32** | milliseconds between keystrokes | [optional] 
**Text** | Pointer to **string** |  | [optional] 

## Methods

### NewKeyboardTypeRequest

`func NewKeyboardTypeRequest() *KeyboardTypeRequest`

NewKeyboardTypeRequest instantiates a new KeyboardTypeRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewKeyboardTypeRequestWithDefaults

`func NewKeyboardTypeRequestWithDefaults() *KeyboardTypeRequest`

NewKeyboardTypeRequestWithDefaults instantiates a new KeyboardTypeRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDelay

`func (o *KeyboardTypeRequest) GetDelay() int32`

GetDelay returns the Delay field if non-nil, zero value otherwise.

### GetDelayOk

`func (o *KeyboardTypeRequest) GetDelayOk() (*int32, bool)`

GetDelayOk returns a tuple with the Delay field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDelay

`func (o *KeyboardTypeRequest) SetDelay(v int32)`

SetDelay sets Delay field to given value.

### HasDelay

`func (o *KeyboardTypeRequest) HasDelay() bool`

HasDelay returns a boolean if a field has been set.

### GetText

`func (o *KeyboardTypeRequest) GetText() string`

GetText returns the Text field if non-nil, zero value otherwise.

### GetTextOk

`func (o *KeyboardTypeRequest) GetTextOk() (*string, bool)`

GetTextOk returns a tuple with the Text field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetText

`func (o *KeyboardTypeRequest) SetText(v string)`

SetText sets Text field to given value.

### HasText

`func (o *KeyboardTypeRequest) HasText() bool`

HasText returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


