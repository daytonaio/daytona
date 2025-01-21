# InstallProviderDTO

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DownloadUrls** | **map[string]string** |  | 
**Name** | **string** |  | 

## Methods

### NewInstallProviderDTO

`func NewInstallProviderDTO(downloadUrls map[string]string, name string, ) *InstallProviderDTO`

NewInstallProviderDTO instantiates a new InstallProviderDTO object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInstallProviderDTOWithDefaults

`func NewInstallProviderDTOWithDefaults() *InstallProviderDTO`

NewInstallProviderDTOWithDefaults instantiates a new InstallProviderDTO object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDownloadUrls

`func (o *InstallProviderDTO) GetDownloadUrls() map[string]string`

GetDownloadUrls returns the DownloadUrls field if non-nil, zero value otherwise.

### GetDownloadUrlsOk

`func (o *InstallProviderDTO) GetDownloadUrlsOk() (*map[string]string, bool)`

GetDownloadUrlsOk returns a tuple with the DownloadUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDownloadUrls

`func (o *InstallProviderDTO) SetDownloadUrls(v map[string]string)`

SetDownloadUrls sets DownloadUrls field to given value.


### GetName

`func (o *InstallProviderDTO) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *InstallProviderDTO) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *InstallProviderDTO) SetName(v string)`

SetName sets Name field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


