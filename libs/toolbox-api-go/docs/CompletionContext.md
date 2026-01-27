# CompletionContext

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**TriggerCharacter** | Pointer to **string** |  | [optional] 
**TriggerKind** | **int32** |  | 

## Methods

### NewCompletionContext

`func NewCompletionContext(triggerKind int32, ) *CompletionContext`

NewCompletionContext instantiates a new CompletionContext object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCompletionContextWithDefaults

`func NewCompletionContextWithDefaults() *CompletionContext`

NewCompletionContextWithDefaults instantiates a new CompletionContext object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTriggerCharacter

`func (o *CompletionContext) GetTriggerCharacter() string`

GetTriggerCharacter returns the TriggerCharacter field if non-nil, zero value otherwise.

### GetTriggerCharacterOk

`func (o *CompletionContext) GetTriggerCharacterOk() (*string, bool)`

GetTriggerCharacterOk returns a tuple with the TriggerCharacter field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTriggerCharacter

`func (o *CompletionContext) SetTriggerCharacter(v string)`

SetTriggerCharacter sets TriggerCharacter field to given value.

### HasTriggerCharacter

`func (o *CompletionContext) HasTriggerCharacter() bool`

HasTriggerCharacter returns a boolean if a field has been set.

### GetTriggerKind

`func (o *CompletionContext) GetTriggerKind() int32`

GetTriggerKind returns the TriggerKind field if non-nil, zero value otherwise.

### GetTriggerKindOk

`func (o *CompletionContext) GetTriggerKindOk() (*int32, bool)`

GetTriggerKindOk returns a tuple with the TriggerKind field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTriggerKind

`func (o *CompletionContext) SetTriggerKind(v int32)`

SetTriggerKind sets TriggerKind field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


