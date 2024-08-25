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

// checks if the SetGitProviderConfig type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &SetGitProviderConfig{}

// SetGitProviderConfig struct for SetGitProviderConfig
type SetGitProviderConfig struct {
	BaseApiUrl *string `json:"baseApiUrl,omitempty"`
	Id string `json:"id"`
	Token string `json:"token"`
	Username *string `json:"username,omitempty"`
	SigningMethod *string `json:"signingMethod,omitempty"`
	SigningKey *string `json:"signingKey,omitempty"`
}

type _SetGitProviderConfig SetGitProviderConfig

// NewSetGitProviderConfig instantiates a new SetGitProviderConfig object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewSetGitProviderConfig(id string, token string) *SetGitProviderConfig {
	this := SetGitProviderConfig{}
	this.Id = id
	this.Token = token
	return &this
}

// NewSetGitProviderConfigWithDefaults instantiates a new SetGitProviderConfig object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewSetGitProviderConfigWithDefaults() *SetGitProviderConfig {
	this := SetGitProviderConfig{}
	return &this
}

// GetBaseApiUrl returns the BaseApiUrl field value if set, zero value otherwise.
func (o *SetGitProviderConfig) GetBaseApiUrl() string {
	if o == nil || IsNil(o.BaseApiUrl) {
		var ret string
		return ret
	}
	return *o.BaseApiUrl
}

// GetBaseApiUrlOk returns a tuple with the BaseApiUrl field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *SetGitProviderConfig) GetBaseApiUrlOk() (*string, bool) {
	if o == nil || IsNil(o.BaseApiUrl) {
		return nil, false
	}
	return o.BaseApiUrl, true
}

// HasBaseApiUrl returns a boolean if a field has been set.
func (o *SetGitProviderConfig) HasBaseApiUrl() bool {
	if o != nil && !IsNil(o.BaseApiUrl) {
		return true
	}

	return false
}

// SetBaseApiUrl gets a reference to the given string and assigns it to the BaseApiUrl field.
func (o *SetGitProviderConfig) SetBaseApiUrl(v string) {
	o.BaseApiUrl = &v
}

// GetId returns the Id field value
func (o *SetGitProviderConfig) GetId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Id
}

// GetIdOk returns a tuple with the Id field value
// and a boolean to check if the value has been set.
func (o *SetGitProviderConfig) GetIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Id, true
}

// SetId sets field value
func (o *SetGitProviderConfig) SetId(v string) {
	o.Id = v
}

// GetToken returns the Token field value
func (o *SetGitProviderConfig) GetToken() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Token
}

// GetTokenOk returns a tuple with the Token field value
// and a boolean to check if the value has been set.
func (o *SetGitProviderConfig) GetTokenOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Token, true
}

// SetToken sets field value
func (o *SetGitProviderConfig) SetToken(v string) {
	o.Token = v
}

// GetUsername returns the Username field value if set, zero value otherwise.
func (o *SetGitProviderConfig) GetUsername() string {
	if o == nil || IsNil(o.Username) {
		var ret string
		return ret
	}
	return *o.Username
}

// GetUsernameOk returns a tuple with the Username field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *SetGitProviderConfig) GetUsernameOk() (*string, bool) {
	if o == nil || IsNil(o.Username) {
		return nil, false
	}
	return o.Username, true
}

// HasUsername returns a boolean if a field has been set.
func (o *SetGitProviderConfig) HasUsername() bool {
	if o != nil && !IsNil(o.Username) {
		return true
	}

	return false
}

// SetUsername gets a reference to the given string and assigns it to the Username field.
func (o *SetGitProviderConfig) SetUsername(v string) {
	o.Username = &v
}

// GetSigningMethod returns the SigningMethod field value if set, zero value otherwise.
func (o *SetGitProviderConfig) GetSigningMethod() string {
	if o == nil || IsNil(o.SigningMethod) {
		var ret string
		return ret
	}
	return *o.SigningMethod
}

// GetSigningMethodOk returns a tuple with the SigningMethod field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *SetGitProviderConfig) GetSigningMethodOk() (*string, bool) {
	if o == nil || IsNil(o.SigningMethod) {
		return nil, false
	}
	return o.SigningMethod, true
}

// HasSigningMethod returns a boolean if a field has been set.
func (o *SetGitProviderConfig) HasSigningMethod() bool {
	if o != nil && !IsNil(o.SigningMethod) {
		return true
	}

	return false
}

// SetSigningMethod gets a reference to the given string and assigns it to the SigningMethod field.
func (o *SetGitProviderConfig) SetSigningMethod(v string) {
	o.SigningMethod = &v
}

// GetSigningKey returns the SigningKey field value if set, zero value otherwise.
func (o *SetGitProviderConfig) GetSigningKey() string {
	if o == nil || IsNil(o.SigningKey) {
		var ret string
		return ret
	}
	return *o.SigningKey
}

// GetSigningKeyOk returns a tuple with the SigningKey field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *SetGitProviderConfig) GetSigningKeyOk() (*string, bool) {
	if o == nil || IsNil(o.SigningKey) {
		return nil, false
	}
	return o.SigningKey, true
}

// HasSigningKey returns a boolean if a field has been set.
func (o *SetGitProviderConfig) HasSigningKey() bool {
	if o != nil && !IsNil(o.SigningKey) {
		return true
	}

	return false
}

// SetSigningKey gets a reference to the given string and assigns it to the SigningKey field.
func (o *SetGitProviderConfig) SetSigningKey(v string) {
	o.SigningKey = &v
}

func (o SetGitProviderConfig) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o SetGitProviderConfig) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.BaseApiUrl) {
		toSerialize["baseApiUrl"] = o.BaseApiUrl
	}
	toSerialize["id"] = o.Id
	toSerialize["token"] = o.Token
	if !IsNil(o.Username) {
		toSerialize["username"] = o.Username
	}
	if !IsNil(o.SigningMethod) {
		toSerialize["signingMethod"] = o.SigningMethod
	}
	if !IsNil(o.SigningKey) {
		toSerialize["signingKey"] = o.SigningKey
	}
	return toSerialize, nil
}

func (o *SetGitProviderConfig) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"id",
		"token",
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

	varSetGitProviderConfig := _SetGitProviderConfig{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varSetGitProviderConfig)

	if err != nil {
		return err
	}

	*o = SetGitProviderConfig(varSetGitProviderConfig)

	return err
}

type NullableSetGitProviderConfig struct {
	value *SetGitProviderConfig
	isSet bool
}

func (v NullableSetGitProviderConfig) Get() *SetGitProviderConfig {
	return v.value
}

func (v *NullableSetGitProviderConfig) Set(val *SetGitProviderConfig) {
	v.value = val
	v.isSet = true
}

func (v NullableSetGitProviderConfig) IsSet() bool {
	return v.isSet
}

func (v *NullableSetGitProviderConfig) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableSetGitProviderConfig(val *SetGitProviderConfig) *NullableSetGitProviderConfig {
	return &NullableSetGitProviderConfig{value: val, isSet: true}
}

func (v NullableSetGitProviderConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableSetGitProviderConfig) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


