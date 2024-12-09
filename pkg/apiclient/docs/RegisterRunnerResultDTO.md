# RegisterRunnerResultDTO

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Alias** | **string** |  | 
**ApiKey** | **string** |  | 
**Id** | **string** |  | 
**Metadata** | Pointer to [**RunnerMetadata**](RunnerMetadata.md) |  | [optional] 

## Methods

### NewRegisterRunnerResultDTO

`func NewRegisterRunnerResultDTO(alias string, apiKey string, id string, ) *RegisterRunnerResultDTO`

NewRegisterRunnerResultDTO instantiates a new RegisterRunnerResultDTO object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRegisterRunnerResultDTOWithDefaults

`func NewRegisterRunnerResultDTOWithDefaults() *RegisterRunnerResultDTO`

NewRegisterRunnerResultDTOWithDefaults instantiates a new RegisterRunnerResultDTO object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAlias

`func (o *RegisterRunnerResultDTO) GetAlias() string`

GetAlias returns the Alias field if non-nil, zero value otherwise.

### GetAliasOk

`func (o *RegisterRunnerResultDTO) GetAliasOk() (*string, bool)`

GetAliasOk returns a tuple with the Alias field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAlias

`func (o *RegisterRunnerResultDTO) SetAlias(v string)`

SetAlias sets Alias field to given value.


### GetApiKey

`func (o *RegisterRunnerResultDTO) GetApiKey() string`

GetApiKey returns the ApiKey field if non-nil, zero value otherwise.

### GetApiKeyOk

`func (o *RegisterRunnerResultDTO) GetApiKeyOk() (*string, bool)`

GetApiKeyOk returns a tuple with the ApiKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetApiKey

`func (o *RegisterRunnerResultDTO) SetApiKey(v string)`

SetApiKey sets ApiKey field to given value.


### GetId

`func (o *RegisterRunnerResultDTO) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *RegisterRunnerResultDTO) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *RegisterRunnerResultDTO) SetId(v string)`

SetId sets Id field to given value.


### GetMetadata

`func (o *RegisterRunnerResultDTO) GetMetadata() RunnerMetadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *RegisterRunnerResultDTO) GetMetadataOk() (*RunnerMetadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *RegisterRunnerResultDTO) SetMetadata(v RunnerMetadata)`

SetMetadata sets Metadata field to given value.

### HasMetadata

`func (o *RegisterRunnerResultDTO) HasMetadata() bool`

HasMetadata returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


