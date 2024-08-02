/*
Daytona Server API

Daytona Server API

API version: 0.24.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package apiclient

import (
	"encoding/json"
)

// checks if the ApiKey type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &ApiKey{}

// ApiKey struct for ApiKey
type ApiKey struct {
	KeyHash *string `json:"keyHash,omitempty"`
	// Project or client name
	Name *string           `json:"name,omitempty"`
	Type *ApikeyApiKeyType `json:"type,omitempty"`
}

// NewApiKey instantiates a new ApiKey object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewApiKey() *ApiKey {
	this := ApiKey{}
	return &this
}

// NewApiKeyWithDefaults instantiates a new ApiKey object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewApiKeyWithDefaults() *ApiKey {
	this := ApiKey{}
	return &this
}

// GetKeyHash returns the KeyHash field value if set, zero value otherwise.
func (o *ApiKey) GetKeyHash() string {
	if o == nil || IsNil(o.KeyHash) {
		var ret string
		return ret
	}
	return *o.KeyHash
}

// GetKeyHashOk returns a tuple with the KeyHash field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ApiKey) GetKeyHashOk() (*string, bool) {
	if o == nil || IsNil(o.KeyHash) {
		return nil, false
	}
	return o.KeyHash, true
}

// HasKeyHash returns a boolean if a field has been set.
func (o *ApiKey) HasKeyHash() bool {
	if o != nil && !IsNil(o.KeyHash) {
		return true
	}

	return false
}

// SetKeyHash gets a reference to the given string and assigns it to the KeyHash field.
func (o *ApiKey) SetKeyHash(v string) {
	o.KeyHash = &v
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *ApiKey) GetName() string {
	if o == nil || IsNil(o.Name) {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ApiKey) GetNameOk() (*string, bool) {
	if o == nil || IsNil(o.Name) {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *ApiKey) HasName() bool {
	if o != nil && !IsNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *ApiKey) SetName(v string) {
	o.Name = &v
}

// GetType returns the Type field value if set, zero value otherwise.
func (o *ApiKey) GetType() ApikeyApiKeyType {
	if o == nil || IsNil(o.Type) {
		var ret ApikeyApiKeyType
		return ret
	}
	return *o.Type
}

// GetTypeOk returns a tuple with the Type field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ApiKey) GetTypeOk() (*ApikeyApiKeyType, bool) {
	if o == nil || IsNil(o.Type) {
		return nil, false
	}
	return o.Type, true
}

// HasType returns a boolean if a field has been set.
func (o *ApiKey) HasType() bool {
	if o != nil && !IsNil(o.Type) {
		return true
	}

	return false
}

// SetType gets a reference to the given ApikeyApiKeyType and assigns it to the Type field.
func (o *ApiKey) SetType(v ApikeyApiKeyType) {
	o.Type = &v
}

func (o ApiKey) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o ApiKey) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.KeyHash) {
		toSerialize["keyHash"] = o.KeyHash
	}
	if !IsNil(o.Name) {
		toSerialize["name"] = o.Name
	}
	if !IsNil(o.Type) {
		toSerialize["type"] = o.Type
	}
	return toSerialize, nil
}

type NullableApiKey struct {
	value *ApiKey
	isSet bool
}

func (v NullableApiKey) Get() *ApiKey {
	return v.value
}

func (v *NullableApiKey) Set(val *ApiKey) {
	v.value = val
	v.isSet = true
}

func (v NullableApiKey) IsSet() bool {
	return v.isSet
}

func (v *NullableApiKey) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableApiKey(val *ApiKey) *NullableApiKey {
	return &NullableApiKey{value: val, isSet: true}
}

func (v NullableApiKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableApiKey) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
