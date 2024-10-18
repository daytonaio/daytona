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

// checks if the CreateProviderTargetDTO type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &CreateProviderTargetDTO{}

// CreateProviderTargetDTO struct for CreateProviderTargetDTO
type CreateProviderTargetDTO struct {
	Name         string               `json:"name"`
	Options      string               `json:"options"`
	ProviderInfo ProviderProviderInfo `json:"providerInfo"`
}

type _CreateProviderTargetDTO CreateProviderTargetDTO

// NewCreateProviderTargetDTO instantiates a new CreateProviderTargetDTO object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewCreateProviderTargetDTO(name string, options string, providerInfo ProviderProviderInfo) *CreateProviderTargetDTO {
	this := CreateProviderTargetDTO{}
	this.Name = name
	this.Options = options
	this.ProviderInfo = providerInfo
	return &this
}

// NewCreateProviderTargetDTOWithDefaults instantiates a new CreateProviderTargetDTO object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewCreateProviderTargetDTOWithDefaults() *CreateProviderTargetDTO {
	this := CreateProviderTargetDTO{}
	return &this
}

// GetName returns the Name field value
func (o *CreateProviderTargetDTO) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *CreateProviderTargetDTO) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *CreateProviderTargetDTO) SetName(v string) {
	o.Name = v
}

// GetOptions returns the Options field value
func (o *CreateProviderTargetDTO) GetOptions() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Options
}

// GetOptionsOk returns a tuple with the Options field value
// and a boolean to check if the value has been set.
func (o *CreateProviderTargetDTO) GetOptionsOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Options, true
}

// SetOptions sets field value
func (o *CreateProviderTargetDTO) SetOptions(v string) {
	o.Options = v
}

// GetProviderInfo returns the ProviderInfo field value
func (o *CreateProviderTargetDTO) GetProviderInfo() ProviderProviderInfo {
	if o == nil {
		var ret ProviderProviderInfo
		return ret
	}

	return o.ProviderInfo
}

// GetProviderInfoOk returns a tuple with the ProviderInfo field value
// and a boolean to check if the value has been set.
func (o *CreateProviderTargetDTO) GetProviderInfoOk() (*ProviderProviderInfo, bool) {
	if o == nil {
		return nil, false
	}
	return &o.ProviderInfo, true
}

// SetProviderInfo sets field value
func (o *CreateProviderTargetDTO) SetProviderInfo(v ProviderProviderInfo) {
	o.ProviderInfo = v
}

func (o CreateProviderTargetDTO) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o CreateProviderTargetDTO) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["name"] = o.Name
	toSerialize["options"] = o.Options
	toSerialize["providerInfo"] = o.ProviderInfo
	return toSerialize, nil
}

func (o *CreateProviderTargetDTO) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"name",
		"options",
		"providerInfo",
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

	varCreateProviderTargetDTO := _CreateProviderTargetDTO{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varCreateProviderTargetDTO)

	if err != nil {
		return err
	}

	*o = CreateProviderTargetDTO(varCreateProviderTargetDTO)

	return err
}

type NullableCreateProviderTargetDTO struct {
	value *CreateProviderTargetDTO
	isSet bool
}

func (v NullableCreateProviderTargetDTO) Get() *CreateProviderTargetDTO {
	return v.value
}

func (v *NullableCreateProviderTargetDTO) Set(val *CreateProviderTargetDTO) {
	v.value = val
	v.isSet = true
}

func (v NullableCreateProviderTargetDTO) IsSet() bool {
	return v.isSet
}

func (v *NullableCreateProviderTargetDTO) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableCreateProviderTargetDTO(val *CreateProviderTargetDTO) *NullableCreateProviderTargetDTO {
	return &NullableCreateProviderTargetDTO{value: val, isSet: true}
}

func (v NullableCreateProviderTargetDTO) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableCreateProviderTargetDTO) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
