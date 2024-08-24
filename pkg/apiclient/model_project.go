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

// checks if the Project type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &Project{}

// Project struct for Project
type Project struct {
	BuildConfig *ProjectBuildConfig `json:"buildConfig,omitempty"`
	Default     bool                `json:"default"`
	EnvVars     map[string]string   `json:"envVars"`
	Identity    string              `json:"identity"`
	Image       string              `json:"image"`
	Name        string              `json:"name"`
	Repository  GitRepository       `json:"repository"`
	State       *ProjectState       `json:"state,omitempty"`
	Target      string              `json:"target"`
	User        string              `json:"user"`
	WorkspaceId string              `json:"workspaceId"`
}

type _Project Project

// NewProject instantiates a new Project object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewProject(default_ bool, envVars map[string]string, identity string, image string, name string, repository GitRepository, target string, user string, workspaceId string) *Project {
	this := Project{}
	this.Default = default_
	this.EnvVars = envVars
	this.Identity = identity
	this.Image = image
	this.Name = name
	this.Repository = repository
	this.Target = target
	this.User = user
	this.WorkspaceId = workspaceId
	return &this
}

// NewProjectWithDefaults instantiates a new Project object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewProjectWithDefaults() *Project {
	this := Project{}
	return &this
}

// GetBuildConfig returns the BuildConfig field value if set, zero value otherwise.
func (o *Project) GetBuildConfig() ProjectBuildConfig {
	if o == nil || IsNil(o.BuildConfig) {
		var ret ProjectBuildConfig
		return ret
	}
	return *o.BuildConfig
}

// GetBuildConfigOk returns a tuple with the BuildConfig field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Project) GetBuildConfigOk() (*ProjectBuildConfig, bool) {
	if o == nil || IsNil(o.BuildConfig) {
		return nil, false
	}
	return o.BuildConfig, true
}

// HasBuildConfig returns a boolean if a field has been set.
func (o *Project) HasBuildConfig() bool {
	if o != nil && !IsNil(o.BuildConfig) {
		return true
	}

	return false
}

// SetBuildConfig gets a reference to the given ProjectBuildConfig and assigns it to the BuildConfig field.
func (o *Project) SetBuildConfig(v ProjectBuildConfig) {
	o.BuildConfig = &v
}

// GetDefault returns the Default field value
func (o *Project) GetDefault() bool {
	if o == nil {
		var ret bool
		return ret
	}

	return o.Default
}

// GetDefaultOk returns a tuple with the Default field value
// and a boolean to check if the value has been set.
func (o *Project) GetDefaultOk() (*bool, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Default, true
}

// SetDefault sets field value
func (o *Project) SetDefault(v bool) {
	o.Default = v
}

// GetEnvVars returns the EnvVars field value
func (o *Project) GetEnvVars() map[string]string {
	if o == nil {
		var ret map[string]string
		return ret
	}

	return o.EnvVars
}

// GetEnvVarsOk returns a tuple with the EnvVars field value
// and a boolean to check if the value has been set.
func (o *Project) GetEnvVarsOk() (*map[string]string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.EnvVars, true
}

// SetEnvVars sets field value
func (o *Project) SetEnvVars(v map[string]string) {
	o.EnvVars = v
}

// GetIdentity returns the Identity field value
func (o *Project) GetIdentity() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Identity
}

// GetIdentityOk returns a tuple with the Identity field value
// and a boolean to check if the value has been set.
func (o *Project) GetIdentityOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Identity, true
}

// SetIdentity sets field value
func (o *Project) SetIdentity(v string) {
	o.Identity = v
}

// GetImage returns the Image field value
func (o *Project) GetImage() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Image
}

// GetImageOk returns a tuple with the Image field value
// and a boolean to check if the value has been set.
func (o *Project) GetImageOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Image, true
}

// SetImage sets field value
func (o *Project) SetImage(v string) {
	o.Image = v
}

// GetName returns the Name field value
func (o *Project) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *Project) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *Project) SetName(v string) {
	o.Name = v
}

// GetRepository returns the Repository field value
func (o *Project) GetRepository() GitRepository {
	if o == nil {
		var ret GitRepository
		return ret
	}

	return o.Repository
}

// GetRepositoryOk returns a tuple with the Repository field value
// and a boolean to check if the value has been set.
func (o *Project) GetRepositoryOk() (*GitRepository, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Repository, true
}

// SetRepository sets field value
func (o *Project) SetRepository(v GitRepository) {
	o.Repository = v
}

// GetState returns the State field value if set, zero value otherwise.
func (o *Project) GetState() ProjectState {
	if o == nil || IsNil(o.State) {
		var ret ProjectState
		return ret
	}
	return *o.State
}

// GetStateOk returns a tuple with the State field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Project) GetStateOk() (*ProjectState, bool) {
	if o == nil || IsNil(o.State) {
		return nil, false
	}
	return o.State, true
}

// HasState returns a boolean if a field has been set.
func (o *Project) HasState() bool {
	if o != nil && !IsNil(o.State) {
		return true
	}

	return false
}

// SetState gets a reference to the given ProjectState and assigns it to the State field.
func (o *Project) SetState(v ProjectState) {
	o.State = &v
}

// GetTarget returns the Target field value
func (o *Project) GetTarget() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Target
}

// GetTargetOk returns a tuple with the Target field value
// and a boolean to check if the value has been set.
func (o *Project) GetTargetOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Target, true
}

// SetTarget sets field value
func (o *Project) SetTarget(v string) {
	o.Target = v
}

// GetUser returns the User field value
func (o *Project) GetUser() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.User
}

// GetUserOk returns a tuple with the User field value
// and a boolean to check if the value has been set.
func (o *Project) GetUserOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.User, true
}

// SetUser sets field value
func (o *Project) SetUser(v string) {
	o.User = v
}

// GetWorkspaceId returns the WorkspaceId field value
func (o *Project) GetWorkspaceId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.WorkspaceId
}

// GetWorkspaceIdOk returns a tuple with the WorkspaceId field value
// and a boolean to check if the value has been set.
func (o *Project) GetWorkspaceIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.WorkspaceId, true
}

// SetWorkspaceId sets field value
func (o *Project) SetWorkspaceId(v string) {
	o.WorkspaceId = v
}

func (o Project) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o Project) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.BuildConfig) {
		toSerialize["buildConfig"] = o.BuildConfig
	}
	toSerialize["default"] = o.Default
	toSerialize["envVars"] = o.EnvVars
	toSerialize["identity"] = o.Identity
	toSerialize["image"] = o.Image
	toSerialize["name"] = o.Name
	toSerialize["repository"] = o.Repository
	if !IsNil(o.State) {
		toSerialize["state"] = o.State
	}
	toSerialize["target"] = o.Target
	toSerialize["user"] = o.User
	toSerialize["workspaceId"] = o.WorkspaceId
	return toSerialize, nil
}

func (o *Project) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"default",
		"envVars",
		"identity",
		"image",
		"name",
		"repository",
		"target",
		"user",
		"workspaceId",
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

	varProject := _Project{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varProject)

	if err != nil {
		return err
	}

	*o = Project(varProject)

	return err
}

type NullableProject struct {
	value *Project
	isSet bool
}

func (v NullableProject) Get() *Project {
	return v.value
}

func (v *NullableProject) Set(val *Project) {
	v.value = val
	v.isSet = true
}

func (v NullableProject) IsSet() bool {
	return v.isSet
}

func (v *NullableProject) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableProject(val *Project) *NullableProject {
	return &NullableProject{value: val, isSet: true}
}

func (v NullableProject) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableProject) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
