# TargetDTO

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Default** | **bool** |  | 
**Id** | **string** |  | 
**Info** | Pointer to [**TargetInfo**](TargetInfo.md) |  | [optional] 
**Name** | **string** |  | 
**Options** | **string** | JSON encoded map of options | 
**ProviderInfo** | [**TargetProviderInfo**](TargetProviderInfo.md) |  | 
**WorkspaceCount** | **int32** |  | 

## Methods

### NewTargetDTO

`func NewTargetDTO(default_ bool, id string, name string, options string, providerInfo TargetProviderInfo, workspaceCount int32, ) *TargetDTO`

NewTargetDTO instantiates a new TargetDTO object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTargetDTOWithDefaults

`func NewTargetDTOWithDefaults() *TargetDTO`

NewTargetDTOWithDefaults instantiates a new TargetDTO object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDefault

`func (o *TargetDTO) GetDefault() bool`

GetDefault returns the Default field if non-nil, zero value otherwise.

### GetDefaultOk

`func (o *TargetDTO) GetDefaultOk() (*bool, bool)`

GetDefaultOk returns a tuple with the Default field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefault

`func (o *TargetDTO) SetDefault(v bool)`

SetDefault sets Default field to given value.


### GetId

`func (o *TargetDTO) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *TargetDTO) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *TargetDTO) SetId(v string)`

SetId sets Id field to given value.


### GetInfo

`func (o *TargetDTO) GetInfo() TargetInfo`

GetInfo returns the Info field if non-nil, zero value otherwise.

### GetInfoOk

`func (o *TargetDTO) GetInfoOk() (*TargetInfo, bool)`

GetInfoOk returns a tuple with the Info field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInfo

`func (o *TargetDTO) SetInfo(v TargetInfo)`

SetInfo sets Info field to given value.

### HasInfo

`func (o *TargetDTO) HasInfo() bool`

HasInfo returns a boolean if a field has been set.

### GetName

`func (o *TargetDTO) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *TargetDTO) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *TargetDTO) SetName(v string)`

SetName sets Name field to given value.


### GetOptions

`func (o *TargetDTO) GetOptions() string`

GetOptions returns the Options field if non-nil, zero value otherwise.

### GetOptionsOk

`func (o *TargetDTO) GetOptionsOk() (*string, bool)`

GetOptionsOk returns a tuple with the Options field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOptions

`func (o *TargetDTO) SetOptions(v string)`

SetOptions sets Options field to given value.


### GetProviderInfo

`func (o *TargetDTO) GetProviderInfo() TargetProviderInfo`

GetProviderInfo returns the ProviderInfo field if non-nil, zero value otherwise.

### GetProviderInfoOk

`func (o *TargetDTO) GetProviderInfoOk() (*TargetProviderInfo, bool)`

GetProviderInfoOk returns a tuple with the ProviderInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProviderInfo

`func (o *TargetDTO) SetProviderInfo(v TargetProviderInfo)`

SetProviderInfo sets ProviderInfo field to given value.


### GetWorkspaceCount

`func (o *TargetDTO) GetWorkspaceCount() int32`

GetWorkspaceCount returns the WorkspaceCount field if non-nil, zero value otherwise.

### GetWorkspaceCountOk

`func (o *TargetDTO) GetWorkspaceCountOk() (*int32, bool)`

GetWorkspaceCountOk returns a tuple with the WorkspaceCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkspaceCount

`func (o *TargetDTO) SetWorkspaceCount(v int32)`

SetWorkspaceCount sets WorkspaceCount field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


