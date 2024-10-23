# GitProvider

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Alias** | **string** |  | 
**BaseApiUrl** | Pointer to **string** |  | [optional] 
**Id** | **string** |  | 
**ProviderId** | **string** |  | 
**SigningKey** | Pointer to **string** |  | [optional] 
**SigningMethod** | Pointer to [**SigningMethod**](SigningMethod.md) |  | [optional] 
**Token** | **string** |  | 
**Username** | **string** |  | 

## Methods

### NewGitProvider

`func NewGitProvider(alias string, id string, providerId string, token string, username string, ) *GitProvider`

NewGitProvider instantiates a new GitProvider object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGitProviderWithDefaults

`func NewGitProviderWithDefaults() *GitProvider`

NewGitProviderWithDefaults instantiates a new GitProvider object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAlias

`func (o *GitProvider) GetAlias() string`

GetAlias returns the Alias field if non-nil, zero value otherwise.

### GetAliasOk

`func (o *GitProvider) GetAliasOk() (*string, bool)`

GetAliasOk returns a tuple with the Alias field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAlias

`func (o *GitProvider) SetAlias(v string)`

SetAlias sets Alias field to given value.


### GetBaseApiUrl

`func (o *GitProvider) GetBaseApiUrl() string`

GetBaseApiUrl returns the BaseApiUrl field if non-nil, zero value otherwise.

### GetBaseApiUrlOk

`func (o *GitProvider) GetBaseApiUrlOk() (*string, bool)`

GetBaseApiUrlOk returns a tuple with the BaseApiUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBaseApiUrl

`func (o *GitProvider) SetBaseApiUrl(v string)`

SetBaseApiUrl sets BaseApiUrl field to given value.

### HasBaseApiUrl

`func (o *GitProvider) HasBaseApiUrl() bool`

HasBaseApiUrl returns a boolean if a field has been set.

### GetId

`func (o *GitProvider) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *GitProvider) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *GitProvider) SetId(v string)`

SetId sets Id field to given value.


### GetProviderId

`func (o *GitProvider) GetProviderId() string`

GetProviderId returns the ProviderId field if non-nil, zero value otherwise.

### GetProviderIdOk

`func (o *GitProvider) GetProviderIdOk() (*string, bool)`

GetProviderIdOk returns a tuple with the ProviderId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProviderId

`func (o *GitProvider) SetProviderId(v string)`

SetProviderId sets ProviderId field to given value.


### GetSigningKey

`func (o *GitProvider) GetSigningKey() string`

GetSigningKey returns the SigningKey field if non-nil, zero value otherwise.

### GetSigningKeyOk

`func (o *GitProvider) GetSigningKeyOk() (*string, bool)`

GetSigningKeyOk returns a tuple with the SigningKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSigningKey

`func (o *GitProvider) SetSigningKey(v string)`

SetSigningKey sets SigningKey field to given value.

### HasSigningKey

`func (o *GitProvider) HasSigningKey() bool`

HasSigningKey returns a boolean if a field has been set.

### GetSigningMethod

`func (o *GitProvider) GetSigningMethod() SigningMethod`

GetSigningMethod returns the SigningMethod field if non-nil, zero value otherwise.

### GetSigningMethodOk

`func (o *GitProvider) GetSigningMethodOk() (*SigningMethod, bool)`

GetSigningMethodOk returns a tuple with the SigningMethod field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSigningMethod

`func (o *GitProvider) SetSigningMethod(v SigningMethod)`

SetSigningMethod sets SigningMethod field to given value.

### HasSigningMethod

`func (o *GitProvider) HasSigningMethod() bool`

HasSigningMethod returns a boolean if a field has been set.

### GetToken

`func (o *GitProvider) GetToken() string`

GetToken returns the Token field if non-nil, zero value otherwise.

### GetTokenOk

`func (o *GitProvider) GetTokenOk() (*string, bool)`

GetTokenOk returns a tuple with the Token field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToken

`func (o *GitProvider) SetToken(v string)`

SetToken sets Token field to given value.


### GetUsername

`func (o *GitProvider) GetUsername() string`

GetUsername returns the Username field if non-nil, zero value otherwise.

### GetUsernameOk

`func (o *GitProvider) GetUsernameOk() (*string, bool)`

GetUsernameOk returns a tuple with the Username field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsername

`func (o *GitProvider) SetUsername(v string)`

SetUsername sets Username field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


