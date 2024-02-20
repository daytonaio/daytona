/*
Daytona Server API

Daytona Server API

API version: 0.1.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package api_client

import (
	"encoding/json"
)

// checks if the TypesWorkspaceProvisioner type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &TypesWorkspaceProvisioner{}

// TypesWorkspaceProvisioner struct for TypesWorkspaceProvisioner
type TypesWorkspaceProvisioner struct {
	Name *string `json:"name,omitempty"`
	Profile *string `json:"profile,omitempty"`
}

// NewTypesWorkspaceProvisioner instantiates a new TypesWorkspaceProvisioner object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewTypesWorkspaceProvisioner() *TypesWorkspaceProvisioner {
	this := TypesWorkspaceProvisioner{}
	return &this
}

// NewTypesWorkspaceProvisionerWithDefaults instantiates a new TypesWorkspaceProvisioner object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewTypesWorkspaceProvisionerWithDefaults() *TypesWorkspaceProvisioner {
	this := TypesWorkspaceProvisioner{}
	return &this
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *TypesWorkspaceProvisioner) GetName() string {
	if o == nil || IsNil(o.Name) {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TypesWorkspaceProvisioner) GetNameOk() (*string, bool) {
	if o == nil || IsNil(o.Name) {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *TypesWorkspaceProvisioner) HasName() bool {
	if o != nil && !IsNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *TypesWorkspaceProvisioner) SetName(v string) {
	o.Name = &v
}

// GetProfile returns the Profile field value if set, zero value otherwise.
func (o *TypesWorkspaceProvisioner) GetProfile() string {
	if o == nil || IsNil(o.Profile) {
		var ret string
		return ret
	}
	return *o.Profile
}

// GetProfileOk returns a tuple with the Profile field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TypesWorkspaceProvisioner) GetProfileOk() (*string, bool) {
	if o == nil || IsNil(o.Profile) {
		return nil, false
	}
	return o.Profile, true
}

// HasProfile returns a boolean if a field has been set.
func (o *TypesWorkspaceProvisioner) HasProfile() bool {
	if o != nil && !IsNil(o.Profile) {
		return true
	}

	return false
}

// SetProfile gets a reference to the given string and assigns it to the Profile field.
func (o *TypesWorkspaceProvisioner) SetProfile(v string) {
	o.Profile = &v
}

func (o TypesWorkspaceProvisioner) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o TypesWorkspaceProvisioner) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Name) {
		toSerialize["name"] = o.Name
	}
	if !IsNil(o.Profile) {
		toSerialize["profile"] = o.Profile
	}
	return toSerialize, nil
}

type NullableTypesWorkspaceProvisioner struct {
	value *TypesWorkspaceProvisioner
	isSet bool
}

func (v NullableTypesWorkspaceProvisioner) Get() *TypesWorkspaceProvisioner {
	return v.value
}

func (v *NullableTypesWorkspaceProvisioner) Set(val *TypesWorkspaceProvisioner) {
	v.value = val
	v.isSet = true
}

func (v NullableTypesWorkspaceProvisioner) IsSet() bool {
	return v.isSet
}

func (v *NullableTypesWorkspaceProvisioner) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableTypesWorkspaceProvisioner(val *TypesWorkspaceProvisioner) *NullableTypesWorkspaceProvisioner {
	return &NullableTypesWorkspaceProvisioner{value: val, isSet: true}
}

func (v NullableTypesWorkspaceProvisioner) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableTypesWorkspaceProvisioner) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


