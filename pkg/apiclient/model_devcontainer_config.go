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

// checks if the DevcontainerConfig type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &DevcontainerConfig{}

// DevcontainerConfig struct for DevcontainerConfig
type DevcontainerConfig struct {
	FilePath string `json:"filePath"`
}

type _DevcontainerConfig DevcontainerConfig

// NewDevcontainerConfig instantiates a new DevcontainerConfig object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewDevcontainerConfig(filePath string) *DevcontainerConfig {
	this := DevcontainerConfig{}
	this.FilePath = filePath
	return &this
}

// NewDevcontainerConfigWithDefaults instantiates a new DevcontainerConfig object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewDevcontainerConfigWithDefaults() *DevcontainerConfig {
	this := DevcontainerConfig{}
	return &this
}

// GetFilePath returns the FilePath field value
func (o *DevcontainerConfig) GetFilePath() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.FilePath
}

// GetFilePathOk returns a tuple with the FilePath field value
// and a boolean to check if the value has been set.
func (o *DevcontainerConfig) GetFilePathOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.FilePath, true
}

// SetFilePath sets field value
func (o *DevcontainerConfig) SetFilePath(v string) {
	o.FilePath = v
}

func (o DevcontainerConfig) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o DevcontainerConfig) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["filePath"] = o.FilePath
	return toSerialize, nil
}

func (o *DevcontainerConfig) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"filePath",
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

	varDevcontainerConfig := _DevcontainerConfig{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varDevcontainerConfig)

	if err != nil {
		return err
	}

	*o = DevcontainerConfig(varDevcontainerConfig)

	return err
}

type NullableDevcontainerConfig struct {
	value *DevcontainerConfig
	isSet bool
}

func (v NullableDevcontainerConfig) Get() *DevcontainerConfig {
	return v.value
}

func (v *NullableDevcontainerConfig) Set(val *DevcontainerConfig) {
	v.value = val
	v.isSet = true
}

func (v NullableDevcontainerConfig) IsSet() bool {
	return v.isSet
}

func (v *NullableDevcontainerConfig) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableDevcontainerConfig(val *DevcontainerConfig) *NullableDevcontainerConfig {
	return &NullableDevcontainerConfig{value: val, isSet: true}
}

func (v NullableDevcontainerConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableDevcontainerConfig) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
