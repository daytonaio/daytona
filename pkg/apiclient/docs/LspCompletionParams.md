# LspCompletionParams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Context** | Pointer to [**GithubComDaytonaioDaytonaPkgAgentToolboxLspCompletionContext**](GithubComDaytonaioDaytonaPkgAgentToolboxLspCompletionContext.md) |  | [optional] 
**LanguageId** | **string** |  | 
**Position** | Pointer to [**GithubComDaytonaioDaytonaPkgAgentToolboxLspPosition**](GithubComDaytonaioDaytonaPkgAgentToolboxLspPosition.md) |  | [optional] 
**TextDocument** | Pointer to [**GithubComDaytonaioDaytonaPkgAgentToolboxLspTextDocumentIdentifier**](GithubComDaytonaioDaytonaPkgAgentToolboxLspTextDocumentIdentifier.md) |  | [optional] 

## Methods

### NewLspCompletionParams

`func NewLspCompletionParams(languageId string, ) *LspCompletionParams`

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

`func (o *LspCompletionParams) GetContext() GithubComDaytonaioDaytonaPkgAgentToolboxLspCompletionContext`

GetContext returns the Context field if non-nil, zero value otherwise.

### GetContextOk

`func (o *LspCompletionParams) GetContextOk() (*GithubComDaytonaioDaytonaPkgAgentToolboxLspCompletionContext, bool)`

GetContextOk returns a tuple with the Context field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContext

`func (o *LspCompletionParams) SetContext(v GithubComDaytonaioDaytonaPkgAgentToolboxLspCompletionContext)`

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


### GetPosition

`func (o *LspCompletionParams) GetPosition() GithubComDaytonaioDaytonaPkgAgentToolboxLspPosition`

GetPosition returns the Position field if non-nil, zero value otherwise.

### GetPositionOk

`func (o *LspCompletionParams) GetPositionOk() (*GithubComDaytonaioDaytonaPkgAgentToolboxLspPosition, bool)`

GetPositionOk returns a tuple with the Position field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPosition

`func (o *LspCompletionParams) SetPosition(v GithubComDaytonaioDaytonaPkgAgentToolboxLspPosition)`

SetPosition sets Position field to given value.

### HasPosition

`func (o *LspCompletionParams) HasPosition() bool`

HasPosition returns a boolean if a field has been set.

### GetTextDocument

`func (o *LspCompletionParams) GetTextDocument() GithubComDaytonaioDaytonaPkgAgentToolboxLspTextDocumentIdentifier`

GetTextDocument returns the TextDocument field if non-nil, zero value otherwise.

### GetTextDocumentOk

`func (o *LspCompletionParams) GetTextDocumentOk() (*GithubComDaytonaioDaytonaPkgAgentToolboxLspTextDocumentIdentifier, bool)`

GetTextDocumentOk returns a tuple with the TextDocument field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTextDocument

`func (o *LspCompletionParams) SetTextDocument(v GithubComDaytonaioDaytonaPkgAgentToolboxLspTextDocumentIdentifier)`

SetTextDocument sets TextDocument field to given value.

### HasTextDocument

`func (o *LspCompletionParams) HasTextDocument() bool`

HasTextDocument returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


