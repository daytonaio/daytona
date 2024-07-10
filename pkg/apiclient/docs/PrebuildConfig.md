# PrebuildConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Branch** | Pointer to **string** | Branch to watch for changes | [optional] 
**CommitInterval** | Pointer to **int32** | Number of commits between each new prebuild | [optional] 
**Key** | Pointer to **string** | Composite key (project-config-name+branch-name) for the prebuild | [optional] 
**ProjectConfig** | Pointer to [**ProjectConfig**](ProjectConfig.md) | Project configuration | [optional] 
**TriggerFiles** | Pointer to **[]string** | Files that should trigger a new prebuild if changed | [optional] 

## Methods

### NewPrebuildConfig

`func NewPrebuildConfig() *PrebuildConfig`

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

### HasBranch

`func (o *PrebuildConfig) HasBranch() bool`

HasBranch returns a boolean if a field has been set.

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

### HasCommitInterval

`func (o *PrebuildConfig) HasCommitInterval() bool`

HasCommitInterval returns a boolean if a field has been set.

### GetKey

`func (o *PrebuildConfig) GetKey() string`

GetKey returns the Key field if non-nil, zero value otherwise.

### GetKeyOk

`func (o *PrebuildConfig) GetKeyOk() (*string, bool)`

GetKeyOk returns a tuple with the Key field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKey

`func (o *PrebuildConfig) SetKey(v string)`

SetKey sets Key field to given value.

### HasKey

`func (o *PrebuildConfig) HasKey() bool`

HasKey returns a boolean if a field has been set.

### GetProjectConfig

`func (o *PrebuildConfig) GetProjectConfig() ProjectConfig`

GetProjectConfig returns the ProjectConfig field if non-nil, zero value otherwise.

### GetProjectConfigOk

`func (o *PrebuildConfig) GetProjectConfigOk() (*ProjectConfig, bool)`

GetProjectConfigOk returns a tuple with the ProjectConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectConfig

`func (o *PrebuildConfig) SetProjectConfig(v ProjectConfig)`

SetProjectConfig sets ProjectConfig field to given value.

### HasProjectConfig

`func (o *PrebuildConfig) HasProjectConfig() bool`

HasProjectConfig returns a boolean if a field has been set.

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

### HasTriggerFiles

`func (o *PrebuildConfig) HasTriggerFiles() bool`

HasTriggerFiles returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


