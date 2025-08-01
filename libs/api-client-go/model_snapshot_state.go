/*
Daytona

Daytona AI platform API Docs

API version: 1.0
Contact: support@daytona.com
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package apiclient

import (
	"encoding/json"
	"fmt"
)

// SnapshotState the model 'SnapshotState'
type SnapshotState string

// List of SnapshotState
const (
	SNAPSHOTSTATE_BUILD_PENDING      SnapshotState = "build_pending"
	SNAPSHOTSTATE_BUILDING           SnapshotState = "building"
	SNAPSHOTSTATE_PENDING            SnapshotState = "pending"
	SNAPSHOTSTATE_PULLING            SnapshotState = "pulling"
	SNAPSHOTSTATE_PENDING_VALIDATION SnapshotState = "pending_validation"
	SNAPSHOTSTATE_VALIDATING         SnapshotState = "validating"
	SNAPSHOTSTATE_ACTIVE             SnapshotState = "active"
	SNAPSHOTSTATE_INACTIVE           SnapshotState = "inactive"
	SNAPSHOTSTATE_ERROR              SnapshotState = "error"
	SNAPSHOTSTATE_BUILD_FAILED       SnapshotState = "build_failed"
	SNAPSHOTSTATE_REMOVING           SnapshotState = "removing"
)

// All allowed values of SnapshotState enum
var AllowedSnapshotStateEnumValues = []SnapshotState{
	"build_pending",
	"building",
	"pending",
	"pulling",
	"pending_validation",
	"validating",
	"active",
	"inactive",
	"error",
	"build_failed",
	"removing",
}

func (v *SnapshotState) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := SnapshotState(value)
	for _, existing := range AllowedSnapshotStateEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid SnapshotState", value)
}

// NewSnapshotStateFromValue returns a pointer to a valid SnapshotState
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewSnapshotStateFromValue(v string) (*SnapshotState, error) {
	ev := SnapshotState(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for SnapshotState: valid values are %v", v, AllowedSnapshotStateEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v SnapshotState) IsValid() bool {
	for _, existing := range AllowedSnapshotStateEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to SnapshotState value
func (v SnapshotState) Ptr() *SnapshotState {
	return &v
}

type NullableSnapshotState struct {
	value *SnapshotState
	isSet bool
}

func (v NullableSnapshotState) Get() *SnapshotState {
	return v.value
}

func (v *NullableSnapshotState) Set(val *SnapshotState) {
	v.value = val
	v.isSet = true
}

func (v NullableSnapshotState) IsSet() bool {
	return v.isSet
}

func (v *NullableSnapshotState) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableSnapshotState(val *SnapshotState) *NullableSnapshotState {
	return &NullableSnapshotState{value: val, isSet: true}
}

func (v NullableSnapshotState) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableSnapshotState) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
