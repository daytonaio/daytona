# CreatePrebuildDTO

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Branch** | Pointer to **string** |  | [optional] 
**CommitInterval** | Pointer to **int32** |  | [optional] 
**Id** | Pointer to **string** |  | [optional] 
**Retention** | **int32** |  | 
**TriggerFiles** | Pointer to **[]string** |  | [optional] 

## Methods

### NewCreatePrebuildDTO

`func NewCreatePrebuildDTO(retention int32, ) *CreatePrebuildDTO`

NewCreatePrebuildDTO instantiates a new CreatePrebuildDTO object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreatePrebuildDTOWithDefaults

`func NewCreatePrebuildDTOWithDefaults() *CreatePrebuildDTO`

NewCreatePrebuildDTOWithDefaults instantiates a new CreatePrebuildDTO object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBranch

`func (o *CreatePrebuildDTO) GetBranch() string`

GetBranch returns the Branch field if non-nil, zero value otherwise.

### GetBranchOk

`func (o *CreatePrebuildDTO) GetBranchOk() (*string, bool)`

GetBranchOk returns a tuple with the Branch field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBranch

`func (o *CreatePrebuildDTO) SetBranch(v string)`

SetBranch sets Branch field to given value.

### HasBranch

`func (o *CreatePrebuildDTO) HasBranch() bool`

HasBranch returns a boolean if a field has been set.

### GetCommitInterval

`func (o *CreatePrebuildDTO) GetCommitInterval() int32`

GetCommitInterval returns the CommitInterval field if non-nil, zero value otherwise.

### GetCommitIntervalOk

`func (o *CreatePrebuildDTO) GetCommitIntervalOk() (*int32, bool)`

GetCommitIntervalOk returns a tuple with the CommitInterval field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommitInterval

`func (o *CreatePrebuildDTO) SetCommitInterval(v int32)`

SetCommitInterval sets CommitInterval field to given value.

### HasCommitInterval

`func (o *CreatePrebuildDTO) HasCommitInterval() bool`

HasCommitInterval returns a boolean if a field has been set.

### GetId

`func (o *CreatePrebuildDTO) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *CreatePrebuildDTO) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *CreatePrebuildDTO) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *CreatePrebuildDTO) HasId() bool`

HasId returns a boolean if a field has been set.

### GetRetention

`func (o *CreatePrebuildDTO) GetRetention() int32`

GetRetention returns the Retention field if non-nil, zero value otherwise.

### GetRetentionOk

`func (o *CreatePrebuildDTO) GetRetentionOk() (*int32, bool)`

GetRetentionOk returns a tuple with the Retention field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRetention

`func (o *CreatePrebuildDTO) SetRetention(v int32)`

SetRetention sets Retention field to given value.


### GetTriggerFiles

`func (o *CreatePrebuildDTO) GetTriggerFiles() []string`

GetTriggerFiles returns the TriggerFiles field if non-nil, zero value otherwise.

### GetTriggerFilesOk

`func (o *CreatePrebuildDTO) GetTriggerFilesOk() (*[]string, bool)`

GetTriggerFilesOk returns a tuple with the TriggerFiles field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTriggerFiles

`func (o *CreatePrebuildDTO) SetTriggerFiles(v []string)`

SetTriggerFiles sets TriggerFiles field to given value.

### HasTriggerFiles

`func (o *CreatePrebuildDTO) HasTriggerFiles() bool`

HasTriggerFiles returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


