# \TargetConfigAPI

All URIs are relative to *http://localhost:3986*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ListTargetConfigs**](TargetConfigAPI.md#ListTargetConfigs) | **Get** /target-config | List target configs
[**RemoveTargetConfig**](TargetConfigAPI.md#RemoveTargetConfig) | **Delete** /target-config/{configName} | Remove a target config
[**SetTargetConfig**](TargetConfigAPI.md#SetTargetConfig) | **Put** /target-config | Set a target config



## ListTargetConfigs

> []TargetConfig ListTargetConfigs(ctx).Execute()

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

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TargetConfigAPI.ListTargetConfigs(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TargetConfigAPI.ListTargetConfigs``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListTargetConfigs`: []TargetConfig
	fmt.Fprintf(os.Stdout, "Response from `TargetConfigAPI.ListTargetConfigs`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiListTargetConfigsRequest struct via the builder pattern


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

> RemoveTargetConfig(ctx, configName).Execute()

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
	configName := "configName_example" // string | Target Config name

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.TargetConfigAPI.RemoveTargetConfig(context.Background(), configName).Execute()
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
**configName** | **string** | Target Config name | 

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


## SetTargetConfig

> TargetConfig SetTargetConfig(ctx).TargetConfig(targetConfig).Execute()

Set a target config



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
	targetConfig := *openapiclient.NewCreateTargetConfigDTO("Name_example", "Options_example", *openapiclient.NewTargetProviderInfo("Name_example", "Version_example")) // CreateTargetConfigDTO | Target config to set

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TargetConfigAPI.SetTargetConfig(context.Background()).TargetConfig(targetConfig).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TargetConfigAPI.SetTargetConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SetTargetConfig`: TargetConfig
	fmt.Fprintf(os.Stdout, "Response from `TargetConfigAPI.SetTargetConfig`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSetTargetConfigRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **targetConfig** | [**CreateTargetConfigDTO**](CreateTargetConfigDTO.md) | Target config to set | 

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

