# \ContainerRegistryAPI

All URIs are relative to *http://localhost:3986*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetContainerRegistry**](ContainerRegistryAPI.md#GetContainerRegistry) | **Get** /container-registry/{server} | Get container registry credentials
[**ListContainerRegistries**](ContainerRegistryAPI.md#ListContainerRegistries) | **Get** /container-registry | List container registries
[**RemoveContainerRegistry**](ContainerRegistryAPI.md#RemoveContainerRegistry) | **Delete** /container-registry/{server} | Remove a container registry credentials
[**SetContainerRegistry**](ContainerRegistryAPI.md#SetContainerRegistry) | **Put** /container-registry/{server} | Set container registry credentials



## GetContainerRegistry

> ContainerRegistry GetContainerRegistry(ctx, server).Execute()

Get container registry credentials



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
	server := "server_example" // string | Container Registry server name

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ContainerRegistryAPI.GetContainerRegistry(context.Background(), server).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ContainerRegistryAPI.GetContainerRegistry``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetContainerRegistry`: ContainerRegistry
	fmt.Fprintf(os.Stdout, "Response from `ContainerRegistryAPI.GetContainerRegistry`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**server** | **string** | Container Registry server name | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetContainerRegistryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


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


## ListContainerRegistries

> []ContainerRegistry ListContainerRegistries(ctx).Execute()

List container registries



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
	resp, r, err := apiClient.ContainerRegistryAPI.ListContainerRegistries(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ContainerRegistryAPI.ListContainerRegistries``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListContainerRegistries`: []ContainerRegistry
	fmt.Fprintf(os.Stdout, "Response from `ContainerRegistryAPI.ListContainerRegistries`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiListContainerRegistriesRequest struct via the builder pattern


### Return type

[**[]ContainerRegistry**](ContainerRegistry.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RemoveContainerRegistry

> RemoveContainerRegistry(ctx, server).Execute()

Remove a container registry credentials



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
	server := "server_example" // string | Container Registry server name

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ContainerRegistryAPI.RemoveContainerRegistry(context.Background(), server).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ContainerRegistryAPI.RemoveContainerRegistry``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**server** | **string** | Container Registry server name | 

### Other Parameters

Other parameters are passed through a pointer to a apiRemoveContainerRegistryRequest struct via the builder pattern


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


## SetContainerRegistry

> SetContainerRegistry(ctx, server).ContainerRegistry(containerRegistry).Execute()

Set container registry credentials



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
	server := "server_example" // string | Container Registry server name
	containerRegistry := *openapiclient.NewContainerRegistry("Password_example", "Server_example", "Username_example") // ContainerRegistry | Container Registry credentials to set

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ContainerRegistryAPI.SetContainerRegistry(context.Background(), server).ContainerRegistry(containerRegistry).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ContainerRegistryAPI.SetContainerRegistry``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**server** | **string** | Container Registry server name | 

### Other Parameters

Other parameters are passed through a pointer to a apiSetContainerRegistryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **containerRegistry** | [**ContainerRegistry**](ContainerRegistry.md) | Container Registry credentials to set | 

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

