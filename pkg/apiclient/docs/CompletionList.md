# CompletionList

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**IsIncomplete** | **bool** |  | 
**Items** | [**[]CompletionItem**](CompletionItem.md) |  | 

## Methods

### NewCompletionList

`func NewCompletionList(isIncomplete bool, items []CompletionItem, ) *CompletionList`

NewCompletionList instantiates a new CompletionList object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCompletionListWithDefaults

`func NewCompletionListWithDefaults() *CompletionList`

NewCompletionListWithDefaults instantiates a new CompletionList object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetIsIncomplete

`func (o *CompletionList) GetIsIncomplete() bool`

GetIsIncomplete returns the IsIncomplete field if non-nil, zero value otherwise.

### GetIsIncompleteOk

`func (o *CompletionList) GetIsIncompleteOk() (*bool, bool)`

GetIsIncompleteOk returns a tuple with the IsIncomplete field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsIncomplete

`func (o *CompletionList) SetIsIncomplete(v bool)`

SetIsIncomplete sets IsIncomplete field to given value.


### GetItems

`func (o *CompletionList) GetItems() []CompletionItem`

GetItems returns the Items field if non-nil, zero value otherwise.

### GetItemsOk

`func (o *CompletionList) GetItemsOk() (*[]CompletionItem, bool)`

GetItemsOk returns a tuple with the Items field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetItems

`func (o *CompletionList) SetItems(v []CompletionItem)`

SetItems sets Items field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


