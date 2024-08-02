# ProjectBuildConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Devcontainer** | Pointer to [**DevcontainerConfig**](DevcontainerConfig.md) |  | [optional] 

## Methods

### NewProjectBuildConfig

`func NewProjectBuildConfig() *ProjectBuildConfig`

NewProjectBuildConfig instantiates a new ProjectBuildConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectBuildConfigWithDefaults

`func NewProjectBuildConfigWithDefaults() *ProjectBuildConfig`

NewProjectBuildConfigWithDefaults instantiates a new ProjectBuildConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDevcontainer

`func (o *ProjectBuildConfig) GetDevcontainer() DevcontainerConfig`

GetDevcontainer returns the Devcontainer field if non-nil, zero value otherwise.

### GetDevcontainerOk

`func (o *ProjectBuildConfig) GetDevcontainerOk() (*DevcontainerConfig, bool)`

GetDevcontainerOk returns a tuple with the Devcontainer field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDevcontainer

`func (o *ProjectBuildConfig) SetDevcontainer(v DevcontainerConfig)`

SetDevcontainer sets Devcontainer field to given value.

### HasDevcontainer

`func (o *ProjectBuildConfig) HasDevcontainer() bool`

HasDevcontainer returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


