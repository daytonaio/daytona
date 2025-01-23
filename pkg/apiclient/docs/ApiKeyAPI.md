# \ApiKeyAPI

All URIs are relative to *http://localhost:3986*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateApiKey**](ApiKeyAPI.md#CreateApiKey) | **Post** /apikey/{apiKeyName} | Create an API key
[**DeleteApiKey**](ApiKeyAPI.md#DeleteApiKey) | **Delete** /apikey/{apiKeyName} | Delete API key
[**ListClientApiKeys**](ApiKeyAPI.md#ListClientApiKeys) | **Get** /apikey | List API keys



## CreateApiKey

> string CreateApiKey(ctx, apiKeyName).Execute()

Create an API key



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
	apiKeyName := "apiKeyName_example" // string | API key name

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ApiKeyAPI.CreateApiKey(context.Background(), apiKeyName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ApiKeyAPI.CreateApiKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateApiKey`: string
	fmt.Fprintf(os.Stdout, "Response from `ApiKeyAPI.CreateApiKey`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**apiKeyName** | **string** | API key name | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateApiKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

**string**

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: text/plain

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteApiKey

> DeleteApiKey(ctx, apiKeyName).Execute()

Delete API key



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
	apiKeyName := "apiKeyName_example" // string | API key name

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ApiKeyAPI.DeleteApiKey(context.Background(), apiKeyName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ApiKeyAPI.DeleteApiKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**apiKeyName** | **string** | API key name | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteApiKeyRequest struct via the builder pattern


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


## ListClientApiKeys

> []ApiKeyViewDTO ListClientApiKeys(ctx).Execute()

List API keys



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
	resp, r, err := apiClient.ApiKeyAPI.ListClientApiKeys(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ApiKeyAPI.ListClientApiKeys``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListClientApiKeys`: []ApiKeyViewDTO
	fmt.Fprintf(os.Stdout, "Response from `ApiKeyAPI.ListClientApiKeys`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiListClientApiKeysRequest struct via the builder pattern


### Return type

[**[]ApiKeyViewDTO**](ApiKeyViewDTO.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

