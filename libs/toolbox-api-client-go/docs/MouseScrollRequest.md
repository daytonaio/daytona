# MouseScrollRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Amount** | Pointer to **int32** |  | [optional] 
**Direction** | Pointer to **string** | up, down | [optional] 
**X** | Pointer to **int32** |  | [optional] 
**Y** | Pointer to **int32** |  | [optional] 

## Methods

### NewMouseScrollRequest

`func NewMouseScrollRequest() *MouseScrollRequest`

NewMouseScrollRequest instantiates a new MouseScrollRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMouseScrollRequestWithDefaults

`func NewMouseScrollRequestWithDefaults() *MouseScrollRequest`

NewMouseScrollRequestWithDefaults instantiates a new MouseScrollRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAmount

`func (o *MouseScrollRequest) GetAmount() int32`

GetAmount returns the Amount field if non-nil, zero value otherwise.

### GetAmountOk

`func (o *MouseScrollRequest) GetAmountOk() (*int32, bool)`

GetAmountOk returns a tuple with the Amount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAmount

`func (o *MouseScrollRequest) SetAmount(v int32)`

SetAmount sets Amount field to given value.

### HasAmount

`func (o *MouseScrollRequest) HasAmount() bool`

HasAmount returns a boolean if a field has been set.

### GetDirection

`func (o *MouseScrollRequest) GetDirection() string`

GetDirection returns the Direction field if non-nil, zero value otherwise.

### GetDirectionOk

`func (o *MouseScrollRequest) GetDirectionOk() (*string, bool)`

GetDirectionOk returns a tuple with the Direction field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDirection

`func (o *MouseScrollRequest) SetDirection(v string)`

SetDirection sets Direction field to given value.

### HasDirection

`func (o *MouseScrollRequest) HasDirection() bool`

HasDirection returns a boolean if a field has been set.

### GetX

`func (o *MouseScrollRequest) GetX() int32`

GetX returns the X field if non-nil, zero value otherwise.

### GetXOk

`func (o *MouseScrollRequest) GetXOk() (*int32, bool)`

GetXOk returns a tuple with the X field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetX

`func (o *MouseScrollRequest) SetX(v int32)`

SetX sets X field to given value.

### HasX

`func (o *MouseScrollRequest) HasX() bool`

HasX returns a boolean if a field has been set.

### GetY

`func (o *MouseScrollRequest) GetY() int32`

GetY returns the Y field if non-nil, zero value otherwise.

### GetYOk

`func (o *MouseScrollRequest) GetYOk() (*int32, bool)`

GetYOk returns a tuple with the Y field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetY

`func (o *MouseScrollRequest) SetY(v int32)`

SetY sets Y field to given value.

### HasY

`func (o *MouseScrollRequest) HasY() bool`

HasY returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


