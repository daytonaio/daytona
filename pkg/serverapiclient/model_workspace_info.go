/*
Daytona Server API

Daytona Server API

API version: 0.1.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package serverapiclient

import (
	"encoding/json"
)

// checks if the WorkspaceInfo type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &WorkspaceInfo{}

// WorkspaceInfo struct for WorkspaceInfo
type WorkspaceInfo struct {
	Name *string `json:"name,omitempty"`
	Projects []ProjectInfo `json:"projects,omitempty"`
	ProviderMetadata *string `json:"providerMetadata,omitempty"`
}

// NewWorkspaceInfo instantiates a new WorkspaceInfo object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewWorkspaceInfo() *WorkspaceInfo {
	this := WorkspaceInfo{}
	return &this
}

// NewWorkspaceInfoWithDefaults instantiates a new WorkspaceInfo object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewWorkspaceInfoWithDefaults() *WorkspaceInfo {
	this := WorkspaceInfo{}
	return &this
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *WorkspaceInfo) GetName() string {
	if o == nil || IsNil(o.Name) {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *WorkspaceInfo) GetNameOk() (*string, bool) {
	if o == nil || IsNil(o.Name) {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *WorkspaceInfo) HasName() bool {
	if o != nil && !IsNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *WorkspaceInfo) SetName(v string) {
	o.Name = &v
}

// GetProjects returns the Projects field value if set, zero value otherwise.
func (o *WorkspaceInfo) GetProjects() []ProjectInfo {
	if o == nil || IsNil(o.Projects) {
		var ret []ProjectInfo
		return ret
	}
	return o.Projects
}

// GetProjectsOk returns a tuple with the Projects field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *WorkspaceInfo) GetProjectsOk() ([]ProjectInfo, bool) {
	if o == nil || IsNil(o.Projects) {
		return nil, false
	}
	return o.Projects, true
}

// HasProjects returns a boolean if a field has been set.
func (o *WorkspaceInfo) HasProjects() bool {
	if o != nil && !IsNil(o.Projects) {
		return true
	}

	return false
}

// SetProjects gets a reference to the given []ProjectInfo and assigns it to the Projects field.
func (o *WorkspaceInfo) SetProjects(v []ProjectInfo) {
	o.Projects = v
}

// GetProviderMetadata returns the ProviderMetadata field value if set, zero value otherwise.
func (o *WorkspaceInfo) GetProviderMetadata() string {
	if o == nil || IsNil(o.ProviderMetadata) {
		var ret string
		return ret
	}
	return *o.ProviderMetadata
}

// GetProviderMetadataOk returns a tuple with the ProviderMetadata field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *WorkspaceInfo) GetProviderMetadataOk() (*string, bool) {
	if o == nil || IsNil(o.ProviderMetadata) {
		return nil, false
	}
	return o.ProviderMetadata, true
}

// HasProviderMetadata returns a boolean if a field has been set.
func (o *WorkspaceInfo) HasProviderMetadata() bool {
	if o != nil && !IsNil(o.ProviderMetadata) {
		return true
	}

	return false
}

// SetProviderMetadata gets a reference to the given string and assigns it to the ProviderMetadata field.
func (o *WorkspaceInfo) SetProviderMetadata(v string) {
	o.ProviderMetadata = &v
}

func (o WorkspaceInfo) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o WorkspaceInfo) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Name) {
		toSerialize["name"] = o.Name
	}
	if !IsNil(o.Projects) {
		toSerialize["projects"] = o.Projects
	}
	if !IsNil(o.ProviderMetadata) {
		toSerialize["providerMetadata"] = o.ProviderMetadata
	}
	return toSerialize, nil
}

type NullableWorkspaceInfo struct {
	value *WorkspaceInfo
	isSet bool
}

func (v NullableWorkspaceInfo) Get() *WorkspaceInfo {
	return v.value
}

func (v *NullableWorkspaceInfo) Set(val *WorkspaceInfo) {
	v.value = val
	v.isSet = true
}

func (v NullableWorkspaceInfo) IsSet() bool {
	return v.isSet
}

func (v *NullableWorkspaceInfo) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableWorkspaceInfo(val *WorkspaceInfo) *NullableWorkspaceInfo {
	return &NullableWorkspaceInfo{value: val, isSet: true}
}

func (v NullableWorkspaceInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableWorkspaceInfo) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


