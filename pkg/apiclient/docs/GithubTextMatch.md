# GithubTextMatch

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Fragment** | Pointer to **string** |  | [optional] 
**Matches** | Pointer to [**[]GithubMatch**](GithubMatch.md) |  | [optional] 
**ObjectType** | Pointer to **string** |  | [optional] 
**ObjectUrl** | Pointer to **string** |  | [optional] 
**Property** | Pointer to **string** |  | [optional] 

## Methods

### NewGithubTextMatch

`func NewGithubTextMatch() *GithubTextMatch`

NewGithubTextMatch instantiates a new GithubTextMatch object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGithubTextMatchWithDefaults

`func NewGithubTextMatchWithDefaults() *GithubTextMatch`

NewGithubTextMatchWithDefaults instantiates a new GithubTextMatch object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFragment

`func (o *GithubTextMatch) GetFragment() string`

GetFragment returns the Fragment field if non-nil, zero value otherwise.

### GetFragmentOk

`func (o *GithubTextMatch) GetFragmentOk() (*string, bool)`

GetFragmentOk returns a tuple with the Fragment field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFragment

`func (o *GithubTextMatch) SetFragment(v string)`

SetFragment sets Fragment field to given value.

### HasFragment

`func (o *GithubTextMatch) HasFragment() bool`

HasFragment returns a boolean if a field has been set.

### GetMatches

`func (o *GithubTextMatch) GetMatches() []GithubMatch`

GetMatches returns the Matches field if non-nil, zero value otherwise.

### GetMatchesOk

`func (o *GithubTextMatch) GetMatchesOk() (*[]GithubMatch, bool)`

GetMatchesOk returns a tuple with the Matches field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMatches

`func (o *GithubTextMatch) SetMatches(v []GithubMatch)`

SetMatches sets Matches field to given value.

### HasMatches

`func (o *GithubTextMatch) HasMatches() bool`

HasMatches returns a boolean if a field has been set.

### GetObjectType

`func (o *GithubTextMatch) GetObjectType() string`

GetObjectType returns the ObjectType field if non-nil, zero value otherwise.

### GetObjectTypeOk

`func (o *GithubTextMatch) GetObjectTypeOk() (*string, bool)`

GetObjectTypeOk returns a tuple with the ObjectType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObjectType

`func (o *GithubTextMatch) SetObjectType(v string)`

SetObjectType sets ObjectType field to given value.

### HasObjectType

`func (o *GithubTextMatch) HasObjectType() bool`

HasObjectType returns a boolean if a field has been set.

### GetObjectUrl

`func (o *GithubTextMatch) GetObjectUrl() string`

GetObjectUrl returns the ObjectUrl field if non-nil, zero value otherwise.

### GetObjectUrlOk

`func (o *GithubTextMatch) GetObjectUrlOk() (*string, bool)`

GetObjectUrlOk returns a tuple with the ObjectUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObjectUrl

`func (o *GithubTextMatch) SetObjectUrl(v string)`

SetObjectUrl sets ObjectUrl field to given value.

### HasObjectUrl

`func (o *GithubTextMatch) HasObjectUrl() bool`

HasObjectUrl returns a boolean if a field has been set.

### GetProperty

`func (o *GithubTextMatch) GetProperty() string`

GetProperty returns the Property field if non-nil, zero value otherwise.

### GetPropertyOk

`func (o *GithubTextMatch) GetPropertyOk() (*string, bool)`

GetPropertyOk returns a tuple with the Property field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProperty

`func (o *GithubTextMatch) SetProperty(v string)`

SetProperty sets Property field to given value.

### HasProperty

`func (o *GithubTextMatch) HasProperty() bool`

HasProperty returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


