# \ProjectConfigAPI

All URIs are relative to *http://localhost:3986*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeleteProjectConfig**](ProjectConfigAPI.md#DeleteProjectConfig) | **Delete** /project-config/{configName} | Delete project config data
[**GetDefaultProjectConfig**](ProjectConfigAPI.md#GetDefaultProjectConfig) | **Get** /project-config/default/{gitUrl} | Get project configs by git url
[**GetProjectConfig**](ProjectConfigAPI.md#GetProjectConfig) | **Get** /project-config/{configName} | Get project config data
[**ListProjectConfigs**](ProjectConfigAPI.md#ListProjectConfigs) | **Get** /project-config | List project configs
[**SetProjectConfig**](ProjectConfigAPI.md#SetProjectConfig) | **Put** /project-config | Set project config data



## DeleteProjectConfig

> DeleteProjectConfig(ctx, configName).Execute()

Delete project config data



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
	configName := "configName_example" // string | Config name

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ProjectConfigAPI.DeleteProjectConfig(context.Background(), configName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectConfigAPI.DeleteProjectConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**configName** | **string** | Config name | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteProjectConfigRequest struct via the builder pattern


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


## GetDefaultProjectConfig

> ProjectConfig GetDefaultProjectConfig(ctx, gitUrl).Execute()

Get project configs by git url



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
	gitUrl := "gitUrl_example" // string | Git URL

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProjectConfigAPI.GetDefaultProjectConfig(context.Background(), gitUrl).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectConfigAPI.GetDefaultProjectConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetDefaultProjectConfig`: ProjectConfig
	fmt.Fprintf(os.Stdout, "Response from `ProjectConfigAPI.GetDefaultProjectConfig`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**gitUrl** | **string** | Git URL | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDefaultProjectConfigRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ProjectConfig**](ProjectConfig.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetProjectConfig

> ProjectConfig GetProjectConfig(ctx, configName).Execute()

Get project config data



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
	configName := "configName_example" // string | Config name

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProjectConfigAPI.GetProjectConfig(context.Background(), configName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectConfigAPI.GetProjectConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetProjectConfig`: ProjectConfig
	fmt.Fprintf(os.Stdout, "Response from `ProjectConfigAPI.GetProjectConfig`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**configName** | **string** | Config name | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetProjectConfigRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ProjectConfig**](ProjectConfig.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListProjectConfigs

> []ProjectConfig ListProjectConfigs(ctx).Execute()

List project configs



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
	resp, r, err := apiClient.ProjectConfigAPI.ListProjectConfigs(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectConfigAPI.ListProjectConfigs``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListProjectConfigs`: []ProjectConfig
	fmt.Fprintf(os.Stdout, "Response from `ProjectConfigAPI.ListProjectConfigs`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiListProjectConfigsRequest struct via the builder pattern


### Return type

[**[]ProjectConfig**](ProjectConfig.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SetProjectConfig

> SetProjectConfig(ctx).ProjectConfig(projectConfig).Execute()

Set project config data



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
	projectConfig := *openapiclient.NewCreateProjectConfigDTO() // CreateProjectConfigDTO | Project config

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ProjectConfigAPI.SetProjectConfig(context.Background()).ProjectConfig(projectConfig).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectConfigAPI.SetProjectConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSetProjectConfigRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **projectConfig** | [**CreateProjectConfigDTO**](CreateProjectConfigDTO.md) | Project config | 

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

