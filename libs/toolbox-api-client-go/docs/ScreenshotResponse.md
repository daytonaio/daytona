# ScreenshotResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CursorPosition** | Pointer to [**Position**](Position.md) |  | [optional] 
**Screenshot** | Pointer to **string** |  | [optional] 
**SizeBytes** | Pointer to **int32** |  | [optional] 

## Methods

### NewScreenshotResponse

`func NewScreenshotResponse() *ScreenshotResponse`

NewScreenshotResponse instantiates a new ScreenshotResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewScreenshotResponseWithDefaults

`func NewScreenshotResponseWithDefaults() *ScreenshotResponse`

NewScreenshotResponseWithDefaults instantiates a new ScreenshotResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCursorPosition

`func (o *ScreenshotResponse) GetCursorPosition() Position`

GetCursorPosition returns the CursorPosition field if non-nil, zero value otherwise.

### GetCursorPositionOk

`func (o *ScreenshotResponse) GetCursorPositionOk() (*Position, bool)`

GetCursorPositionOk returns a tuple with the CursorPosition field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCursorPosition

`func (o *ScreenshotResponse) SetCursorPosition(v Position)`

SetCursorPosition sets CursorPosition field to given value.

### HasCursorPosition

`func (o *ScreenshotResponse) HasCursorPosition() bool`

HasCursorPosition returns a boolean if a field has been set.

### GetScreenshot

`func (o *ScreenshotResponse) GetScreenshot() string`

GetScreenshot returns the Screenshot field if non-nil, zero value otherwise.

### GetScreenshotOk

`func (o *ScreenshotResponse) GetScreenshotOk() (*string, bool)`

GetScreenshotOk returns a tuple with the Screenshot field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScreenshot

`func (o *ScreenshotResponse) SetScreenshot(v string)`

SetScreenshot sets Screenshot field to given value.

### HasScreenshot

`func (o *ScreenshotResponse) HasScreenshot() bool`

HasScreenshot returns a boolean if a field has been set.

### GetSizeBytes

`func (o *ScreenshotResponse) GetSizeBytes() int32`

GetSizeBytes returns the SizeBytes field if non-nil, zero value otherwise.

### GetSizeBytesOk

`func (o *ScreenshotResponse) GetSizeBytesOk() (*int32, bool)`

GetSizeBytesOk returns a tuple with the SizeBytes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSizeBytes

`func (o *ScreenshotResponse) SetSizeBytes(v int32)`

SetSizeBytes sets SizeBytes field to given value.

### HasSizeBytes

`func (o *ScreenshotResponse) HasSizeBytes() bool`

HasSizeBytes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


