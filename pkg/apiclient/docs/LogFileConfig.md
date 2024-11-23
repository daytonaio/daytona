# LogFileConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Compress** | Pointer to **bool** |  | [optional] 
**LocalTime** | Pointer to **bool** |  | [optional] 
**MaxAge** | **int32** |  | 
**MaxBackups** | **int32** |  | 
**MaxSize** | **int32** |  | 
**Path** | **string** |  | 

## Methods

### NewLogFileConfig

`func NewLogFileConfig(maxAge int32, maxBackups int32, maxSize int32, path string, ) *LogFileConfig`

NewLogFileConfig instantiates a new LogFileConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewLogFileConfigWithDefaults

`func NewLogFileConfigWithDefaults() *LogFileConfig`

NewLogFileConfigWithDefaults instantiates a new LogFileConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCompress

`func (o *LogFileConfig) GetCompress() bool`

GetCompress returns the Compress field if non-nil, zero value otherwise.

### GetCompressOk

`func (o *LogFileConfig) GetCompressOk() (*bool, bool)`

GetCompressOk returns a tuple with the Compress field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompress

`func (o *LogFileConfig) SetCompress(v bool)`

SetCompress sets Compress field to given value.

### HasCompress

`func (o *LogFileConfig) HasCompress() bool`

HasCompress returns a boolean if a field has been set.

### GetLocalTime

`func (o *LogFileConfig) GetLocalTime() bool`

GetLocalTime returns the LocalTime field if non-nil, zero value otherwise.

### GetLocalTimeOk

`func (o *LogFileConfig) GetLocalTimeOk() (*bool, bool)`

GetLocalTimeOk returns a tuple with the LocalTime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLocalTime

`func (o *LogFileConfig) SetLocalTime(v bool)`

SetLocalTime sets LocalTime field to given value.

### HasLocalTime

`func (o *LogFileConfig) HasLocalTime() bool`

HasLocalTime returns a boolean if a field has been set.

### GetMaxAge

`func (o *LogFileConfig) GetMaxAge() int32`

GetMaxAge returns the MaxAge field if non-nil, zero value otherwise.

### GetMaxAgeOk

`func (o *LogFileConfig) GetMaxAgeOk() (*int32, bool)`

GetMaxAgeOk returns a tuple with the MaxAge field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxAge

`func (o *LogFileConfig) SetMaxAge(v int32)`

SetMaxAge sets MaxAge field to given value.


### GetMaxBackups

`func (o *LogFileConfig) GetMaxBackups() int32`

GetMaxBackups returns the MaxBackups field if non-nil, zero value otherwise.

### GetMaxBackupsOk

`func (o *LogFileConfig) GetMaxBackupsOk() (*int32, bool)`

GetMaxBackupsOk returns a tuple with the MaxBackups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxBackups

`func (o *LogFileConfig) SetMaxBackups(v int32)`

SetMaxBackups sets MaxBackups field to given value.


### GetMaxSize

`func (o *LogFileConfig) GetMaxSize() int32`

GetMaxSize returns the MaxSize field if non-nil, zero value otherwise.

### GetMaxSizeOk

`func (o *LogFileConfig) GetMaxSizeOk() (*int32, bool)`

GetMaxSizeOk returns a tuple with the MaxSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxSize

`func (o *LogFileConfig) SetMaxSize(v int32)`

SetMaxSize sets MaxSize field to given value.


### GetPath

`func (o *LogFileConfig) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *LogFileConfig) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *LogFileConfig) SetPath(v string)`

SetPath sets Path field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


