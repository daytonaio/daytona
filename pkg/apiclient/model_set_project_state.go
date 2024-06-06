/*
Daytona Server API

Daytona Server API

API version: 0.1.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package apiclient

import (
	"encoding/json"
)

// checks if the SetProjectState type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &SetProjectState{}

// SetProjectState struct for SetProjectState
type SetProjectState struct {
	GitStatus *GitStatus `json:"gitStatus,omitempty"`
	Uptime    *int32     `json:"uptime,omitempty"`
}

// NewSetProjectState instantiates a new SetProjectState object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewSetProjectState() *SetProjectState {
	this := SetProjectState{}
	return &this
}

// NewSetProjectStateWithDefaults instantiates a new SetProjectState object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewSetProjectStateWithDefaults() *SetProjectState {
	this := SetProjectState{}
	return &this
}

// GetGitStatus returns the GitStatus field value if set, zero value otherwise.
func (o *SetProjectState) GetGitStatus() GitStatus {
	if o == nil || IsNil(o.GitStatus) {
		var ret GitStatus
		return ret
	}
	return *o.GitStatus
}

// GetGitStatusOk returns a tuple with the GitStatus field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *SetProjectState) GetGitStatusOk() (*GitStatus, bool) {
	if o == nil || IsNil(o.GitStatus) {
		return nil, false
	}
	return o.GitStatus, true
}

// HasGitStatus returns a boolean if a field has been set.
func (o *SetProjectState) HasGitStatus() bool {
	if o != nil && !IsNil(o.GitStatus) {
		return true
	}

	return false
}

// SetGitStatus gets a reference to the given GitStatus and assigns it to the GitStatus field.
func (o *SetProjectState) SetGitStatus(v GitStatus) {
	o.GitStatus = &v
}

// GetUptime returns the Uptime field value if set, zero value otherwise.
func (o *SetProjectState) GetUptime() int32 {
	if o == nil || IsNil(o.Uptime) {
		var ret int32
		return ret
	}
	return *o.Uptime
}

// GetUptimeOk returns a tuple with the Uptime field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *SetProjectState) GetUptimeOk() (*int32, bool) {
	if o == nil || IsNil(o.Uptime) {
		return nil, false
	}
	return o.Uptime, true
}

// HasUptime returns a boolean if a field has been set.
func (o *SetProjectState) HasUptime() bool {
	if o != nil && !IsNil(o.Uptime) {
		return true
	}

	return false
}

// SetUptime gets a reference to the given int32 and assigns it to the Uptime field.
func (o *SetProjectState) SetUptime(v int32) {
	o.Uptime = &v
}

func (o SetProjectState) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o SetProjectState) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.GitStatus) {
		toSerialize["gitStatus"] = o.GitStatus
	}
	if !IsNil(o.Uptime) {
		toSerialize["uptime"] = o.Uptime
	}
	return toSerialize, nil
}

type NullableSetProjectState struct {
	value *SetProjectState
	isSet bool
}

func (v NullableSetProjectState) Get() *SetProjectState {
	return v.value
}

func (v *NullableSetProjectState) Set(val *SetProjectState) {
	v.value = val
	v.isSet = true
}

func (v NullableSetProjectState) IsSet() bool {
	return v.isSet
}

func (v *NullableSetProjectState) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableSetProjectState(val *SetProjectState) *NullableSetProjectState {
	return &NullableSetProjectState{value: val, isSet: true}
}

func (v NullableSetProjectState) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableSetProjectState) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
