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

// checks if the GoogleGolangOrgProtobufTypesKnownStructpbValue type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &GoogleGolangOrgProtobufTypesKnownStructpbValue{}

// GoogleGolangOrgProtobufTypesKnownStructpbValue struct for GoogleGolangOrgProtobufTypesKnownStructpbValue
type GoogleGolangOrgProtobufTypesKnownStructpbValue struct {
	// The kind of value.  Types that are assignable to Kind:   *Value_NullValue  *Value_NumberValue  *Value_StringValue  *Value_BoolValue  *Value_StructValue  *Value_ListValue
	Kind map[string]interface{} `json:"kind,omitempty"`
}

// NewGoogleGolangOrgProtobufTypesKnownStructpbValue instantiates a new GoogleGolangOrgProtobufTypesKnownStructpbValue object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewGoogleGolangOrgProtobufTypesKnownStructpbValue() *GoogleGolangOrgProtobufTypesKnownStructpbValue {
	this := GoogleGolangOrgProtobufTypesKnownStructpbValue{}
	return &this
}

// NewGoogleGolangOrgProtobufTypesKnownStructpbValueWithDefaults instantiates a new GoogleGolangOrgProtobufTypesKnownStructpbValue object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewGoogleGolangOrgProtobufTypesKnownStructpbValueWithDefaults() *GoogleGolangOrgProtobufTypesKnownStructpbValue {
	this := GoogleGolangOrgProtobufTypesKnownStructpbValue{}
	return &this
}

// GetKind returns the Kind field value if set, zero value otherwise.
func (o *GoogleGolangOrgProtobufTypesKnownStructpbValue) GetKind() map[string]interface{} {
	if o == nil || IsNil(o.Kind) {
		var ret map[string]interface{}
		return ret
	}
	return o.Kind
}

// GetKindOk returns a tuple with the Kind field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GoogleGolangOrgProtobufTypesKnownStructpbValue) GetKindOk() (map[string]interface{}, bool) {
	if o == nil || IsNil(o.Kind) {
		return map[string]interface{}{}, false
	}
	return o.Kind, true
}

// HasKind returns a boolean if a field has been set.
func (o *GoogleGolangOrgProtobufTypesKnownStructpbValue) HasKind() bool {
	if o != nil && !IsNil(o.Kind) {
		return true
	}

	return false
}

// SetKind gets a reference to the given map[string]interface{} and assigns it to the Kind field.
func (o *GoogleGolangOrgProtobufTypesKnownStructpbValue) SetKind(v map[string]interface{}) {
	o.Kind = v
}

func (o GoogleGolangOrgProtobufTypesKnownStructpbValue) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o GoogleGolangOrgProtobufTypesKnownStructpbValue) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Kind) {
		toSerialize["kind"] = o.Kind
	}
	return toSerialize, nil
}

type NullableGoogleGolangOrgProtobufTypesKnownStructpbValue struct {
	value *GoogleGolangOrgProtobufTypesKnownStructpbValue
	isSet bool
}

func (v NullableGoogleGolangOrgProtobufTypesKnownStructpbValue) Get() *GoogleGolangOrgProtobufTypesKnownStructpbValue {
	return v.value
}

func (v *NullableGoogleGolangOrgProtobufTypesKnownStructpbValue) Set(val *GoogleGolangOrgProtobufTypesKnownStructpbValue) {
	v.value = val
	v.isSet = true
}

func (v NullableGoogleGolangOrgProtobufTypesKnownStructpbValue) IsSet() bool {
	return v.isSet
}

func (v *NullableGoogleGolangOrgProtobufTypesKnownStructpbValue) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableGoogleGolangOrgProtobufTypesKnownStructpbValue(val *GoogleGolangOrgProtobufTypesKnownStructpbValue) *NullableGoogleGolangOrgProtobufTypesKnownStructpbValue {
	return &NullableGoogleGolangOrgProtobufTypesKnownStructpbValue{value: val, isSet: true}
}

func (v NullableGoogleGolangOrgProtobufTypesKnownStructpbValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableGoogleGolangOrgProtobufTypesKnownStructpbValue) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


