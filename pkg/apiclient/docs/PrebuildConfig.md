# PrebuildConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Branch** | **string** |  | 
**CommitInterval** | **int32** |  | 
**Id** | **string** |  | 
**Retention** | **int32** |  | 
**TriggerFiles** | **[]string** |  | 

## Methods

### NewPrebuildConfig

`func NewPrebuildConfig(branch string, commitInterval int32, id string, retention int32, triggerFiles []string, ) *PrebuildConfig`

NewPrebuildConfig instantiates a new PrebuildConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPrebuildConfigWithDefaults

`func NewPrebuildConfigWithDefaults() *PrebuildConfig`

NewPrebuildConfigWithDefaults instantiates a new PrebuildConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBranch

`func (o *PrebuildConfig) GetBranch() string`

GetBranch returns the Branch field if non-nil, zero value otherwise.

### GetBranchOk

`func (o *PrebuildConfig) GetBranchOk() (*string, bool)`

GetBranchOk returns a tuple with the Branch field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBranch

`func (o *PrebuildConfig) SetBranch(v string)`

SetBranch sets Branch field to given value.


### GetCommitInterval

`func (o *PrebuildConfig) GetCommitInterval() int32`

GetCommitInterval returns the CommitInterval field if non-nil, zero value otherwise.

### GetCommitIntervalOk

`func (o *PrebuildConfig) GetCommitIntervalOk() (*int32, bool)`

GetCommitIntervalOk returns a tuple with the CommitInterval field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommitInterval

`func (o *PrebuildConfig) SetCommitInterval(v int32)`

SetCommitInterval sets CommitInterval field to given value.


### GetId

`func (o *PrebuildConfig) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *PrebuildConfig) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *PrebuildConfig) SetId(v string)`

SetId sets Id field to given value.


### GetRetention

`func (o *PrebuildConfig) GetRetention() int32`

GetRetention returns the Retention field if non-nil, zero value otherwise.

### GetRetentionOk

`func (o *PrebuildConfig) GetRetentionOk() (*int32, bool)`

GetRetentionOk returns a tuple with the Retention field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRetention

`func (o *PrebuildConfig) SetRetention(v int32)`

SetRetention sets Retention field to given value.


### GetTriggerFiles

`func (o *PrebuildConfig) GetTriggerFiles() []string`

GetTriggerFiles returns the TriggerFiles field if non-nil, zero value otherwise.

### GetTriggerFilesOk

`func (o *PrebuildConfig) GetTriggerFilesOk() (*[]string, bool)`

GetTriggerFilesOk returns a tuple with the TriggerFiles field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTriggerFiles

`func (o *PrebuildConfig) SetTriggerFiles(v []string)`

SetTriggerFiles sets TriggerFiles field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


