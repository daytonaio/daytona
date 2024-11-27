# Job

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Action** | [**ModelsJobAction**](ModelsJobAction.md) |  | 
**CreatedAt** | **string** |  | 
**Error** | Pointer to **string** |  | [optional] 
**Id** | **string** |  | 
**ResourceId** | **string** |  | 
**ResourceType** | [**ResourceType**](ResourceType.md) |  | 
**State** | [**JobState**](JobState.md) |  | 
**UpdatedAt** | **string** |  | 

## Methods

### NewJob

`func NewJob(action ModelsJobAction, createdAt string, id string, resourceId string, resourceType ResourceType, state JobState, updatedAt string, ) *Job`

NewJob instantiates a new Job object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewJobWithDefaults

`func NewJobWithDefaults() *Job`

NewJobWithDefaults instantiates a new Job object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAction

`func (o *Job) GetAction() ModelsJobAction`

GetAction returns the Action field if non-nil, zero value otherwise.

### GetActionOk

`func (o *Job) GetActionOk() (*ModelsJobAction, bool)`

GetActionOk returns a tuple with the Action field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAction

`func (o *Job) SetAction(v ModelsJobAction)`

SetAction sets Action field to given value.


### GetCreatedAt

`func (o *Job) GetCreatedAt() string`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Job) GetCreatedAtOk() (*string, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Job) SetCreatedAt(v string)`

SetCreatedAt sets CreatedAt field to given value.


### GetError

`func (o *Job) GetError() string`

GetError returns the Error field if non-nil, zero value otherwise.

### GetErrorOk

`func (o *Job) GetErrorOk() (*string, bool)`

GetErrorOk returns a tuple with the Error field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetError

`func (o *Job) SetError(v string)`

SetError sets Error field to given value.

### HasError

`func (o *Job) HasError() bool`

HasError returns a boolean if a field has been set.

### GetId

`func (o *Job) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Job) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Job) SetId(v string)`

SetId sets Id field to given value.


### GetResourceId

`func (o *Job) GetResourceId() string`

GetResourceId returns the ResourceId field if non-nil, zero value otherwise.

### GetResourceIdOk

`func (o *Job) GetResourceIdOk() (*string, bool)`

GetResourceIdOk returns a tuple with the ResourceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceId

`func (o *Job) SetResourceId(v string)`

SetResourceId sets ResourceId field to given value.


### GetResourceType

`func (o *Job) GetResourceType() ResourceType`

GetResourceType returns the ResourceType field if non-nil, zero value otherwise.

### GetResourceTypeOk

`func (o *Job) GetResourceTypeOk() (*ResourceType, bool)`

GetResourceTypeOk returns a tuple with the ResourceType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceType

`func (o *Job) SetResourceType(v ResourceType)`

SetResourceType sets ResourceType field to given value.


### GetState

`func (o *Job) GetState() JobState`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *Job) GetStateOk() (*JobState, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *Job) SetState(v JobState)`

SetState sets State field to given value.


### GetUpdatedAt

`func (o *Job) GetUpdatedAt() string`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *Job) GetUpdatedAtOk() (*string, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *Job) SetUpdatedAt(v string)`

SetUpdatedAt sets UpdatedAt field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


