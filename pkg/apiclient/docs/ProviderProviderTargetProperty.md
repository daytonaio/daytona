# ProviderProviderTargetProperty

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DefaultValue** | Pointer to **string** | DefaultValue is converted into the appropriate type based on the Type If the property is a FilePath, the DefaultValue is a path to a directory | [optional] 
**Description** | Pointer to **string** | Brief description of the property | [optional] 
**DisabledPredicate** | Pointer to **string** | A regex string matched with the name of the target to determine if the property should be disabled If the regex matches the target name, the property will be disabled E.g. \&quot;^local$\&quot; will disable the property for the local target | [optional] 
**InputMasked** | Pointer to **bool** |  | [optional] 
**Options** | Pointer to **[]string** | Options is only used if the Type is ProviderTargetPropertyTypeOption | [optional] 
**Suggestions** | Pointer to **[]string** | Suggestions is an optional list of auto-complete values to assist the user while filling the field | [optional] 
**Type** | Pointer to [**ProviderProviderTargetPropertyType**](ProviderProviderTargetPropertyType.md) |  | [optional] 

## Methods

### NewProviderProviderTargetProperty

`func NewProviderProviderTargetProperty() *ProviderProviderTargetProperty`

NewProviderProviderTargetProperty instantiates a new ProviderProviderTargetProperty object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProviderProviderTargetPropertyWithDefaults

`func NewProviderProviderTargetPropertyWithDefaults() *ProviderProviderTargetProperty`

NewProviderProviderTargetPropertyWithDefaults instantiates a new ProviderProviderTargetProperty object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDefaultValue

`func (o *ProviderProviderTargetProperty) GetDefaultValue() string`

GetDefaultValue returns the DefaultValue field if non-nil, zero value otherwise.

### GetDefaultValueOk

`func (o *ProviderProviderTargetProperty) GetDefaultValueOk() (*string, bool)`

GetDefaultValueOk returns a tuple with the DefaultValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultValue

`func (o *ProviderProviderTargetProperty) SetDefaultValue(v string)`

SetDefaultValue sets DefaultValue field to given value.

### HasDefaultValue

`func (o *ProviderProviderTargetProperty) HasDefaultValue() bool`

HasDefaultValue returns a boolean if a field has been set.

### GetDescription

`func (o *ProviderProviderTargetProperty) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *ProviderProviderTargetProperty) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *ProviderProviderTargetProperty) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *ProviderProviderTargetProperty) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetDisabledPredicate

`func (o *ProviderProviderTargetProperty) GetDisabledPredicate() string`

GetDisabledPredicate returns the DisabledPredicate field if non-nil, zero value otherwise.

### GetDisabledPredicateOk

`func (o *ProviderProviderTargetProperty) GetDisabledPredicateOk() (*string, bool)`

GetDisabledPredicateOk returns a tuple with the DisabledPredicate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDisabledPredicate

`func (o *ProviderProviderTargetProperty) SetDisabledPredicate(v string)`

SetDisabledPredicate sets DisabledPredicate field to given value.

### HasDisabledPredicate

`func (o *ProviderProviderTargetProperty) HasDisabledPredicate() bool`

HasDisabledPredicate returns a boolean if a field has been set.

### GetInputMasked

`func (o *ProviderProviderTargetProperty) GetInputMasked() bool`

GetInputMasked returns the InputMasked field if non-nil, zero value otherwise.

### GetInputMaskedOk

`func (o *ProviderProviderTargetProperty) GetInputMaskedOk() (*bool, bool)`

GetInputMaskedOk returns a tuple with the InputMasked field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInputMasked

`func (o *ProviderProviderTargetProperty) SetInputMasked(v bool)`

SetInputMasked sets InputMasked field to given value.

### HasInputMasked

`func (o *ProviderProviderTargetProperty) HasInputMasked() bool`

HasInputMasked returns a boolean if a field has been set.

### GetOptions

`func (o *ProviderProviderTargetProperty) GetOptions() []string`

GetOptions returns the Options field if non-nil, zero value otherwise.

### GetOptionsOk

`func (o *ProviderProviderTargetProperty) GetOptionsOk() (*[]string, bool)`

GetOptionsOk returns a tuple with the Options field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOptions

`func (o *ProviderProviderTargetProperty) SetOptions(v []string)`

SetOptions sets Options field to given value.

### HasOptions

`func (o *ProviderProviderTargetProperty) HasOptions() bool`

HasOptions returns a boolean if a field has been set.

### GetSuggestions

`func (o *ProviderProviderTargetProperty) GetSuggestions() []string`

GetSuggestions returns the Suggestions field if non-nil, zero value otherwise.

### GetSuggestionsOk

`func (o *ProviderProviderTargetProperty) GetSuggestionsOk() (*[]string, bool)`

GetSuggestionsOk returns a tuple with the Suggestions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSuggestions

`func (o *ProviderProviderTargetProperty) SetSuggestions(v []string)`

SetSuggestions sets Suggestions field to given value.

### HasSuggestions

`func (o *ProviderProviderTargetProperty) HasSuggestions() bool`

HasSuggestions returns a boolean if a field has been set.

### GetType

`func (o *ProviderProviderTargetProperty) GetType() ProviderProviderTargetPropertyType`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *ProviderProviderTargetProperty) GetTypeOk() (*ProviderProviderTargetPropertyType, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *ProviderProviderTargetProperty) SetType(v ProviderProviderTargetPropertyType)`

SetType sets Type field to given value.

### HasType

`func (o *ProviderProviderTargetProperty) HasType() bool`

HasType returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


