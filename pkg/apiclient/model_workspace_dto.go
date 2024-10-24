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

// checks if the WorkspaceDTO type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &WorkspaceDTO{}

// WorkspaceDTO struct for WorkspaceDTO
type WorkspaceDTO struct {
	Id           string         `json:"id"`
	Info         *WorkspaceInfo `json:"info,omitempty"`
	Name         string         `json:"name"`
	Projects     []Project      `json:"projects"`
	TargetConfig string         `json:"targetConfig"`
}

type _WorkspaceDTO WorkspaceDTO

// NewWorkspaceDTO instantiates a new WorkspaceDTO object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewWorkspaceDTO(id string, name string, projects []Project, targetConfig string) *WorkspaceDTO {
	this := WorkspaceDTO{}
	this.Id = id
	this.Name = name
	this.Projects = projects
	this.TargetConfig = targetConfig
	return &this
}

// NewWorkspaceDTOWithDefaults instantiates a new WorkspaceDTO object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewWorkspaceDTOWithDefaults() *WorkspaceDTO {
	this := WorkspaceDTO{}
	return &this
}

// GetId returns the Id field value
func (o *WorkspaceDTO) GetId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Id
}

// GetIdOk returns a tuple with the Id field value
// and a boolean to check if the value has been set.
func (o *WorkspaceDTO) GetIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Id, true
}

// SetId sets field value
func (o *WorkspaceDTO) SetId(v string) {
	o.Id = v
}

// GetInfo returns the Info field value if set, zero value otherwise.
func (o *WorkspaceDTO) GetInfo() WorkspaceInfo {
	if o == nil || IsNil(o.Info) {
		var ret WorkspaceInfo
		return ret
	}
	return *o.Info
}

// GetInfoOk returns a tuple with the Info field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *WorkspaceDTO) GetInfoOk() (*WorkspaceInfo, bool) {
	if o == nil || IsNil(o.Info) {
		return nil, false
	}
	return o.Info, true
}

// HasInfo returns a boolean if a field has been set.
func (o *WorkspaceDTO) HasInfo() bool {
	if o != nil && !IsNil(o.Info) {
		return true
	}

	return false
}

// SetInfo gets a reference to the given WorkspaceInfo and assigns it to the Info field.
func (o *WorkspaceDTO) SetInfo(v WorkspaceInfo) {
	o.Info = &v
}

// GetName returns the Name field value
func (o *WorkspaceDTO) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *WorkspaceDTO) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *WorkspaceDTO) SetName(v string) {
	o.Name = v
}

// GetProjects returns the Projects field value
func (o *WorkspaceDTO) GetProjects() []Project {
	if o == nil {
		var ret []Project
		return ret
	}

	return o.Projects
}

// GetProjectsOk returns a tuple with the Projects field value
// and a boolean to check if the value has been set.
func (o *WorkspaceDTO) GetProjectsOk() ([]Project, bool) {
	if o == nil {
		return nil, false
	}
	return o.Projects, true
}

// SetProjects sets field value
func (o *WorkspaceDTO) SetProjects(v []Project) {
	o.Projects = v
}

// GetTargetConfig returns the TargetConfig field value
func (o *WorkspaceDTO) GetTargetConfig() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.TargetConfig
}

// GetTargetConfigOk returns a tuple with the TargetConfig field value
// and a boolean to check if the value has been set.
func (o *WorkspaceDTO) GetTargetConfigOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.TargetConfig, true
}

// SetTargetConfig sets field value
func (o *WorkspaceDTO) SetTargetConfig(v string) {
	o.TargetConfig = v
}

func (o WorkspaceDTO) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o WorkspaceDTO) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["id"] = o.Id
	if !IsNil(o.Info) {
		toSerialize["info"] = o.Info
	}
	toSerialize["name"] = o.Name
	toSerialize["projects"] = o.Projects
	toSerialize["targetConfig"] = o.TargetConfig
	return toSerialize, nil
}

func (o *WorkspaceDTO) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"id",
		"name",
		"projects",
		"targetConfig",
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

	varWorkspaceDTO := _WorkspaceDTO{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varWorkspaceDTO)

	if err != nil {
		return err
	}

	*o = WorkspaceDTO(varWorkspaceDTO)

	return err
}

type NullableWorkspaceDTO struct {
	value *WorkspaceDTO
	isSet bool
}

func (v NullableWorkspaceDTO) Get() *WorkspaceDTO {
	return v.value
}

func (v *NullableWorkspaceDTO) Set(val *WorkspaceDTO) {
	v.value = val
	v.isSet = true
}

func (v NullableWorkspaceDTO) IsSet() bool {
	return v.isSet
}

func (v *NullableWorkspaceDTO) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableWorkspaceDTO(val *WorkspaceDTO) *NullableWorkspaceDTO {
	return &NullableWorkspaceDTO{value: val, isSet: true}
}

func (v NullableWorkspaceDTO) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableWorkspaceDTO) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
