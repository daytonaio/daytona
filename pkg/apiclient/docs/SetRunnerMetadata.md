# SetRunnerMetadata

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Providers** | [**[]ProviderInfo**](ProviderInfo.md) |  | 
**RunningJobs** | Pointer to **int32** |  | [optional] 
**Uptime** | **int32** |  | 

## Methods

### NewSetRunnerMetadata

`func NewSetRunnerMetadata(providers []ProviderInfo, uptime int32, ) *SetRunnerMetadata`

NewSetRunnerMetadata instantiates a new SetRunnerMetadata object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSetRunnerMetadataWithDefaults

`func NewSetRunnerMetadataWithDefaults() *SetRunnerMetadata`

NewSetRunnerMetadataWithDefaults instantiates a new SetRunnerMetadata object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetProviders

`func (o *SetRunnerMetadata) GetProviders() []ProviderInfo`

GetProviders returns the Providers field if non-nil, zero value otherwise.

### GetProvidersOk

`func (o *SetRunnerMetadata) GetProvidersOk() (*[]ProviderInfo, bool)`

GetProvidersOk returns a tuple with the Providers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProviders

`func (o *SetRunnerMetadata) SetProviders(v []ProviderInfo)`

SetProviders sets Providers field to given value.


### GetRunningJobs

`func (o *SetRunnerMetadata) GetRunningJobs() int32`

GetRunningJobs returns the RunningJobs field if non-nil, zero value otherwise.

### GetRunningJobsOk

`func (o *SetRunnerMetadata) GetRunningJobsOk() (*int32, bool)`

GetRunningJobsOk returns a tuple with the RunningJobs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunningJobs

`func (o *SetRunnerMetadata) SetRunningJobs(v int32)`

SetRunningJobs sets RunningJobs field to given value.

### HasRunningJobs

`func (o *SetRunnerMetadata) HasRunningJobs() bool`

HasRunningJobs returns a boolean if a field has been set.

### GetUptime

`func (o *SetRunnerMetadata) GetUptime() int32`

GetUptime returns the Uptime field if non-nil, zero value otherwise.

### GetUptimeOk

`func (o *SetRunnerMetadata) GetUptimeOk() (*int32, bool)`

GetUptimeOk returns a tuple with the Uptime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUptime

`func (o *SetRunnerMetadata) SetUptime(v int32)`

SetUptime sets Uptime field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


