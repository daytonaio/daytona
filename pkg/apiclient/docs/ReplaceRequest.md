# ReplaceRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Files** | **[]string** |  | 
**NewValue** | **string** |  | 
**Pattern** | **string** |  | 

## Methods

### NewReplaceRequest

`func NewReplaceRequest(files []string, newValue string, pattern string, ) *ReplaceRequest`

NewReplaceRequest instantiates a new ReplaceRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewReplaceRequestWithDefaults

`func NewReplaceRequestWithDefaults() *ReplaceRequest`

NewReplaceRequestWithDefaults instantiates a new ReplaceRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFiles

`func (o *ReplaceRequest) GetFiles() []string`

GetFiles returns the Files field if non-nil, zero value otherwise.

### GetFilesOk

`func (o *ReplaceRequest) GetFilesOk() (*[]string, bool)`

GetFilesOk returns a tuple with the Files field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFiles

`func (o *ReplaceRequest) SetFiles(v []string)`

SetFiles sets Files field to given value.


### GetNewValue

`func (o *ReplaceRequest) GetNewValue() string`

GetNewValue returns the NewValue field if non-nil, zero value otherwise.

### GetNewValueOk

`func (o *ReplaceRequest) GetNewValueOk() (*string, bool)`

GetNewValueOk returns a tuple with the NewValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNewValue

`func (o *ReplaceRequest) SetNewValue(v string)`

SetNewValue sets NewValue field to given value.


### GetPattern

`func (o *ReplaceRequest) GetPattern() string`

GetPattern returns the Pattern field if non-nil, zero value otherwise.

### GetPatternOk

`func (o *ReplaceRequest) GetPatternOk() (*string, bool)`

GetPatternOk returns a tuple with the Pattern field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPattern

`func (o *ReplaceRequest) SetPattern(v string)`

SetPattern sets Pattern field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


