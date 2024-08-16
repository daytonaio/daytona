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

// checks if the GitRepository type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &GitRepository{}

// GitRepository struct for GitRepository
type GitRepository struct {
	Branch      *string      `json:"branch,omitempty"`
	Clonetarget *CloneTarget `json:"clonetarget,omitempty"`
	Id          string       `json:"id"`
	Name        string       `json:"name"`
	Owner       string       `json:"owner"`
	Path        *string      `json:"path,omitempty"`
	PrNumber    *int32       `json:"prNumber,omitempty"`
	Sha         string       `json:"sha"`
	Source      string       `json:"source"`
	Url         string       `json:"url"`
}

type _GitRepository GitRepository

// NewGitRepository instantiates a new GitRepository object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewGitRepository(id string, name string, owner string, sha string, source string, url string) *GitRepository {
	this := GitRepository{}
	this.Id = id
	this.Name = name
	this.Owner = owner
	this.Sha = sha
	this.Source = source
	this.Url = url
	return &this
}

// NewGitRepositoryWithDefaults instantiates a new GitRepository object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewGitRepositoryWithDefaults() *GitRepository {
	this := GitRepository{}
	return &this
}

// GetBranch returns the Branch field value if set, zero value otherwise.
func (o *GitRepository) GetBranch() string {
	if o == nil || IsNil(o.Branch) {
		var ret string
		return ret
	}
	return *o.Branch
}

// GetBranchOk returns a tuple with the Branch field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GitRepository) GetBranchOk() (*string, bool) {
	if o == nil || IsNil(o.Branch) {
		return nil, false
	}
	return o.Branch, true
}

// HasBranch returns a boolean if a field has been set.
func (o *GitRepository) HasBranch() bool {
	if o != nil && !IsNil(o.Branch) {
		return true
	}

	return false
}

// SetBranch gets a reference to the given string and assigns it to the Branch field.
func (o *GitRepository) SetBranch(v string) {
	o.Branch = &v
}

// GetClonetarget returns the Clonetarget field value if set, zero value otherwise.
func (o *GitRepository) GetClonetarget() CloneTarget {
	if o == nil || IsNil(o.Clonetarget) {
		var ret CloneTarget
		return ret
	}
	return *o.Clonetarget
}

// GetClonetargetOk returns a tuple with the Clonetarget field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GitRepository) GetClonetargetOk() (*CloneTarget, bool) {
	if o == nil || IsNil(o.Clonetarget) {
		return nil, false
	}
	return o.Clonetarget, true
}

// HasClonetarget returns a boolean if a field has been set.
func (o *GitRepository) HasClonetarget() bool {
	if o != nil && !IsNil(o.Clonetarget) {
		return true
	}

	return false
}

// SetClonetarget gets a reference to the given CloneTarget and assigns it to the Clonetarget field.
func (o *GitRepository) SetClonetarget(v CloneTarget) {
	o.Clonetarget = &v
}

// GetId returns the Id field value
func (o *GitRepository) GetId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Id
}

// GetIdOk returns a tuple with the Id field value
// and a boolean to check if the value has been set.
func (o *GitRepository) GetIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Id, true
}

// SetId sets field value
func (o *GitRepository) SetId(v string) {
	o.Id = v
}

// GetName returns the Name field value
func (o *GitRepository) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *GitRepository) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *GitRepository) SetName(v string) {
	o.Name = v
}

// GetOwner returns the Owner field value
func (o *GitRepository) GetOwner() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Owner
}

// GetOwnerOk returns a tuple with the Owner field value
// and a boolean to check if the value has been set.
func (o *GitRepository) GetOwnerOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Owner, true
}

// SetOwner sets field value
func (o *GitRepository) SetOwner(v string) {
	o.Owner = v
}

// GetPath returns the Path field value if set, zero value otherwise.
func (o *GitRepository) GetPath() string {
	if o == nil || IsNil(o.Path) {
		var ret string
		return ret
	}
	return *o.Path
}

// GetPathOk returns a tuple with the Path field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GitRepository) GetPathOk() (*string, bool) {
	if o == nil || IsNil(o.Path) {
		return nil, false
	}
	return o.Path, true
}

// HasPath returns a boolean if a field has been set.
func (o *GitRepository) HasPath() bool {
	if o != nil && !IsNil(o.Path) {
		return true
	}

	return false
}

// SetPath gets a reference to the given string and assigns it to the Path field.
func (o *GitRepository) SetPath(v string) {
	o.Path = &v
}

// GetPrNumber returns the PrNumber field value if set, zero value otherwise.
func (o *GitRepository) GetPrNumber() int32 {
	if o == nil || IsNil(o.PrNumber) {
		var ret int32
		return ret
	}
	return *o.PrNumber
}

// GetPrNumberOk returns a tuple with the PrNumber field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GitRepository) GetPrNumberOk() (*int32, bool) {
	if o == nil || IsNil(o.PrNumber) {
		return nil, false
	}
	return o.PrNumber, true
}

// HasPrNumber returns a boolean if a field has been set.
func (o *GitRepository) HasPrNumber() bool {
	if o != nil && !IsNil(o.PrNumber) {
		return true
	}

	return false
}

// SetPrNumber gets a reference to the given int32 and assigns it to the PrNumber field.
func (o *GitRepository) SetPrNumber(v int32) {
	o.PrNumber = &v
}

// GetSha returns the Sha field value
func (o *GitRepository) GetSha() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Sha
}

// GetShaOk returns a tuple with the Sha field value
// and a boolean to check if the value has been set.
func (o *GitRepository) GetShaOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Sha, true
}

// SetSha sets field value
func (o *GitRepository) SetSha(v string) {
	o.Sha = v
}

// GetSource returns the Source field value
func (o *GitRepository) GetSource() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Source
}

// GetSourceOk returns a tuple with the Source field value
// and a boolean to check if the value has been set.
func (o *GitRepository) GetSourceOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Source, true
}

// SetSource sets field value
func (o *GitRepository) SetSource(v string) {
	o.Source = v
}

// GetUrl returns the Url field value
func (o *GitRepository) GetUrl() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Url
}

// GetUrlOk returns a tuple with the Url field value
// and a boolean to check if the value has been set.
func (o *GitRepository) GetUrlOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Url, true
}

// SetUrl sets field value
func (o *GitRepository) SetUrl(v string) {
	o.Url = v
}

func (o GitRepository) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o GitRepository) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Branch) {
		toSerialize["branch"] = o.Branch
	}
	if !IsNil(o.Clonetarget) {
		toSerialize["clonetarget"] = o.Clonetarget
	}
	toSerialize["id"] = o.Id
	toSerialize["name"] = o.Name
	toSerialize["owner"] = o.Owner
	if !IsNil(o.Path) {
		toSerialize["path"] = o.Path
	}
	if !IsNil(o.PrNumber) {
		toSerialize["prNumber"] = o.PrNumber
	}
	toSerialize["sha"] = o.Sha
	toSerialize["source"] = o.Source
	toSerialize["url"] = o.Url
	return toSerialize, nil
}

func (o *GitRepository) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"id",
		"name",
		"owner",
		"sha",
		"source",
		"url",
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

	varGitRepository := _GitRepository{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varGitRepository)

	if err != nil {
		return err
	}

	*o = GitRepository(varGitRepository)

	return err
}

type NullableGitRepository struct {
	value *GitRepository
	isSet bool
}

func (v NullableGitRepository) Get() *GitRepository {
	return v.value
}

func (v *NullableGitRepository) Set(val *GitRepository) {
	v.value = val
	v.isSet = true
}

func (v NullableGitRepository) IsSet() bool {
	return v.isSet
}

func (v *NullableGitRepository) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableGitRepository(val *GitRepository) *NullableGitRepository {
	return &NullableGitRepository{value: val, isSet: true}
}

func (v NullableGitRepository) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableGitRepository) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
