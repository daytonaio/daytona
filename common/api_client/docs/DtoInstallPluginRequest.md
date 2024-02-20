# DtoInstallPluginRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DownloadUrls** | Pointer to **map[string]string** |  | [optional] 
**Name** | Pointer to **string** |  | [optional] 

## Methods

### NewDtoInstallPluginRequest

`func NewDtoInstallPluginRequest() *DtoInstallPluginRequest`

NewDtoInstallPluginRequest instantiates a new DtoInstallPluginRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDtoInstallPluginRequestWithDefaults

`func NewDtoInstallPluginRequestWithDefaults() *DtoInstallPluginRequest`

NewDtoInstallPluginRequestWithDefaults instantiates a new DtoInstallPluginRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDownloadUrls

`func (o *DtoInstallPluginRequest) GetDownloadUrls() map[string]string`

GetDownloadUrls returns the DownloadUrls field if non-nil, zero value otherwise.

### GetDownloadUrlsOk

`func (o *DtoInstallPluginRequest) GetDownloadUrlsOk() (*map[string]string, bool)`

GetDownloadUrlsOk returns a tuple with the DownloadUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDownloadUrls

`func (o *DtoInstallPluginRequest) SetDownloadUrls(v map[string]string)`

SetDownloadUrls sets DownloadUrls field to given value.

### HasDownloadUrls

`func (o *DtoInstallPluginRequest) HasDownloadUrls() bool`

HasDownloadUrls returns a boolean if a field has been set.

### GetName

`func (o *DtoInstallPluginRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DtoInstallPluginRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DtoInstallPluginRequest) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *DtoInstallPluginRequest) HasName() bool`

HasName returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


