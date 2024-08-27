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

// checks if the GitStatus type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &GitStatus{}

// GitStatus struct for GitStatus
type GitStatus struct {
	CurrentBranch string       `json:"currentBranch"`
	FileStatus    []FileStatus `json:"fileStatus"`
}

type _GitStatus GitStatus

// NewGitStatus instantiates a new GitStatus object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewGitStatus(currentBranch string, fileStatus []FileStatus) *GitStatus {
	this := GitStatus{}
	this.CurrentBranch = currentBranch
	this.FileStatus = fileStatus
	return &this
}

// NewGitStatusWithDefaults instantiates a new GitStatus object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewGitStatusWithDefaults() *GitStatus {
	this := GitStatus{}
	return &this
}

// GetCurrentBranch returns the CurrentBranch field value
func (o *GitStatus) GetCurrentBranch() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.CurrentBranch
}

// GetCurrentBranchOk returns a tuple with the CurrentBranch field value
// and a boolean to check if the value has been set.
func (o *GitStatus) GetCurrentBranchOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.CurrentBranch, true
}

// SetCurrentBranch sets field value
func (o *GitStatus) SetCurrentBranch(v string) {
	o.CurrentBranch = v
}

// GetFileStatus returns the FileStatus field value
func (o *GitStatus) GetFileStatus() []FileStatus {
	if o == nil {
		var ret []FileStatus
		return ret
	}

	return o.FileStatus
}

// GetFileStatusOk returns a tuple with the FileStatus field value
// and a boolean to check if the value has been set.
func (o *GitStatus) GetFileStatusOk() ([]FileStatus, bool) {
	if o == nil {
		return nil, false
	}
	return o.FileStatus, true
}

// SetFileStatus sets field value
func (o *GitStatus) SetFileStatus(v []FileStatus) {
	o.FileStatus = v
}

func (o GitStatus) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o GitStatus) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["currentBranch"] = o.CurrentBranch
	toSerialize["fileStatus"] = o.FileStatus
	return toSerialize, nil
}

func (o *GitStatus) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"currentBranch",
		"fileStatus",
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

	varGitStatus := _GitStatus{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varGitStatus)

	if err != nil {
		return err
	}

	*o = GitStatus(varGitStatus)

	return err
}

type NullableGitStatus struct {
	value *GitStatus
	isSet bool
}

func (v NullableGitStatus) Get() *GitStatus {
	return v.value
}

func (v *NullableGitStatus) Set(val *GitStatus) {
	v.value = val
	v.isSet = true
}

func (v NullableGitStatus) IsSet() bool {
	return v.isSet
}

func (v *NullableGitStatus) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableGitStatus(val *GitStatus) *NullableGitStatus {
	return &NullableGitStatus{value: val, isSet: true}
}

func (v NullableGitStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableGitStatus) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
