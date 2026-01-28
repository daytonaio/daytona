# MouseClickRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Button** | Pointer to **string** | left, right, middle | [optional] 
**Double** | Pointer to **bool** |  | [optional] 
**X** | Pointer to **int32** |  | [optional] 
**Y** | Pointer to **int32** |  | [optional] 

## Methods

### NewMouseClickRequest

`func NewMouseClickRequest() *MouseClickRequest`

NewMouseClickRequest instantiates a new MouseClickRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMouseClickRequestWithDefaults

`func NewMouseClickRequestWithDefaults() *MouseClickRequest`

NewMouseClickRequestWithDefaults instantiates a new MouseClickRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetButton

`func (o *MouseClickRequest) GetButton() string`

GetButton returns the Button field if non-nil, zero value otherwise.

### GetButtonOk

`func (o *MouseClickRequest) GetButtonOk() (*string, bool)`

GetButtonOk returns a tuple with the Button field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetButton

`func (o *MouseClickRequest) SetButton(v string)`

SetButton sets Button field to given value.

### HasButton

`func (o *MouseClickRequest) HasButton() bool`

HasButton returns a boolean if a field has been set.

### GetDouble

`func (o *MouseClickRequest) GetDouble() bool`

GetDouble returns the Double field if non-nil, zero value otherwise.

### GetDoubleOk

`func (o *MouseClickRequest) GetDoubleOk() (*bool, bool)`

GetDoubleOk returns a tuple with the Double field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDouble

`func (o *MouseClickRequest) SetDouble(v bool)`

SetDouble sets Double field to given value.

### HasDouble

`func (o *MouseClickRequest) HasDouble() bool`

HasDouble returns a boolean if a field has been set.

### GetX

`func (o *MouseClickRequest) GetX() int32`

GetX returns the X field if non-nil, zero value otherwise.

### GetXOk

`func (o *MouseClickRequest) GetXOk() (*int32, bool)`

GetXOk returns a tuple with the X field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetX

`func (o *MouseClickRequest) SetX(v int32)`

SetX sets X field to given value.

### HasX

`func (o *MouseClickRequest) HasX() bool`

HasX returns a boolean if a field has been set.

### GetY

`func (o *MouseClickRequest) GetY() int32`

GetY returns the Y field if non-nil, zero value otherwise.

### GetYOk

`func (o *MouseClickRequest) GetYOk() (*int32, bool)`

GetYOk returns a tuple with the Y field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetY

`func (o *MouseClickRequest) SetY(v int32)`

SetY sets Y field to given value.

### HasY

`func (o *MouseClickRequest) HasY() bool`

HasY returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


