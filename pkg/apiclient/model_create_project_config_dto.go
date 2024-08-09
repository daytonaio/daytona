/*
Daytona Server API

Daytona Server API

API version: v0.0.0-dev
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package apiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// checks if the CreateProjectConfigDTO type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &CreateProjectConfigDTO{}

// CreateProjectConfigDTO struct for CreateProjectConfigDTO
type CreateProjectConfigDTO struct {
	BuildConfig *ProjectBuildConfig          `json:"buildConfig,omitempty"`
	EnvVars     map[string]string            `json:"envVars"`
	Image       *string                      `json:"image,omitempty"`
	Name        string                       `json:"name"`
	Source      CreateProjectConfigSourceDTO `json:"source"`
	User        *string                      `json:"user,omitempty"`
}

type _CreateProjectConfigDTO CreateProjectConfigDTO

// NewCreateProjectConfigDTO instantiates a new CreateProjectConfigDTO object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewCreateProjectConfigDTO(envVars map[string]string, name string, source CreateProjectConfigSourceDTO) *CreateProjectConfigDTO {
	this := CreateProjectConfigDTO{}
	this.EnvVars = envVars
	this.Name = name
	this.Source = source
	return &this
}

// NewCreateProjectConfigDTOWithDefaults instantiates a new CreateProjectConfigDTO object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewCreateProjectConfigDTOWithDefaults() *CreateProjectConfigDTO {
	this := CreateProjectConfigDTO{}
	return &this
}

// GetBuildConfig returns the BuildConfig field value if set, zero value otherwise.
func (o *CreateProjectConfigDTO) GetBuildConfig() ProjectBuildConfig {
	if o == nil || IsNil(o.BuildConfig) {
		var ret ProjectBuildConfig
		return ret
	}
	return *o.BuildConfig
}

// GetBuildConfigOk returns a tuple with the BuildConfig field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateProjectConfigDTO) GetBuildConfigOk() (*ProjectBuildConfig, bool) {
	if o == nil || IsNil(o.BuildConfig) {
		return nil, false
	}
	return o.BuildConfig, true
}

// HasBuildConfig returns a boolean if a field has been set.
func (o *CreateProjectConfigDTO) HasBuildConfig() bool {
	if o != nil && !IsNil(o.BuildConfig) {
		return true
	}

	return false
}

// SetBuildConfig gets a reference to the given ProjectBuildConfig and assigns it to the BuildConfig field.
func (o *CreateProjectConfigDTO) SetBuildConfig(v ProjectBuildConfig) {
	o.BuildConfig = &v
}

// GetEnvVars returns the EnvVars field value
func (o *CreateProjectConfigDTO) GetEnvVars() map[string]string {
	if o == nil {
		var ret map[string]string
		return ret
	}

	return o.EnvVars
}

// GetEnvVarsOk returns a tuple with the EnvVars field value
// and a boolean to check if the value has been set.
func (o *CreateProjectConfigDTO) GetEnvVarsOk() (*map[string]string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.EnvVars, true
}

// SetEnvVars sets field value
func (o *CreateProjectConfigDTO) SetEnvVars(v map[string]string) {
	o.EnvVars = v
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

// GetName returns the Name field value
func (o *CreateProjectConfigDTO) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *CreateProjectConfigDTO) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *CreateProjectConfigDTO) SetName(v string) {
	o.Name = v
}

// GetSource returns the Source field value
func (o *CreateProjectConfigDTO) GetSource() CreateProjectConfigSourceDTO {
	if o == nil {
		var ret CreateProjectConfigSourceDTO
		return ret
	}

	return o.Source
}

// GetSourceOk returns a tuple with the Source field value
// and a boolean to check if the value has been set.
func (o *CreateProjectConfigDTO) GetSourceOk() (*CreateProjectConfigSourceDTO, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Source, true
}

// SetSource sets field value
func (o *CreateProjectConfigDTO) SetSource(v CreateProjectConfigSourceDTO) {
	o.Source = v
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
	if !IsNil(o.BuildConfig) {
		toSerialize["buildConfig"] = o.BuildConfig
	}
	toSerialize["envVars"] = o.EnvVars
	if !IsNil(o.Image) {
		toSerialize["image"] = o.Image
	}
	toSerialize["name"] = o.Name
	toSerialize["source"] = o.Source
	if !IsNil(o.User) {
		toSerialize["user"] = o.User
	}
	return toSerialize, nil
}

func (o *CreateProjectConfigDTO) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"envVars",
		"name",
		"source",
	}

	allProperties := make(map[string]interface{})

	err = json.Unmarshal(data, &allProperties)

	if err != nil {
		return err
	}

	for _, requiredProperty := range requiredProperties {
		if _, exists := allProperties[requiredProperty]; !exists {
			return fmt.Errorf("no value given for required property %v", requiredProperty)
		}
	}

	varCreateProjectConfigDTO := _CreateProjectConfigDTO{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varCreateProjectConfigDTO)

	if err != nil {
		return err
	}

	*o = CreateProjectConfigDTO(varCreateProjectConfigDTO)

	return err
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
