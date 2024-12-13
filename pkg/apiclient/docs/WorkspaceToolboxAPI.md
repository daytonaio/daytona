# \WorkspaceToolboxAPI

All URIs are relative to *http://localhost:3986*

Method | HTTP request | Description
------------- | ------------- | -------------
[**FsCreateFolder**](WorkspaceToolboxAPI.md#FsCreateFolder) | **Post** /workspace/{workspaceId}/{projectId}/toolbox/files/folder | Create folder
[**FsDeleteFile**](WorkspaceToolboxAPI.md#FsDeleteFile) | **Delete** /workspace/{workspaceId}/{projectId}/toolbox/files | Delete file
[**FsDownloadFile**](WorkspaceToolboxAPI.md#FsDownloadFile) | **Get** /workspace/{workspaceId}/{projectId}/toolbox/files/download | Download file
[**FsFindInFiles**](WorkspaceToolboxAPI.md#FsFindInFiles) | **Get** /workspace/{workspaceId}/{projectId}/toolbox/files/find | Search for text/pattern in files
[**FsGetFileDetails**](WorkspaceToolboxAPI.md#FsGetFileDetails) | **Get** /workspace/{workspaceId}/{projectId}/toolbox/files/info | Get file info
[**FsListFiles**](WorkspaceToolboxAPI.md#FsListFiles) | **Get** /workspace/{workspaceId}/{projectId}/toolbox/files | List files
[**FsMoveFile**](WorkspaceToolboxAPI.md#FsMoveFile) | **Post** /workspace/{workspaceId}/{projectId}/toolbox/files/move | Create folder
[**FsReplaceInFiles**](WorkspaceToolboxAPI.md#FsReplaceInFiles) | **Post** /workspace/{workspaceId}/{projectId}/toolbox/files/replace | Repleace text/pattern in files
[**FsSearchFiles**](WorkspaceToolboxAPI.md#FsSearchFiles) | **Get** /workspace/{workspaceId}/{projectId}/toolbox/files/search | Search for files
[**FsSetFilePermissions**](WorkspaceToolboxAPI.md#FsSetFilePermissions) | **Post** /workspace/{workspaceId}/{projectId}/toolbox/files/permissions | Set file owner/group/permissions
[**FsUploadFile**](WorkspaceToolboxAPI.md#FsUploadFile) | **Post** /workspace/{workspaceId}/{projectId}/toolbox/files/upload | Upload file
[**GetProjectDir**](WorkspaceToolboxAPI.md#GetProjectDir) | **Get** /workspace/{workspaceId}/{projectId}/toolbox/project-dir | Get project dir
[**GitAddFiles**](WorkspaceToolboxAPI.md#GitAddFiles) | **Post** /workspace/{workspaceId}/{projectId}/toolbox/git/add | Add files
[**GitBranchList**](WorkspaceToolboxAPI.md#GitBranchList) | **Get** /workspace/{workspaceId}/{projectId}/toolbox/git/branches | Get branch list
[**GitCloneRepository**](WorkspaceToolboxAPI.md#GitCloneRepository) | **Post** /workspace/{workspaceId}/{projectId}/toolbox/git/clone | Clone git repository
[**GitCommitChanges**](WorkspaceToolboxAPI.md#GitCommitChanges) | **Post** /workspace/{workspaceId}/{projectId}/toolbox/git/commit | Commit changes
[**GitCommitHistory**](WorkspaceToolboxAPI.md#GitCommitHistory) | **Get** /workspace/{workspaceId}/{projectId}/toolbox/git/history | Get commit history
[**GitCreateBranch**](WorkspaceToolboxAPI.md#GitCreateBranch) | **Post** /workspace/{workspaceId}/{projectId}/toolbox/git/branches | Create branch
[**GitGitStatus**](WorkspaceToolboxAPI.md#GitGitStatus) | **Get** /workspace/{workspaceId}/{projectId}/toolbox/git/status | Get git status
[**GitPullChanges**](WorkspaceToolboxAPI.md#GitPullChanges) | **Post** /workspace/{workspaceId}/{projectId}/toolbox/git/pull | Pull changes
[**GitPushChanges**](WorkspaceToolboxAPI.md#GitPushChanges) | **Post** /workspace/{workspaceId}/{projectId}/toolbox/git/push | Push changes
[**LspCompletions**](WorkspaceToolboxAPI.md#LspCompletions) | **Post** /workspace/{workspaceId}/{projectId}/toolbox/lsp/completions | Get Lsp Completions
[**LspDidClose**](WorkspaceToolboxAPI.md#LspDidClose) | **Post** /workspace/{workspaceId}/{projectId}/toolbox/lsp/did-close | Call Lsp DidClose
[**LspDidOpen**](WorkspaceToolboxAPI.md#LspDidOpen) | **Post** /workspace/{workspaceId}/{projectId}/toolbox/lsp/did-open | Call Lsp DidOpen
[**LspDocumentSymbols**](WorkspaceToolboxAPI.md#LspDocumentSymbols) | **Get** /workspace/{workspaceId}/{projectId}/toolbox/lsp/document-symbols | Call Lsp DocumentSymbols
[**LspStart**](WorkspaceToolboxAPI.md#LspStart) | **Post** /workspace/{workspaceId}/{projectId}/toolbox/lsp/start | Start Lsp server
[**LspStop**](WorkspaceToolboxAPI.md#LspStop) | **Post** /workspace/{workspaceId}/{projectId}/toolbox/lsp/stop | Stop Lsp server
[**LspWorkspaceSymbols**](WorkspaceToolboxAPI.md#LspWorkspaceSymbols) | **Get** /workspace/{workspaceId}/{projectId}/toolbox/lsp/workspace-symbols | Call Lsp WorkspaceSymbols
[**ProcessExecuteCommand**](WorkspaceToolboxAPI.md#ProcessExecuteCommand) | **Post** /workspace/{workspaceId}/{projectId}/toolbox/process/execute | Execute command



## FsCreateFolder

> FsCreateFolder(ctx, workspaceId, projectId).Path(path).Mode(mode).Execute()

Create folder



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
	path := "path_example" // string | Path
	mode := "mode_example" // string | Mode

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkspaceToolboxAPI.FsCreateFolder(context.Background(), workspaceId, projectId).Path(path).Mode(mode).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceToolboxAPI.FsCreateFolder``: %v\n", err)
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

Other parameters are passed through a pointer to a apiFsCreateFolderRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **path** | **string** | Path | 
 **mode** | **string** | Mode | 

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


## FsDeleteFile

> FsDeleteFile(ctx, workspaceId, projectId).Path(path).Execute()

Delete file



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
	path := "path_example" // string | Path

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkspaceToolboxAPI.FsDeleteFile(context.Background(), workspaceId, projectId).Path(path).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceToolboxAPI.FsDeleteFile``: %v\n", err)
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

Other parameters are passed through a pointer to a apiFsDeleteFileRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **path** | **string** | Path | 

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


## FsDownloadFile

> *os.File FsDownloadFile(ctx, workspaceId, projectId).Path(path).Execute()

Download file



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
	path := "path_example" // string | Path

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkspaceToolboxAPI.FsDownloadFile(context.Background(), workspaceId, projectId).Path(path).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceToolboxAPI.FsDownloadFile``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `FsDownloadFile`: *os.File
	fmt.Fprintf(os.Stdout, "Response from `WorkspaceToolboxAPI.FsDownloadFile`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workspaceId** | **string** | Workspace ID or Name | 
**projectId** | **string** | Project ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiFsDownloadFileRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **path** | **string** | Path | 

### Return type

[***os.File**](*os.File.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## FsFindInFiles

> []Match FsFindInFiles(ctx, workspaceId, projectId).Path(path).Pattern(pattern).Execute()

Search for text/pattern in files



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
	path := "path_example" // string | Path
	pattern := "pattern_example" // string | Pattern

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkspaceToolboxAPI.FsFindInFiles(context.Background(), workspaceId, projectId).Path(path).Pattern(pattern).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceToolboxAPI.FsFindInFiles``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `FsFindInFiles`: []Match
	fmt.Fprintf(os.Stdout, "Response from `WorkspaceToolboxAPI.FsFindInFiles`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workspaceId** | **string** | Workspace ID or Name | 
**projectId** | **string** | Project ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiFsFindInFilesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **path** | **string** | Path | 
 **pattern** | **string** | Pattern | 

### Return type

[**[]Match**](Match.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## FsGetFileDetails

> FileInfo FsGetFileDetails(ctx, workspaceId, projectId).Path(path).Execute()

Get file info



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
	path := "path_example" // string | Path

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkspaceToolboxAPI.FsGetFileDetails(context.Background(), workspaceId, projectId).Path(path).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceToolboxAPI.FsGetFileDetails``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `FsGetFileDetails`: FileInfo
	fmt.Fprintf(os.Stdout, "Response from `WorkspaceToolboxAPI.FsGetFileDetails`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workspaceId** | **string** | Workspace ID or Name | 
**projectId** | **string** | Project ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiFsGetFileDetailsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **path** | **string** | Path | 

### Return type

[**FileInfo**](FileInfo.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## FsListFiles

> []FileInfo FsListFiles(ctx, workspaceId, projectId).Path(path).Execute()

List files



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
	path := "path_example" // string | Path (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkspaceToolboxAPI.FsListFiles(context.Background(), workspaceId, projectId).Path(path).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceToolboxAPI.FsListFiles``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `FsListFiles`: []FileInfo
	fmt.Fprintf(os.Stdout, "Response from `WorkspaceToolboxAPI.FsListFiles`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workspaceId** | **string** | Workspace ID or Name | 
**projectId** | **string** | Project ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiFsListFilesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **path** | **string** | Path | 

### Return type

[**[]FileInfo**](FileInfo.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## FsMoveFile

> FsMoveFile(ctx, workspaceId, projectId).Source(source).Destination(destination).Execute()

Create folder



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
	source := "source_example" // string | Source path
	destination := "destination_example" // string | Destination path

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkspaceToolboxAPI.FsMoveFile(context.Background(), workspaceId, projectId).Source(source).Destination(destination).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceToolboxAPI.FsMoveFile``: %v\n", err)
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

Other parameters are passed through a pointer to a apiFsMoveFileRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **source** | **string** | Source path | 
 **destination** | **string** | Destination path | 

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


## FsReplaceInFiles

> []ReplaceResult FsReplaceInFiles(ctx, workspaceId, projectId).Replace(replace).Execute()

Repleace text/pattern in files



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
	replace := *openapiclient.NewReplaceRequest([]string{"Files_example"}, "NewValue_example", "Pattern_example") // ReplaceRequest | ReplaceParams

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkspaceToolboxAPI.FsReplaceInFiles(context.Background(), workspaceId, projectId).Replace(replace).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceToolboxAPI.FsReplaceInFiles``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `FsReplaceInFiles`: []ReplaceResult
	fmt.Fprintf(os.Stdout, "Response from `WorkspaceToolboxAPI.FsReplaceInFiles`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workspaceId** | **string** | Workspace ID or Name | 
**projectId** | **string** | Project ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiFsReplaceInFilesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **replace** | [**ReplaceRequest**](ReplaceRequest.md) | ReplaceParams | 

### Return type

[**[]ReplaceResult**](ReplaceResult.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## FsSearchFiles

> SearchFilesResponse FsSearchFiles(ctx, workspaceId, projectId).Path(path).Pattern(pattern).Execute()

Search for files



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
	path := "path_example" // string | Path
	pattern := "pattern_example" // string | Pattern

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkspaceToolboxAPI.FsSearchFiles(context.Background(), workspaceId, projectId).Path(path).Pattern(pattern).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceToolboxAPI.FsSearchFiles``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `FsSearchFiles`: SearchFilesResponse
	fmt.Fprintf(os.Stdout, "Response from `WorkspaceToolboxAPI.FsSearchFiles`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workspaceId** | **string** | Workspace ID or Name | 
**projectId** | **string** | Project ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiFsSearchFilesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **path** | **string** | Path | 
 **pattern** | **string** | Pattern | 

### Return type

[**SearchFilesResponse**](SearchFilesResponse.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## FsSetFilePermissions

> FsSetFilePermissions(ctx, workspaceId, projectId).Path(path).Owner(owner).Group(group).Mode(mode).Execute()

Set file owner/group/permissions



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
	path := "path_example" // string | Path
	owner := "owner_example" // string | Owner (optional)
	group := "group_example" // string | Group (optional)
	mode := "mode_example" // string | Mode (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkspaceToolboxAPI.FsSetFilePermissions(context.Background(), workspaceId, projectId).Path(path).Owner(owner).Group(group).Mode(mode).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceToolboxAPI.FsSetFilePermissions``: %v\n", err)
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

Other parameters are passed through a pointer to a apiFsSetFilePermissionsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **path** | **string** | Path | 
 **owner** | **string** | Owner | 
 **group** | **string** | Group | 
 **mode** | **string** | Mode | 

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


## FsUploadFile

> FsUploadFile(ctx, workspaceId, projectId).Path(path).File(file).Execute()

Upload file



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
	path := "path_example" // string | Path
	file := os.NewFile(1234, "some_file") // *os.File | File

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkspaceToolboxAPI.FsUploadFile(context.Background(), workspaceId, projectId).Path(path).File(file).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceToolboxAPI.FsUploadFile``: %v\n", err)
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

Other parameters are passed through a pointer to a apiFsUploadFileRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **path** | **string** | Path | 
 **file** | ***os.File** | File | 

### Return type

 (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: multipart/form-data
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetProjectDir

> ProjectDirResponse GetProjectDir(ctx, workspaceId, projectId).Execute()

Get project dir



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
	resp, r, err := apiClient.WorkspaceToolboxAPI.GetProjectDir(context.Background(), workspaceId, projectId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceToolboxAPI.GetProjectDir``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetProjectDir`: ProjectDirResponse
	fmt.Fprintf(os.Stdout, "Response from `WorkspaceToolboxAPI.GetProjectDir`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workspaceId** | **string** | Workspace ID or Name | 
**projectId** | **string** | Project ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetProjectDirRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**ProjectDirResponse**](ProjectDirResponse.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GitAddFiles

> GitAddFiles(ctx, workspaceId, projectId).Params(params).Execute()

Add files



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
	params := *openapiclient.NewGitAddRequest([]string{"Files_example"}, "Path_example") // GitAddRequest | GitAddRequest

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkspaceToolboxAPI.GitAddFiles(context.Background(), workspaceId, projectId).Params(params).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceToolboxAPI.GitAddFiles``: %v\n", err)
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

Other parameters are passed through a pointer to a apiGitAddFilesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **params** | [**GitAddRequest**](GitAddRequest.md) | GitAddRequest | 

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


## GitBranchList

> ListBranchResponse GitBranchList(ctx, workspaceId, projectId).Path(path).Execute()

Get branch list



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
	path := "path_example" // string | Path to git repository

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkspaceToolboxAPI.GitBranchList(context.Background(), workspaceId, projectId).Path(path).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceToolboxAPI.GitBranchList``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GitBranchList`: ListBranchResponse
	fmt.Fprintf(os.Stdout, "Response from `WorkspaceToolboxAPI.GitBranchList`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workspaceId** | **string** | Workspace ID or Name | 
**projectId** | **string** | Project ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGitBranchListRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **path** | **string** | Path to git repository | 

### Return type

[**ListBranchResponse**](ListBranchResponse.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GitCloneRepository

> GitCloneRepository(ctx, workspaceId, projectId).Params(params).Execute()

Clone git repository



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
	params := *openapiclient.NewGitCloneRequest("Path_example", "Url_example") // GitCloneRequest | GitCloneRequest

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkspaceToolboxAPI.GitCloneRepository(context.Background(), workspaceId, projectId).Params(params).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceToolboxAPI.GitCloneRepository``: %v\n", err)
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

Other parameters are passed through a pointer to a apiGitCloneRepositoryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **params** | [**GitCloneRequest**](GitCloneRequest.md) | GitCloneRequest | 

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


## GitCommitChanges

> GitCommitResponse GitCommitChanges(ctx, workspaceId, projectId).Params(params).Execute()

Commit changes



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
	params := *openapiclient.NewGitCommitRequest("Author_example", "Email_example", "Message_example", "Path_example") // GitCommitRequest | GitCommitRequest

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkspaceToolboxAPI.GitCommitChanges(context.Background(), workspaceId, projectId).Params(params).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceToolboxAPI.GitCommitChanges``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GitCommitChanges`: GitCommitResponse
	fmt.Fprintf(os.Stdout, "Response from `WorkspaceToolboxAPI.GitCommitChanges`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workspaceId** | **string** | Workspace ID or Name | 
**projectId** | **string** | Project ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGitCommitChangesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **params** | [**GitCommitRequest**](GitCommitRequest.md) | GitCommitRequest | 

### Return type

[**GitCommitResponse**](GitCommitResponse.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GitCommitHistory

> []GitCommitInfo GitCommitHistory(ctx, workspaceId, projectId).Path(path).Execute()

Get commit history



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
	path := "path_example" // string | Path to git repository

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkspaceToolboxAPI.GitCommitHistory(context.Background(), workspaceId, projectId).Path(path).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceToolboxAPI.GitCommitHistory``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GitCommitHistory`: []GitCommitInfo
	fmt.Fprintf(os.Stdout, "Response from `WorkspaceToolboxAPI.GitCommitHistory`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workspaceId** | **string** | Workspace ID or Name | 
**projectId** | **string** | Project ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGitCommitHistoryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **path** | **string** | Path to git repository | 

### Return type

[**[]GitCommitInfo**](GitCommitInfo.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GitCreateBranch

> GitCreateBranch(ctx, workspaceId, projectId).Params(params).Execute()

Create branch



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
	params := *openapiclient.NewGitBranchRequest("Name_example", "Path_example") // GitBranchRequest | GitBranchRequest

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkspaceToolboxAPI.GitCreateBranch(context.Background(), workspaceId, projectId).Params(params).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceToolboxAPI.GitCreateBranch``: %v\n", err)
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

Other parameters are passed through a pointer to a apiGitCreateBranchRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **params** | [**GitBranchRequest**](GitBranchRequest.md) | GitBranchRequest | 

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


## GitGitStatus

> GitStatus GitGitStatus(ctx, workspaceId, projectId).Path(path).Execute()

Get git status



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
	path := "path_example" // string | Path to git repository

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkspaceToolboxAPI.GitGitStatus(context.Background(), workspaceId, projectId).Path(path).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceToolboxAPI.GitGitStatus``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GitGitStatus`: GitStatus
	fmt.Fprintf(os.Stdout, "Response from `WorkspaceToolboxAPI.GitGitStatus`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workspaceId** | **string** | Workspace ID or Name | 
**projectId** | **string** | Project ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGitGitStatusRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **path** | **string** | Path to git repository | 

### Return type

[**GitStatus**](GitStatus.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GitPullChanges

> GitPullChanges(ctx, workspaceId, projectId).Params(params).Execute()

Pull changes



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
	params := *openapiclient.NewGitRepoRequest("Path_example") // GitRepoRequest | Git pull request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkspaceToolboxAPI.GitPullChanges(context.Background(), workspaceId, projectId).Params(params).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceToolboxAPI.GitPullChanges``: %v\n", err)
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

Other parameters are passed through a pointer to a apiGitPullChangesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **params** | [**GitRepoRequest**](GitRepoRequest.md) | Git pull request | 

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


## GitPushChanges

> GitPushChanges(ctx, workspaceId, projectId).Params(params).Execute()

Push changes



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
	params := *openapiclient.NewGitRepoRequest("Path_example") // GitRepoRequest | Git push request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkspaceToolboxAPI.GitPushChanges(context.Background(), workspaceId, projectId).Params(params).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceToolboxAPI.GitPushChanges``: %v\n", err)
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

Other parameters are passed through a pointer to a apiGitPushChangesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **params** | [**GitRepoRequest**](GitRepoRequest.md) | Git push request | 

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


## LspCompletions

> CompletionList LspCompletions(ctx, workspaceId, projectId).Params(params).Execute()

Get Lsp Completions



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
	params := *openapiclient.NewLspCompletionParams("LanguageId_example", "PathToProject_example", *openapiclient.NewPosition(int32(123), int32(123)), "Uri_example") // LspCompletionParams | LspCompletionParams

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkspaceToolboxAPI.LspCompletions(context.Background(), workspaceId, projectId).Params(params).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceToolboxAPI.LspCompletions``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `LspCompletions`: CompletionList
	fmt.Fprintf(os.Stdout, "Response from `WorkspaceToolboxAPI.LspCompletions`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workspaceId** | **string** | Workspace ID or Name | 
**projectId** | **string** | Project ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiLspCompletionsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **params** | [**LspCompletionParams**](LspCompletionParams.md) | LspCompletionParams | 

### Return type

[**CompletionList**](CompletionList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## LspDidClose

> LspDidClose(ctx, workspaceId, projectId).Params(params).Execute()

Call Lsp DidClose



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
	params := *openapiclient.NewLspDocumentRequest("LanguageId_example", "PathToProject_example", "Uri_example") // LspDocumentRequest | LspDocumentRequest

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkspaceToolboxAPI.LspDidClose(context.Background(), workspaceId, projectId).Params(params).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceToolboxAPI.LspDidClose``: %v\n", err)
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

Other parameters are passed through a pointer to a apiLspDidCloseRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **params** | [**LspDocumentRequest**](LspDocumentRequest.md) | LspDocumentRequest | 

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


## LspDidOpen

> LspDidOpen(ctx, workspaceId, projectId).Params(params).Execute()

Call Lsp DidOpen



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
	params := *openapiclient.NewLspDocumentRequest("LanguageId_example", "PathToProject_example", "Uri_example") // LspDocumentRequest | LspDocumentRequest

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkspaceToolboxAPI.LspDidOpen(context.Background(), workspaceId, projectId).Params(params).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceToolboxAPI.LspDidOpen``: %v\n", err)
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

Other parameters are passed through a pointer to a apiLspDidOpenRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **params** | [**LspDocumentRequest**](LspDocumentRequest.md) | LspDocumentRequest | 

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


## LspDocumentSymbols

> []LspSymbol LspDocumentSymbols(ctx, workspaceId, projectId).LanguageId(languageId).PathToProject(pathToProject).Uri(uri).Execute()

Call Lsp DocumentSymbols



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
	languageId := "languageId_example" // string | Language ID
	pathToProject := "pathToProject_example" // string | Path to project
	uri := "uri_example" // string | Document Uri

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkspaceToolboxAPI.LspDocumentSymbols(context.Background(), workspaceId, projectId).LanguageId(languageId).PathToProject(pathToProject).Uri(uri).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceToolboxAPI.LspDocumentSymbols``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `LspDocumentSymbols`: []LspSymbol
	fmt.Fprintf(os.Stdout, "Response from `WorkspaceToolboxAPI.LspDocumentSymbols`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workspaceId** | **string** | Workspace ID or Name | 
**projectId** | **string** | Project ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiLspDocumentSymbolsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **languageId** | **string** | Language ID | 
 **pathToProject** | **string** | Path to project | 
 **uri** | **string** | Document Uri | 

### Return type

[**[]LspSymbol**](LspSymbol.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## LspStart

> LspStart(ctx, workspaceId, projectId).Params(params).Execute()

Start Lsp server



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
	params := *openapiclient.NewLspServerRequest("LanguageId_example", "PathToProject_example") // LspServerRequest | LspServerRequest

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkspaceToolboxAPI.LspStart(context.Background(), workspaceId, projectId).Params(params).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceToolboxAPI.LspStart``: %v\n", err)
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

Other parameters are passed through a pointer to a apiLspStartRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **params** | [**LspServerRequest**](LspServerRequest.md) | LspServerRequest | 

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


## LspStop

> LspStop(ctx, workspaceId, projectId).Params(params).Execute()

Stop Lsp server



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
	params := *openapiclient.NewLspServerRequest("LanguageId_example", "PathToProject_example") // LspServerRequest | LspServerRequest

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkspaceToolboxAPI.LspStop(context.Background(), workspaceId, projectId).Params(params).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceToolboxAPI.LspStop``: %v\n", err)
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

Other parameters are passed through a pointer to a apiLspStopRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **params** | [**LspServerRequest**](LspServerRequest.md) | LspServerRequest | 

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


## LspWorkspaceSymbols

> []LspSymbol LspWorkspaceSymbols(ctx, workspaceId, projectId).LanguageId(languageId).PathToProject(pathToProject).Query(query).Execute()

Call Lsp WorkspaceSymbols



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
	languageId := "languageId_example" // string | Language ID
	pathToProject := "pathToProject_example" // string | Path to project
	query := "query_example" // string | Symbol Query

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkspaceToolboxAPI.LspWorkspaceSymbols(context.Background(), workspaceId, projectId).LanguageId(languageId).PathToProject(pathToProject).Query(query).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceToolboxAPI.LspWorkspaceSymbols``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `LspWorkspaceSymbols`: []LspSymbol
	fmt.Fprintf(os.Stdout, "Response from `WorkspaceToolboxAPI.LspWorkspaceSymbols`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workspaceId** | **string** | Workspace ID or Name | 
**projectId** | **string** | Project ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiLspWorkspaceSymbolsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **languageId** | **string** | Language ID | 
 **pathToProject** | **string** | Path to project | 
 **query** | **string** | Symbol Query | 

### Return type

[**[]LspSymbol**](LspSymbol.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ProcessExecuteCommand

> ExecuteResponse ProcessExecuteCommand(ctx, workspaceId, projectId).Params(params).Execute()

Execute command



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
	params := *openapiclient.NewExecuteRequest("Command_example") // ExecuteRequest | Execute command request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkspaceToolboxAPI.ProcessExecuteCommand(context.Background(), workspaceId, projectId).Params(params).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceToolboxAPI.ProcessExecuteCommand``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ProcessExecuteCommand`: ExecuteResponse
	fmt.Fprintf(os.Stdout, "Response from `WorkspaceToolboxAPI.ProcessExecuteCommand`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workspaceId** | **string** | Workspace ID or Name | 
**projectId** | **string** | Project ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiProcessExecuteCommandRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **params** | [**ExecuteRequest**](ExecuteRequest.md) | Execute command request | 

### Return type

[**ExecuteResponse**](ExecuteResponse.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

