# SetGitProviderConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BaseApiUrl** | Pointer to **string** |  | [optional] 
**Id** | **string** |  | 
**Token** | **string** |  | 
**TokenIdentity** | Pointer to **string** |  | [optional] 
**TokenScope** | Pointer to **string** |  | [optional] 
**TokenScopeType** | Pointer to [**GitproviderTokenScopeType**](GitproviderTokenScopeType.md) |  | [optional] 
**Username** | Pointer to **string** |  | [optional] 

## Methods

### NewSetGitProviderConfig

`func NewSetGitProviderConfig(id string, token string, ) *SetGitProviderConfig`

NewSetGitProviderConfig instantiates a new SetGitProviderConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSetGitProviderConfigWithDefaults

`func NewSetGitProviderConfigWithDefaults() *SetGitProviderConfig`

NewSetGitProviderConfigWithDefaults instantiates a new SetGitProviderConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBaseApiUrl

`func (o *SetGitProviderConfig) GetBaseApiUrl() string`

GetBaseApiUrl returns the BaseApiUrl field if non-nil, zero value otherwise.

### GetBaseApiUrlOk

`func (o *SetGitProviderConfig) GetBaseApiUrlOk() (*string, bool)`

GetBaseApiUrlOk returns a tuple with the BaseApiUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBaseApiUrl

`func (o *SetGitProviderConfig) SetBaseApiUrl(v string)`

SetBaseApiUrl sets BaseApiUrl field to given value.

### HasBaseApiUrl

`func (o *SetGitProviderConfig) HasBaseApiUrl() bool`

HasBaseApiUrl returns a boolean if a field has been set.

### GetId

`func (o *SetGitProviderConfig) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *SetGitProviderConfig) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *SetGitProviderConfig) SetId(v string)`

SetId sets Id field to given value.


### GetToken

`func (o *SetGitProviderConfig) GetToken() string`

GetToken returns the Token field if non-nil, zero value otherwise.

### GetTokenOk

`func (o *SetGitProviderConfig) GetTokenOk() (*string, bool)`

GetTokenOk returns a tuple with the Token field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToken

`func (o *SetGitProviderConfig) SetToken(v string)`

SetToken sets Token field to given value.


### GetTokenIdentity

`func (o *SetGitProviderConfig) GetTokenIdentity() string`

GetTokenIdentity returns the TokenIdentity field if non-nil, zero value otherwise.

### GetTokenIdentityOk

`func (o *SetGitProviderConfig) GetTokenIdentityOk() (*string, bool)`

GetTokenIdentityOk returns a tuple with the TokenIdentity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTokenIdentity

`func (o *SetGitProviderConfig) SetTokenIdentity(v string)`

SetTokenIdentity sets TokenIdentity field to given value.

### HasTokenIdentity

`func (o *SetGitProviderConfig) HasTokenIdentity() bool`

HasTokenIdentity returns a boolean if a field has been set.

### GetTokenScope

`func (o *SetGitProviderConfig) GetTokenScope() string`

GetTokenScope returns the TokenScope field if non-nil, zero value otherwise.

### GetTokenScopeOk

`func (o *SetGitProviderConfig) GetTokenScopeOk() (*string, bool)`

GetTokenScopeOk returns a tuple with the TokenScope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTokenScope

`func (o *SetGitProviderConfig) SetTokenScope(v string)`

SetTokenScope sets TokenScope field to given value.

### HasTokenScope

`func (o *SetGitProviderConfig) HasTokenScope() bool`

HasTokenScope returns a boolean if a field has been set.

### GetTokenScopeType

`func (o *SetGitProviderConfig) GetTokenScopeType() GitproviderTokenScopeType`

GetTokenScopeType returns the TokenScopeType field if non-nil, zero value otherwise.

### GetTokenScopeTypeOk

`func (o *SetGitProviderConfig) GetTokenScopeTypeOk() (*GitproviderTokenScopeType, bool)`

GetTokenScopeTypeOk returns a tuple with the TokenScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTokenScopeType

`func (o *SetGitProviderConfig) SetTokenScopeType(v GitproviderTokenScopeType)`

SetTokenScopeType sets TokenScopeType field to given value.

### HasTokenScopeType

`func (o *SetGitProviderConfig) HasTokenScopeType() bool`

HasTokenScopeType returns a boolean if a field has been set.

### GetUsername

`func (o *SetGitProviderConfig) GetUsername() string`

GetUsername returns the Username field if non-nil, zero value otherwise.

### GetUsernameOk

`func (o *SetGitProviderConfig) GetUsernameOk() (*string, bool)`

GetUsernameOk returns a tuple with the Username field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsername

`func (o *SetGitProviderConfig) SetUsername(v string)`

SetUsername sets Username field to given value.

### HasUsername

`func (o *SetGitProviderConfig) HasUsername() bool`

HasUsername returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


