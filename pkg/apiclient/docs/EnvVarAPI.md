# \EnvVarAPI

All URIs are relative to *http://localhost:3986*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeleteEnvironmentVariable**](EnvVarAPI.md#DeleteEnvironmentVariable) | **Delete** /env/{key} | Delete environment variable
[**ListEnvironmentVariables**](EnvVarAPI.md#ListEnvironmentVariables) | **Get** /env | List environment variables
[**SetEnvironmentVariable**](EnvVarAPI.md#SetEnvironmentVariable) | **Put** /env | Set environment variable



## DeleteEnvironmentVariable

> DeleteEnvironmentVariable(ctx, key).Execute()

Delete environment variable



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID/apiclient"
)

func main() {
	key := "key_example" // string | Environment Variable Key

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.EnvVarAPI.DeleteEnvironmentVariable(context.Background(), key).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `EnvVarAPI.DeleteEnvironmentVariable``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**key** | **string** | Environment Variable Key | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteEnvironmentVariableRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

 (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListEnvironmentVariables

> []EnvironmentVariable ListEnvironmentVariables(ctx).Execute()

List environment variables



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID/apiclient"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.EnvVarAPI.ListEnvironmentVariables(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `EnvVarAPI.ListEnvironmentVariables``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListEnvironmentVariables`: []EnvironmentVariable
	fmt.Fprintf(os.Stdout, "Response from `EnvVarAPI.ListEnvironmentVariables`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiListEnvironmentVariablesRequest struct via the builder pattern


### Return type

[**[]EnvironmentVariable**](EnvironmentVariable.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SetEnvironmentVariable

> SetEnvironmentVariable(ctx).EnvironmentVariable(environmentVariable).Execute()

Set environment variable



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID/apiclient"
)

func main() {
	environmentVariable := *openapiclient.NewEnvironmentVariable("Key_example", "Value_example") // EnvironmentVariable | Environment Variable

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.EnvVarAPI.SetEnvironmentVariable(context.Background()).EnvironmentVariable(environmentVariable).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `EnvVarAPI.SetEnvironmentVariable``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSetEnvironmentVariableRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **environmentVariable** | [**EnvironmentVariable**](EnvironmentVariable.md) | Environment Variable | 

### Return type

 (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

