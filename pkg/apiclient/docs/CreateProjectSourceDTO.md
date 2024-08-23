# CreateProjectSourceDTO

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Repository** | [**GitRepository**](GitRepository.md) |  | 

## Methods

### NewCreateProjectSourceDTO

`func NewCreateProjectSourceDTO(repository GitRepository, ) *CreateProjectSourceDTO`

NewCreateProjectSourceDTO instantiates a new CreateProjectSourceDTO object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateProjectSourceDTOWithDefaults

`func NewCreateProjectSourceDTOWithDefaults() *CreateProjectSourceDTO`

NewCreateProjectSourceDTOWithDefaults instantiates a new CreateProjectSourceDTO object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRepository

`func (o *CreateProjectSourceDTO) GetRepository() GitRepository`

GetRepository returns the Repository field if non-nil, zero value otherwise.

### GetRepositoryOk

`func (o *CreateProjectSourceDTO) GetRepositoryOk() (*GitRepository, bool)`

GetRepositoryOk returns a tuple with the Repository field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRepository

`func (o *CreateProjectSourceDTO) SetRepository(v GitRepository)`

SetRepository sets Repository field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


