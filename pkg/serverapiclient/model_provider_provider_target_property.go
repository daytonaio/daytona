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

// checks if the ProviderProviderTargetProperty type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &ProviderProviderTargetProperty{}

// ProviderProviderTargetProperty struct for ProviderProviderTargetProperty
type ProviderProviderTargetProperty struct {
	// DefaultValue is converted into the appropriate type based on the Type
	DefaultValue *string `json:"defaultValue,omitempty"`
	// Options is only used if the Type is ProviderTargetPropertyTypeOption
	Options []string `json:"options,omitempty"`
	Type *ProviderProviderTargetPropertyType `json:"type,omitempty"`
}

// NewProviderProviderTargetProperty instantiates a new ProviderProviderTargetProperty object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewProviderProviderTargetProperty() *ProviderProviderTargetProperty {
	this := ProviderProviderTargetProperty{}
	return &this
}

// NewProviderProviderTargetPropertyWithDefaults instantiates a new ProviderProviderTargetProperty object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewProviderProviderTargetPropertyWithDefaults() *ProviderProviderTargetProperty {
	this := ProviderProviderTargetProperty{}
	return &this
}

// GetDefaultValue returns the DefaultValue field value if set, zero value otherwise.
func (o *ProviderProviderTargetProperty) GetDefaultValue() string {
	if o == nil || IsNil(o.DefaultValue) {
		var ret string
		return ret
	}
	return *o.DefaultValue
}

// GetDefaultValueOk returns a tuple with the DefaultValue field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProviderProviderTargetProperty) GetDefaultValueOk() (*string, bool) {
	if o == nil || IsNil(o.DefaultValue) {
		return nil, false
	}
	return o.DefaultValue, true
}

// HasDefaultValue returns a boolean if a field has been set.
func (o *ProviderProviderTargetProperty) HasDefaultValue() bool {
	if o != nil && !IsNil(o.DefaultValue) {
		return true
	}

	return false
}

// SetDefaultValue gets a reference to the given string and assigns it to the DefaultValue field.
func (o *ProviderProviderTargetProperty) SetDefaultValue(v string) {
	o.DefaultValue = &v
}

// GetOptions returns the Options field value if set, zero value otherwise.
func (o *ProviderProviderTargetProperty) GetOptions() []string {
	if o == nil || IsNil(o.Options) {
		var ret []string
		return ret
	}
	return o.Options
}

// GetOptionsOk returns a tuple with the Options field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProviderProviderTargetProperty) GetOptionsOk() ([]string, bool) {
	if o == nil || IsNil(o.Options) {
		return nil, false
	}
	return o.Options, true
}

// HasOptions returns a boolean if a field has been set.
func (o *ProviderProviderTargetProperty) HasOptions() bool {
	if o != nil && !IsNil(o.Options) {
		return true
	}

	return false
}

// SetOptions gets a reference to the given []string and assigns it to the Options field.
func (o *ProviderProviderTargetProperty) SetOptions(v []string) {
	o.Options = v
}

// GetType returns the Type field value if set, zero value otherwise.
func (o *ProviderProviderTargetProperty) GetType() ProviderProviderTargetPropertyType {
	if o == nil || IsNil(o.Type) {
		var ret ProviderProviderTargetPropertyType
		return ret
	}
	return *o.Type
}

// GetTypeOk returns a tuple with the Type field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProviderProviderTargetProperty) GetTypeOk() (*ProviderProviderTargetPropertyType, bool) {
	if o == nil || IsNil(o.Type) {
		return nil, false
	}
	return o.Type, true
}

// HasType returns a boolean if a field has been set.
func (o *ProviderProviderTargetProperty) HasType() bool {
	if o != nil && !IsNil(o.Type) {
		return true
	}

	return false
}

// SetType gets a reference to the given ProviderProviderTargetPropertyType and assigns it to the Type field.
func (o *ProviderProviderTargetProperty) SetType(v ProviderProviderTargetPropertyType) {
	o.Type = &v
}

func (o ProviderProviderTargetProperty) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o ProviderProviderTargetProperty) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.DefaultValue) {
		toSerialize["defaultValue"] = o.DefaultValue
	}
	if !IsNil(o.Options) {
		toSerialize["options"] = o.Options
	}
	if !IsNil(o.Type) {
		toSerialize["type"] = o.Type
	}
	return toSerialize, nil
}

type NullableProviderProviderTargetProperty struct {
	value *ProviderProviderTargetProperty
	isSet bool
}

func (v NullableProviderProviderTargetProperty) Get() *ProviderProviderTargetProperty {
	return v.value
}

func (v *NullableProviderProviderTargetProperty) Set(val *ProviderProviderTargetProperty) {
	v.value = val
	v.isSet = true
}

func (v NullableProviderProviderTargetProperty) IsSet() bool {
	return v.isSet
}

func (v *NullableProviderProviderTargetProperty) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableProviderProviderTargetProperty(val *ProviderProviderTargetProperty) *NullableProviderProviderTargetProperty {
	return &NullableProviderProviderTargetProperty{value: val, isSet: true}
}

func (v NullableProviderProviderTargetProperty) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableProviderProviderTargetProperty) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


