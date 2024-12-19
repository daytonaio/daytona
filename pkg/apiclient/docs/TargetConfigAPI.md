# \TargetConfigAPI

All URIs are relative to *http://localhost:3986*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AddTargetConfig**](TargetConfigAPI.md#AddTargetConfig) | **Put** /target-config | Add a target config
[**ListTargetConfigs**](TargetConfigAPI.md#ListTargetConfigs) | **Get** /target-config | List target configs
[**RemoveTargetConfig**](TargetConfigAPI.md#RemoveTargetConfig) | **Delete** /target-config/{configId} | Remove a target config



## AddTargetConfig

> TargetConfig AddTargetConfig(ctx).TargetConfig(targetConfig).ShowOptions(showOptions).Execute()

Add a target config



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
	targetConfig := *openapiclient.NewAddTargetConfigDTO("Name_example", "Options_example", *openapiclient.NewProviderInfo("Name_example", "RunnerId_example", "RunnerName_example", map[string]TargetConfigProperty{"key": *openapiclient.NewTargetConfigProperty()}, "Version_example")) // AddTargetConfigDTO | Target config to add
	showOptions := true // bool | Show target config options (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TargetConfigAPI.AddTargetConfig(context.Background()).TargetConfig(targetConfig).ShowOptions(showOptions).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TargetConfigAPI.AddTargetConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AddTargetConfig`: TargetConfig
	fmt.Fprintf(os.Stdout, "Response from `TargetConfigAPI.AddTargetConfig`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAddTargetConfigRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **targetConfig** | [**AddTargetConfigDTO**](AddTargetConfigDTO.md) | Target config to add | 
 **showOptions** | **bool** | Show target config options | 

### Return type

[**TargetConfig**](TargetConfig.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListTargetConfigs

> []TargetConfig ListTargetConfigs(ctx).ShowOptions(showOptions).Execute()

List target configs



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
	showOptions := true // bool | Show target config options (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TargetConfigAPI.ListTargetConfigs(context.Background()).ShowOptions(showOptions).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TargetConfigAPI.ListTargetConfigs``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListTargetConfigs`: []TargetConfig
	fmt.Fprintf(os.Stdout, "Response from `TargetConfigAPI.ListTargetConfigs`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListTargetConfigsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **showOptions** | **bool** | Show target config options | 

### Return type

[**[]TargetConfig**](TargetConfig.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RemoveTargetConfig

> RemoveTargetConfig(ctx, configId).Execute()

Remove a target config



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
	configId := "configId_example" // string | Target Config Id

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.TargetConfigAPI.RemoveTargetConfig(context.Background(), configId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TargetConfigAPI.RemoveTargetConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**configId** | **string** | Target Config Id | 

### Other Parameters

Other parameters are passed through a pointer to a apiRemoveTargetConfigRequest struct via the builder pattern


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

