# RunnerMetadata

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Providers** | [**[]ProviderInfo**](ProviderInfo.md) |  | 
**RunnerId** | **string** |  | 
**RunningJobs** | Pointer to **int32** |  | [optional] 
**UpdatedAt** | **string** |  | 
**Uptime** | **int32** |  | 

## Methods

### NewRunnerMetadata

`func NewRunnerMetadata(providers []ProviderInfo, runnerId string, updatedAt string, uptime int32, ) *RunnerMetadata`

NewRunnerMetadata instantiates a new RunnerMetadata object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRunnerMetadataWithDefaults

`func NewRunnerMetadataWithDefaults() *RunnerMetadata`

NewRunnerMetadataWithDefaults instantiates a new RunnerMetadata object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetProviders

`func (o *RunnerMetadata) GetProviders() []ProviderInfo`

GetProviders returns the Providers field if non-nil, zero value otherwise.

### GetProvidersOk

`func (o *RunnerMetadata) GetProvidersOk() (*[]ProviderInfo, bool)`

GetProvidersOk returns a tuple with the Providers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProviders

`func (o *RunnerMetadata) SetProviders(v []ProviderInfo)`

SetProviders sets Providers field to given value.


### GetRunnerId

`func (o *RunnerMetadata) GetRunnerId() string`

GetRunnerId returns the RunnerId field if non-nil, zero value otherwise.

### GetRunnerIdOk

`func (o *RunnerMetadata) GetRunnerIdOk() (*string, bool)`

GetRunnerIdOk returns a tuple with the RunnerId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunnerId

`func (o *RunnerMetadata) SetRunnerId(v string)`

SetRunnerId sets RunnerId field to given value.


### GetRunningJobs

`func (o *RunnerMetadata) GetRunningJobs() int32`

GetRunningJobs returns the RunningJobs field if non-nil, zero value otherwise.

### GetRunningJobsOk

`func (o *RunnerMetadata) GetRunningJobsOk() (*int32, bool)`

GetRunningJobsOk returns a tuple with the RunningJobs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunningJobs

`func (o *RunnerMetadata) SetRunningJobs(v int32)`

SetRunningJobs sets RunningJobs field to given value.

### HasRunningJobs

`func (o *RunnerMetadata) HasRunningJobs() bool`

HasRunningJobs returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *RunnerMetadata) GetUpdatedAt() string`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *RunnerMetadata) GetUpdatedAtOk() (*string, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *RunnerMetadata) SetUpdatedAt(v string)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetUptime

`func (o *RunnerMetadata) GetUptime() int32`

GetUptime returns the Uptime field if non-nil, zero value otherwise.

### GetUptimeOk

`func (o *RunnerMetadata) GetUptimeOk() (*int32, bool)`

GetUptimeOk returns a tuple with the Uptime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUptime

`func (o *RunnerMetadata) SetUptime(v int32)`

SetUptime sets Uptime field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


