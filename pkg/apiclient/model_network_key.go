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

// checks if the NetworkKey type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &NetworkKey{}

// NetworkKey struct for NetworkKey
type NetworkKey struct {
	Key string `json:"key"`
}

type _NetworkKey NetworkKey

// NewNetworkKey instantiates a new NetworkKey object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewNetworkKey(key string) *NetworkKey {
	this := NetworkKey{}
	this.Key = key
	return &this
}

// NewNetworkKeyWithDefaults instantiates a new NetworkKey object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewNetworkKeyWithDefaults() *NetworkKey {
	this := NetworkKey{}
	return &this
}

// GetKey returns the Key field value
func (o *NetworkKey) GetKey() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Key
}

// GetKeyOk returns a tuple with the Key field value
// and a boolean to check if the value has been set.
func (o *NetworkKey) GetKeyOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Key, true
}

// SetKey sets field value
func (o *NetworkKey) SetKey(v string) {
	o.Key = v
}

func (o NetworkKey) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o NetworkKey) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["key"] = o.Key
	return toSerialize, nil
}

func (o *NetworkKey) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"key",
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

	varNetworkKey := _NetworkKey{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varNetworkKey)

	if err != nil {
		return err
	}

	*o = NetworkKey(varNetworkKey)

	return err
}

type NullableNetworkKey struct {
	value *NetworkKey
	isSet bool
}

func (v NullableNetworkKey) Get() *NetworkKey {
	return v.value
}

func (v *NullableNetworkKey) Set(val *NetworkKey) {
	v.value = val
	v.isSet = true
}

func (v NullableNetworkKey) IsSet() bool {
	return v.isSet
}

func (v *NullableNetworkKey) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableNetworkKey(val *NetworkKey) *NullableNetworkKey {
	return &NullableNetworkKey{value: val, isSet: true}
}

func (v NullableNetworkKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableNetworkKey) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
