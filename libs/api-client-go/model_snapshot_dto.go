/*
Daytona

Daytona AI platform API Docs

API version: 1.0
Contact: support@daytona.com
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package apiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

// checks if the SnapshotDto type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &SnapshotDto{}

// SnapshotDto struct for SnapshotDto
type SnapshotDto struct {
	Id             string          `json:"id"`
	OrganizationId *string         `json:"organizationId,omitempty"`
	General        bool            `json:"general"`
	Name           string          `json:"name"`
	ImageName      *string         `json:"imageName,omitempty"`
	State          SnapshotState   `json:"state"`
	Size           NullableFloat32 `json:"size"`
	Entrypoint     []string        `json:"entrypoint"`
	Cpu            float32         `json:"cpu"`
	Gpu            float32         `json:"gpu"`
	Mem            float32         `json:"mem"`
	Disk           float32         `json:"disk"`
	ErrorReason    NullableString  `json:"errorReason"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
	LastUsedAt     NullableTime    `json:"lastUsedAt"`
	// Build information for the snapshot
	BuildInfo *BuildInfo `json:"buildInfo,omitempty"`
}

type _SnapshotDto SnapshotDto

// NewSnapshotDto instantiates a new SnapshotDto object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewSnapshotDto(id string, general bool, name string, state SnapshotState, size NullableFloat32, entrypoint []string, cpu float32, gpu float32, mem float32, disk float32, errorReason NullableString, createdAt time.Time, updatedAt time.Time, lastUsedAt NullableTime) *SnapshotDto {
	this := SnapshotDto{}
	this.Id = id
	this.General = general
	this.Name = name
	this.State = state
	this.Size = size
	this.Entrypoint = entrypoint
	this.Cpu = cpu
	this.Gpu = gpu
	this.Mem = mem
	this.Disk = disk
	this.ErrorReason = errorReason
	this.CreatedAt = createdAt
	this.UpdatedAt = updatedAt
	this.LastUsedAt = lastUsedAt
	return &this
}

// NewSnapshotDtoWithDefaults instantiates a new SnapshotDto object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewSnapshotDtoWithDefaults() *SnapshotDto {
	this := SnapshotDto{}
	return &this
}

// GetId returns the Id field value
func (o *SnapshotDto) GetId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Id
}

// GetIdOk returns a tuple with the Id field value
// and a boolean to check if the value has been set.
func (o *SnapshotDto) GetIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Id, true
}

// SetId sets field value
func (o *SnapshotDto) SetId(v string) {
	o.Id = v
}

// GetOrganizationId returns the OrganizationId field value if set, zero value otherwise.
func (o *SnapshotDto) GetOrganizationId() string {
	if o == nil || IsNil(o.OrganizationId) {
		var ret string
		return ret
	}
	return *o.OrganizationId
}

// GetOrganizationIdOk returns a tuple with the OrganizationId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *SnapshotDto) GetOrganizationIdOk() (*string, bool) {
	if o == nil || IsNil(o.OrganizationId) {
		return nil, false
	}
	return o.OrganizationId, true
}

// HasOrganizationId returns a boolean if a field has been set.
func (o *SnapshotDto) HasOrganizationId() bool {
	if o != nil && !IsNil(o.OrganizationId) {
		return true
	}

	return false
}

// SetOrganizationId gets a reference to the given string and assigns it to the OrganizationId field.
func (o *SnapshotDto) SetOrganizationId(v string) {
	o.OrganizationId = &v
}

// GetGeneral returns the General field value
func (o *SnapshotDto) GetGeneral() bool {
	if o == nil {
		var ret bool
		return ret
	}

	return o.General
}

// GetGeneralOk returns a tuple with the General field value
// and a boolean to check if the value has been set.
func (o *SnapshotDto) GetGeneralOk() (*bool, bool) {
	if o == nil {
		return nil, false
	}
	return &o.General, true
}

// SetGeneral sets field value
func (o *SnapshotDto) SetGeneral(v bool) {
	o.General = v
}

// GetName returns the Name field value
func (o *SnapshotDto) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *SnapshotDto) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *SnapshotDto) SetName(v string) {
	o.Name = v
}

// GetImageName returns the ImageName field value if set, zero value otherwise.
func (o *SnapshotDto) GetImageName() string {
	if o == nil || IsNil(o.ImageName) {
		var ret string
		return ret
	}
	return *o.ImageName
}

// GetImageNameOk returns a tuple with the ImageName field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *SnapshotDto) GetImageNameOk() (*string, bool) {
	if o == nil || IsNil(o.ImageName) {
		return nil, false
	}
	return o.ImageName, true
}

// HasImageName returns a boolean if a field has been set.
func (o *SnapshotDto) HasImageName() bool {
	if o != nil && !IsNil(o.ImageName) {
		return true
	}

	return false
}

// SetImageName gets a reference to the given string and assigns it to the ImageName field.
func (o *SnapshotDto) SetImageName(v string) {
	o.ImageName = &v
}

// GetState returns the State field value
func (o *SnapshotDto) GetState() SnapshotState {
	if o == nil {
		var ret SnapshotState
		return ret
	}

	return o.State
}

// GetStateOk returns a tuple with the State field value
// and a boolean to check if the value has been set.
func (o *SnapshotDto) GetStateOk() (*SnapshotState, bool) {
	if o == nil {
		return nil, false
	}
	return &o.State, true
}

// SetState sets field value
func (o *SnapshotDto) SetState(v SnapshotState) {
	o.State = v
}

// GetSize returns the Size field value
// If the value is explicit nil, the zero value for float32 will be returned
func (o *SnapshotDto) GetSize() float32 {
	if o == nil || o.Size.Get() == nil {
		var ret float32
		return ret
	}

	return *o.Size.Get()
}

// GetSizeOk returns a tuple with the Size field value
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *SnapshotDto) GetSizeOk() (*float32, bool) {
	if o == nil {
		return nil, false
	}
	return o.Size.Get(), o.Size.IsSet()
}

// SetSize sets field value
func (o *SnapshotDto) SetSize(v float32) {
	o.Size.Set(&v)
}

// GetEntrypoint returns the Entrypoint field value
// If the value is explicit nil, the zero value for []string will be returned
func (o *SnapshotDto) GetEntrypoint() []string {
	if o == nil {
		var ret []string
		return ret
	}

	return o.Entrypoint
}

// GetEntrypointOk returns a tuple with the Entrypoint field value
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *SnapshotDto) GetEntrypointOk() ([]string, bool) {
	if o == nil || IsNil(o.Entrypoint) {
		return nil, false
	}
	return o.Entrypoint, true
}

// SetEntrypoint sets field value
func (o *SnapshotDto) SetEntrypoint(v []string) {
	o.Entrypoint = v
}

// GetCpu returns the Cpu field value
func (o *SnapshotDto) GetCpu() float32 {
	if o == nil {
		var ret float32
		return ret
	}

	return o.Cpu
}

// GetCpuOk returns a tuple with the Cpu field value
// and a boolean to check if the value has been set.
func (o *SnapshotDto) GetCpuOk() (*float32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Cpu, true
}

// SetCpu sets field value
func (o *SnapshotDto) SetCpu(v float32) {
	o.Cpu = v
}

// GetGpu returns the Gpu field value
func (o *SnapshotDto) GetGpu() float32 {
	if o == nil {
		var ret float32
		return ret
	}

	return o.Gpu
}

// GetGpuOk returns a tuple with the Gpu field value
// and a boolean to check if the value has been set.
func (o *SnapshotDto) GetGpuOk() (*float32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Gpu, true
}

// SetGpu sets field value
func (o *SnapshotDto) SetGpu(v float32) {
	o.Gpu = v
}

// GetMem returns the Mem field value
func (o *SnapshotDto) GetMem() float32 {
	if o == nil {
		var ret float32
		return ret
	}

	return o.Mem
}

// GetMemOk returns a tuple with the Mem field value
// and a boolean to check if the value has been set.
func (o *SnapshotDto) GetMemOk() (*float32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Mem, true
}

// SetMem sets field value
func (o *SnapshotDto) SetMem(v float32) {
	o.Mem = v
}

// GetDisk returns the Disk field value
func (o *SnapshotDto) GetDisk() float32 {
	if o == nil {
		var ret float32
		return ret
	}

	return o.Disk
}

// GetDiskOk returns a tuple with the Disk field value
// and a boolean to check if the value has been set.
func (o *SnapshotDto) GetDiskOk() (*float32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Disk, true
}

// SetDisk sets field value
func (o *SnapshotDto) SetDisk(v float32) {
	o.Disk = v
}

// GetErrorReason returns the ErrorReason field value
// If the value is explicit nil, the zero value for string will be returned
func (o *SnapshotDto) GetErrorReason() string {
	if o == nil || o.ErrorReason.Get() == nil {
		var ret string
		return ret
	}

	return *o.ErrorReason.Get()
}

// GetErrorReasonOk returns a tuple with the ErrorReason field value
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *SnapshotDto) GetErrorReasonOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return o.ErrorReason.Get(), o.ErrorReason.IsSet()
}

// SetErrorReason sets field value
func (o *SnapshotDto) SetErrorReason(v string) {
	o.ErrorReason.Set(&v)
}

// GetCreatedAt returns the CreatedAt field value
func (o *SnapshotDto) GetCreatedAt() time.Time {
	if o == nil {
		var ret time.Time
		return ret
	}

	return o.CreatedAt
}

// GetCreatedAtOk returns a tuple with the CreatedAt field value
// and a boolean to check if the value has been set.
func (o *SnapshotDto) GetCreatedAtOk() (*time.Time, bool) {
	if o == nil {
		return nil, false
	}
	return &o.CreatedAt, true
}

// SetCreatedAt sets field value
func (o *SnapshotDto) SetCreatedAt(v time.Time) {
	o.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt field value
func (o *SnapshotDto) GetUpdatedAt() time.Time {
	if o == nil {
		var ret time.Time
		return ret
	}

	return o.UpdatedAt
}

// GetUpdatedAtOk returns a tuple with the UpdatedAt field value
// and a boolean to check if the value has been set.
func (o *SnapshotDto) GetUpdatedAtOk() (*time.Time, bool) {
	if o == nil {
		return nil, false
	}
	return &o.UpdatedAt, true
}

// SetUpdatedAt sets field value
func (o *SnapshotDto) SetUpdatedAt(v time.Time) {
	o.UpdatedAt = v
}

// GetLastUsedAt returns the LastUsedAt field value
// If the value is explicit nil, the zero value for time.Time will be returned
func (o *SnapshotDto) GetLastUsedAt() time.Time {
	if o == nil || o.LastUsedAt.Get() == nil {
		var ret time.Time
		return ret
	}

	return *o.LastUsedAt.Get()
}

// GetLastUsedAtOk returns a tuple with the LastUsedAt field value
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *SnapshotDto) GetLastUsedAtOk() (*time.Time, bool) {
	if o == nil {
		return nil, false
	}
	return o.LastUsedAt.Get(), o.LastUsedAt.IsSet()
}

// SetLastUsedAt sets field value
func (o *SnapshotDto) SetLastUsedAt(v time.Time) {
	o.LastUsedAt.Set(&v)
}

// GetBuildInfo returns the BuildInfo field value if set, zero value otherwise.
func (o *SnapshotDto) GetBuildInfo() BuildInfo {
	if o == nil || IsNil(o.BuildInfo) {
		var ret BuildInfo
		return ret
	}
	return *o.BuildInfo
}

// GetBuildInfoOk returns a tuple with the BuildInfo field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *SnapshotDto) GetBuildInfoOk() (*BuildInfo, bool) {
	if o == nil || IsNil(o.BuildInfo) {
		return nil, false
	}
	return o.BuildInfo, true
}

// HasBuildInfo returns a boolean if a field has been set.
func (o *SnapshotDto) HasBuildInfo() bool {
	if o != nil && !IsNil(o.BuildInfo) {
		return true
	}

	return false
}

// SetBuildInfo gets a reference to the given BuildInfo and assigns it to the BuildInfo field.
func (o *SnapshotDto) SetBuildInfo(v BuildInfo) {
	o.BuildInfo = &v
}

func (o SnapshotDto) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o SnapshotDto) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["id"] = o.Id
	if !IsNil(o.OrganizationId) {
		toSerialize["organizationId"] = o.OrganizationId
	}
	toSerialize["general"] = o.General
	toSerialize["name"] = o.Name
	if !IsNil(o.ImageName) {
		toSerialize["imageName"] = o.ImageName
	}
	toSerialize["state"] = o.State
	toSerialize["size"] = o.Size.Get()
	if o.Entrypoint != nil {
		toSerialize["entrypoint"] = o.Entrypoint
	}
	toSerialize["cpu"] = o.Cpu
	toSerialize["gpu"] = o.Gpu
	toSerialize["mem"] = o.Mem
	toSerialize["disk"] = o.Disk
	toSerialize["errorReason"] = o.ErrorReason.Get()
	toSerialize["createdAt"] = o.CreatedAt
	toSerialize["updatedAt"] = o.UpdatedAt
	toSerialize["lastUsedAt"] = o.LastUsedAt.Get()
	if !IsNil(o.BuildInfo) {
		toSerialize["buildInfo"] = o.BuildInfo
	}
	return toSerialize, nil
}

func (o *SnapshotDto) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"id",
		"general",
		"name",
		"state",
		"size",
		"entrypoint",
		"cpu",
		"gpu",
		"mem",
		"disk",
		"errorReason",
		"createdAt",
		"updatedAt",
		"lastUsedAt",
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

	varSnapshotDto := _SnapshotDto{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varSnapshotDto)

	if err != nil {
		return err
	}

	*o = SnapshotDto(varSnapshotDto)

	return err
}

type NullableSnapshotDto struct {
	value *SnapshotDto
	isSet bool
}

func (v NullableSnapshotDto) Get() *SnapshotDto {
	return v.value
}

func (v *NullableSnapshotDto) Set(val *SnapshotDto) {
	v.value = val
	v.isSet = true
}

func (v NullableSnapshotDto) IsSet() bool {
	return v.isSet
}

func (v *NullableSnapshotDto) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableSnapshotDto(val *SnapshotDto) *NullableSnapshotDto {
	return &NullableSnapshotDto{value: val, isSet: true}
}

func (v NullableSnapshotDto) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableSnapshotDto) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
