# CreateBuildDTO

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Branch** | Pointer to **string** |  | [optional] 
**PrebuildId** | Pointer to **string** |  | [optional] 
**ProjectConfigName** | **string** |  | 

## Methods

### NewCreateBuildDTO

`func NewCreateBuildDTO(projectConfigName string, ) *CreateBuildDTO`

NewCreateBuildDTO instantiates a new CreateBuildDTO object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateBuildDTOWithDefaults

`func NewCreateBuildDTOWithDefaults() *CreateBuildDTO`

NewCreateBuildDTOWithDefaults instantiates a new CreateBuildDTO object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBranch

`func (o *CreateBuildDTO) GetBranch() string`

GetBranch returns the Branch field if non-nil, zero value otherwise.

### GetBranchOk

`func (o *CreateBuildDTO) GetBranchOk() (*string, bool)`

GetBranchOk returns a tuple with the Branch field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBranch

`func (o *CreateBuildDTO) SetBranch(v string)`

SetBranch sets Branch field to given value.

### HasBranch

`func (o *CreateBuildDTO) HasBranch() bool`

HasBranch returns a boolean if a field has been set.

### GetPrebuildId

`func (o *CreateBuildDTO) GetPrebuildId() string`

GetPrebuildId returns the PrebuildId field if non-nil, zero value otherwise.

### GetPrebuildIdOk

`func (o *CreateBuildDTO) GetPrebuildIdOk() (*string, bool)`

GetPrebuildIdOk returns a tuple with the PrebuildId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrebuildId

`func (o *CreateBuildDTO) SetPrebuildId(v string)`

SetPrebuildId sets PrebuildId field to given value.

### HasPrebuildId

`func (o *CreateBuildDTO) HasPrebuildId() bool`

HasPrebuildId returns a boolean if a field has been set.

### GetProjectConfigName

`func (o *CreateBuildDTO) GetProjectConfigName() string`

GetProjectConfigName returns the ProjectConfigName field if non-nil, zero value otherwise.

### GetProjectConfigNameOk

`func (o *CreateBuildDTO) GetProjectConfigNameOk() (*string, bool)`

GetProjectConfigNameOk returns a tuple with the ProjectConfigName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectConfigName

`func (o *CreateBuildDTO) SetProjectConfigName(v string)`

SetProjectConfigName sets ProjectConfigName field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


