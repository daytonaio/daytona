/*
Daytona Server API

Daytona Server API

API version: 0.24.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package apiclient

import (
	"encoding/json"
)

// checks if the InstallProviderRequest type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &InstallProviderRequest{}

// InstallProviderRequest struct for InstallProviderRequest
type InstallProviderRequest struct {
	DownloadUrls *map[string]string `json:"downloadUrls,omitempty"`
	Name         *string            `json:"name,omitempty"`
}

// NewInstallProviderRequest instantiates a new InstallProviderRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewInstallProviderRequest() *InstallProviderRequest {
	this := InstallProviderRequest{}
	return &this
}

// NewInstallProviderRequestWithDefaults instantiates a new InstallProviderRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewInstallProviderRequestWithDefaults() *InstallProviderRequest {
	this := InstallProviderRequest{}
	return &this
}

// GetDownloadUrls returns the DownloadUrls field value if set, zero value otherwise.
func (o *InstallProviderRequest) GetDownloadUrls() map[string]string {
	if o == nil || IsNil(o.DownloadUrls) {
		var ret map[string]string
		return ret
	}
	return *o.DownloadUrls
}

// GetDownloadUrlsOk returns a tuple with the DownloadUrls field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *InstallProviderRequest) GetDownloadUrlsOk() (*map[string]string, bool) {
	if o == nil || IsNil(o.DownloadUrls) {
		return nil, false
	}
	return o.DownloadUrls, true
}

// HasDownloadUrls returns a boolean if a field has been set.
func (o *InstallProviderRequest) HasDownloadUrls() bool {
	if o != nil && !IsNil(o.DownloadUrls) {
		return true
	}

	return false
}

// SetDownloadUrls gets a reference to the given map[string]string and assigns it to the DownloadUrls field.
func (o *InstallProviderRequest) SetDownloadUrls(v map[string]string) {
	o.DownloadUrls = &v
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *InstallProviderRequest) GetName() string {
	if o == nil || IsNil(o.Name) {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *InstallProviderRequest) GetNameOk() (*string, bool) {
	if o == nil || IsNil(o.Name) {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *InstallProviderRequest) HasName() bool {
	if o != nil && !IsNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *InstallProviderRequest) SetName(v string) {
	o.Name = &v
}

func (o InstallProviderRequest) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o InstallProviderRequest) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.DownloadUrls) {
		toSerialize["downloadUrls"] = o.DownloadUrls
	}
	if !IsNil(o.Name) {
		toSerialize["name"] = o.Name
	}
	return toSerialize, nil
}

type NullableInstallProviderRequest struct {
	value *InstallProviderRequest
	isSet bool
}

func (v NullableInstallProviderRequest) Get() *InstallProviderRequest {
	return v.value
}

func (v *NullableInstallProviderRequest) Set(val *InstallProviderRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableInstallProviderRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableInstallProviderRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableInstallProviderRequest(val *InstallProviderRequest) *NullableInstallProviderRequest {
	return &NullableInstallProviderRequest{value: val, isSet: true}
}

func (v NullableInstallProviderRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableInstallProviderRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
