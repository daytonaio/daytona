# KeyboardHotkeyRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Keys** | Pointer to **string** | e.g., \&quot;ctrl+c\&quot;, \&quot;cmd+v\&quot; | [optional] 

## Methods

### NewKeyboardHotkeyRequest

`func NewKeyboardHotkeyRequest() *KeyboardHotkeyRequest`

NewKeyboardHotkeyRequest instantiates a new KeyboardHotkeyRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewKeyboardHotkeyRequestWithDefaults

`func NewKeyboardHotkeyRequestWithDefaults() *KeyboardHotkeyRequest`

NewKeyboardHotkeyRequestWithDefaults instantiates a new KeyboardHotkeyRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetKeys

`func (o *KeyboardHotkeyRequest) GetKeys() string`

GetKeys returns the Keys field if non-nil, zero value otherwise.

### GetKeysOk

`func (o *KeyboardHotkeyRequest) GetKeysOk() (*string, bool)`

GetKeysOk returns a tuple with the Keys field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKeys

`func (o *KeyboardHotkeyRequest) SetKeys(v string)`

SetKeys sets Keys field to given value.

### HasKeys

`func (o *KeyboardHotkeyRequest) HasKeys() bool`

HasKeys returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


