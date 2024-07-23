# CreateProjectConfigDTO

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Build** | Pointer to [**ProjectBuildConfig**](ProjectBuildConfig.md) |  | [optional] 
**EnvVars** | Pointer to **map[string]string** |  | [optional] 
**Image** | Pointer to **string** |  | [optional] 
**Name** | Pointer to **string** |  | [optional] 
**Source** | Pointer to [**CreateProjectConfigSourceDTO**](CreateProjectConfigSourceDTO.md) |  | [optional] 
**User** | Pointer to **string** |  | [optional] 

## Methods

### NewCreateProjectConfigDTO

`func NewCreateProjectConfigDTO() *CreateProjectConfigDTO`

NewCreateProjectConfigDTO instantiates a new CreateProjectConfigDTO object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateProjectConfigDTOWithDefaults

`func NewCreateProjectConfigDTOWithDefaults() *CreateProjectConfigDTO`

NewCreateProjectConfigDTOWithDefaults instantiates a new CreateProjectConfigDTO object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBuild

`func (o *CreateProjectConfigDTO) GetBuild() ProjectBuildConfig`

GetBuild returns the Build field if non-nil, zero value otherwise.

### GetBuildOk

`func (o *CreateProjectConfigDTO) GetBuildOk() (*ProjectBuildConfig, bool)`

GetBuildOk returns a tuple with the Build field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBuild

`func (o *CreateProjectConfigDTO) SetBuild(v ProjectBuildConfig)`

SetBuild sets Build field to given value.

### HasBuild

`func (o *CreateProjectConfigDTO) HasBuild() bool`

HasBuild returns a boolean if a field has been set.

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

### HasEnvVars

`func (o *CreateProjectConfigDTO) HasEnvVars() bool`

HasEnvVars returns a boolean if a field has been set.

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

### HasName

`func (o *CreateProjectConfigDTO) HasName() bool`

HasName returns a boolean if a field has been set.

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

### HasSource

`func (o *CreateProjectConfigDTO) HasSource() bool`

HasSource returns a boolean if a field has been set.

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


