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

// checks if the GitNamespace type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &GitNamespace{}

// GitNamespace struct for GitNamespace
type GitNamespace struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type _GitNamespace GitNamespace

// NewGitNamespace instantiates a new GitNamespace object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewGitNamespace(id string, name string) *GitNamespace {
	this := GitNamespace{}
	this.Id = id
	this.Name = name
	return &this
}

// NewGitNamespaceWithDefaults instantiates a new GitNamespace object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewGitNamespaceWithDefaults() *GitNamespace {
	this := GitNamespace{}
	return &this
}

// GetId returns the Id field value
func (o *GitNamespace) GetId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Id
}

// GetIdOk returns a tuple with the Id field value
// and a boolean to check if the value has been set.
func (o *GitNamespace) GetIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Id, true
}

// SetId sets field value
func (o *GitNamespace) SetId(v string) {
	o.Id = v
}

// GetName returns the Name field value
func (o *GitNamespace) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *GitNamespace) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *GitNamespace) SetName(v string) {
	o.Name = v
}

func (o GitNamespace) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o GitNamespace) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["id"] = o.Id
	toSerialize["name"] = o.Name
	return toSerialize, nil
}

func (o *GitNamespace) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"id",
		"name",
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

	varGitNamespace := _GitNamespace{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varGitNamespace)

	if err != nil {
		return err
	}

	*o = GitNamespace(varGitNamespace)

	return err
}

type NullableGitNamespace struct {
	value *GitNamespace
	isSet bool
}

func (v NullableGitNamespace) Get() *GitNamespace {
	return v.value
}

func (v *NullableGitNamespace) Set(val *GitNamespace) {
	v.value = val
	v.isSet = true
}

func (v NullableGitNamespace) IsSet() bool {
	return v.isSet
}

func (v *NullableGitNamespace) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableGitNamespace(val *GitNamespace) *NullableGitNamespace {
	return &NullableGitNamespace{value: val, isSet: true}
}

func (v NullableGitNamespace) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableGitNamespace) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
