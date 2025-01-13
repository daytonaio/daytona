# \ContainerRegistryAPI

All URIs are relative to *http://localhost:3986*

Method | HTTP request | Description
------------- | ------------- | -------------
[**FindContainerRegistry**](ContainerRegistryAPI.md#FindContainerRegistry) | **Get** /container-registry/{server} | Find container registry



## FindContainerRegistry

> ContainerRegistry FindContainerRegistry(ctx, server).WorkspaceId(workspaceId).Execute()

Find container registry



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
	server := "server_example" // string | Container registry server
	workspaceId := "workspaceId_example" // string | Workspace ID or Name (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ContainerRegistryAPI.FindContainerRegistry(context.Background(), server).WorkspaceId(workspaceId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ContainerRegistryAPI.FindContainerRegistry``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `FindContainerRegistry`: ContainerRegistry
	fmt.Fprintf(os.Stdout, "Response from `ContainerRegistryAPI.FindContainerRegistry`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**server** | **string** | Container registry server | 

### Other Parameters

Other parameters are passed through a pointer to a apiFindContainerRegistryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **workspaceId** | **string** | Workspace ID or Name | 

### Return type

[**ContainerRegistry**](ContainerRegistry.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

