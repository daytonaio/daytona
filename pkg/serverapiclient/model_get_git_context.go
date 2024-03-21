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

// checks if the GetGitContext type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &GetGitContext{}

// GetGitContext struct for GetGitContext
type GetGitContext struct {
	Url *string `json:"url,omitempty"`
}

// NewGetGitContext instantiates a new GetGitContext object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewGetGitContext() *GetGitContext {
	this := GetGitContext{}
	return &this
}

// NewGetGitContextWithDefaults instantiates a new GetGitContext object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewGetGitContextWithDefaults() *GetGitContext {
	this := GetGitContext{}
	return &this
}

// GetUrl returns the Url field value if set, zero value otherwise.
func (o *GetGitContext) GetUrl() string {
	if o == nil || IsNil(o.Url) {
		var ret string
		return ret
	}
	return *o.Url
}

// GetUrlOk returns a tuple with the Url field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GetGitContext) GetUrlOk() (*string, bool) {
	if o == nil || IsNil(o.Url) {
		return nil, false
	}
	return o.Url, true
}

// HasUrl returns a boolean if a field has been set.
func (o *GetGitContext) HasUrl() bool {
	if o != nil && !IsNil(o.Url) {
		return true
	}

	return false
}

// SetUrl gets a reference to the given string and assigns it to the Url field.
func (o *GetGitContext) SetUrl(v string) {
	o.Url = &v
}

func (o GetGitContext) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o GetGitContext) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Url) {
		toSerialize["url"] = o.Url
	}
	return toSerialize, nil
}

type NullableGetGitContext struct {
	value *GetGitContext
	isSet bool
}

func (v NullableGetGitContext) Get() *GetGitContext {
	return v.value
}

func (v *NullableGetGitContext) Set(val *GetGitContext) {
	v.value = val
	v.isSet = true
}

func (v NullableGetGitContext) IsSet() bool {
	return v.isSet
}

func (v *NullableGetGitContext) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableGetGitContext(val *GetGitContext) *NullableGetGitContext {
	return &NullableGetGitContext{value: val, isSet: true}
}

func (v NullableGetGitContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableGetGitContext) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
