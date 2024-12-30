# CompletionItem

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Detail** | Pointer to **string** |  | [optional] 
**Documentation** | Pointer to **map[string]interface{}** |  | [optional] 
**FilterText** | Pointer to **string** |  | [optional] 
**InsertText** | Pointer to **string** |  | [optional] 
**Kind** | Pointer to **int32** |  | [optional] 
**Label** | **string** |  | 
**SortText** | Pointer to **string** |  | [optional] 

## Methods

### NewCompletionItem

`func NewCompletionItem(label string, ) *CompletionItem`

NewCompletionItem instantiates a new CompletionItem object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCompletionItemWithDefaults

`func NewCompletionItemWithDefaults() *CompletionItem`

NewCompletionItemWithDefaults instantiates a new CompletionItem object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDetail

`func (o *CompletionItem) GetDetail() string`

GetDetail returns the Detail field if non-nil, zero value otherwise.

### GetDetailOk

`func (o *CompletionItem) GetDetailOk() (*string, bool)`

GetDetailOk returns a tuple with the Detail field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDetail

`func (o *CompletionItem) SetDetail(v string)`

SetDetail sets Detail field to given value.

### HasDetail

`func (o *CompletionItem) HasDetail() bool`

HasDetail returns a boolean if a field has been set.

### GetDocumentation

`func (o *CompletionItem) GetDocumentation() map[string]interface{}`

GetDocumentation returns the Documentation field if non-nil, zero value otherwise.

### GetDocumentationOk

`func (o *CompletionItem) GetDocumentationOk() (*map[string]interface{}, bool)`

GetDocumentationOk returns a tuple with the Documentation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDocumentation

`func (o *CompletionItem) SetDocumentation(v map[string]interface{})`

SetDocumentation sets Documentation field to given value.

### HasDocumentation

`func (o *CompletionItem) HasDocumentation() bool`

HasDocumentation returns a boolean if a field has been set.

### GetFilterText

`func (o *CompletionItem) GetFilterText() string`

GetFilterText returns the FilterText field if non-nil, zero value otherwise.

### GetFilterTextOk

`func (o *CompletionItem) GetFilterTextOk() (*string, bool)`

GetFilterTextOk returns a tuple with the FilterText field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFilterText

`func (o *CompletionItem) SetFilterText(v string)`

SetFilterText sets FilterText field to given value.

### HasFilterText

`func (o *CompletionItem) HasFilterText() bool`

HasFilterText returns a boolean if a field has been set.

### GetInsertText

`func (o *CompletionItem) GetInsertText() string`

GetInsertText returns the InsertText field if non-nil, zero value otherwise.

### GetInsertTextOk

`func (o *CompletionItem) GetInsertTextOk() (*string, bool)`

GetInsertTextOk returns a tuple with the InsertText field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInsertText

`func (o *CompletionItem) SetInsertText(v string)`

SetInsertText sets InsertText field to given value.

### HasInsertText

`func (o *CompletionItem) HasInsertText() bool`

HasInsertText returns a boolean if a field has been set.

### GetKind

`func (o *CompletionItem) GetKind() int32`

GetKind returns the Kind field if non-nil, zero value otherwise.

### GetKindOk

`func (o *CompletionItem) GetKindOk() (*int32, bool)`

GetKindOk returns a tuple with the Kind field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKind

`func (o *CompletionItem) SetKind(v int32)`

SetKind sets Kind field to given value.

### HasKind

`func (o *CompletionItem) HasKind() bool`

HasKind returns a boolean if a field has been set.

### GetLabel

`func (o *CompletionItem) GetLabel() string`

GetLabel returns the Label field if non-nil, zero value otherwise.

### GetLabelOk

`func (o *CompletionItem) GetLabelOk() (*string, bool)`

GetLabelOk returns a tuple with the Label field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabel

`func (o *CompletionItem) SetLabel(v string)`

SetLabel sets Label field to given value.


### GetSortText

`func (o *CompletionItem) GetSortText() string`

GetSortText returns the SortText field if non-nil, zero value otherwise.

### GetSortTextOk

`func (o *CompletionItem) GetSortTextOk() (*string, bool)`

GetSortTextOk returns a tuple with the SortText field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSortText

`func (o *CompletionItem) SetSortText(v string)`

SetSortText sets SortText field to given value.

### HasSortText

`func (o *CompletionItem) HasSortText() bool`

HasSortText returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


