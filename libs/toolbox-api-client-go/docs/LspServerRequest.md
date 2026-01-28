# LspServerRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**LanguageId** | **string** |  | 
**PathToProject** | **string** |  | 

## Methods

### NewLspServerRequest

`func NewLspServerRequest(languageId string, pathToProject string, ) *LspServerRequest`

NewLspServerRequest instantiates a new LspServerRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewLspServerRequestWithDefaults

`func NewLspServerRequestWithDefaults() *LspServerRequest`

NewLspServerRequestWithDefaults instantiates a new LspServerRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLanguageId

`func (o *LspServerRequest) GetLanguageId() string`

GetLanguageId returns the LanguageId field if non-nil, zero value otherwise.

### GetLanguageIdOk

`func (o *LspServerRequest) GetLanguageIdOk() (*string, bool)`

GetLanguageIdOk returns a tuple with the LanguageId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLanguageId

`func (o *LspServerRequest) SetLanguageId(v string)`

SetLanguageId sets LanguageId field to given value.


### GetPathToProject

`func (o *LspServerRequest) GetPathToProject() string`

GetPathToProject returns the PathToProject field if non-nil, zero value otherwise.

### GetPathToProjectOk

`func (o *LspServerRequest) GetPathToProjectOk() (*string, bool)`

GetPathToProjectOk returns a tuple with the PathToProject field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPathToProject

`func (o *LspServerRequest) SetPathToProject(v string)`

SetPathToProject sets PathToProject field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


