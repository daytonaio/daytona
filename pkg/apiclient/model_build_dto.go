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

// checks if the BuildDTO type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &BuildDTO{}

// BuildDTO struct for BuildDTO
type BuildDTO struct {
	BuildConfig     *BuildConfig      `json:"buildConfig,omitempty"`
	ContainerConfig ContainerConfig   `json:"containerConfig"`
	CreatedAt       string            `json:"createdAt"`
	EnvVars         map[string]string `json:"envVars"`
	Id              string            `json:"id"`
	Image           *string           `json:"image,omitempty"`
	LastJob         *Job              `json:"lastJob,omitempty"`
	LastJobId       *string           `json:"lastJobId,omitempty"`
	PrebuildId      *string           `json:"prebuildId,omitempty"`
	Repository      GitRepository     `json:"repository"`
	State           ResourceState     `json:"state"`
	UpdatedAt       string            `json:"updatedAt"`
	User            *string           `json:"user,omitempty"`
}

type _BuildDTO BuildDTO

// NewBuildDTO instantiates a new BuildDTO object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewBuildDTO(containerConfig ContainerConfig, createdAt string, envVars map[string]string, id string, repository GitRepository, state ResourceState, updatedAt string) *BuildDTO {
	this := BuildDTO{}
	this.ContainerConfig = containerConfig
	this.CreatedAt = createdAt
	this.EnvVars = envVars
	this.Id = id
	this.Repository = repository
	this.State = state
	this.UpdatedAt = updatedAt
	return &this
}

// NewBuildDTOWithDefaults instantiates a new BuildDTO object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewBuildDTOWithDefaults() *BuildDTO {
	this := BuildDTO{}
	return &this
}

// GetBuildConfig returns the BuildConfig field value if set, zero value otherwise.
func (o *BuildDTO) GetBuildConfig() BuildConfig {
	if o == nil || IsNil(o.BuildConfig) {
		var ret BuildConfig
		return ret
	}
	return *o.BuildConfig
}

// GetBuildConfigOk returns a tuple with the BuildConfig field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *BuildDTO) GetBuildConfigOk() (*BuildConfig, bool) {
	if o == nil || IsNil(o.BuildConfig) {
		return nil, false
	}
	return o.BuildConfig, true
}

// HasBuildConfig returns a boolean if a field has been set.
func (o *BuildDTO) HasBuildConfig() bool {
	if o != nil && !IsNil(o.BuildConfig) {
		return true
	}

	return false
}

// SetBuildConfig gets a reference to the given BuildConfig and assigns it to the BuildConfig field.
func (o *BuildDTO) SetBuildConfig(v BuildConfig) {
	o.BuildConfig = &v
}

// GetContainerConfig returns the ContainerConfig field value
func (o *BuildDTO) GetContainerConfig() ContainerConfig {
	if o == nil {
		var ret ContainerConfig
		return ret
	}

	return o.ContainerConfig
}

// GetContainerConfigOk returns a tuple with the ContainerConfig field value
// and a boolean to check if the value has been set.
func (o *BuildDTO) GetContainerConfigOk() (*ContainerConfig, bool) {
	if o == nil {
		return nil, false
	}
	return &o.ContainerConfig, true
}

// SetContainerConfig sets field value
func (o *BuildDTO) SetContainerConfig(v ContainerConfig) {
	o.ContainerConfig = v
}

// GetCreatedAt returns the CreatedAt field value
func (o *BuildDTO) GetCreatedAt() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.CreatedAt
}

// GetCreatedAtOk returns a tuple with the CreatedAt field value
// and a boolean to check if the value has been set.
func (o *BuildDTO) GetCreatedAtOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.CreatedAt, true
}

// SetCreatedAt sets field value
func (o *BuildDTO) SetCreatedAt(v string) {
	o.CreatedAt = v
}

// GetEnvVars returns the EnvVars field value
func (o *BuildDTO) GetEnvVars() map[string]string {
	if o == nil {
		var ret map[string]string
		return ret
	}

	return o.EnvVars
}

// GetEnvVarsOk returns a tuple with the EnvVars field value
// and a boolean to check if the value has been set.
func (o *BuildDTO) GetEnvVarsOk() (*map[string]string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.EnvVars, true
}

// SetEnvVars sets field value
func (o *BuildDTO) SetEnvVars(v map[string]string) {
	o.EnvVars = v
}

// GetId returns the Id field value
func (o *BuildDTO) GetId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Id
}

// GetIdOk returns a tuple with the Id field value
// and a boolean to check if the value has been set.
func (o *BuildDTO) GetIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Id, true
}

// SetId sets field value
func (o *BuildDTO) SetId(v string) {
	o.Id = v
}

// GetImage returns the Image field value if set, zero value otherwise.
func (o *BuildDTO) GetImage() string {
	if o == nil || IsNil(o.Image) {
		var ret string
		return ret
	}
	return *o.Image
}

// GetImageOk returns a tuple with the Image field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *BuildDTO) GetImageOk() (*string, bool) {
	if o == nil || IsNil(o.Image) {
		return nil, false
	}
	return o.Image, true
}

// HasImage returns a boolean if a field has been set.
func (o *BuildDTO) HasImage() bool {
	if o != nil && !IsNil(o.Image) {
		return true
	}

	return false
}

// SetImage gets a reference to the given string and assigns it to the Image field.
func (o *BuildDTO) SetImage(v string) {
	o.Image = &v
}

// GetLastJob returns the LastJob field value if set, zero value otherwise.
func (o *BuildDTO) GetLastJob() Job {
	if o == nil || IsNil(o.LastJob) {
		var ret Job
		return ret
	}
	return *o.LastJob
}

// GetLastJobOk returns a tuple with the LastJob field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *BuildDTO) GetLastJobOk() (*Job, bool) {
	if o == nil || IsNil(o.LastJob) {
		return nil, false
	}
	return o.LastJob, true
}

// HasLastJob returns a boolean if a field has been set.
func (o *BuildDTO) HasLastJob() bool {
	if o != nil && !IsNil(o.LastJob) {
		return true
	}

	return false
}

// SetLastJob gets a reference to the given Job and assigns it to the LastJob field.
func (o *BuildDTO) SetLastJob(v Job) {
	o.LastJob = &v
}

// GetLastJobId returns the LastJobId field value if set, zero value otherwise.
func (o *BuildDTO) GetLastJobId() string {
	if o == nil || IsNil(o.LastJobId) {
		var ret string
		return ret
	}
	return *o.LastJobId
}

// GetLastJobIdOk returns a tuple with the LastJobId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *BuildDTO) GetLastJobIdOk() (*string, bool) {
	if o == nil || IsNil(o.LastJobId) {
		return nil, false
	}
	return o.LastJobId, true
}

// HasLastJobId returns a boolean if a field has been set.
func (o *BuildDTO) HasLastJobId() bool {
	if o != nil && !IsNil(o.LastJobId) {
		return true
	}

	return false
}

// SetLastJobId gets a reference to the given string and assigns it to the LastJobId field.
func (o *BuildDTO) SetLastJobId(v string) {
	o.LastJobId = &v
}

// GetPrebuildId returns the PrebuildId field value if set, zero value otherwise.
func (o *BuildDTO) GetPrebuildId() string {
	if o == nil || IsNil(o.PrebuildId) {
		var ret string
		return ret
	}
	return *o.PrebuildId
}

// GetPrebuildIdOk returns a tuple with the PrebuildId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *BuildDTO) GetPrebuildIdOk() (*string, bool) {
	if o == nil || IsNil(o.PrebuildId) {
		return nil, false
	}
	return o.PrebuildId, true
}

// HasPrebuildId returns a boolean if a field has been set.
func (o *BuildDTO) HasPrebuildId() bool {
	if o != nil && !IsNil(o.PrebuildId) {
		return true
	}

	return false
}

// SetPrebuildId gets a reference to the given string and assigns it to the PrebuildId field.
func (o *BuildDTO) SetPrebuildId(v string) {
	o.PrebuildId = &v
}

// GetRepository returns the Repository field value
func (o *BuildDTO) GetRepository() GitRepository {
	if o == nil {
		var ret GitRepository
		return ret
	}

	return o.Repository
}

// GetRepositoryOk returns a tuple with the Repository field value
// and a boolean to check if the value has been set.
func (o *BuildDTO) GetRepositoryOk() (*GitRepository, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Repository, true
}

// SetRepository sets field value
func (o *BuildDTO) SetRepository(v GitRepository) {
	o.Repository = v
}

// GetState returns the State field value
func (o *BuildDTO) GetState() ResourceState {
	if o == nil {
		var ret ResourceState
		return ret
	}

	return o.State
}

// GetStateOk returns a tuple with the State field value
// and a boolean to check if the value has been set.
func (o *BuildDTO) GetStateOk() (*ResourceState, bool) {
	if o == nil {
		return nil, false
	}
	return &o.State, true
}

// SetState sets field value
func (o *BuildDTO) SetState(v ResourceState) {
	o.State = v
}

// GetUpdatedAt returns the UpdatedAt field value
func (o *BuildDTO) GetUpdatedAt() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.UpdatedAt
}

// GetUpdatedAtOk returns a tuple with the UpdatedAt field value
// and a boolean to check if the value has been set.
func (o *BuildDTO) GetUpdatedAtOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.UpdatedAt, true
}

// SetUpdatedAt sets field value
func (o *BuildDTO) SetUpdatedAt(v string) {
	o.UpdatedAt = v
}

// GetUser returns the User field value if set, zero value otherwise.
func (o *BuildDTO) GetUser() string {
	if o == nil || IsNil(o.User) {
		var ret string
		return ret
	}
	return *o.User
}

// GetUserOk returns a tuple with the User field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *BuildDTO) GetUserOk() (*string, bool) {
	if o == nil || IsNil(o.User) {
		return nil, false
	}
	return o.User, true
}

// HasUser returns a boolean if a field has been set.
func (o *BuildDTO) HasUser() bool {
	if o != nil && !IsNil(o.User) {
		return true
	}

	return false
}

// SetUser gets a reference to the given string and assigns it to the User field.
func (o *BuildDTO) SetUser(v string) {
	o.User = &v
}

func (o BuildDTO) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o BuildDTO) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.BuildConfig) {
		toSerialize["buildConfig"] = o.BuildConfig
	}
	toSerialize["containerConfig"] = o.ContainerConfig
	toSerialize["createdAt"] = o.CreatedAt
	toSerialize["envVars"] = o.EnvVars
	toSerialize["id"] = o.Id
	if !IsNil(o.Image) {
		toSerialize["image"] = o.Image
	}
	if !IsNil(o.LastJob) {
		toSerialize["lastJob"] = o.LastJob
	}
	if !IsNil(o.LastJobId) {
		toSerialize["lastJobId"] = o.LastJobId
	}
	if !IsNil(o.PrebuildId) {
		toSerialize["prebuildId"] = o.PrebuildId
	}
	toSerialize["repository"] = o.Repository
	toSerialize["state"] = o.State
	toSerialize["updatedAt"] = o.UpdatedAt
	if !IsNil(o.User) {
		toSerialize["user"] = o.User
	}
	return toSerialize, nil
}

func (o *BuildDTO) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"containerConfig",
		"createdAt",
		"envVars",
		"id",
		"repository",
		"state",
		"updatedAt",
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

	varBuildDTO := _BuildDTO{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varBuildDTO)

	if err != nil {
		return err
	}

	*o = BuildDTO(varBuildDTO)

	return err
}

type NullableBuildDTO struct {
	value *BuildDTO
	isSet bool
}

func (v NullableBuildDTO) Get() *BuildDTO {
	return v.value
}

func (v *NullableBuildDTO) Set(val *BuildDTO) {
	v.value = val
	v.isSet = true
}

func (v NullableBuildDTO) IsSet() bool {
	return v.isSet
}

func (v *NullableBuildDTO) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableBuildDTO(val *BuildDTO) *NullableBuildDTO {
	return &NullableBuildDTO{value: val, isSet: true}
}

func (v NullableBuildDTO) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableBuildDTO) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
