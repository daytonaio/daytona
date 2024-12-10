# InstallProviderDTO

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**ProviderDownloadUrlsDTO** | **map[string]string** |  | 

## Methods

### NewInstallProviderDTO

`func NewInstallProviderDTO(name string, providerDownloadUrlsDTO map[string]string, ) *InstallProviderDTO`

NewInstallProviderDTO instantiates a new InstallProviderDTO object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInstallProviderDTOWithDefaults

`func NewInstallProviderDTOWithDefaults() *InstallProviderDTO`

NewInstallProviderDTOWithDefaults instantiates a new InstallProviderDTO object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

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


### GetProviderDownloadUrlsDTO

`func (o *InstallProviderDTO) GetProviderDownloadUrlsDTO() map[string]string`

GetProviderDownloadUrlsDTO returns the ProviderDownloadUrlsDTO field if non-nil, zero value otherwise.

### GetProviderDownloadUrlsDTOOk

`func (o *InstallProviderDTO) GetProviderDownloadUrlsDTOOk() (*map[string]string, bool)`

GetProviderDownloadUrlsDTOOk returns a tuple with the ProviderDownloadUrlsDTO field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProviderDownloadUrlsDTO

`func (o *InstallProviderDTO) SetProviderDownloadUrlsDTO(v map[string]string)`

SetProviderDownloadUrlsDTO sets ProviderDownloadUrlsDTO field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


