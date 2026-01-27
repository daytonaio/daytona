# PtySessionInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Active** | Pointer to **bool** |  | [optional] 
**Cols** | Pointer to **int32** |  | [optional] 
**CreatedAt** | Pointer to **string** |  | [optional] 
**Cwd** | Pointer to **string** |  | [optional] 
**Envs** | Pointer to **map[string]string** |  | [optional] 
**Id** | Pointer to **string** |  | [optional] 
**LazyStart** | Pointer to **bool** | Whether this session uses lazy start | [optional] 
**Rows** | Pointer to **int32** |  | [optional] 

## Methods

### NewPtySessionInfo

`func NewPtySessionInfo() *PtySessionInfo`

NewPtySessionInfo instantiates a new PtySessionInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPtySessionInfoWithDefaults

`func NewPtySessionInfoWithDefaults() *PtySessionInfo`

NewPtySessionInfoWithDefaults instantiates a new PtySessionInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActive

`func (o *PtySessionInfo) GetActive() bool`

GetActive returns the Active field if non-nil, zero value otherwise.

### GetActiveOk

`func (o *PtySessionInfo) GetActiveOk() (*bool, bool)`

GetActiveOk returns a tuple with the Active field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActive

`func (o *PtySessionInfo) SetActive(v bool)`

SetActive sets Active field to given value.

### HasActive

`func (o *PtySessionInfo) HasActive() bool`

HasActive returns a boolean if a field has been set.

### GetCols

`func (o *PtySessionInfo) GetCols() int32`

GetCols returns the Cols field if non-nil, zero value otherwise.

### GetColsOk

`func (o *PtySessionInfo) GetColsOk() (*int32, bool)`

GetColsOk returns a tuple with the Cols field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCols

`func (o *PtySessionInfo) SetCols(v int32)`

SetCols sets Cols field to given value.

### HasCols

`func (o *PtySessionInfo) HasCols() bool`

HasCols returns a boolean if a field has been set.

### GetCreatedAt

`func (o *PtySessionInfo) GetCreatedAt() string`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *PtySessionInfo) GetCreatedAtOk() (*string, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *PtySessionInfo) SetCreatedAt(v string)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *PtySessionInfo) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetCwd

`func (o *PtySessionInfo) GetCwd() string`

GetCwd returns the Cwd field if non-nil, zero value otherwise.

### GetCwdOk

`func (o *PtySessionInfo) GetCwdOk() (*string, bool)`

GetCwdOk returns a tuple with the Cwd field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCwd

`func (o *PtySessionInfo) SetCwd(v string)`

SetCwd sets Cwd field to given value.

### HasCwd

`func (o *PtySessionInfo) HasCwd() bool`

HasCwd returns a boolean if a field has been set.

### GetEnvs

`func (o *PtySessionInfo) GetEnvs() map[string]string`

GetEnvs returns the Envs field if non-nil, zero value otherwise.

### GetEnvsOk

`func (o *PtySessionInfo) GetEnvsOk() (*map[string]string, bool)`

GetEnvsOk returns a tuple with the Envs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvs

`func (o *PtySessionInfo) SetEnvs(v map[string]string)`

SetEnvs sets Envs field to given value.

### HasEnvs

`func (o *PtySessionInfo) HasEnvs() bool`

HasEnvs returns a boolean if a field has been set.

### GetId

`func (o *PtySessionInfo) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *PtySessionInfo) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *PtySessionInfo) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *PtySessionInfo) HasId() bool`

HasId returns a boolean if a field has been set.

### GetLazyStart

`func (o *PtySessionInfo) GetLazyStart() bool`

GetLazyStart returns the LazyStart field if non-nil, zero value otherwise.

### GetLazyStartOk

`func (o *PtySessionInfo) GetLazyStartOk() (*bool, bool)`

GetLazyStartOk returns a tuple with the LazyStart field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLazyStart

`func (o *PtySessionInfo) SetLazyStart(v bool)`

SetLazyStart sets LazyStart field to given value.

### HasLazyStart

`func (o *PtySessionInfo) HasLazyStart() bool`

HasLazyStart returns a boolean if a field has been set.

### GetRows

`func (o *PtySessionInfo) GetRows() int32`

GetRows returns the Rows field if non-nil, zero value otherwise.

### GetRowsOk

`func (o *PtySessionInfo) GetRowsOk() (*int32, bool)`

GetRowsOk returns a tuple with the Rows field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRows

`func (o *PtySessionInfo) SetRows(v int32)`

SetRows sets Rows field to given value.

### HasRows

`func (o *PtySessionInfo) HasRows() bool`

HasRows returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


