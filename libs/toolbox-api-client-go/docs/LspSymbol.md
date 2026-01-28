# LspSymbol

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Kind** | **int32** |  | 
**Location** | [**LspLocation**](LspLocation.md) |  | 
**Name** | **string** |  | 

## Methods

### NewLspSymbol

`func NewLspSymbol(kind int32, location LspLocation, name string, ) *LspSymbol`

NewLspSymbol instantiates a new LspSymbol object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewLspSymbolWithDefaults

`func NewLspSymbolWithDefaults() *LspSymbol`

NewLspSymbolWithDefaults instantiates a new LspSymbol object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetKind

`func (o *LspSymbol) GetKind() int32`

GetKind returns the Kind field if non-nil, zero value otherwise.

### GetKindOk

`func (o *LspSymbol) GetKindOk() (*int32, bool)`

GetKindOk returns a tuple with the Kind field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKind

`func (o *LspSymbol) SetKind(v int32)`

SetKind sets Kind field to given value.


### GetLocation

`func (o *LspSymbol) GetLocation() LspLocation`

GetLocation returns the Location field if non-nil, zero value otherwise.

### GetLocationOk

`func (o *LspSymbol) GetLocationOk() (*LspLocation, bool)`

GetLocationOk returns a tuple with the Location field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLocation

`func (o *LspSymbol) SetLocation(v LspLocation)`

SetLocation sets Location field to given value.


### GetName

`func (o *LspSymbol) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *LspSymbol) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *LspSymbol) SetName(v string)`

SetName sets Name field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


