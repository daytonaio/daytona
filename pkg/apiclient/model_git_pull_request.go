/*
Daytona Server API

Daytona Server API

API version: v0.0.0-dev
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package apiclient

import (
	"encoding/json"
)

// checks if the GitPullRequest type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &GitPullRequest{}

// GitPullRequest struct for GitPullRequest
type GitPullRequest struct {
	Branch          *string `json:"branch,omitempty"`
	Name            *string `json:"name,omitempty"`
	Sha             *string `json:"sha,omitempty"`
	SourceRepoId    *string `json:"sourceRepoId,omitempty"`
	SourceRepoName  *string `json:"sourceRepoName,omitempty"`
	SourceRepoOwner *string `json:"sourceRepoOwner,omitempty"`
	SourceRepoUrl   *string `json:"sourceRepoUrl,omitempty"`
}

// NewGitPullRequest instantiates a new GitPullRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewGitPullRequest() *GitPullRequest {
	this := GitPullRequest{}
	return &this
}

// NewGitPullRequestWithDefaults instantiates a new GitPullRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewGitPullRequestWithDefaults() *GitPullRequest {
	this := GitPullRequest{}
	return &this
}

// GetBranch returns the Branch field value if set, zero value otherwise.
func (o *GitPullRequest) GetBranch() string {
	if o == nil || IsNil(o.Branch) {
		var ret string
		return ret
	}
	return *o.Branch
}

// GetBranchOk returns a tuple with the Branch field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GitPullRequest) GetBranchOk() (*string, bool) {
	if o == nil || IsNil(o.Branch) {
		return nil, false
	}
	return o.Branch, true
}

// HasBranch returns a boolean if a field has been set.
func (o *GitPullRequest) HasBranch() bool {
	if o != nil && !IsNil(o.Branch) {
		return true
	}

	return false
}

// SetBranch gets a reference to the given string and assigns it to the Branch field.
func (o *GitPullRequest) SetBranch(v string) {
	o.Branch = &v
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *GitPullRequest) GetName() string {
	if o == nil || IsNil(o.Name) {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GitPullRequest) GetNameOk() (*string, bool) {
	if o == nil || IsNil(o.Name) {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *GitPullRequest) HasName() bool {
	if o != nil && !IsNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *GitPullRequest) SetName(v string) {
	o.Name = &v
}

// GetSha returns the Sha field value if set, zero value otherwise.
func (o *GitPullRequest) GetSha() string {
	if o == nil || IsNil(o.Sha) {
		var ret string
		return ret
	}
	return *o.Sha
}

// GetShaOk returns a tuple with the Sha field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GitPullRequest) GetShaOk() (*string, bool) {
	if o == nil || IsNil(o.Sha) {
		return nil, false
	}
	return o.Sha, true
}

// HasSha returns a boolean if a field has been set.
func (o *GitPullRequest) HasSha() bool {
	if o != nil && !IsNil(o.Sha) {
		return true
	}

	return false
}

// SetSha gets a reference to the given string and assigns it to the Sha field.
func (o *GitPullRequest) SetSha(v string) {
	o.Sha = &v
}

// GetSourceRepoId returns the SourceRepoId field value if set, zero value otherwise.
func (o *GitPullRequest) GetSourceRepoId() string {
	if o == nil || IsNil(o.SourceRepoId) {
		var ret string
		return ret
	}
	return *o.SourceRepoId
}

// GetSourceRepoIdOk returns a tuple with the SourceRepoId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GitPullRequest) GetSourceRepoIdOk() (*string, bool) {
	if o == nil || IsNil(o.SourceRepoId) {
		return nil, false
	}
	return o.SourceRepoId, true
}

// HasSourceRepoId returns a boolean if a field has been set.
func (o *GitPullRequest) HasSourceRepoId() bool {
	if o != nil && !IsNil(o.SourceRepoId) {
		return true
	}

	return false
}

// SetSourceRepoId gets a reference to the given string and assigns it to the SourceRepoId field.
func (o *GitPullRequest) SetSourceRepoId(v string) {
	o.SourceRepoId = &v
}

// GetSourceRepoName returns the SourceRepoName field value if set, zero value otherwise.
func (o *GitPullRequest) GetSourceRepoName() string {
	if o == nil || IsNil(o.SourceRepoName) {
		var ret string
		return ret
	}
	return *o.SourceRepoName
}

// GetSourceRepoNameOk returns a tuple with the SourceRepoName field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GitPullRequest) GetSourceRepoNameOk() (*string, bool) {
	if o == nil || IsNil(o.SourceRepoName) {
		return nil, false
	}
	return o.SourceRepoName, true
}

// HasSourceRepoName returns a boolean if a field has been set.
func (o *GitPullRequest) HasSourceRepoName() bool {
	if o != nil && !IsNil(o.SourceRepoName) {
		return true
	}

	return false
}

// SetSourceRepoName gets a reference to the given string and assigns it to the SourceRepoName field.
func (o *GitPullRequest) SetSourceRepoName(v string) {
	o.SourceRepoName = &v
}

// GetSourceRepoOwner returns the SourceRepoOwner field value if set, zero value otherwise.
func (o *GitPullRequest) GetSourceRepoOwner() string {
	if o == nil || IsNil(o.SourceRepoOwner) {
		var ret string
		return ret
	}
	return *o.SourceRepoOwner
}

// GetSourceRepoOwnerOk returns a tuple with the SourceRepoOwner field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GitPullRequest) GetSourceRepoOwnerOk() (*string, bool) {
	if o == nil || IsNil(o.SourceRepoOwner) {
		return nil, false
	}
	return o.SourceRepoOwner, true
}

// HasSourceRepoOwner returns a boolean if a field has been set.
func (o *GitPullRequest) HasSourceRepoOwner() bool {
	if o != nil && !IsNil(o.SourceRepoOwner) {
		return true
	}

	return false
}

// SetSourceRepoOwner gets a reference to the given string and assigns it to the SourceRepoOwner field.
func (o *GitPullRequest) SetSourceRepoOwner(v string) {
	o.SourceRepoOwner = &v
}

// GetSourceRepoUrl returns the SourceRepoUrl field value if set, zero value otherwise.
func (o *GitPullRequest) GetSourceRepoUrl() string {
	if o == nil || IsNil(o.SourceRepoUrl) {
		var ret string
		return ret
	}
	return *o.SourceRepoUrl
}

// GetSourceRepoUrlOk returns a tuple with the SourceRepoUrl field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GitPullRequest) GetSourceRepoUrlOk() (*string, bool) {
	if o == nil || IsNil(o.SourceRepoUrl) {
		return nil, false
	}
	return o.SourceRepoUrl, true
}

// HasSourceRepoUrl returns a boolean if a field has been set.
func (o *GitPullRequest) HasSourceRepoUrl() bool {
	if o != nil && !IsNil(o.SourceRepoUrl) {
		return true
	}

	return false
}

// SetSourceRepoUrl gets a reference to the given string and assigns it to the SourceRepoUrl field.
func (o *GitPullRequest) SetSourceRepoUrl(v string) {
	o.SourceRepoUrl = &v
}

func (o GitPullRequest) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o GitPullRequest) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Branch) {
		toSerialize["branch"] = o.Branch
	}
	if !IsNil(o.Name) {
		toSerialize["name"] = o.Name
	}
	if !IsNil(o.Sha) {
		toSerialize["sha"] = o.Sha
	}
	if !IsNil(o.SourceRepoId) {
		toSerialize["sourceRepoId"] = o.SourceRepoId
	}
	if !IsNil(o.SourceRepoName) {
		toSerialize["sourceRepoName"] = o.SourceRepoName
	}
	if !IsNil(o.SourceRepoOwner) {
		toSerialize["sourceRepoOwner"] = o.SourceRepoOwner
	}
	if !IsNil(o.SourceRepoUrl) {
		toSerialize["sourceRepoUrl"] = o.SourceRepoUrl
	}
	return toSerialize, nil
}

type NullableGitPullRequest struct {
	value *GitPullRequest
	isSet bool
}

func (v NullableGitPullRequest) Get() *GitPullRequest {
	return v.value
}

func (v *NullableGitPullRequest) Set(val *GitPullRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableGitPullRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableGitPullRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableGitPullRequest(val *GitPullRequest) *NullableGitPullRequest {
	return &NullableGitPullRequest{value: val, isSet: true}
}

func (v NullableGitPullRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableGitPullRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
