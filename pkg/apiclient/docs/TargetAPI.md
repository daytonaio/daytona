# \TargetAPI

All URIs are relative to *http://localhost:3986*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateTarget**](TargetAPI.md#CreateTarget) | **Post** /target | Create a target
[**GetTarget**](TargetAPI.md#GetTarget) | **Get** /target/{targetId} | Get target info
[**ListTargets**](TargetAPI.md#ListTargets) | **Get** /target | List targets
[**RemoveTarget**](TargetAPI.md#RemoveTarget) | **Delete** /target/{targetId} | Remove target
[**SetProjectState**](TargetAPI.md#SetProjectState) | **Post** /target/{targetId}/{projectId}/state | Set project state
[**StartProject**](TargetAPI.md#StartProject) | **Post** /target/{targetId}/{projectId}/start | Start project
[**StartTarget**](TargetAPI.md#StartTarget) | **Post** /target/{targetId}/start | Start target
[**StopProject**](TargetAPI.md#StopProject) | **Post** /target/{targetId}/{projectId}/stop | Stop project
[**StopTarget**](TargetAPI.md#StopTarget) | **Post** /target/{targetId}/stop | Stop target



## CreateTarget

> Target CreateTarget(ctx).Target(target).Execute()

Create a target



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
	target := *openapiclient.NewCreateTargetDTO("Id_example", "Name_example", []openapiclient.CreateProjectDTO{*openapiclient.NewCreateProjectDTO(map[string]string{"key": "Inner_example"}, "Name_example", *openapiclient.NewCreateProjectSourceDTO(*openapiclient.NewGitRepository("Branch_example", "Id_example", "Name_example", "Owner_example", "Sha_example", "Source_example", "Url_example")))}, "TargetConfig_example") // CreateTargetDTO | Create target

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TargetAPI.CreateTarget(context.Background()).Target(target).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TargetAPI.CreateTarget``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateTarget`: Target
	fmt.Fprintf(os.Stdout, "Response from `TargetAPI.CreateTarget`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateTargetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **target** | [**CreateTargetDTO**](CreateTargetDTO.md) | Create target | 

### Return type

[**Target**](Target.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetTarget

> TargetDTO GetTarget(ctx, targetId).Verbose(verbose).Execute()

Get target info



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
	targetId := "targetId_example" // string | Target ID or Name
	verbose := true // bool | Verbose (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TargetAPI.GetTarget(context.Background(), targetId).Verbose(verbose).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TargetAPI.GetTarget``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetTarget`: TargetDTO
	fmt.Fprintf(os.Stdout, "Response from `TargetAPI.GetTarget`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**targetId** | **string** | Target ID or Name | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetTargetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **verbose** | **bool** | Verbose | 

### Return type

[**TargetDTO**](TargetDTO.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListTargets

> []TargetDTO ListTargets(ctx).Verbose(verbose).Execute()

List targets



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
	verbose := true // bool | Verbose (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TargetAPI.ListTargets(context.Background()).Verbose(verbose).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TargetAPI.ListTargets``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListTargets`: []TargetDTO
	fmt.Fprintf(os.Stdout, "Response from `TargetAPI.ListTargets`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListTargetsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **verbose** | **bool** | Verbose | 

### Return type

[**[]TargetDTO**](TargetDTO.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RemoveTarget

> RemoveTarget(ctx, targetId).Force(force).Execute()

Remove target



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
	targetId := "targetId_example" // string | Target ID
	force := true // bool | Force (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.TargetAPI.RemoveTarget(context.Background(), targetId).Force(force).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TargetAPI.RemoveTarget``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**targetId** | **string** | Target ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiRemoveTargetRequest struct via the builder pattern


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


## SetProjectState

> SetProjectState(ctx, targetId, projectId).SetState(setState).Execute()

Set project state



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
	targetId := "targetId_example" // string | Target ID or Name
	projectId := "projectId_example" // string | Project ID
	setState := *openapiclient.NewSetProjectState(int32(123)) // SetProjectState | Set State

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.TargetAPI.SetProjectState(context.Background(), targetId, projectId).SetState(setState).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TargetAPI.SetProjectState``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**targetId** | **string** | Target ID or Name | 
**projectId** | **string** | Project ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiSetProjectStateRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **setState** | [**SetProjectState**](SetProjectState.md) | Set State | 

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


## StartProject

> StartProject(ctx, targetId, projectId).Execute()

Start project



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
	targetId := "targetId_example" // string | Target ID or Name
	projectId := "projectId_example" // string | Project ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.TargetAPI.StartProject(context.Background(), targetId, projectId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TargetAPI.StartProject``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**targetId** | **string** | Target ID or Name | 
**projectId** | **string** | Project ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiStartProjectRequest struct via the builder pattern


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


## StartTarget

> StartTarget(ctx, targetId).Execute()

Start target



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
	targetId := "targetId_example" // string | Target ID or Name

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.TargetAPI.StartTarget(context.Background(), targetId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TargetAPI.StartTarget``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**targetId** | **string** | Target ID or Name | 

### Other Parameters

Other parameters are passed through a pointer to a apiStartTargetRequest struct via the builder pattern


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


## StopProject

> StopProject(ctx, targetId, projectId).Execute()

Stop project



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
	targetId := "targetId_example" // string | Target ID or Name
	projectId := "projectId_example" // string | Project ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.TargetAPI.StopProject(context.Background(), targetId, projectId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TargetAPI.StopProject``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**targetId** | **string** | Target ID or Name | 
**projectId** | **string** | Project ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiStopProjectRequest struct via the builder pattern


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


## StopTarget

> StopTarget(ctx, targetId).Execute()

Stop target



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
	targetId := "targetId_example" // string | Target ID or Name

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.TargetAPI.StopTarget(context.Background(), targetId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TargetAPI.StopTarget``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**targetId** | **string** | Target ID or Name | 

### Other Parameters

Other parameters are passed through a pointer to a apiStopTargetRequest struct via the builder pattern


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

