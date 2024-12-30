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

// checks if the LspRange type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &LspRange{}

// LspRange struct for LspRange
type LspRange struct {
	End   LspPosition `json:"end"`
	Start LspPosition `json:"start"`
}

type _LspRange LspRange

// NewLspRange instantiates a new LspRange object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewLspRange(end LspPosition, start LspPosition) *LspRange {
	this := LspRange{}
	this.End = end
	this.Start = start
	return &this
}

// NewLspRangeWithDefaults instantiates a new LspRange object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewLspRangeWithDefaults() *LspRange {
	this := LspRange{}
	return &this
}

// GetEnd returns the End field value
func (o *LspRange) GetEnd() LspPosition {
	if o == nil {
		var ret LspPosition
		return ret
	}

	return o.End
}

// GetEndOk returns a tuple with the End field value
// and a boolean to check if the value has been set.
func (o *LspRange) GetEndOk() (*LspPosition, bool) {
	if o == nil {
		return nil, false
	}
	return &o.End, true
}

// SetEnd sets field value
func (o *LspRange) SetEnd(v LspPosition) {
	o.End = v
}

// GetStart returns the Start field value
func (o *LspRange) GetStart() LspPosition {
	if o == nil {
		var ret LspPosition
		return ret
	}

	return o.Start
}

// GetStartOk returns a tuple with the Start field value
// and a boolean to check if the value has been set.
func (o *LspRange) GetStartOk() (*LspPosition, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Start, true
}

// SetStart sets field value
func (o *LspRange) SetStart(v LspPosition) {
	o.Start = v
}

func (o LspRange) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o LspRange) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["end"] = o.End
	toSerialize["start"] = o.Start
	return toSerialize, nil
}

func (o *LspRange) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"end",
		"start",
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

	varLspRange := _LspRange{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varLspRange)

	if err != nil {
		return err
	}

	*o = LspRange(varLspRange)

	return err
}

type NullableLspRange struct {
	value *LspRange
	isSet bool
}

func (v NullableLspRange) Get() *LspRange {
	return v.value
}

func (v *NullableLspRange) Set(val *LspRange) {
	v.value = val
	v.isSet = true
}

func (v NullableLspRange) IsSet() bool {
	return v.isSet
}

func (v *NullableLspRange) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableLspRange(val *LspRange) *NullableLspRange {
	return &NullableLspRange{value: val, isSet: true}
}

func (v NullableLspRange) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableLspRange) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
