# CreateProjectDTO

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BuildConfig** | Pointer to [**BuildConfig**](BuildConfig.md) |  | [optional] 
**EnvVars** | **map[string]string** |  | 
**GitProviderConfigId** | Pointer to **string** |  | [optional] 
**Image** | Pointer to **string** |  | [optional] 
**Name** | **string** |  | 
**Source** | [**CreateProjectSourceDTO**](CreateProjectSourceDTO.md) |  | 
**User** | Pointer to **string** |  | [optional] 

## Methods

### NewCreateProjectDTO

`func NewCreateProjectDTO(envVars map[string]string, name string, source CreateProjectSourceDTO, ) *CreateProjectDTO`

NewCreateProjectDTO instantiates a new CreateProjectDTO object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateProjectDTOWithDefaults

`func NewCreateProjectDTOWithDefaults() *CreateProjectDTO`

NewCreateProjectDTOWithDefaults instantiates a new CreateProjectDTO object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBuildConfig

`func (o *CreateProjectDTO) GetBuildConfig() BuildConfig`

GetBuildConfig returns the BuildConfig field if non-nil, zero value otherwise.

### GetBuildConfigOk

`func (o *CreateProjectDTO) GetBuildConfigOk() (*BuildConfig, bool)`

GetBuildConfigOk returns a tuple with the BuildConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBuildConfig

`func (o *CreateProjectDTO) SetBuildConfig(v BuildConfig)`

SetBuildConfig sets BuildConfig field to given value.

### HasBuildConfig

`func (o *CreateProjectDTO) HasBuildConfig() bool`

HasBuildConfig returns a boolean if a field has been set.

### GetEnvVars

`func (o *CreateProjectDTO) GetEnvVars() map[string]string`

GetEnvVars returns the EnvVars field if non-nil, zero value otherwise.

### GetEnvVarsOk

`func (o *CreateProjectDTO) GetEnvVarsOk() (*map[string]string, bool)`

GetEnvVarsOk returns a tuple with the EnvVars field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvVars

`func (o *CreateProjectDTO) SetEnvVars(v map[string]string)`

SetEnvVars sets EnvVars field to given value.


### GetGitProviderConfigId

`func (o *CreateProjectDTO) GetGitProviderConfigId() string`

GetGitProviderConfigId returns the GitProviderConfigId field if non-nil, zero value otherwise.

### GetGitProviderConfigIdOk

`func (o *CreateProjectDTO) GetGitProviderConfigIdOk() (*string, bool)`

GetGitProviderConfigIdOk returns a tuple with the GitProviderConfigId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGitProviderConfigId

`func (o *CreateProjectDTO) SetGitProviderConfigId(v string)`

SetGitProviderConfigId sets GitProviderConfigId field to given value.

### HasGitProviderConfigId

`func (o *CreateProjectDTO) HasGitProviderConfigId() bool`

HasGitProviderConfigId returns a boolean if a field has been set.

### GetImage

`func (o *CreateProjectDTO) GetImage() string`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *CreateProjectDTO) GetImageOk() (*string, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *CreateProjectDTO) SetImage(v string)`

SetImage sets Image field to given value.

### HasImage

`func (o *CreateProjectDTO) HasImage() bool`

HasImage returns a boolean if a field has been set.

### GetName

`func (o *CreateProjectDTO) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *CreateProjectDTO) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *CreateProjectDTO) SetName(v string)`

SetName sets Name field to given value.


### GetSource

`func (o *CreateProjectDTO) GetSource() CreateProjectSourceDTO`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *CreateProjectDTO) GetSourceOk() (*CreateProjectSourceDTO, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *CreateProjectDTO) SetSource(v CreateProjectSourceDTO)`

SetSource sets Source field to given value.


### GetUser

`func (o *CreateProjectDTO) GetUser() string`

GetUser returns the User field if non-nil, zero value otherwise.

### GetUserOk

`func (o *CreateProjectDTO) GetUserOk() (*string, bool)`

GetUserOk returns a tuple with the User field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUser

`func (o *CreateProjectDTO) SetUser(v string)`

SetUser sets User field to given value.

### HasUser

`func (o *CreateProjectDTO) HasUser() bool`

HasUser returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


