# GitRepository

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Branch** | **string** |  | 
**CloneTarget** | Pointer to [**CloneTarget**](CloneTarget.md) |  | [optional] 
**Id** | **string** |  | 
**Name** | **string** |  | 
**Owner** | **string** |  | 
**Path** | Pointer to **string** |  | [optional] 
**PrNumber** | Pointer to **int32** |  | [optional] 
**Sha** | **string** |  | 
**Source** | **string** |  | 
**Url** | **string** |  | 

## Methods

### NewGitRepository

`func NewGitRepository(branch string, id string, name string, owner string, sha string, source string, url string, ) *GitRepository`

NewGitRepository instantiates a new GitRepository object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGitRepositoryWithDefaults

`func NewGitRepositoryWithDefaults() *GitRepository`

NewGitRepositoryWithDefaults instantiates a new GitRepository object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBranch

`func (o *GitRepository) GetBranch() string`

GetBranch returns the Branch field if non-nil, zero value otherwise.

### GetBranchOk

`func (o *GitRepository) GetBranchOk() (*string, bool)`

GetBranchOk returns a tuple with the Branch field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBranch

`func (o *GitRepository) SetBranch(v string)`

SetBranch sets Branch field to given value.


### GetCloneTarget

`func (o *GitRepository) GetCloneTarget() CloneTarget`

GetCloneTarget returns the CloneTarget field if non-nil, zero value otherwise.

### GetCloneTargetOk

`func (o *GitRepository) GetCloneTargetOk() (*CloneTarget, bool)`

GetCloneTargetOk returns a tuple with the CloneTarget field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCloneTarget

`func (o *GitRepository) SetCloneTarget(v CloneTarget)`

SetCloneTarget sets CloneTarget field to given value.

### HasCloneTarget

`func (o *GitRepository) HasCloneTarget() bool`

HasCloneTarget returns a boolean if a field has been set.

### GetId

`func (o *GitRepository) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *GitRepository) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *GitRepository) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *GitRepository) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *GitRepository) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *GitRepository) SetName(v string)`

SetName sets Name field to given value.


### GetOwner

`func (o *GitRepository) GetOwner() string`

GetOwner returns the Owner field if non-nil, zero value otherwise.

### GetOwnerOk

`func (o *GitRepository) GetOwnerOk() (*string, bool)`

GetOwnerOk returns a tuple with the Owner field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOwner

`func (o *GitRepository) SetOwner(v string)`

SetOwner sets Owner field to given value.


### GetPath

`func (o *GitRepository) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *GitRepository) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *GitRepository) SetPath(v string)`

SetPath sets Path field to given value.

### HasPath

`func (o *GitRepository) HasPath() bool`

HasPath returns a boolean if a field has been set.

### GetPrNumber

`func (o *GitRepository) GetPrNumber() int32`

GetPrNumber returns the PrNumber field if non-nil, zero value otherwise.

### GetPrNumberOk

`func (o *GitRepository) GetPrNumberOk() (*int32, bool)`

GetPrNumberOk returns a tuple with the PrNumber field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrNumber

`func (o *GitRepository) SetPrNumber(v int32)`

SetPrNumber sets PrNumber field to given value.

### HasPrNumber

`func (o *GitRepository) HasPrNumber() bool`

HasPrNumber returns a boolean if a field has been set.

### GetSha

`func (o *GitRepository) GetSha() string`

GetSha returns the Sha field if non-nil, zero value otherwise.

### GetShaOk

`func (o *GitRepository) GetShaOk() (*string, bool)`

GetShaOk returns a tuple with the Sha field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSha

`func (o *GitRepository) SetSha(v string)`

SetSha sets Sha field to given value.


### GetSource

`func (o *GitRepository) GetSource() string`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *GitRepository) GetSourceOk() (*string, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *GitRepository) SetSource(v string)`

SetSource sets Source field to given value.


### GetUrl

`func (o *GitRepository) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *GitRepository) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *GitRepository) SetUrl(v string)`

SetUrl sets Url field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


