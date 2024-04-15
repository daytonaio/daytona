# CreateWorkspaceRequestProjectSource

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Repository** | Pointer to [**GitRepository**](GitRepository.md) |  | [optional] 

## Methods

### NewCreateWorkspaceRequestProjectSource

`func NewCreateWorkspaceRequestProjectSource() *CreateWorkspaceRequestProjectSource`

NewCreateWorkspaceRequestProjectSource instantiates a new CreateWorkspaceRequestProjectSource object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateWorkspaceRequestProjectSourceWithDefaults

`func NewCreateWorkspaceRequestProjectSourceWithDefaults() *CreateWorkspaceRequestProjectSource`

NewCreateWorkspaceRequestProjectSourceWithDefaults instantiates a new CreateWorkspaceRequestProjectSource object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRepository

`func (o *CreateWorkspaceRequestProjectSource) GetRepository() GitRepository`

GetRepository returns the Repository field if non-nil, zero value otherwise.

### GetRepositoryOk

`func (o *CreateWorkspaceRequestProjectSource) GetRepositoryOk() (*GitRepository, bool)`

GetRepositoryOk returns a tuple with the Repository field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRepository

`func (o *CreateWorkspaceRequestProjectSource) SetRepository(v GitRepository)`

SetRepository sets Repository field to given value.

### HasRepository

`func (o *CreateWorkspaceRequestProjectSource) HasRepository() bool`

HasRepository returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


