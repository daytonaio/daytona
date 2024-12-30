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

// checks if the CreateProjectSourceDTO type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &CreateProjectSourceDTO{}

// CreateProjectSourceDTO struct for CreateProjectSourceDTO
type CreateProjectSourceDTO struct {
	Repository GitRepository `json:"repository"`
}

type _CreateProjectSourceDTO CreateProjectSourceDTO

// NewCreateProjectSourceDTO instantiates a new CreateProjectSourceDTO object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewCreateProjectSourceDTO(repository GitRepository) *CreateProjectSourceDTO {
	this := CreateProjectSourceDTO{}
	this.Repository = repository
	return &this
}

// NewCreateProjectSourceDTOWithDefaults instantiates a new CreateProjectSourceDTO object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewCreateProjectSourceDTOWithDefaults() *CreateProjectSourceDTO {
	this := CreateProjectSourceDTO{}
	return &this
}

// GetRepository returns the Repository field value
func (o *CreateProjectSourceDTO) GetRepository() GitRepository {
	if o == nil {
		var ret GitRepository
		return ret
	}

	return o.Repository
}

// GetRepositoryOk returns a tuple with the Repository field value
// and a boolean to check if the value has been set.
func (o *CreateProjectSourceDTO) GetRepositoryOk() (*GitRepository, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Repository, true
}

// SetRepository sets field value
func (o *CreateProjectSourceDTO) SetRepository(v GitRepository) {
	o.Repository = v
}

func (o CreateProjectSourceDTO) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o CreateProjectSourceDTO) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["repository"] = o.Repository
	return toSerialize, nil
}

func (o *CreateProjectSourceDTO) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"repository",
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

	varCreateProjectSourceDTO := _CreateProjectSourceDTO{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varCreateProjectSourceDTO)

	if err != nil {
		return err
	}

	*o = CreateProjectSourceDTO(varCreateProjectSourceDTO)

	return err
}

type NullableCreateProjectSourceDTO struct {
	value *CreateProjectSourceDTO
	isSet bool
}

func (v NullableCreateProjectSourceDTO) Get() *CreateProjectSourceDTO {
	return v.value
}

func (v *NullableCreateProjectSourceDTO) Set(val *CreateProjectSourceDTO) {
	v.value = val
	v.isSet = true
}

func (v NullableCreateProjectSourceDTO) IsSet() bool {
	return v.isSet
}

func (v *NullableCreateProjectSourceDTO) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableCreateProjectSourceDTO(val *CreateProjectSourceDTO) *NullableCreateProjectSourceDTO {
	return &NullableCreateProjectSourceDTO{value: val, isSet: true}
}

func (v NullableCreateProjectSourceDTO) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableCreateProjectSourceDTO) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
