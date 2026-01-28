# LspDocumentRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**LanguageId** | **string** |  | 
**PathToProject** | **string** |  | 
**Uri** | **string** |  | 

## Methods

### NewLspDocumentRequest

`func NewLspDocumentRequest(languageId string, pathToProject string, uri string, ) *LspDocumentRequest`

NewLspDocumentRequest instantiates a new LspDocumentRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewLspDocumentRequestWithDefaults

`func NewLspDocumentRequestWithDefaults() *LspDocumentRequest`

NewLspDocumentRequestWithDefaults instantiates a new LspDocumentRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLanguageId

`func (o *LspDocumentRequest) GetLanguageId() string`

GetLanguageId returns the LanguageId field if non-nil, zero value otherwise.

### GetLanguageIdOk

`func (o *LspDocumentRequest) GetLanguageIdOk() (*string, bool)`

GetLanguageIdOk returns a tuple with the LanguageId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLanguageId

`func (o *LspDocumentRequest) SetLanguageId(v string)`

SetLanguageId sets LanguageId field to given value.


### GetPathToProject

`func (o *LspDocumentRequest) GetPathToProject() string`

GetPathToProject returns the PathToProject field if non-nil, zero value otherwise.

### GetPathToProjectOk

`func (o *LspDocumentRequest) GetPathToProjectOk() (*string, bool)`

GetPathToProjectOk returns a tuple with the PathToProject field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPathToProject

`func (o *LspDocumentRequest) SetPathToProject(v string)`

SetPathToProject sets PathToProject field to given value.


### GetUri

`func (o *LspDocumentRequest) GetUri() string`

GetUri returns the Uri field if non-nil, zero value otherwise.

### GetUriOk

`func (o *LspDocumentRequest) GetUriOk() (*string, bool)`

GetUriOk returns a tuple with the Uri field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUri

`func (o *LspDocumentRequest) SetUri(v string)`

SetUri sets Uri field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


