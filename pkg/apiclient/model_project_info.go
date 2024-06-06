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

// checks if the ProjectInfo type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &ProjectInfo{}

// ProjectInfo struct for ProjectInfo
type ProjectInfo struct {
	Created          *string `json:"created,omitempty"`
	IsRunning        *bool   `json:"isRunning,omitempty"`
	Name             *string `json:"name,omitempty"`
	ProviderMetadata *string `json:"providerMetadata,omitempty"`
	WorkspaceId      *string `json:"workspaceId,omitempty"`
}

// NewProjectInfo instantiates a new ProjectInfo object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewProjectInfo() *ProjectInfo {
	this := ProjectInfo{}
	return &this
}

// NewProjectInfoWithDefaults instantiates a new ProjectInfo object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewProjectInfoWithDefaults() *ProjectInfo {
	this := ProjectInfo{}
	return &this
}

// GetCreated returns the Created field value if set, zero value otherwise.
func (o *ProjectInfo) GetCreated() string {
	if o == nil || IsNil(o.Created) {
		var ret string
		return ret
	}
	return *o.Created
}

// GetCreatedOk returns a tuple with the Created field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProjectInfo) GetCreatedOk() (*string, bool) {
	if o == nil || IsNil(o.Created) {
		return nil, false
	}
	return o.Created, true
}

// HasCreated returns a boolean if a field has been set.
func (o *ProjectInfo) HasCreated() bool {
	if o != nil && !IsNil(o.Created) {
		return true
	}

	return false
}

// SetCreated gets a reference to the given string and assigns it to the Created field.
func (o *ProjectInfo) SetCreated(v string) {
	o.Created = &v
}

// GetIsRunning returns the IsRunning field value if set, zero value otherwise.
func (o *ProjectInfo) GetIsRunning() bool {
	if o == nil || IsNil(o.IsRunning) {
		var ret bool
		return ret
	}
	return *o.IsRunning
}

// GetIsRunningOk returns a tuple with the IsRunning field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProjectInfo) GetIsRunningOk() (*bool, bool) {
	if o == nil || IsNil(o.IsRunning) {
		return nil, false
	}
	return o.IsRunning, true
}

// HasIsRunning returns a boolean if a field has been set.
func (o *ProjectInfo) HasIsRunning() bool {
	if o != nil && !IsNil(o.IsRunning) {
		return true
	}

	return false
}

// SetIsRunning gets a reference to the given bool and assigns it to the IsRunning field.
func (o *ProjectInfo) SetIsRunning(v bool) {
	o.IsRunning = &v
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *ProjectInfo) GetName() string {
	if o == nil || IsNil(o.Name) {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProjectInfo) GetNameOk() (*string, bool) {
	if o == nil || IsNil(o.Name) {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *ProjectInfo) HasName() bool {
	if o != nil && !IsNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *ProjectInfo) SetName(v string) {
	o.Name = &v
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

// GetWorkspaceId returns the WorkspaceId field value if set, zero value otherwise.
func (o *ProjectInfo) GetWorkspaceId() string {
	if o == nil || IsNil(o.WorkspaceId) {
		var ret string
		return ret
	}
	return *o.WorkspaceId
}

// GetWorkspaceIdOk returns a tuple with the WorkspaceId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProjectInfo) GetWorkspaceIdOk() (*string, bool) {
	if o == nil || IsNil(o.WorkspaceId) {
		return nil, false
	}
	return o.WorkspaceId, true
}

// HasWorkspaceId returns a boolean if a field has been set.
func (o *ProjectInfo) HasWorkspaceId() bool {
	if o != nil && !IsNil(o.WorkspaceId) {
		return true
	}

	return false
}

// SetWorkspaceId gets a reference to the given string and assigns it to the WorkspaceId field.
func (o *ProjectInfo) SetWorkspaceId(v string) {
	o.WorkspaceId = &v
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
	if !IsNil(o.Created) {
		toSerialize["created"] = o.Created
	}
	if !IsNil(o.IsRunning) {
		toSerialize["isRunning"] = o.IsRunning
	}
	if !IsNil(o.Name) {
		toSerialize["name"] = o.Name
	}
	if !IsNil(o.ProviderMetadata) {
		toSerialize["providerMetadata"] = o.ProviderMetadata
	}
	if !IsNil(o.WorkspaceId) {
		toSerialize["workspaceId"] = o.WorkspaceId
	}
	return toSerialize, nil
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
