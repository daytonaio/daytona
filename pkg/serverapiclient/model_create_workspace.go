/*
Daytona Server API

Daytona Server API

API version: 0.1.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package serverapiclient

import (
	"encoding/json"
)

// checks if the CreateWorkspace type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &CreateWorkspace{}

// CreateWorkspace struct for CreateWorkspace
type CreateWorkspace struct {
	Id           *string         `json:"id,omitempty"`
	Name         *string         `json:"name,omitempty"`
	Repositories []GitRepository `json:"repositories,omitempty"`
	Target       *string         `json:"target,omitempty"`
}

// NewCreateWorkspace instantiates a new CreateWorkspace object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewCreateWorkspace() *CreateWorkspace {
	this := CreateWorkspace{}
	return &this
}

// NewCreateWorkspaceWithDefaults instantiates a new CreateWorkspace object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewCreateWorkspaceWithDefaults() *CreateWorkspace {
	this := CreateWorkspace{}
	return &this
}

// GetId returns the Id field value if set, zero value otherwise.
func (o *CreateWorkspace) GetId() string {
	if o == nil || IsNil(o.Id) {
		var ret string
		return ret
	}
	return *o.Id
}

// GetIdOk returns a tuple with the Id field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateWorkspace) GetIdOk() (*string, bool) {
	if o == nil || IsNil(o.Id) {
		return nil, false
	}
	return o.Id, true
}

// HasId returns a boolean if a field has been set.
func (o *CreateWorkspace) HasId() bool {
	if o != nil && !IsNil(o.Id) {
		return true
	}

	return false
}

// SetId gets a reference to the given string and assigns it to the Id field.
func (o *CreateWorkspace) SetId(v string) {
	o.Id = &v
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *CreateWorkspace) GetName() string {
	if o == nil || IsNil(o.Name) {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateWorkspace) GetNameOk() (*string, bool) {
	if o == nil || IsNil(o.Name) {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *CreateWorkspace) HasName() bool {
	if o != nil && !IsNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *CreateWorkspace) SetName(v string) {
	o.Name = &v
}

// GetRepositories returns the Repositories field value if set, zero value otherwise.
func (o *CreateWorkspace) GetRepositories() []GitRepository {
	if o == nil || IsNil(o.Repositories) {
		var ret []GitRepository
		return ret
	}
	return o.Repositories
}

// GetRepositoriesOk returns a tuple with the Repositories field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateWorkspace) GetRepositoriesOk() ([]GitRepository, bool) {
	if o == nil || IsNil(o.Repositories) {
		return nil, false
	}
	return o.Repositories, true
}

// HasRepositories returns a boolean if a field has been set.
func (o *CreateWorkspace) HasRepositories() bool {
	if o != nil && !IsNil(o.Repositories) {
		return true
	}

	return false
}

// SetRepositories gets a reference to the given []GitRepository and assigns it to the Repositories field.
func (o *CreateWorkspace) SetRepositories(v []GitRepository) {
	o.Repositories = v
}

// GetTarget returns the Target field value if set, zero value otherwise.
func (o *CreateWorkspace) GetTarget() string {
	if o == nil || IsNil(o.Target) {
		var ret string
		return ret
	}
	return *o.Target
}

// GetTargetOk returns a tuple with the Target field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateWorkspace) GetTargetOk() (*string, bool) {
	if o == nil || IsNil(o.Target) {
		return nil, false
	}
	return o.Target, true
}

// HasTarget returns a boolean if a field has been set.
func (o *CreateWorkspace) HasTarget() bool {
	if o != nil && !IsNil(o.Target) {
		return true
	}

	return false
}

// SetTarget gets a reference to the given string and assigns it to the Target field.
func (o *CreateWorkspace) SetTarget(v string) {
	o.Target = &v
}

func (o CreateWorkspace) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o CreateWorkspace) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Id) {
		toSerialize["id"] = o.Id
	}
	if !IsNil(o.Name) {
		toSerialize["name"] = o.Name
	}
	if !IsNil(o.Repositories) {
		toSerialize["repositories"] = o.Repositories
	}
	if !IsNil(o.Target) {
		toSerialize["target"] = o.Target
	}
	return toSerialize, nil
}

type NullableCreateWorkspace struct {
	value *CreateWorkspace
	isSet bool
}

func (v NullableCreateWorkspace) Get() *CreateWorkspace {
	return v.value
}

func (v *NullableCreateWorkspace) Set(val *CreateWorkspace) {
	v.value = val
	v.isSet = true
}

func (v NullableCreateWorkspace) IsSet() bool {
	return v.isSet
}

func (v *NullableCreateWorkspace) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableCreateWorkspace(val *CreateWorkspace) *NullableCreateWorkspace {
	return &NullableCreateWorkspace{value: val, isSet: true}
}

func (v NullableCreateWorkspace) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableCreateWorkspace) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
