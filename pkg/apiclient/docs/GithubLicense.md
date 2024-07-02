# GithubLicense

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Body** | Pointer to **string** |  | [optional] 
**Conditions** | Pointer to **[]string** |  | [optional] 
**Description** | Pointer to **string** |  | [optional] 
**Featured** | Pointer to **bool** |  | [optional] 
**HtmlUrl** | Pointer to **string** |  | [optional] 
**Implementation** | Pointer to **string** |  | [optional] 
**Key** | Pointer to **string** |  | [optional] 
**Limitations** | Pointer to **[]string** |  | [optional] 
**Name** | Pointer to **string** |  | [optional] 
**Permissions** | Pointer to **[]string** |  | [optional] 
**SpdxId** | Pointer to **string** |  | [optional] 
**Url** | Pointer to **string** |  | [optional] 

## Methods

### NewGithubLicense

`func NewGithubLicense() *GithubLicense`

NewGithubLicense instantiates a new GithubLicense object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGithubLicenseWithDefaults

`func NewGithubLicenseWithDefaults() *GithubLicense`

NewGithubLicenseWithDefaults instantiates a new GithubLicense object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBody

`func (o *GithubLicense) GetBody() string`

GetBody returns the Body field if non-nil, zero value otherwise.

### GetBodyOk

`func (o *GithubLicense) GetBodyOk() (*string, bool)`

GetBodyOk returns a tuple with the Body field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBody

`func (o *GithubLicense) SetBody(v string)`

SetBody sets Body field to given value.

### HasBody

`func (o *GithubLicense) HasBody() bool`

HasBody returns a boolean if a field has been set.

### GetConditions

`func (o *GithubLicense) GetConditions() []string`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *GithubLicense) GetConditionsOk() (*[]string, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *GithubLicense) SetConditions(v []string)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *GithubLicense) HasConditions() bool`

HasConditions returns a boolean if a field has been set.

### GetDescription

`func (o *GithubLicense) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *GithubLicense) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *GithubLicense) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *GithubLicense) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetFeatured

`func (o *GithubLicense) GetFeatured() bool`

GetFeatured returns the Featured field if non-nil, zero value otherwise.

### GetFeaturedOk

`func (o *GithubLicense) GetFeaturedOk() (*bool, bool)`

GetFeaturedOk returns a tuple with the Featured field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFeatured

`func (o *GithubLicense) SetFeatured(v bool)`

SetFeatured sets Featured field to given value.

### HasFeatured

`func (o *GithubLicense) HasFeatured() bool`

HasFeatured returns a boolean if a field has been set.

### GetHtmlUrl

`func (o *GithubLicense) GetHtmlUrl() string`

GetHtmlUrl returns the HtmlUrl field if non-nil, zero value otherwise.

### GetHtmlUrlOk

`func (o *GithubLicense) GetHtmlUrlOk() (*string, bool)`

GetHtmlUrlOk returns a tuple with the HtmlUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHtmlUrl

`func (o *GithubLicense) SetHtmlUrl(v string)`

SetHtmlUrl sets HtmlUrl field to given value.

### HasHtmlUrl

`func (o *GithubLicense) HasHtmlUrl() bool`

HasHtmlUrl returns a boolean if a field has been set.

### GetImplementation

`func (o *GithubLicense) GetImplementation() string`

GetImplementation returns the Implementation field if non-nil, zero value otherwise.

### GetImplementationOk

`func (o *GithubLicense) GetImplementationOk() (*string, bool)`

GetImplementationOk returns a tuple with the Implementation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImplementation

`func (o *GithubLicense) SetImplementation(v string)`

SetImplementation sets Implementation field to given value.

### HasImplementation

`func (o *GithubLicense) HasImplementation() bool`

HasImplementation returns a boolean if a field has been set.

### GetKey

`func (o *GithubLicense) GetKey() string`

GetKey returns the Key field if non-nil, zero value otherwise.

### GetKeyOk

`func (o *GithubLicense) GetKeyOk() (*string, bool)`

GetKeyOk returns a tuple with the Key field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKey

`func (o *GithubLicense) SetKey(v string)`

SetKey sets Key field to given value.

### HasKey

`func (o *GithubLicense) HasKey() bool`

HasKey returns a boolean if a field has been set.

### GetLimitations

`func (o *GithubLicense) GetLimitations() []string`

GetLimitations returns the Limitations field if non-nil, zero value otherwise.

### GetLimitationsOk

`func (o *GithubLicense) GetLimitationsOk() (*[]string, bool)`

GetLimitationsOk returns a tuple with the Limitations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLimitations

`func (o *GithubLicense) SetLimitations(v []string)`

SetLimitations sets Limitations field to given value.

### HasLimitations

`func (o *GithubLicense) HasLimitations() bool`

HasLimitations returns a boolean if a field has been set.

### GetName

`func (o *GithubLicense) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *GithubLicense) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *GithubLicense) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *GithubLicense) HasName() bool`

HasName returns a boolean if a field has been set.

### GetPermissions

`func (o *GithubLicense) GetPermissions() []string`

GetPermissions returns the Permissions field if non-nil, zero value otherwise.

### GetPermissionsOk

`func (o *GithubLicense) GetPermissionsOk() (*[]string, bool)`

GetPermissionsOk returns a tuple with the Permissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissions

`func (o *GithubLicense) SetPermissions(v []string)`

SetPermissions sets Permissions field to given value.

### HasPermissions

`func (o *GithubLicense) HasPermissions() bool`

HasPermissions returns a boolean if a field has been set.

### GetSpdxId

`func (o *GithubLicense) GetSpdxId() string`

GetSpdxId returns the SpdxId field if non-nil, zero value otherwise.

### GetSpdxIdOk

`func (o *GithubLicense) GetSpdxIdOk() (*string, bool)`

GetSpdxIdOk returns a tuple with the SpdxId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpdxId

`func (o *GithubLicense) SetSpdxId(v string)`

SetSpdxId sets SpdxId field to given value.

### HasSpdxId

`func (o *GithubLicense) HasSpdxId() bool`

HasSpdxId returns a boolean if a field has been set.

### GetUrl

`func (o *GithubLicense) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *GithubLicense) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *GithubLicense) SetUrl(v string)`

SetUrl sets Url field to given value.

### HasUrl

`func (o *GithubLicense) HasUrl() bool`

HasUrl returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


