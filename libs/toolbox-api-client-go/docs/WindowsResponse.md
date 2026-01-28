# WindowsResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Windows** | Pointer to [**[]WindowInfo**](WindowInfo.md) |  | [optional] 

## Methods

### NewWindowsResponse

`func NewWindowsResponse() *WindowsResponse`

NewWindowsResponse instantiates a new WindowsResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWindowsResponseWithDefaults

`func NewWindowsResponseWithDefaults() *WindowsResponse`

NewWindowsResponseWithDefaults instantiates a new WindowsResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetWindows

`func (o *WindowsResponse) GetWindows() []WindowInfo`

GetWindows returns the Windows field if non-nil, zero value otherwise.

### GetWindowsOk

`func (o *WindowsResponse) GetWindowsOk() (*[]WindowInfo, bool)`

GetWindowsOk returns a tuple with the Windows field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWindows

`func (o *WindowsResponse) SetWindows(v []WindowInfo)`

SetWindows sets Windows field to given value.

### HasWindows

`func (o *WindowsResponse) HasWindows() bool`

HasWindows returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


