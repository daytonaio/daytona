# PtyCreateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Cols** | Pointer to **int32** |  | [optional] 
**Cwd** | Pointer to **string** |  | [optional] 
**Envs** | Pointer to **map[string]string** |  | [optional] 
**Id** | Pointer to **string** |  | [optional] 
**LazyStart** | Pointer to **bool** | Don&#39;t start PTY until first client connects | [optional] 
**Rows** | Pointer to **int32** |  | [optional] 

## Methods

### NewPtyCreateRequest

`func NewPtyCreateRequest() *PtyCreateRequest`

NewPtyCreateRequest instantiates a new PtyCreateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPtyCreateRequestWithDefaults

`func NewPtyCreateRequestWithDefaults() *PtyCreateRequest`

NewPtyCreateRequestWithDefaults instantiates a new PtyCreateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCols

`func (o *PtyCreateRequest) GetCols() int32`

GetCols returns the Cols field if non-nil, zero value otherwise.

### GetColsOk

`func (o *PtyCreateRequest) GetColsOk() (*int32, bool)`

GetColsOk returns a tuple with the Cols field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCols

`func (o *PtyCreateRequest) SetCols(v int32)`

SetCols sets Cols field to given value.

### HasCols

`func (o *PtyCreateRequest) HasCols() bool`

HasCols returns a boolean if a field has been set.

### GetCwd

`func (o *PtyCreateRequest) GetCwd() string`

GetCwd returns the Cwd field if non-nil, zero value otherwise.

### GetCwdOk

`func (o *PtyCreateRequest) GetCwdOk() (*string, bool)`

GetCwdOk returns a tuple with the Cwd field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCwd

`func (o *PtyCreateRequest) SetCwd(v string)`

SetCwd sets Cwd field to given value.

### HasCwd

`func (o *PtyCreateRequest) HasCwd() bool`

HasCwd returns a boolean if a field has been set.

### GetEnvs

`func (o *PtyCreateRequest) GetEnvs() map[string]string`

GetEnvs returns the Envs field if non-nil, zero value otherwise.

### GetEnvsOk

`func (o *PtyCreateRequest) GetEnvsOk() (*map[string]string, bool)`

GetEnvsOk returns a tuple with the Envs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvs

`func (o *PtyCreateRequest) SetEnvs(v map[string]string)`

SetEnvs sets Envs field to given value.

### HasEnvs

`func (o *PtyCreateRequest) HasEnvs() bool`

HasEnvs returns a boolean if a field has been set.

### GetId

`func (o *PtyCreateRequest) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *PtyCreateRequest) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *PtyCreateRequest) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *PtyCreateRequest) HasId() bool`

HasId returns a boolean if a field has been set.

### GetLazyStart

`func (o *PtyCreateRequest) GetLazyStart() bool`

GetLazyStart returns the LazyStart field if non-nil, zero value otherwise.

### GetLazyStartOk

`func (o *PtyCreateRequest) GetLazyStartOk() (*bool, bool)`

GetLazyStartOk returns a tuple with the LazyStart field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLazyStart

`func (o *PtyCreateRequest) SetLazyStart(v bool)`

SetLazyStart sets LazyStart field to given value.

### HasLazyStart

`func (o *PtyCreateRequest) HasLazyStart() bool`

HasLazyStart returns a boolean if a field has been set.

### GetRows

`func (o *PtyCreateRequest) GetRows() int32`

GetRows returns the Rows field if non-nil, zero value otherwise.

### GetRowsOk

`func (o *PtyCreateRequest) GetRowsOk() (*int32, bool)`

GetRowsOk returns a tuple with the Rows field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRows

`func (o *PtyCreateRequest) SetRows(v int32)`

SetRows sets Rows field to given value.

### HasRows

`func (o *PtyCreateRequest) HasRows() bool`

HasRows returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


