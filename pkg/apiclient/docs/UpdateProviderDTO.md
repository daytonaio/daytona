# UpdateProviderDTO

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DownloadUrls** | **map[string]string** |  | 
**Version** | **string** |  | 

## Methods

### NewUpdateProviderDTO

`func NewUpdateProviderDTO(downloadUrls map[string]string, version string, ) *UpdateProviderDTO`

NewUpdateProviderDTO instantiates a new UpdateProviderDTO object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateProviderDTOWithDefaults

`func NewUpdateProviderDTOWithDefaults() *UpdateProviderDTO`

NewUpdateProviderDTOWithDefaults instantiates a new UpdateProviderDTO object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDownloadUrls

`func (o *UpdateProviderDTO) GetDownloadUrls() map[string]string`

GetDownloadUrls returns the DownloadUrls field if non-nil, zero value otherwise.

### GetDownloadUrlsOk

`func (o *UpdateProviderDTO) GetDownloadUrlsOk() (*map[string]string, bool)`

GetDownloadUrlsOk returns a tuple with the DownloadUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDownloadUrls

`func (o *UpdateProviderDTO) SetDownloadUrls(v map[string]string)`

SetDownloadUrls sets DownloadUrls field to given value.


### GetVersion

`func (o *UpdateProviderDTO) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *UpdateProviderDTO) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *UpdateProviderDTO) SetVersion(v string)`

SetVersion sets Version field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


