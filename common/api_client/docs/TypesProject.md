# TypesProject

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AuthKey** | Pointer to **string** |  | [optional] 
**Name** | Pointer to **string** |  | [optional] 
**Repository** | Pointer to [**TypesRepository**](TypesRepository.md) |  | [optional] 
**WorkspaceId** | Pointer to **string** |  | [optional] 

## Methods

### NewTypesProject

`func NewTypesProject() *TypesProject`

NewTypesProject instantiates a new TypesProject object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTypesProjectWithDefaults

`func NewTypesProjectWithDefaults() *TypesProject`

NewTypesProjectWithDefaults instantiates a new TypesProject object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAuthKey

`func (o *TypesProject) GetAuthKey() string`

GetAuthKey returns the AuthKey field if non-nil, zero value otherwise.

### GetAuthKeyOk

`func (o *TypesProject) GetAuthKeyOk() (*string, bool)`

GetAuthKeyOk returns a tuple with the AuthKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthKey

`func (o *TypesProject) SetAuthKey(v string)`

SetAuthKey sets AuthKey field to given value.

### HasAuthKey

`func (o *TypesProject) HasAuthKey() bool`

HasAuthKey returns a boolean if a field has been set.

### GetName

`func (o *TypesProject) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *TypesProject) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *TypesProject) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *TypesProject) HasName() bool`

HasName returns a boolean if a field has been set.

### GetRepository

`func (o *TypesProject) GetRepository() TypesRepository`

GetRepository returns the Repository field if non-nil, zero value otherwise.

### GetRepositoryOk

`func (o *TypesProject) GetRepositoryOk() (*TypesRepository, bool)`

GetRepositoryOk returns a tuple with the Repository field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRepository

`func (o *TypesProject) SetRepository(v TypesRepository)`

SetRepository sets Repository field to given value.

### HasRepository

`func (o *TypesProject) HasRepository() bool`

HasRepository returns a boolean if a field has been set.

### GetWorkspaceId

`func (o *TypesProject) GetWorkspaceId() string`

GetWorkspaceId returns the WorkspaceId field if non-nil, zero value otherwise.

### GetWorkspaceIdOk

`func (o *TypesProject) GetWorkspaceIdOk() (*string, bool)`

GetWorkspaceIdOk returns a tuple with the WorkspaceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkspaceId

`func (o *TypesProject) SetWorkspaceId(v string)`

SetWorkspaceId sets WorkspaceId field to given value.

### HasWorkspaceId

`func (o *TypesProject) HasWorkspaceId() bool`

HasWorkspaceId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


