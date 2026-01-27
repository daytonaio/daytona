# CreateContextRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Cwd** | Pointer to **string** |  | [optional] 
**Language** | Pointer to **string** |  | [optional] 

## Methods

### NewCreateContextRequest

`func NewCreateContextRequest() *CreateContextRequest`

NewCreateContextRequest instantiates a new CreateContextRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateContextRequestWithDefaults

`func NewCreateContextRequestWithDefaults() *CreateContextRequest`

NewCreateContextRequestWithDefaults instantiates a new CreateContextRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCwd

`func (o *CreateContextRequest) GetCwd() string`

GetCwd returns the Cwd field if non-nil, zero value otherwise.

### GetCwdOk

`func (o *CreateContextRequest) GetCwdOk() (*string, bool)`

GetCwdOk returns a tuple with the Cwd field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCwd

`func (o *CreateContextRequest) SetCwd(v string)`

SetCwd sets Cwd field to given value.

### HasCwd

`func (o *CreateContextRequest) HasCwd() bool`

HasCwd returns a boolean if a field has been set.

### GetLanguage

`func (o *CreateContextRequest) GetLanguage() string`

GetLanguage returns the Language field if non-nil, zero value otherwise.

### GetLanguageOk

`func (o *CreateContextRequest) GetLanguageOk() (*string, bool)`

GetLanguageOk returns a tuple with the Language field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLanguage

`func (o *CreateContextRequest) SetLanguage(v string)`

SetLanguage sets Language field to given value.

### HasLanguage

`func (o *CreateContextRequest) HasLanguage() bool`

HasLanguage returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


