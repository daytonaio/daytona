/*
Daytona Server API

Daytona Server API

API version: v0.0.0-dev
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package apiclient

import (
	"encoding/json"
	"bytes"
	"fmt"
)

// checks if the GitUser type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &GitUser{}

// GitUser struct for GitUser
type GitUser struct {
	Email string `json:"email"`
	Id string `json:"id"`
	Name string `json:"name"`
	Username string `json:"username"`
}

type _GitUser GitUser

// NewGitUser instantiates a new GitUser object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewGitUser(email string, id string, name string, username string) *GitUser {
	this := GitUser{}
	this.Email = email
	this.Id = id
	this.Name = name
	this.Username = username
	return &this
}

// NewGitUserWithDefaults instantiates a new GitUser object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewGitUserWithDefaults() *GitUser {
	this := GitUser{}
	return &this
}

// GetEmail returns the Email field value
func (o *GitUser) GetEmail() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Email
}

// GetEmailOk returns a tuple with the Email field value
// and a boolean to check if the value has been set.
func (o *GitUser) GetEmailOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Email, true
}

// SetEmail sets field value
func (o *GitUser) SetEmail(v string) {
	o.Email = v
}

// GetId returns the Id field value
func (o *GitUser) GetId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Id
}

// GetIdOk returns a tuple with the Id field value
// and a boolean to check if the value has been set.
func (o *GitUser) GetIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Id, true
}

// SetId sets field value
func (o *GitUser) SetId(v string) {
	o.Id = v
}

// GetName returns the Name field value
func (o *GitUser) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *GitUser) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *GitUser) SetName(v string) {
	o.Name = v
}

// GetUsername returns the Username field value
func (o *GitUser) GetUsername() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Username
}

// GetUsernameOk returns a tuple with the Username field value
// and a boolean to check if the value has been set.
func (o *GitUser) GetUsernameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Username, true
}

// SetUsername sets field value
func (o *GitUser) SetUsername(v string) {
	o.Username = v
}

func (o GitUser) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o GitUser) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["email"] = o.Email
	toSerialize["id"] = o.Id
	toSerialize["name"] = o.Name
	toSerialize["username"] = o.Username
	return toSerialize, nil
}

func (o *GitUser) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"email",
		"id",
		"name",
		"username",
	}

	allProperties := make(map[string]interface{})

	err = json.Unmarshal(data, &allProperties)

	if err != nil {
		return err;
	}

	for _, requiredProperty := range(requiredProperties) {
		if _, exists := allProperties[requiredProperty]; !exists {
			return fmt.Errorf("no value given for required property %v", requiredProperty)
		}
	}

	varGitUser := _GitUser{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varGitUser)

	if err != nil {
		return err
	}

	*o = GitUser(varGitUser)

	return err
}

type NullableGitUser struct {
	value *GitUser
	isSet bool
}

func (v NullableGitUser) Get() *GitUser {
	return v.value
}

func (v *NullableGitUser) Set(val *GitUser) {
	v.value = val
	v.isSet = true
}

func (v NullableGitUser) IsSet() bool {
	return v.isSet
}

func (v *NullableGitUser) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableGitUser(val *GitUser) *NullableGitUser {
	return &NullableGitUser{value: val, isSet: true}
}

func (v NullableGitUser) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableGitUser) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


