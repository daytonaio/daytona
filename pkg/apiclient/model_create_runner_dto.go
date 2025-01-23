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

// checks if the CreateRunnerDTO type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &CreateRunnerDTO{}

// CreateRunnerDTO struct for CreateRunnerDTO
type CreateRunnerDTO struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type _CreateRunnerDTO CreateRunnerDTO

// NewCreateRunnerDTO instantiates a new CreateRunnerDTO object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewCreateRunnerDTO(id string, name string) *CreateRunnerDTO {
	this := CreateRunnerDTO{}
	this.Id = id
	this.Name = name
	return &this
}

// NewCreateRunnerDTOWithDefaults instantiates a new CreateRunnerDTO object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewCreateRunnerDTOWithDefaults() *CreateRunnerDTO {
	this := CreateRunnerDTO{}
	return &this
}

// GetId returns the Id field value
func (o *CreateRunnerDTO) GetId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Id
}

// GetIdOk returns a tuple with the Id field value
// and a boolean to check if the value has been set.
func (o *CreateRunnerDTO) GetIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Id, true
}

// SetId sets field value
func (o *CreateRunnerDTO) SetId(v string) {
	o.Id = v
}

// GetName returns the Name field value
func (o *CreateRunnerDTO) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *CreateRunnerDTO) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *CreateRunnerDTO) SetName(v string) {
	o.Name = v
}

func (o CreateRunnerDTO) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o CreateRunnerDTO) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["id"] = o.Id
	toSerialize["name"] = o.Name
	return toSerialize, nil
}

func (o *CreateRunnerDTO) UnmarshalJSON(data []byte) (err error) {
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

	varCreateRunnerDTO := _CreateRunnerDTO{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varCreateRunnerDTO)

	if err != nil {
		return err
	}

	*o = CreateRunnerDTO(varCreateRunnerDTO)

	return err
}

type NullableCreateRunnerDTO struct {
	value *CreateRunnerDTO
	isSet bool
}

func (v NullableCreateRunnerDTO) Get() *CreateRunnerDTO {
	return v.value
}

func (v *NullableCreateRunnerDTO) Set(val *CreateRunnerDTO) {
	v.value = val
	v.isSet = true
}

func (v NullableCreateRunnerDTO) IsSet() bool {
	return v.isSet
}

func (v *NullableCreateRunnerDTO) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableCreateRunnerDTO(val *CreateRunnerDTO) *NullableCreateRunnerDTO {
	return &NullableCreateRunnerDTO{value: val, isSet: true}
}

func (v NullableCreateRunnerDTO) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableCreateRunnerDTO) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
