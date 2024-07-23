# \WorkspaceAPI

All URIs are relative to *http://localhost:3986*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateWorkspace**](WorkspaceAPI.md#CreateWorkspace) | **Post** /workspace | Create a workspace
[**GetWorkspace**](WorkspaceAPI.md#GetWorkspace) | **Get** /workspace/{workspaceId} | Get workspace info
[**ListWorkspaces**](WorkspaceAPI.md#ListWorkspaces) | **Get** /workspace | List workspaces
[**RemoveWorkspace**](WorkspaceAPI.md#RemoveWorkspace) | **Delete** /workspace/{workspaceId} | Remove workspace
[**SetProjectState**](WorkspaceAPI.md#SetProjectState) | **Post** /workspace/{workspaceId}/{projectId}/state | Set project state
[**StartProject**](WorkspaceAPI.md#StartProject) | **Post** /workspace/{workspaceId}/{projectId}/start | Start project
[**StartWorkspace**](WorkspaceAPI.md#StartWorkspace) | **Post** /workspace/{workspaceId}/start | Start workspace
[**StopProject**](WorkspaceAPI.md#StopProject) | **Post** /workspace/{workspaceId}/{projectId}/stop | Stop project
[**StopWorkspace**](WorkspaceAPI.md#StopWorkspace) | **Post** /workspace/{workspaceId}/stop | Stop workspace



## CreateWorkspace

> Workspace CreateWorkspace(ctx).Workspace(workspace).Execute()

Create a workspace



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
	workspace := *openapiclient.NewCreateWorkspaceDTO([]openapiclient.CreateProjectDTO{*openapiclient.NewCreateProjectDTO()}) // CreateWorkspaceDTO | Create workspace

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkspaceAPI.CreateWorkspace(context.Background()).Workspace(workspace).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceAPI.CreateWorkspace``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateWorkspace`: Workspace
	fmt.Fprintf(os.Stdout, "Response from `WorkspaceAPI.CreateWorkspace`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateWorkspaceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **workspace** | [**CreateWorkspaceDTO**](CreateWorkspaceDTO.md) | Create workspace | 

### Return type

[**Workspace**](Workspace.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetWorkspace

> WorkspaceDTO GetWorkspace(ctx, workspaceId).Execute()

Get workspace info



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
	workspaceId := "workspaceId_example" // string | Workspace ID or Name

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkspaceAPI.GetWorkspace(context.Background(), workspaceId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceAPI.GetWorkspace``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWorkspace`: WorkspaceDTO
	fmt.Fprintf(os.Stdout, "Response from `WorkspaceAPI.GetWorkspace`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workspaceId** | **string** | Workspace ID or Name | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkspaceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**WorkspaceDTO**](WorkspaceDTO.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListWorkspaces

> []WorkspaceDTO ListWorkspaces(ctx).Verbose(verbose).Execute()

List workspaces



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
	resp, r, err := apiClient.WorkspaceAPI.ListWorkspaces(context.Background()).Verbose(verbose).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceAPI.ListWorkspaces``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListWorkspaces`: []WorkspaceDTO
	fmt.Fprintf(os.Stdout, "Response from `WorkspaceAPI.ListWorkspaces`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListWorkspacesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **verbose** | **bool** | Verbose | 

### Return type

[**[]WorkspaceDTO**](WorkspaceDTO.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RemoveWorkspace

> RemoveWorkspace(ctx, workspaceId).Force(force).Execute()

Remove workspace



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
	workspaceId := "workspaceId_example" // string | Workspace ID
	force := true // bool | Force (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkspaceAPI.RemoveWorkspace(context.Background(), workspaceId).Force(force).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceAPI.RemoveWorkspace``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workspaceId** | **string** | Workspace ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiRemoveWorkspaceRequest struct via the builder pattern


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

> SetProjectState(ctx, workspaceId, projectId).SetState(setState).Execute()

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
	workspaceId := "workspaceId_example" // string | Workspace ID or Name
	projectId := "projectId_example" // string | Project ID
	setState := *openapiclient.NewSetProjectState() // SetProjectState | Set State

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkspaceAPI.SetProjectState(context.Background(), workspaceId, projectId).SetState(setState).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceAPI.SetProjectState``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workspaceId** | **string** | Workspace ID or Name | 
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

> StartProject(ctx, workspaceId, projectId).Execute()

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
	workspaceId := "workspaceId_example" // string | Workspace ID or Name
	projectId := "projectId_example" // string | Project ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkspaceAPI.StartProject(context.Background(), workspaceId, projectId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceAPI.StartProject``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workspaceId** | **string** | Workspace ID or Name | 
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


## StartWorkspace

> StartWorkspace(ctx, workspaceId).Execute()

Start workspace



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
	workspaceId := "workspaceId_example" // string | Workspace ID or Name

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkspaceAPI.StartWorkspace(context.Background(), workspaceId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceAPI.StartWorkspace``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workspaceId** | **string** | Workspace ID or Name | 

### Other Parameters

Other parameters are passed through a pointer to a apiStartWorkspaceRequest struct via the builder pattern


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

> StopProject(ctx, workspaceId, projectId).Execute()

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
	workspaceId := "workspaceId_example" // string | Workspace ID or Name
	projectId := "projectId_example" // string | Project ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkspaceAPI.StopProject(context.Background(), workspaceId, projectId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceAPI.StopProject``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workspaceId** | **string** | Workspace ID or Name | 
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


## StopWorkspace

> StopWorkspace(ctx, workspaceId).Execute()

Stop workspace



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
	workspaceId := "workspaceId_example" // string | Workspace ID or Name

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkspaceAPI.StopWorkspace(context.Background(), workspaceId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceAPI.StopWorkspace``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workspaceId** | **string** | Workspace ID or Name | 

### Other Parameters

Other parameters are passed through a pointer to a apiStopWorkspaceRequest struct via the builder pattern


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

