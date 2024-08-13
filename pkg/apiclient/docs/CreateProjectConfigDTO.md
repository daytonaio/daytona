# CreateProjectConfigDTO

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BuildConfig** | Pointer to [**ProjectBuildConfig**](ProjectBuildConfig.md) |  | [optional] 
**EnvVars** | **map[string]string** |  | 
**Image** | Pointer to **string** |  | [optional] 
**Name** | **string** |  | 
**Source** | [**CreateProjectConfigSourceDTO**](CreateProjectConfigSourceDTO.md) |  | 
**User** | Pointer to **string** |  | [optional] 

## Methods

### NewCreateProjectConfigDTO

`func NewCreateProjectConfigDTO(envVars map[string]string, name string, source CreateProjectConfigSourceDTO, ) *CreateProjectConfigDTO`

NewCreateProjectConfigDTO instantiates a new CreateProjectConfigDTO object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateProjectConfigDTOWithDefaults

`func NewCreateProjectConfigDTOWithDefaults() *CreateProjectConfigDTO`

NewCreateProjectConfigDTOWithDefaults instantiates a new CreateProjectConfigDTO object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBuildConfig

`func (o *CreateProjectConfigDTO) GetBuildConfig() ProjectBuildConfig`

GetBuildConfig returns the BuildConfig field if non-nil, zero value otherwise.

### GetBuildConfigOk

`func (o *CreateProjectConfigDTO) GetBuildConfigOk() (*ProjectBuildConfig, bool)`

GetBuildConfigOk returns a tuple with the BuildConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBuildConfig

`func (o *CreateProjectConfigDTO) SetBuildConfig(v ProjectBuildConfig)`

SetBuildConfig sets BuildConfig field to given value.

### HasBuildConfig

`func (o *CreateProjectConfigDTO) HasBuildConfig() bool`

HasBuildConfig returns a boolean if a field has been set.

### GetEnvVars

`func (o *CreateProjectConfigDTO) GetEnvVars() map[string]string`

GetEnvVars returns the EnvVars field if non-nil, zero value otherwise.

### GetEnvVarsOk

`func (o *CreateProjectConfigDTO) GetEnvVarsOk() (*map[string]string, bool)`

GetEnvVarsOk returns a tuple with the EnvVars field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvVars

`func (o *CreateProjectConfigDTO) SetEnvVars(v map[string]string)`

SetEnvVars sets EnvVars field to given value.


### GetImage

`func (o *CreateProjectConfigDTO) GetImage() string`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *CreateProjectConfigDTO) GetImageOk() (*string, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *CreateProjectConfigDTO) SetImage(v string)`

SetImage sets Image field to given value.

### HasImage

`func (o *CreateProjectConfigDTO) HasImage() bool`

HasImage returns a boolean if a field has been set.

### GetName

`func (o *CreateProjectConfigDTO) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *CreateProjectConfigDTO) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *CreateProjectConfigDTO) SetName(v string)`

SetName sets Name field to given value.


### GetSource

`func (o *CreateProjectConfigDTO) GetSource() CreateProjectConfigSourceDTO`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *CreateProjectConfigDTO) GetSourceOk() (*CreateProjectConfigSourceDTO, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *CreateProjectConfigDTO) SetSource(v CreateProjectConfigSourceDTO)`

SetSource sets Source field to given value.


### GetUser

`func (o *CreateProjectConfigDTO) GetUser() string`

GetUser returns the User field if non-nil, zero value otherwise.

### GetUserOk

`func (o *CreateProjectConfigDTO) GetUserOk() (*string, bool)`

GetUserOk returns a tuple with the User field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUser

`func (o *CreateProjectConfigDTO) SetUser(v string)`

SetUser sets User field to given value.

### HasUser

`func (o *CreateProjectConfigDTO) HasUser() bool`

HasUser returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


