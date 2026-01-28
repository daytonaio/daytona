# \InterpreterAPI

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateInterpreterContext**](InterpreterAPI.md#CreateInterpreterContext) | **Post** /process/interpreter/context | Create a new interpreter context
[**DeleteInterpreterContext**](InterpreterAPI.md#DeleteInterpreterContext) | **Delete** /process/interpreter/context/{id} | Delete an interpreter context
[**ExecuteInterpreterCode**](InterpreterAPI.md#ExecuteInterpreterCode) | **Get** /process/interpreter/execute | Execute code in an interpreter context
[**ListInterpreterContexts**](InterpreterAPI.md#ListInterpreterContexts) | **Get** /process/interpreter/context | List all user-created interpreter contexts



## CreateInterpreterContext

> InterpreterContext CreateInterpreterContext(ctx).Request(request).Execute()

Create a new interpreter context



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-client-go"
)

func main() {
	request := *openapiclient.NewCreateContextRequest() // CreateContextRequest | Context configuration

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InterpreterAPI.CreateInterpreterContext(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InterpreterAPI.CreateInterpreterContext``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateInterpreterContext`: InterpreterContext
	fmt.Fprintf(os.Stdout, "Response from `InterpreterAPI.CreateInterpreterContext`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateInterpreterContextRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**CreateContextRequest**](CreateContextRequest.md) | Context configuration | 

### Return type

[**InterpreterContext**](InterpreterContext.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteInterpreterContext

> map[string]string DeleteInterpreterContext(ctx, id).Execute()

Delete an interpreter context



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-client-go"
)

func main() {
	id := "id_example" // string | Context ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InterpreterAPI.DeleteInterpreterContext(context.Background(), id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InterpreterAPI.DeleteInterpreterContext``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteInterpreterContext`: map[string]string
	fmt.Fprintf(os.Stdout, "Response from `InterpreterAPI.DeleteInterpreterContext`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | Context ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteInterpreterContextRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

**map[string]string**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ExecuteInterpreterCode

> ExecuteInterpreterCode(ctx).Execute()

Execute code in an interpreter context



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-client-go"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.InterpreterAPI.ExecuteInterpreterCode(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InterpreterAPI.ExecuteInterpreterCode``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiExecuteInterpreterCodeRequest struct via the builder pattern


### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListInterpreterContexts

> ListContextsResponse ListInterpreterContexts(ctx).Execute()

List all user-created interpreter contexts



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-client-go"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InterpreterAPI.ListInterpreterContexts(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InterpreterAPI.ListInterpreterContexts``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListInterpreterContexts`: ListContextsResponse
	fmt.Fprintf(os.Stdout, "Response from `InterpreterAPI.ListInterpreterContexts`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiListInterpreterContextsRequest struct via the builder pattern


### Return type

[**ListContextsResponse**](ListContextsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

