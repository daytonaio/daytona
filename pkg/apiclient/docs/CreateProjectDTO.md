# CreateProjectDTO

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ExistingConfig** | Pointer to [**ExistingConfigDTO**](ExistingConfigDTO.md) |  | [optional] 
**NewConfig** | Pointer to [**CreateProjectConfigDTO**](CreateProjectConfigDTO.md) |  | [optional] 

## Methods

### NewCreateProjectDTO

`func NewCreateProjectDTO() *CreateProjectDTO`

NewCreateProjectDTO instantiates a new CreateProjectDTO object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateProjectDTOWithDefaults

`func NewCreateProjectDTOWithDefaults() *CreateProjectDTO`

NewCreateProjectDTOWithDefaults instantiates a new CreateProjectDTO object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetExistingConfig

`func (o *CreateProjectDTO) GetExistingConfig() ExistingConfigDTO`

GetExistingConfig returns the ExistingConfig field if non-nil, zero value otherwise.

### GetExistingConfigOk

`func (o *CreateProjectDTO) GetExistingConfigOk() (*ExistingConfigDTO, bool)`

GetExistingConfigOk returns a tuple with the ExistingConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExistingConfig

`func (o *CreateProjectDTO) SetExistingConfig(v ExistingConfigDTO)`

SetExistingConfig sets ExistingConfig field to given value.

### HasExistingConfig

`func (o *CreateProjectDTO) HasExistingConfig() bool`

HasExistingConfig returns a boolean if a field has been set.

### GetNewConfig

`func (o *CreateProjectDTO) GetNewConfig() CreateProjectConfigDTO`

GetNewConfig returns the NewConfig field if non-nil, zero value otherwise.

### GetNewConfigOk

`func (o *CreateProjectDTO) GetNewConfigOk() (*CreateProjectConfigDTO, bool)`

GetNewConfigOk returns a tuple with the NewConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNewConfig

`func (o *CreateProjectDTO) SetNewConfig(v CreateProjectConfigDTO)`

SetNewConfig sets NewConfig field to given value.

### HasNewConfig

`func (o *CreateProjectDTO) HasNewConfig() bool`

HasNewConfig returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


