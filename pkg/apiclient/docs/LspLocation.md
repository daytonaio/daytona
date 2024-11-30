# LspLocation

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Range** | [**LspRange**](LspRange.md) |  | 
**Uri** | **string** |  | 

## Methods

### NewLspLocation

`func NewLspLocation(range_ LspRange, uri string, ) *LspLocation`

NewLspLocation instantiates a new LspLocation object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewLspLocationWithDefaults

`func NewLspLocationWithDefaults() *LspLocation`

NewLspLocationWithDefaults instantiates a new LspLocation object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRange

`func (o *LspLocation) GetRange() LspRange`

GetRange returns the Range field if non-nil, zero value otherwise.

### GetRangeOk

`func (o *LspLocation) GetRangeOk() (*LspRange, bool)`

GetRangeOk returns a tuple with the Range field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRange

`func (o *LspLocation) SetRange(v LspRange)`

SetRange sets Range field to given value.


### GetUri

`func (o *LspLocation) GetUri() string`

GetUri returns the Uri field if non-nil, zero value otherwise.

### GetUriOk

`func (o *LspLocation) GetUriOk() (*string, bool)`

GetUriOk returns a tuple with the Uri field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUri

`func (o *LspLocation) SetUri(v string)`

SetUri sets Uri field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


