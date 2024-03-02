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

// checks if the WorkspaceProvider type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &WorkspaceProvider{}

// WorkspaceProvider struct for WorkspaceProvider
type WorkspaceProvider struct {
	Name *string `json:"name,omitempty"`
	Profile *string `json:"profile,omitempty"`
}

// NewWorkspaceProvider instantiates a new WorkspaceProvider object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewWorkspaceProvider() *WorkspaceProvider {
	this := WorkspaceProvider{}
	return &this
}

// NewWorkspaceProviderWithDefaults instantiates a new WorkspaceProvider object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewWorkspaceProviderWithDefaults() *WorkspaceProvider {
	this := WorkspaceProvider{}
	return &this
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *WorkspaceProvider) GetName() string {
	if o == nil || IsNil(o.Name) {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *WorkspaceProvider) GetNameOk() (*string, bool) {
	if o == nil || IsNil(o.Name) {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *WorkspaceProvider) HasName() bool {
	if o != nil && !IsNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *WorkspaceProvider) SetName(v string) {
	o.Name = &v
}

// GetProfile returns the Profile field value if set, zero value otherwise.
func (o *WorkspaceProvider) GetProfile() string {
	if o == nil || IsNil(o.Profile) {
		var ret string
		return ret
	}
	return *o.Profile
}

// GetProfileOk returns a tuple with the Profile field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *WorkspaceProvider) GetProfileOk() (*string, bool) {
	if o == nil || IsNil(o.Profile) {
		return nil, false
	}
	return o.Profile, true
}

// HasProfile returns a boolean if a field has been set.
func (o *WorkspaceProvider) HasProfile() bool {
	if o != nil && !IsNil(o.Profile) {
		return true
	}

	return false
}

// SetProfile gets a reference to the given string and assigns it to the Profile field.
func (o *WorkspaceProvider) SetProfile(v string) {
	o.Profile = &v
}

func (o WorkspaceProvider) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o WorkspaceProvider) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Name) {
		toSerialize["name"] = o.Name
	}
	if !IsNil(o.Profile) {
		toSerialize["profile"] = o.Profile
	}
	return toSerialize, nil
}

type NullableWorkspaceProvider struct {
	value *WorkspaceProvider
	isSet bool
}

func (v NullableWorkspaceProvider) Get() *WorkspaceProvider {
	return v.value
}

func (v *NullableWorkspaceProvider) Set(val *WorkspaceProvider) {
	v.value = val
	v.isSet = true
}

func (v NullableWorkspaceProvider) IsSet() bool {
	return v.isSet
}

func (v *NullableWorkspaceProvider) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableWorkspaceProvider(val *WorkspaceProvider) *NullableWorkspaceProvider {
	return &NullableWorkspaceProvider{value: val, isSet: true}
}

func (v NullableWorkspaceProvider) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableWorkspaceProvider) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


