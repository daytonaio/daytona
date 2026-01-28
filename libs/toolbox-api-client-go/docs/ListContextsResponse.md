# ListContextsResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Contexts** | [**[]InterpreterContext**](InterpreterContext.md) |  | 

## Methods

### NewListContextsResponse

`func NewListContextsResponse(contexts []InterpreterContext, ) *ListContextsResponse`

NewListContextsResponse instantiates a new ListContextsResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewListContextsResponseWithDefaults

`func NewListContextsResponseWithDefaults() *ListContextsResponse`

NewListContextsResponseWithDefaults instantiates a new ListContextsResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetContexts

`func (o *ListContextsResponse) GetContexts() []InterpreterContext`

GetContexts returns the Contexts field if non-nil, zero value otherwise.

### GetContextsOk

`func (o *ListContextsResponse) GetContextsOk() (*[]InterpreterContext, bool)`

GetContextsOk returns a tuple with the Contexts field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContexts

`func (o *ListContextsResponse) SetContexts(v []InterpreterContext)`

SetContexts sets Contexts field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


