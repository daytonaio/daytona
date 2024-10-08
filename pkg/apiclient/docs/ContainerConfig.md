# ContainerConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Image** | **string** |  | 
**User** | **string** |  | 

## Methods

### NewContainerConfig

`func NewContainerConfig(image string, user string, ) *ContainerConfig`

NewContainerConfig instantiates a new ContainerConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewContainerConfigWithDefaults

`func NewContainerConfigWithDefaults() *ContainerConfig`

NewContainerConfigWithDefaults instantiates a new ContainerConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetImage

`func (o *ContainerConfig) GetImage() string`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *ContainerConfig) GetImageOk() (*string, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *ContainerConfig) SetImage(v string)`

SetImage sets Image field to given value.


### GetUser

`func (o *ContainerConfig) GetUser() string`

GetUser returns the User field if non-nil, zero value otherwise.

### GetUserOk

`func (o *ContainerConfig) GetUserOk() (*string, bool)`

GetUserOk returns a tuple with the User field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUser

`func (o *ContainerConfig) SetUser(v string)`

SetUser sets User field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


