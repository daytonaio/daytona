# \PrebuildAPI

All URIs are relative to *http://localhost:3986*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeletePrebuild**](PrebuildAPI.md#DeletePrebuild) | **Delete** /workspace-template/{templateName}/prebuild/{prebuildId} | Delete prebuild
[**FindPrebuild**](PrebuildAPI.md#FindPrebuild) | **Get** /workspace-template/{templateName}/prebuild/{prebuildId} | Find prebuild
[**ListPrebuilds**](PrebuildAPI.md#ListPrebuilds) | **Get** /workspace-template/prebuild | List prebuilds
[**ListPrebuildsForWorkspaceTemplate**](PrebuildAPI.md#ListPrebuildsForWorkspaceTemplate) | **Get** /workspace-template/{templateName}/prebuild | List prebuilds for workspace template
[**ProcessGitEvent**](PrebuildAPI.md#ProcessGitEvent) | **Post** /workspace-template/prebuild/process-git-event | ProcessGitEvent
[**SavePrebuild**](PrebuildAPI.md#SavePrebuild) | **Put** /workspace-template/{templateName}/prebuild | Save prebuild



## DeletePrebuild

> DeletePrebuild(ctx, templateName, prebuildId).Force(force).Execute()

Delete prebuild



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
	templateName := "templateName_example" // string | Workspace template name
	prebuildId := "prebuildId_example" // string | Prebuild ID
	force := true // bool | Force (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PrebuildAPI.DeletePrebuild(context.Background(), templateName, prebuildId).Force(force).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PrebuildAPI.DeletePrebuild``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**templateName** | **string** | Workspace template name | 
**prebuildId** | **string** | Prebuild ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeletePrebuildRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **force** | **bool** | Force | 

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


## FindPrebuild

> PrebuildDTO FindPrebuild(ctx, templateName, prebuildId).Execute()

Find prebuild



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
	templateName := "templateName_example" // string | Workspace template name
	prebuildId := "prebuildId_example" // string | Prebuild ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PrebuildAPI.FindPrebuild(context.Background(), templateName, prebuildId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PrebuildAPI.FindPrebuild``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `FindPrebuild`: PrebuildDTO
	fmt.Fprintf(os.Stdout, "Response from `PrebuildAPI.FindPrebuild`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**templateName** | **string** | Workspace template name | 
**prebuildId** | **string** | Prebuild ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiFindPrebuildRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**PrebuildDTO**](PrebuildDTO.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListPrebuilds

> []PrebuildDTO ListPrebuilds(ctx).Execute()

List prebuilds



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
	resp, r, err := apiClient.PrebuildAPI.ListPrebuilds(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PrebuildAPI.ListPrebuilds``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListPrebuilds`: []PrebuildDTO
	fmt.Fprintf(os.Stdout, "Response from `PrebuildAPI.ListPrebuilds`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiListPrebuildsRequest struct via the builder pattern


### Return type

[**[]PrebuildDTO**](PrebuildDTO.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListPrebuildsForWorkspaceTemplate

> []PrebuildDTO ListPrebuildsForWorkspaceTemplate(ctx, templateName).Execute()

List prebuilds for workspace template



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
	templateName := "templateName_example" // string | Template name

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PrebuildAPI.ListPrebuildsForWorkspaceTemplate(context.Background(), templateName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PrebuildAPI.ListPrebuildsForWorkspaceTemplate``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListPrebuildsForWorkspaceTemplate`: []PrebuildDTO
	fmt.Fprintf(os.Stdout, "Response from `PrebuildAPI.ListPrebuildsForWorkspaceTemplate`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**templateName** | **string** | Template name | 

### Other Parameters

Other parameters are passed through a pointer to a apiListPrebuildsForWorkspaceTemplateRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]PrebuildDTO**](PrebuildDTO.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ProcessGitEvent

> ProcessGitEvent(ctx).Body(body).Execute()

ProcessGitEvent



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
	body := map[string]interface{}{ ... } // map[string]interface{} | Webhook event

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PrebuildAPI.ProcessGitEvent(context.Background()).Body(body).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PrebuildAPI.ProcessGitEvent``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiProcessGitEventRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | **map[string]interface{}** | Webhook event | 

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


## SavePrebuild

> string SavePrebuild(ctx, templateName).Prebuild(prebuild).Execute()

Save prebuild



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
	templateName := "templateName_example" // string | Template name
	prebuild := *openapiclient.NewCreatePrebuildDTO(int32(123)) // CreatePrebuildDTO | Prebuild

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PrebuildAPI.SavePrebuild(context.Background(), templateName).Prebuild(prebuild).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PrebuildAPI.SavePrebuild``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SavePrebuild`: string
	fmt.Fprintf(os.Stdout, "Response from `PrebuildAPI.SavePrebuild`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**templateName** | **string** | Template name | 

### Other Parameters

Other parameters are passed through a pointer to a apiSavePrebuildRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **prebuild** | [**CreatePrebuildDTO**](CreatePrebuildDTO.md) | Prebuild | 

### Return type

**string**

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

