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

// checks if the ProjectInfo type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &ProjectInfo{}

// ProjectInfo struct for ProjectInfo
type ProjectInfo struct {
	Created          string  `json:"created"`
	IsRunning        bool    `json:"isRunning"`
	Name             string  `json:"name"`
	ProviderMetadata *string `json:"providerMetadata,omitempty"`
	TargetId         string  `json:"targetId"`
}

type _ProjectInfo ProjectInfo

// NewProjectInfo instantiates a new ProjectInfo object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewProjectInfo(created string, isRunning bool, name string, targetId string) *ProjectInfo {
	this := ProjectInfo{}
	this.Created = created
	this.IsRunning = isRunning
	this.Name = name
	this.TargetId = targetId
	return &this
}

// NewProjectInfoWithDefaults instantiates a new ProjectInfo object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewProjectInfoWithDefaults() *ProjectInfo {
	this := ProjectInfo{}
	return &this
}

// GetCreated returns the Created field value
func (o *ProjectInfo) GetCreated() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Created
}

// GetCreatedOk returns a tuple with the Created field value
// and a boolean to check if the value has been set.
func (o *ProjectInfo) GetCreatedOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Created, true
}

// SetCreated sets field value
func (o *ProjectInfo) SetCreated(v string) {
	o.Created = v
}

// GetIsRunning returns the IsRunning field value
func (o *ProjectInfo) GetIsRunning() bool {
	if o == nil {
		var ret bool
		return ret
	}

	return o.IsRunning
}

// GetIsRunningOk returns a tuple with the IsRunning field value
// and a boolean to check if the value has been set.
func (o *ProjectInfo) GetIsRunningOk() (*bool, bool) {
	if o == nil {
		return nil, false
	}
	return &o.IsRunning, true
}

// SetIsRunning sets field value
func (o *ProjectInfo) SetIsRunning(v bool) {
	o.IsRunning = v
}

// GetName returns the Name field value
func (o *ProjectInfo) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *ProjectInfo) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *ProjectInfo) SetName(v string) {
	o.Name = v
}

// GetProviderMetadata returns the ProviderMetadata field value if set, zero value otherwise.
func (o *ProjectInfo) GetProviderMetadata() string {
	if o == nil || IsNil(o.ProviderMetadata) {
		var ret string
		return ret
	}
	return *o.ProviderMetadata
}

// GetProviderMetadataOk returns a tuple with the ProviderMetadata field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProjectInfo) GetProviderMetadataOk() (*string, bool) {
	if o == nil || IsNil(o.ProviderMetadata) {
		return nil, false
	}
	return o.ProviderMetadata, true
}

// HasProviderMetadata returns a boolean if a field has been set.
func (o *ProjectInfo) HasProviderMetadata() bool {
	if o != nil && !IsNil(o.ProviderMetadata) {
		return true
	}

	return false
}

// SetProviderMetadata gets a reference to the given string and assigns it to the ProviderMetadata field.
func (o *ProjectInfo) SetProviderMetadata(v string) {
	o.ProviderMetadata = &v
}

// GetTargetId returns the TargetId field value
func (o *ProjectInfo) GetTargetId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.TargetId
}

// GetTargetIdOk returns a tuple with the TargetId field value
// and a boolean to check if the value has been set.
func (o *ProjectInfo) GetTargetIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.TargetId, true
}

// SetTargetId sets field value
func (o *ProjectInfo) SetTargetId(v string) {
	o.TargetId = v
}

func (o ProjectInfo) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o ProjectInfo) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["created"] = o.Created
	toSerialize["isRunning"] = o.IsRunning
	toSerialize["name"] = o.Name
	if !IsNil(o.ProviderMetadata) {
		toSerialize["providerMetadata"] = o.ProviderMetadata
	}
	toSerialize["targetId"] = o.TargetId
	return toSerialize, nil
}

func (o *ProjectInfo) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"created",
		"isRunning",
		"name",
		"targetId",
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

	varProjectInfo := _ProjectInfo{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varProjectInfo)

	if err != nil {
		return err
	}

	*o = ProjectInfo(varProjectInfo)

	return err
}

type NullableProjectInfo struct {
	value *ProjectInfo
	isSet bool
}

func (v NullableProjectInfo) Get() *ProjectInfo {
	return v.value
}

func (v *NullableProjectInfo) Set(val *ProjectInfo) {
	v.value = val
	v.isSet = true
}

func (v NullableProjectInfo) IsSet() bool {
	return v.isSet
}

func (v *NullableProjectInfo) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableProjectInfo(val *ProjectInfo) *NullableProjectInfo {
	return &NullableProjectInfo{value: val, isSet: true}
}

func (v NullableProjectInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableProjectInfo) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
