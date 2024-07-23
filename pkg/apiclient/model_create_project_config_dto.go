/*
Daytona Server API

Daytona Server API

API version: 0.1.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package apiclient

import (
	"encoding/json"
)

// checks if the CreateProjectConfigDTO type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &CreateProjectConfigDTO{}

// CreateProjectConfigDTO struct for CreateProjectConfigDTO
type CreateProjectConfigDTO struct {
	Build   *ProjectBuildConfig           `json:"build,omitempty"`
	EnvVars *map[string]string            `json:"envVars,omitempty"`
	Image   *string                       `json:"image,omitempty"`
	Name    *string                       `json:"name,omitempty"`
	Source  *CreateProjectConfigSourceDTO `json:"source,omitempty"`
	User    *string                       `json:"user,omitempty"`
}

// NewCreateProjectConfigDTO instantiates a new CreateProjectConfigDTO object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewCreateProjectConfigDTO() *CreateProjectConfigDTO {
	this := CreateProjectConfigDTO{}
	return &this
}

// NewCreateProjectConfigDTOWithDefaults instantiates a new CreateProjectConfigDTO object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewCreateProjectConfigDTOWithDefaults() *CreateProjectConfigDTO {
	this := CreateProjectConfigDTO{}
	return &this
}

// GetBuild returns the Build field value if set, zero value otherwise.
func (o *CreateProjectConfigDTO) GetBuild() ProjectBuildConfig {
	if o == nil || IsNil(o.Build) {
		var ret ProjectBuildConfig
		return ret
	}
	return *o.Build
}

// GetBuildOk returns a tuple with the Build field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateProjectConfigDTO) GetBuildOk() (*ProjectBuildConfig, bool) {
	if o == nil || IsNil(o.Build) {
		return nil, false
	}
	return o.Build, true
}

// HasBuild returns a boolean if a field has been set.
func (o *CreateProjectConfigDTO) HasBuild() bool {
	if o != nil && !IsNil(o.Build) {
		return true
	}

	return false
}

// SetBuild gets a reference to the given ProjectBuildConfig and assigns it to the Build field.
func (o *CreateProjectConfigDTO) SetBuild(v ProjectBuildConfig) {
	o.Build = &v
}

// GetEnvVars returns the EnvVars field value if set, zero value otherwise.
func (o *CreateProjectConfigDTO) GetEnvVars() map[string]string {
	if o == nil || IsNil(o.EnvVars) {
		var ret map[string]string
		return ret
	}
	return *o.EnvVars
}

// GetEnvVarsOk returns a tuple with the EnvVars field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateProjectConfigDTO) GetEnvVarsOk() (*map[string]string, bool) {
	if o == nil || IsNil(o.EnvVars) {
		return nil, false
	}
	return o.EnvVars, true
}

// HasEnvVars returns a boolean if a field has been set.
func (o *CreateProjectConfigDTO) HasEnvVars() bool {
	if o != nil && !IsNil(o.EnvVars) {
		return true
	}

	return false
}

// SetEnvVars gets a reference to the given map[string]string and assigns it to the EnvVars field.
func (o *CreateProjectConfigDTO) SetEnvVars(v map[string]string) {
	o.EnvVars = &v
}

// GetImage returns the Image field value if set, zero value otherwise.
func (o *CreateProjectConfigDTO) GetImage() string {
	if o == nil || IsNil(o.Image) {
		var ret string
		return ret
	}
	return *o.Image
}

// GetImageOk returns a tuple with the Image field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateProjectConfigDTO) GetImageOk() (*string, bool) {
	if o == nil || IsNil(o.Image) {
		return nil, false
	}
	return o.Image, true
}

// HasImage returns a boolean if a field has been set.
func (o *CreateProjectConfigDTO) HasImage() bool {
	if o != nil && !IsNil(o.Image) {
		return true
	}

	return false
}

// SetImage gets a reference to the given string and assigns it to the Image field.
func (o *CreateProjectConfigDTO) SetImage(v string) {
	o.Image = &v
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *CreateProjectConfigDTO) GetName() string {
	if o == nil || IsNil(o.Name) {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateProjectConfigDTO) GetNameOk() (*string, bool) {
	if o == nil || IsNil(o.Name) {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *CreateProjectConfigDTO) HasName() bool {
	if o != nil && !IsNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *CreateProjectConfigDTO) SetName(v string) {
	o.Name = &v
}

// GetSource returns the Source field value if set, zero value otherwise.
func (o *CreateProjectConfigDTO) GetSource() CreateProjectConfigSourceDTO {
	if o == nil || IsNil(o.Source) {
		var ret CreateProjectConfigSourceDTO
		return ret
	}
	return *o.Source
}

// GetSourceOk returns a tuple with the Source field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateProjectConfigDTO) GetSourceOk() (*CreateProjectConfigSourceDTO, bool) {
	if o == nil || IsNil(o.Source) {
		return nil, false
	}
	return o.Source, true
}

// HasSource returns a boolean if a field has been set.
func (o *CreateProjectConfigDTO) HasSource() bool {
	if o != nil && !IsNil(o.Source) {
		return true
	}

	return false
}

// SetSource gets a reference to the given CreateProjectConfigSourceDTO and assigns it to the Source field.
func (o *CreateProjectConfigDTO) SetSource(v CreateProjectConfigSourceDTO) {
	o.Source = &v
}

// GetUser returns the User field value if set, zero value otherwise.
func (o *CreateProjectConfigDTO) GetUser() string {
	if o == nil || IsNil(o.User) {
		var ret string
		return ret
	}
	return *o.User
}

// GetUserOk returns a tuple with the User field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateProjectConfigDTO) GetUserOk() (*string, bool) {
	if o == nil || IsNil(o.User) {
		return nil, false
	}
	return o.User, true
}

// HasUser returns a boolean if a field has been set.
func (o *CreateProjectConfigDTO) HasUser() bool {
	if o != nil && !IsNil(o.User) {
		return true
	}

	return false
}

// SetUser gets a reference to the given string and assigns it to the User field.
func (o *CreateProjectConfigDTO) SetUser(v string) {
	o.User = &v
}

func (o CreateProjectConfigDTO) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o CreateProjectConfigDTO) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Build) {
		toSerialize["build"] = o.Build
	}
	if !IsNil(o.EnvVars) {
		toSerialize["envVars"] = o.EnvVars
	}
	if !IsNil(o.Image) {
		toSerialize["image"] = o.Image
	}
	if !IsNil(o.Name) {
		toSerialize["name"] = o.Name
	}
	if !IsNil(o.Source) {
		toSerialize["source"] = o.Source
	}
	if !IsNil(o.User) {
		toSerialize["user"] = o.User
	}
	return toSerialize, nil
}

type NullableCreateProjectConfigDTO struct {
	value *CreateProjectConfigDTO
	isSet bool
}

func (v NullableCreateProjectConfigDTO) Get() *CreateProjectConfigDTO {
	return v.value
}

func (v *NullableCreateProjectConfigDTO) Set(val *CreateProjectConfigDTO) {
	v.value = val
	v.isSet = true
}

func (v NullableCreateProjectConfigDTO) IsSet() bool {
	return v.isSet
}

func (v *NullableCreateProjectConfigDTO) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableCreateProjectConfigDTO(val *CreateProjectConfigDTO) *NullableCreateProjectConfigDTO {
	return &NullableCreateProjectConfigDTO{value: val, isSet: true}
}

func (v NullableCreateProjectConfigDTO) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableCreateProjectConfigDTO) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
