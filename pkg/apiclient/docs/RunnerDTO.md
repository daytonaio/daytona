# RunnerDTO

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** |  | 
**Metadata** | Pointer to [**RunnerMetadata**](RunnerMetadata.md) |  | [optional] 
**Name** | **string** |  | 
**State** | **string** |  | 

## Methods

### NewRunnerDTO

`func NewRunnerDTO(id string, name string, state string, ) *RunnerDTO`

NewRunnerDTO instantiates a new RunnerDTO object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRunnerDTOWithDefaults

`func NewRunnerDTOWithDefaults() *RunnerDTO`

NewRunnerDTOWithDefaults instantiates a new RunnerDTO object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *RunnerDTO) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *RunnerDTO) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *RunnerDTO) SetId(v string)`

SetId sets Id field to given value.


### GetMetadata

`func (o *RunnerDTO) GetMetadata() RunnerMetadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *RunnerDTO) GetMetadataOk() (*RunnerMetadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *RunnerDTO) SetMetadata(v RunnerMetadata)`

SetMetadata sets Metadata field to given value.

### HasMetadata

`func (o *RunnerDTO) HasMetadata() bool`

HasMetadata returns a boolean if a field has been set.

### GetName

`func (o *RunnerDTO) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *RunnerDTO) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *RunnerDTO) SetName(v string)`

SetName sets Name field to given value.


### GetState

`func (o *RunnerDTO) GetState() string`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *RunnerDTO) GetStateOk() (*string, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *RunnerDTO) SetState(v string)`

SetState sets State field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


