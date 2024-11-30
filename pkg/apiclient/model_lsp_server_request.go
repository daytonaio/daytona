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

// checks if the LspServerRequest type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &LspServerRequest{}

// LspServerRequest struct for LspServerRequest
type LspServerRequest struct {
	LanguageId    string `json:"languageId"`
	PathToProject string `json:"pathToProject"`
}

type _LspServerRequest LspServerRequest

// NewLspServerRequest instantiates a new LspServerRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewLspServerRequest(languageId string, pathToProject string) *LspServerRequest {
	this := LspServerRequest{}
	this.LanguageId = languageId
	this.PathToProject = pathToProject
	return &this
}

// NewLspServerRequestWithDefaults instantiates a new LspServerRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewLspServerRequestWithDefaults() *LspServerRequest {
	this := LspServerRequest{}
	return &this
}

// GetLanguageId returns the LanguageId field value
func (o *LspServerRequest) GetLanguageId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.LanguageId
}

// GetLanguageIdOk returns a tuple with the LanguageId field value
// and a boolean to check if the value has been set.
func (o *LspServerRequest) GetLanguageIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.LanguageId, true
}

// SetLanguageId sets field value
func (o *LspServerRequest) SetLanguageId(v string) {
	o.LanguageId = v
}

// GetPathToProject returns the PathToProject field value
func (o *LspServerRequest) GetPathToProject() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.PathToProject
}

// GetPathToProjectOk returns a tuple with the PathToProject field value
// and a boolean to check if the value has been set.
func (o *LspServerRequest) GetPathToProjectOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.PathToProject, true
}

// SetPathToProject sets field value
func (o *LspServerRequest) SetPathToProject(v string) {
	o.PathToProject = v
}

func (o LspServerRequest) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o LspServerRequest) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["languageId"] = o.LanguageId
	toSerialize["pathToProject"] = o.PathToProject
	return toSerialize, nil
}

func (o *LspServerRequest) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"languageId",
		"pathToProject",
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

	varLspServerRequest := _LspServerRequest{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varLspServerRequest)

	if err != nil {
		return err
	}

	*o = LspServerRequest(varLspServerRequest)

	return err
}

type NullableLspServerRequest struct {
	value *LspServerRequest
	isSet bool
}

func (v NullableLspServerRequest) Get() *LspServerRequest {
	return v.value
}

func (v *NullableLspServerRequest) Set(val *LspServerRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableLspServerRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableLspServerRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableLspServerRequest(val *LspServerRequest) *NullableLspServerRequest {
	return &NullableLspServerRequest{value: val, isSet: true}
}

func (v NullableLspServerRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableLspServerRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
