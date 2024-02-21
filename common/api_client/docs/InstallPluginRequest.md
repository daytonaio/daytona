# InstallPluginRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DownloadUrls** | Pointer to **map[string]string** |  | [optional] 
**Name** | Pointer to **string** |  | [optional] 

## Methods

### NewInstallPluginRequest

`func NewInstallPluginRequest() *InstallPluginRequest`

NewInstallPluginRequest instantiates a new InstallPluginRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInstallPluginRequestWithDefaults

`func NewInstallPluginRequestWithDefaults() *InstallPluginRequest`

NewInstallPluginRequestWithDefaults instantiates a new InstallPluginRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDownloadUrls

`func (o *InstallPluginRequest) GetDownloadUrls() map[string]string`

GetDownloadUrls returns the DownloadUrls field if non-nil, zero value otherwise.

### GetDownloadUrlsOk

`func (o *InstallPluginRequest) GetDownloadUrlsOk() (*map[string]string, bool)`

GetDownloadUrlsOk returns a tuple with the DownloadUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDownloadUrls

`func (o *InstallPluginRequest) SetDownloadUrls(v map[string]string)`

SetDownloadUrls sets DownloadUrls field to given value.

### HasDownloadUrls

`func (o *InstallPluginRequest) HasDownloadUrls() bool`

HasDownloadUrls returns a boolean if a field has been set.

### GetName

`func (o *InstallPluginRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *InstallPluginRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *InstallPluginRequest) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *InstallPluginRequest) HasName() bool`

HasName returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


