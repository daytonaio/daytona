# LspCompletionParams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Context** | Pointer to [**CompletionContext**](CompletionContext.md) |  | [optional] 
**LanguageId** | **string** |  | 
**PathToProject** | **string** |  | 
**Position** | [**LspPosition**](LspPosition.md) |  | 
**Uri** | **string** |  | 

## Methods

### NewLspCompletionParams

`func NewLspCompletionParams(languageId string, pathToProject string, position LspPosition, uri string, ) *LspCompletionParams`

NewLspCompletionParams instantiates a new LspCompletionParams object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewLspCompletionParamsWithDefaults

`func NewLspCompletionParamsWithDefaults() *LspCompletionParams`

NewLspCompletionParamsWithDefaults instantiates a new LspCompletionParams object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetContext

`func (o *LspCompletionParams) GetContext() CompletionContext`

GetContext returns the Context field if non-nil, zero value otherwise.

### GetContextOk

`func (o *LspCompletionParams) GetContextOk() (*CompletionContext, bool)`

GetContextOk returns a tuple with the Context field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContext

`func (o *LspCompletionParams) SetContext(v CompletionContext)`

SetContext sets Context field to given value.

### HasContext

`func (o *LspCompletionParams) HasContext() bool`

HasContext returns a boolean if a field has been set.

### GetLanguageId

`func (o *LspCompletionParams) GetLanguageId() string`

GetLanguageId returns the LanguageId field if non-nil, zero value otherwise.

### GetLanguageIdOk

`func (o *LspCompletionParams) GetLanguageIdOk() (*string, bool)`

GetLanguageIdOk returns a tuple with the LanguageId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLanguageId

`func (o *LspCompletionParams) SetLanguageId(v string)`

SetLanguageId sets LanguageId field to given value.


### GetPathToProject

`func (o *LspCompletionParams) GetPathToProject() string`

GetPathToProject returns the PathToProject field if non-nil, zero value otherwise.

### GetPathToProjectOk

`func (o *LspCompletionParams) GetPathToProjectOk() (*string, bool)`

GetPathToProjectOk returns a tuple with the PathToProject field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPathToProject

`func (o *LspCompletionParams) SetPathToProject(v string)`

SetPathToProject sets PathToProject field to given value.


### GetPosition

`func (o *LspCompletionParams) GetPosition() LspPosition`

GetPosition returns the Position field if non-nil, zero value otherwise.

### GetPositionOk

`func (o *LspCompletionParams) GetPositionOk() (*LspPosition, bool)`

GetPositionOk returns a tuple with the Position field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPosition

`func (o *LspCompletionParams) SetPosition(v LspPosition)`

SetPosition sets Position field to given value.


### GetUri

`func (o *LspCompletionParams) GetUri() string`

GetUri returns the Uri field if non-nil, zero value otherwise.

### GetUriOk

`func (o *LspCompletionParams) GetUriOk() (*string, bool)`

GetUriOk returns a tuple with the Uri field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUri

`func (o *LspCompletionParams) SetUri(v string)`

SetUri sets Uri field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


