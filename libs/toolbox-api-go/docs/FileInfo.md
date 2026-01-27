# FileInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Group** | **string** |  | 
**IsDir** | **bool** |  | 
**ModTime** | **string** |  | 
**Mode** | **string** |  | 
**Name** | **string** |  | 
**Owner** | **string** |  | 
**Permissions** | **string** |  | 
**Size** | **int32** |  | 

## Methods

### NewFileInfo

`func NewFileInfo(group string, isDir bool, modTime string, mode string, name string, owner string, permissions string, size int32, ) *FileInfo`

NewFileInfo instantiates a new FileInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFileInfoWithDefaults

`func NewFileInfoWithDefaults() *FileInfo`

NewFileInfoWithDefaults instantiates a new FileInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGroup

`func (o *FileInfo) GetGroup() string`

GetGroup returns the Group field if non-nil, zero value otherwise.

### GetGroupOk

`func (o *FileInfo) GetGroupOk() (*string, bool)`

GetGroupOk returns a tuple with the Group field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroup

`func (o *FileInfo) SetGroup(v string)`

SetGroup sets Group field to given value.


### GetIsDir

`func (o *FileInfo) GetIsDir() bool`

GetIsDir returns the IsDir field if non-nil, zero value otherwise.

### GetIsDirOk

`func (o *FileInfo) GetIsDirOk() (*bool, bool)`

GetIsDirOk returns a tuple with the IsDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsDir

`func (o *FileInfo) SetIsDir(v bool)`

SetIsDir sets IsDir field to given value.


### GetModTime

`func (o *FileInfo) GetModTime() string`

GetModTime returns the ModTime field if non-nil, zero value otherwise.

### GetModTimeOk

`func (o *FileInfo) GetModTimeOk() (*string, bool)`

GetModTimeOk returns a tuple with the ModTime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetModTime

`func (o *FileInfo) SetModTime(v string)`

SetModTime sets ModTime field to given value.


### GetMode

`func (o *FileInfo) GetMode() string`

GetMode returns the Mode field if non-nil, zero value otherwise.

### GetModeOk

`func (o *FileInfo) GetModeOk() (*string, bool)`

GetModeOk returns a tuple with the Mode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMode

`func (o *FileInfo) SetMode(v string)`

SetMode sets Mode field to given value.


### GetName

`func (o *FileInfo) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *FileInfo) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *FileInfo) SetName(v string)`

SetName sets Name field to given value.


### GetOwner

`func (o *FileInfo) GetOwner() string`

GetOwner returns the Owner field if non-nil, zero value otherwise.

### GetOwnerOk

`func (o *FileInfo) GetOwnerOk() (*string, bool)`

GetOwnerOk returns a tuple with the Owner field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOwner

`func (o *FileInfo) SetOwner(v string)`

SetOwner sets Owner field to given value.


### GetPermissions

`func (o *FileInfo) GetPermissions() string`

GetPermissions returns the Permissions field if non-nil, zero value otherwise.

### GetPermissionsOk

`func (o *FileInfo) GetPermissionsOk() (*string, bool)`

GetPermissionsOk returns a tuple with the Permissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissions

`func (o *FileInfo) SetPermissions(v string)`

SetPermissions sets Permissions field to given value.


### GetSize

`func (o *FileInfo) GetSize() int32`

GetSize returns the Size field if non-nil, zero value otherwise.

### GetSizeOk

`func (o *FileInfo) GetSizeOk() (*int32, bool)`

GetSizeOk returns a tuple with the Size field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSize

`func (o *FileInfo) SetSize(v int32)`

SetSize sets Size field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


