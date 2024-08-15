/*
Daytona Server API

Daytona Server API

API version: v0.0.0-dev
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package apiclient

import (
	"encoding/json"
	"fmt"
)

// GitproviderTokenScopeType the model 'GitproviderTokenScopeType'
type GitproviderTokenScopeType string

// List of gitprovider.TokenScopeType
const (
	TokenScopeTypeGlobal       GitproviderTokenScopeType = "GLOBAL"
	TokenScopeTypeOrganization GitproviderTokenScopeType = "ORGANIZATION"
	TokenScopeTypeRepository   GitproviderTokenScopeType = "REPOSITORY"
)

// All allowed values of GitproviderTokenScopeType enum
var AllowedGitproviderTokenScopeTypeEnumValues = []GitproviderTokenScopeType{
	"GLOBAL",
	"ORGANIZATION",
	"REPOSITORY",
}

func (v *GitproviderTokenScopeType) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := GitproviderTokenScopeType(value)
	for _, existing := range AllowedGitproviderTokenScopeTypeEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid GitproviderTokenScopeType", value)
}

// NewGitproviderTokenScopeTypeFromValue returns a pointer to a valid GitproviderTokenScopeType
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewGitproviderTokenScopeTypeFromValue(v string) (*GitproviderTokenScopeType, error) {
	ev := GitproviderTokenScopeType(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for GitproviderTokenScopeType: valid values are %v", v, AllowedGitproviderTokenScopeTypeEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v GitproviderTokenScopeType) IsValid() bool {
	for _, existing := range AllowedGitproviderTokenScopeTypeEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to gitprovider.TokenScopeType value
func (v GitproviderTokenScopeType) Ptr() *GitproviderTokenScopeType {
	return &v
}

type NullableGitproviderTokenScopeType struct {
	value *GitproviderTokenScopeType
	isSet bool
}

func (v NullableGitproviderTokenScopeType) Get() *GitproviderTokenScopeType {
	return v.value
}

func (v *NullableGitproviderTokenScopeType) Set(val *GitproviderTokenScopeType) {
	v.value = val
	v.isSet = true
}

func (v NullableGitproviderTokenScopeType) IsSet() bool {
	return v.isSet
}

func (v *NullableGitproviderTokenScopeType) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableGitproviderTokenScopeType(val *GitproviderTokenScopeType) *NullableGitproviderTokenScopeType {
	return &NullableGitproviderTokenScopeType{value: val, isSet: true}
}

func (v NullableGitproviderTokenScopeType) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableGitproviderTokenScopeType) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
