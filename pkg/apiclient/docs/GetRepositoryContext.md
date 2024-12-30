# GetRepositoryContext

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Branch** | Pointer to **string** |  | [optional] 
**Id** | Pointer to **string** |  | [optional] 
**Name** | Pointer to **string** |  | [optional] 
**Owner** | Pointer to **string** |  | [optional] 
**Path** | Pointer to **string** |  | [optional] 
**PrNumber** | Pointer to **int32** |  | [optional] 
**Sha** | Pointer to **string** |  | [optional] 
**Source** | Pointer to **string** |  | [optional] 
**Url** | **string** |  | 

## Methods

### NewGetRepositoryContext

`func NewGetRepositoryContext(url string, ) *GetRepositoryContext`

NewGetRepositoryContext instantiates a new GetRepositoryContext object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetRepositoryContextWithDefaults

`func NewGetRepositoryContextWithDefaults() *GetRepositoryContext`

NewGetRepositoryContextWithDefaults instantiates a new GetRepositoryContext object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBranch

`func (o *GetRepositoryContext) GetBranch() string`

GetBranch returns the Branch field if non-nil, zero value otherwise.

### GetBranchOk

`func (o *GetRepositoryContext) GetBranchOk() (*string, bool)`

GetBranchOk returns a tuple with the Branch field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBranch

`func (o *GetRepositoryContext) SetBranch(v string)`

SetBranch sets Branch field to given value.

### HasBranch

`func (o *GetRepositoryContext) HasBranch() bool`

HasBranch returns a boolean if a field has been set.

### GetId

`func (o *GetRepositoryContext) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *GetRepositoryContext) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *GetRepositoryContext) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *GetRepositoryContext) HasId() bool`

HasId returns a boolean if a field has been set.

### GetName

`func (o *GetRepositoryContext) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *GetRepositoryContext) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *GetRepositoryContext) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *GetRepositoryContext) HasName() bool`

HasName returns a boolean if a field has been set.

### GetOwner

`func (o *GetRepositoryContext) GetOwner() string`

GetOwner returns the Owner field if non-nil, zero value otherwise.

### GetOwnerOk

`func (o *GetRepositoryContext) GetOwnerOk() (*string, bool)`

GetOwnerOk returns a tuple with the Owner field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOwner

`func (o *GetRepositoryContext) SetOwner(v string)`

SetOwner sets Owner field to given value.

### HasOwner

`func (o *GetRepositoryContext) HasOwner() bool`

HasOwner returns a boolean if a field has been set.

### GetPath

`func (o *GetRepositoryContext) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *GetRepositoryContext) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *GetRepositoryContext) SetPath(v string)`

SetPath sets Path field to given value.

### HasPath

`func (o *GetRepositoryContext) HasPath() bool`

HasPath returns a boolean if a field has been set.

### GetPrNumber

`func (o *GetRepositoryContext) GetPrNumber() int32`

GetPrNumber returns the PrNumber field if non-nil, zero value otherwise.

### GetPrNumberOk

`func (o *GetRepositoryContext) GetPrNumberOk() (*int32, bool)`

GetPrNumberOk returns a tuple with the PrNumber field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrNumber

`func (o *GetRepositoryContext) SetPrNumber(v int32)`

SetPrNumber sets PrNumber field to given value.

### HasPrNumber

`func (o *GetRepositoryContext) HasPrNumber() bool`

HasPrNumber returns a boolean if a field has been set.

### GetSha

`func (o *GetRepositoryContext) GetSha() string`

GetSha returns the Sha field if non-nil, zero value otherwise.

### GetShaOk

`func (o *GetRepositoryContext) GetShaOk() (*string, bool)`

GetShaOk returns a tuple with the Sha field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSha

`func (o *GetRepositoryContext) SetSha(v string)`

SetSha sets Sha field to given value.

### HasSha

`func (o *GetRepositoryContext) HasSha() bool`

HasSha returns a boolean if a field has been set.

### GetSource

`func (o *GetRepositoryContext) GetSource() string`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *GetRepositoryContext) GetSourceOk() (*string, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *GetRepositoryContext) SetSource(v string)`

SetSource sets Source field to given value.

### HasSource

`func (o *GetRepositoryContext) HasSource() bool`

HasSource returns a boolean if a field has been set.

### GetUrl

`func (o *GetRepositoryContext) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *GetRepositoryContext) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *GetRepositoryContext) SetUrl(v string)`

SetUrl sets Url field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


