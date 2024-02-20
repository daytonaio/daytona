/*
Daytona Server API

Daytona Server API

API version: 0.1.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package api_client

import (
	"encoding/json"
)

// checks if the Workspace type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &Workspace{}

// Workspace struct for Workspace
type Workspace struct {
	Id *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
	Projects []TypesProject `json:"projects,omitempty"`
	Provisioner *TypesWorkspaceProvisioner `json:"provisioner,omitempty"`
}

// NewWorkspace instantiates a new Workspace object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewWorkspace() *Workspace {
	this := Workspace{}
	return &this
}

// NewWorkspaceWithDefaults instantiates a new Workspace object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewWorkspaceWithDefaults() *Workspace {
	this := Workspace{}
	return &this
}

// GetId returns the Id field value if set, zero value otherwise.
func (o *Workspace) GetId() string {
	if o == nil || IsNil(o.Id) {
		var ret string
		return ret
	}
	return *o.Id
}

// GetIdOk returns a tuple with the Id field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Workspace) GetIdOk() (*string, bool) {
	if o == nil || IsNil(o.Id) {
		return nil, false
	}
	return o.Id, true
}

// HasId returns a boolean if a field has been set.
func (o *Workspace) HasId() bool {
	if o != nil && !IsNil(o.Id) {
		return true
	}

	return false
}

// SetId gets a reference to the given string and assigns it to the Id field.
func (o *Workspace) SetId(v string) {
	o.Id = &v
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *Workspace) GetName() string {
	if o == nil || IsNil(o.Name) {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Workspace) GetNameOk() (*string, bool) {
	if o == nil || IsNil(o.Name) {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *Workspace) HasName() bool {
	if o != nil && !IsNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *Workspace) SetName(v string) {
	o.Name = &v
}

// GetProjects returns the Projects field value if set, zero value otherwise.
func (o *Workspace) GetProjects() []TypesProject {
	if o == nil || IsNil(o.Projects) {
		var ret []TypesProject
		return ret
	}
	return o.Projects
}

// GetProjectsOk returns a tuple with the Projects field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Workspace) GetProjectsOk() ([]TypesProject, bool) {
	if o == nil || IsNil(o.Projects) {
		return nil, false
	}
	return o.Projects, true
}

// HasProjects returns a boolean if a field has been set.
func (o *Workspace) HasProjects() bool {
	if o != nil && !IsNil(o.Projects) {
		return true
	}

	return false
}

// SetProjects gets a reference to the given []TypesProject and assigns it to the Projects field.
func (o *Workspace) SetProjects(v []TypesProject) {
	o.Projects = v
}

// GetProvisioner returns the Provisioner field value if set, zero value otherwise.
func (o *Workspace) GetProvisioner() TypesWorkspaceProvisioner {
	if o == nil || IsNil(o.Provisioner) {
		var ret TypesWorkspaceProvisioner
		return ret
	}
	return *o.Provisioner
}

// GetProvisionerOk returns a tuple with the Provisioner field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Workspace) GetProvisionerOk() (*TypesWorkspaceProvisioner, bool) {
	if o == nil || IsNil(o.Provisioner) {
		return nil, false
	}
	return o.Provisioner, true
}

// HasProvisioner returns a boolean if a field has been set.
func (o *Workspace) HasProvisioner() bool {
	if o != nil && !IsNil(o.Provisioner) {
		return true
	}

	return false
}

// SetProvisioner gets a reference to the given TypesWorkspaceProvisioner and assigns it to the Provisioner field.
func (o *Workspace) SetProvisioner(v TypesWorkspaceProvisioner) {
	o.Provisioner = &v
}

func (o Workspace) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o Workspace) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Id) {
		toSerialize["id"] = o.Id
	}
	if !IsNil(o.Name) {
		toSerialize["name"] = o.Name
	}
	if !IsNil(o.Projects) {
		toSerialize["projects"] = o.Projects
	}
	if !IsNil(o.Provisioner) {
		toSerialize["provisioner"] = o.Provisioner
	}
	return toSerialize, nil
}

type NullableWorkspace struct {
	value *Workspace
	isSet bool
}

func (v NullableWorkspace) Get() *Workspace {
	return v.value
}

func (v *NullableWorkspace) Set(val *Workspace) {
	v.value = val
	v.isSet = true
}

func (v NullableWorkspace) IsSet() bool {
	return v.isSet
}

func (v *NullableWorkspace) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableWorkspace(val *Workspace) *NullableWorkspace {
	return &NullableWorkspace{value: val, isSet: true}
}

func (v NullableWorkspace) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableWorkspace) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


