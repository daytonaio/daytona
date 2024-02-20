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

// checks if the DtoAgentServicePlugin type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &DtoAgentServicePlugin{}

// DtoAgentServicePlugin struct for DtoAgentServicePlugin
type DtoAgentServicePlugin struct {
	Name *string `json:"name,omitempty"`
	Version *string `json:"version,omitempty"`
}

// NewDtoAgentServicePlugin instantiates a new DtoAgentServicePlugin object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewDtoAgentServicePlugin() *DtoAgentServicePlugin {
	this := DtoAgentServicePlugin{}
	return &this
}

// NewDtoAgentServicePluginWithDefaults instantiates a new DtoAgentServicePlugin object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewDtoAgentServicePluginWithDefaults() *DtoAgentServicePlugin {
	this := DtoAgentServicePlugin{}
	return &this
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *DtoAgentServicePlugin) GetName() string {
	if o == nil || IsNil(o.Name) {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *DtoAgentServicePlugin) GetNameOk() (*string, bool) {
	if o == nil || IsNil(o.Name) {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *DtoAgentServicePlugin) HasName() bool {
	if o != nil && !IsNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *DtoAgentServicePlugin) SetName(v string) {
	o.Name = &v
}

// GetVersion returns the Version field value if set, zero value otherwise.
func (o *DtoAgentServicePlugin) GetVersion() string {
	if o == nil || IsNil(o.Version) {
		var ret string
		return ret
	}
	return *o.Version
}

// GetVersionOk returns a tuple with the Version field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *DtoAgentServicePlugin) GetVersionOk() (*string, bool) {
	if o == nil || IsNil(o.Version) {
		return nil, false
	}
	return o.Version, true
}

// HasVersion returns a boolean if a field has been set.
func (o *DtoAgentServicePlugin) HasVersion() bool {
	if o != nil && !IsNil(o.Version) {
		return true
	}

	return false
}

// SetVersion gets a reference to the given string and assigns it to the Version field.
func (o *DtoAgentServicePlugin) SetVersion(v string) {
	o.Version = &v
}

func (o DtoAgentServicePlugin) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o DtoAgentServicePlugin) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Name) {
		toSerialize["name"] = o.Name
	}
	if !IsNil(o.Version) {
		toSerialize["version"] = o.Version
	}
	return toSerialize, nil
}

type NullableDtoAgentServicePlugin struct {
	value *DtoAgentServicePlugin
	isSet bool
}

func (v NullableDtoAgentServicePlugin) Get() *DtoAgentServicePlugin {
	return v.value
}

func (v *NullableDtoAgentServicePlugin) Set(val *DtoAgentServicePlugin) {
	v.value = val
	v.isSet = true
}

func (v NullableDtoAgentServicePlugin) IsSet() bool {
	return v.isSet
}

func (v *NullableDtoAgentServicePlugin) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableDtoAgentServicePlugin(val *DtoAgentServicePlugin) *NullableDtoAgentServicePlugin {
	return &NullableDtoAgentServicePlugin{value: val, isSet: true}
}

func (v NullableDtoAgentServicePlugin) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableDtoAgentServicePlugin) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


