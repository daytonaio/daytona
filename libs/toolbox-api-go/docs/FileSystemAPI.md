# \FileSystemAPI

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateFolder**](FileSystemAPI.md#CreateFolder) | **Post** /files/folder | Create a folder
[**DeleteFile**](FileSystemAPI.md#DeleteFile) | **Delete** /files | Delete a file or directory
[**DownloadFile**](FileSystemAPI.md#DownloadFile) | **Get** /files/download | Download a file
[**DownloadFiles**](FileSystemAPI.md#DownloadFiles) | **Post** /files/bulk-download | Download multiple files
[**FindInFiles**](FileSystemAPI.md#FindInFiles) | **Get** /files/find | Find text in files
[**GetFileInfo**](FileSystemAPI.md#GetFileInfo) | **Get** /files/info | Get file information
[**ListFiles**](FileSystemAPI.md#ListFiles) | **Get** /files | List files and directories
[**MoveFile**](FileSystemAPI.md#MoveFile) | **Post** /files/move | Move or rename file/directory
[**ReplaceInFiles**](FileSystemAPI.md#ReplaceInFiles) | **Post** /files/replace | Replace text in files
[**SearchFiles**](FileSystemAPI.md#SearchFiles) | **Get** /files/search | Search files by pattern
[**SetFilePermissions**](FileSystemAPI.md#SetFilePermissions) | **Post** /files/permissions | Set file permissions
[**UploadFile**](FileSystemAPI.md#UploadFile) | **Post** /files/upload | Upload a file
[**UploadFiles**](FileSystemAPI.md#UploadFiles) | **Post** /files/bulk-upload | Upload multiple files



## CreateFolder

> CreateFolder(ctx).Path(path).Mode(mode).Execute()

Create a folder



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	path := "path_example" // string | Folder path to create
	mode := "mode_example" // string | Octal permission mode (default: 0755)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.FileSystemAPI.CreateFolder(context.Background()).Path(path).Mode(mode).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FileSystemAPI.CreateFolder``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateFolderRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **path** | **string** | Folder path to create | 
 **mode** | **string** | Octal permission mode (default: 0755) | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteFile

> DeleteFile(ctx).Path(path).Recursive(recursive).Execute()

Delete a file or directory



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	path := "path_example" // string | File or directory path to delete
	recursive := true // bool | Enable recursive deletion for directories (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.FileSystemAPI.DeleteFile(context.Background()).Path(path).Recursive(recursive).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FileSystemAPI.DeleteFile``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeleteFileRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **path** | **string** | File or directory path to delete | 
 **recursive** | **bool** | Enable recursive deletion for directories | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DownloadFile

> *os.File DownloadFile(ctx).Path(path).Execute()

Download a file



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	path := "path_example" // string | File path to download

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.FileSystemAPI.DownloadFile(context.Background()).Path(path).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FileSystemAPI.DownloadFile``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DownloadFile`: *os.File
	fmt.Fprintf(os.Stdout, "Response from `FileSystemAPI.DownloadFile`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDownloadFileRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **path** | **string** | File path to download | 

### Return type

[***os.File**](*os.File.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/octet-stream

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DownloadFiles

> map[string]map[string]interface{} DownloadFiles(ctx).DownloadFiles(downloadFiles).Execute()

Download multiple files



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	downloadFiles := *openapiclient.NewFilesDownloadRequest([]string{"Paths_example"}) // FilesDownloadRequest | Paths of files to download

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.FileSystemAPI.DownloadFiles(context.Background()).DownloadFiles(downloadFiles).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FileSystemAPI.DownloadFiles``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DownloadFiles`: map[string]map[string]interface{}
	fmt.Fprintf(os.Stdout, "Response from `FileSystemAPI.DownloadFiles`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDownloadFilesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **downloadFiles** | [**FilesDownloadRequest**](FilesDownloadRequest.md) | Paths of files to download | 

### Return type

**map[string]map[string]interface{}**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: multipart/form-data

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## FindInFiles

> []Match FindInFiles(ctx).Path(path).Pattern(pattern).Execute()

Find text in files



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	path := "path_example" // string | Directory path to search in
	pattern := "pattern_example" // string | Text pattern to search for

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.FileSystemAPI.FindInFiles(context.Background()).Path(path).Pattern(pattern).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FileSystemAPI.FindInFiles``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `FindInFiles`: []Match
	fmt.Fprintf(os.Stdout, "Response from `FileSystemAPI.FindInFiles`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiFindInFilesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **path** | **string** | Directory path to search in | 
 **pattern** | **string** | Text pattern to search for | 

### Return type

[**[]Match**](Match.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetFileInfo

> FileInfo GetFileInfo(ctx).Path(path).Execute()

Get file information



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	path := "path_example" // string | File or directory path

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.FileSystemAPI.GetFileInfo(context.Background()).Path(path).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FileSystemAPI.GetFileInfo``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetFileInfo`: FileInfo
	fmt.Fprintf(os.Stdout, "Response from `FileSystemAPI.GetFileInfo`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetFileInfoRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **path** | **string** | File or directory path | 

### Return type

[**FileInfo**](FileInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListFiles

> []FileInfo ListFiles(ctx).Path(path).Execute()

List files and directories



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	path := "path_example" // string | Directory path to list (defaults to working directory) (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.FileSystemAPI.ListFiles(context.Background()).Path(path).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FileSystemAPI.ListFiles``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListFiles`: []FileInfo
	fmt.Fprintf(os.Stdout, "Response from `FileSystemAPI.ListFiles`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListFilesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **path** | **string** | Directory path to list (defaults to working directory) | 

### Return type

[**[]FileInfo**](FileInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## MoveFile

> MoveFile(ctx).Source(source).Destination(destination).Execute()

Move or rename file/directory



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	source := "source_example" // string | Source file or directory path
	destination := "destination_example" // string | Destination file or directory path

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.FileSystemAPI.MoveFile(context.Background()).Source(source).Destination(destination).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FileSystemAPI.MoveFile``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiMoveFileRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **source** | **string** | Source file or directory path | 
 **destination** | **string** | Destination file or directory path | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ReplaceInFiles

> []ReplaceResult ReplaceInFiles(ctx).Request(request).Execute()

Replace text in files



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	request := *openapiclient.NewReplaceRequest([]string{"Files_example"}, "NewValue_example", "Pattern_example") // ReplaceRequest | Replace request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.FileSystemAPI.ReplaceInFiles(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FileSystemAPI.ReplaceInFiles``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ReplaceInFiles`: []ReplaceResult
	fmt.Fprintf(os.Stdout, "Response from `FileSystemAPI.ReplaceInFiles`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiReplaceInFilesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**ReplaceRequest**](ReplaceRequest.md) | Replace request | 

### Return type

[**[]ReplaceResult**](ReplaceResult.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SearchFiles

> SearchFilesResponse SearchFiles(ctx).Path(path).Pattern(pattern).Execute()

Search files by pattern



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	path := "path_example" // string | Directory path to search in
	pattern := "pattern_example" // string | File pattern to match (e.g., *.txt, *.go)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.FileSystemAPI.SearchFiles(context.Background()).Path(path).Pattern(pattern).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FileSystemAPI.SearchFiles``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SearchFiles`: SearchFilesResponse
	fmt.Fprintf(os.Stdout, "Response from `FileSystemAPI.SearchFiles`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSearchFilesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **path** | **string** | Directory path to search in | 
 **pattern** | **string** | File pattern to match (e.g., *.txt, *.go) | 

### Return type

[**SearchFilesResponse**](SearchFilesResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SetFilePermissions

> SetFilePermissions(ctx).Path(path).Owner(owner).Group(group).Mode(mode).Execute()

Set file permissions



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	path := "path_example" // string | File or directory path
	owner := "owner_example" // string | Owner (username or UID) (optional)
	group := "group_example" // string | Group (group name or GID) (optional)
	mode := "mode_example" // string | File mode in octal format (e.g., 0755) (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.FileSystemAPI.SetFilePermissions(context.Background()).Path(path).Owner(owner).Group(group).Mode(mode).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FileSystemAPI.SetFilePermissions``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSetFilePermissionsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **path** | **string** | File or directory path | 
 **owner** | **string** | Owner (username or UID) | 
 **group** | **string** | Group (group name or GID) | 
 **mode** | **string** | File mode in octal format (e.g., 0755) | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UploadFile

> map[string]map[string]interface{} UploadFile(ctx).Path(path).File(file).Execute()

Upload a file



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	path := "path_example" // string | Destination path for the uploaded file
	file := os.NewFile(1234, "some_file") // *os.File | File to upload

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.FileSystemAPI.UploadFile(context.Background()).Path(path).File(file).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FileSystemAPI.UploadFile``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UploadFile`: map[string]map[string]interface{}
	fmt.Fprintf(os.Stdout, "Response from `FileSystemAPI.UploadFile`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUploadFileRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **path** | **string** | Destination path for the uploaded file | 
 **file** | ***os.File** | File to upload | 

### Return type

**map[string]map[string]interface{}**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: multipart/form-data
- **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UploadFiles

> UploadFiles(ctx).Execute()

Upload multiple files



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.FileSystemAPI.UploadFiles(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FileSystemAPI.UploadFiles``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiUploadFilesRequest struct via the builder pattern


### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

