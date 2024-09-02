# BuildConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CachedBuild** | Pointer to [**CachedBuild**](CachedBuild.md) |  | [optional] 
**Devcontainer** | Pointer to [**DevcontainerConfig**](DevcontainerConfig.md) |  | [optional] 

## Methods

### NewBuildConfig

`func NewBuildConfig() *BuildConfig`

NewBuildConfig instantiates a new BuildConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBuildConfigWithDefaults

`func NewBuildConfigWithDefaults() *BuildConfig`

NewBuildConfigWithDefaults instantiates a new BuildConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCachedBuild

`func (o *BuildConfig) GetCachedBuild() CachedBuild`

GetCachedBuild returns the CachedBuild field if non-nil, zero value otherwise.

### GetCachedBuildOk

`func (o *BuildConfig) GetCachedBuildOk() (*CachedBuild, bool)`

GetCachedBuildOk returns a tuple with the CachedBuild field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCachedBuild

`func (o *BuildConfig) SetCachedBuild(v CachedBuild)`

SetCachedBuild sets CachedBuild field to given value.

### HasCachedBuild

`func (o *BuildConfig) HasCachedBuild() bool`

HasCachedBuild returns a boolean if a field has been set.

### GetDevcontainer

`func (o *BuildConfig) GetDevcontainer() DevcontainerConfig`

GetDevcontainer returns the Devcontainer field if non-nil, zero value otherwise.

### GetDevcontainerOk

`func (o *BuildConfig) GetDevcontainerOk() (*DevcontainerConfig, bool)`

GetDevcontainerOk returns a tuple with the Devcontainer field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDevcontainer

`func (o *BuildConfig) SetDevcontainer(v DevcontainerConfig)`

SetDevcontainer sets Devcontainer field to given value.

### HasDevcontainer

`func (o *BuildConfig) HasDevcontainer() bool`

HasDevcontainer returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


