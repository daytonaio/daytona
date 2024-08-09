# Build

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Hash** | Pointer to **string** |  | [optional] 
**Id** | Pointer to **string** |  | [optional] 
**ProjectConfig** | Pointer to [**ProjectConfig**](ProjectConfig.md) |  | [optional] 
**State** | Pointer to [**BuildBuildState**](BuildBuildState.md) |  | [optional] 

## Methods

### NewBuild

`func NewBuild() *Build`

NewBuild instantiates a new Build object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBuildWithDefaults

`func NewBuildWithDefaults() *Build`

NewBuildWithDefaults instantiates a new Build object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetHash

`func (o *Build) GetHash() string`

GetHash returns the Hash field if non-nil, zero value otherwise.

### GetHashOk

`func (o *Build) GetHashOk() (*string, bool)`

GetHashOk returns a tuple with the Hash field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHash

`func (o *Build) SetHash(v string)`

SetHash sets Hash field to given value.

### HasHash

`func (o *Build) HasHash() bool`

HasHash returns a boolean if a field has been set.

### GetId

`func (o *Build) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Build) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Build) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *Build) HasId() bool`

HasId returns a boolean if a field has been set.

### GetProjectConfig

`func (o *Build) GetProjectConfig() ProjectConfig`

GetProjectConfig returns the ProjectConfig field if non-nil, zero value otherwise.

### GetProjectConfigOk

`func (o *Build) GetProjectConfigOk() (*ProjectConfig, bool)`

GetProjectConfigOk returns a tuple with the ProjectConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectConfig

`func (o *Build) SetProjectConfig(v ProjectConfig)`

SetProjectConfig sets ProjectConfig field to given value.

### HasProjectConfig

`func (o *Build) HasProjectConfig() bool`

HasProjectConfig returns a boolean if a field has been set.

### GetState

`func (o *Build) GetState() BuildBuildState`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *Build) GetStateOk() (*BuildBuildState, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *Build) SetState(v BuildBuildState)`

SetState sets State field to given value.

### HasState

`func (o *Build) HasState() bool`

HasState returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


