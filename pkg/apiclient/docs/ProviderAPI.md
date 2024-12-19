# \ProviderAPI

All URIs are relative to *http://localhost:3986*

Method | HTTP request | Description
------------- | ------------- | -------------
[**InstallProvider**](ProviderAPI.md#InstallProvider) | **Post** /runner/{runnerId}/provider/install | Install provider
[**ListProviders**](ProviderAPI.md#ListProviders) | **Get** /runner/provider | List providers
[**UninstallProvider**](ProviderAPI.md#UninstallProvider) | **Post** /runner/{runnerId}/provider/{providerName}/uninstall | Uninstall provider
[**UpdateProvider**](ProviderAPI.md#UpdateProvider) | **Post** /runner/{runnerId}/provider/{providerName}/update | Update provider



## InstallProvider

> InstallProvider(ctx, runnerId).InstallProviderDto(installProviderDto).Execute()

Install provider



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
	runnerId := "runnerId_example" // string | Runner ID
	installProviderDto := *openapiclient.NewInstallProviderDTO(map[string]string{"key": "Inner_example"}, "Name_example") // InstallProviderDTO | Install provider

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ProviderAPI.InstallProvider(context.Background(), runnerId).InstallProviderDto(installProviderDto).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProviderAPI.InstallProvider``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**runnerId** | **string** | Runner ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiInstallProviderRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **installProviderDto** | [**InstallProviderDTO**](InstallProviderDTO.md) | Install provider | 

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


## ListProviders

> []ProviderInfo ListProviders(ctx).RunnerId(runnerId).Execute()

List providers



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
	runnerId := "runnerId_example" // string | Runner ID (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProviderAPI.ListProviders(context.Background()).RunnerId(runnerId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProviderAPI.ListProviders``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListProviders`: []ProviderInfo
	fmt.Fprintf(os.Stdout, "Response from `ProviderAPI.ListProviders`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListProvidersRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **runnerId** | **string** | Runner ID | 

### Return type

[**[]ProviderInfo**](ProviderInfo.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UninstallProvider

> UninstallProvider(ctx, runnerId, providerName).Execute()

Uninstall provider



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
	runnerId := "runnerId_example" // string | Runner ID
	providerName := "providerName_example" // string | Provider name

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ProviderAPI.UninstallProvider(context.Background(), runnerId, providerName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProviderAPI.UninstallProvider``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**runnerId** | **string** | Runner ID | 
**providerName** | **string** | Provider name | 

### Other Parameters

Other parameters are passed through a pointer to a apiUninstallProviderRequest struct via the builder pattern


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


## UpdateProvider

> UpdateProvider(ctx, runnerId, providerName).DownloadUrls(downloadUrls).Execute()

Update provider



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
	runnerId := "runnerId_example" // string | Runner ID
	providerName := "providerName_example" // string | Provider name
	downloadUrls := map[string]string{"key": "Inner_example"} // map[string]string | Provider download URLs

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ProviderAPI.UpdateProvider(context.Background(), runnerId, providerName).DownloadUrls(downloadUrls).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProviderAPI.UpdateProvider``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**runnerId** | **string** | Runner ID | 
**providerName** | **string** | Provider name | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateProviderRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **downloadUrls** | **map[string]string** | Provider download URLs | 

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

