# \GitAPI

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AddFiles**](GitAPI.md#AddFiles) | **Post** /git/add | Add files to Git staging
[**CheckoutBranch**](GitAPI.md#CheckoutBranch) | **Post** /git/checkout | Checkout branch or commit
[**CloneRepository**](GitAPI.md#CloneRepository) | **Post** /git/clone | Clone a Git repository
[**CommitChanges**](GitAPI.md#CommitChanges) | **Post** /git/commit | Commit changes
[**CreateBranch**](GitAPI.md#CreateBranch) | **Post** /git/branches | Create a new branch
[**DeleteBranch**](GitAPI.md#DeleteBranch) | **Delete** /git/branches | Delete a branch
[**GetCommitHistory**](GitAPI.md#GetCommitHistory) | **Get** /git/history | Get commit history
[**GetStatus**](GitAPI.md#GetStatus) | **Get** /git/status | Get Git status
[**ListBranches**](GitAPI.md#ListBranches) | **Get** /git/branches | List branches
[**PullChanges**](GitAPI.md#PullChanges) | **Post** /git/pull | Pull changes from remote
[**PushChanges**](GitAPI.md#PushChanges) | **Post** /git/push | Push changes to remote



## AddFiles

> AddFiles(ctx).Request(request).Execute()

Add files to Git staging



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-client-go"
)

func main() {
	request := *openapiclient.NewGitAddRequest([]string{"Files_example"}, "Path_example") // GitAddRequest | Add files request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.GitAPI.AddFiles(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GitAPI.AddFiles``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAddFilesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**GitAddRequest**](GitAddRequest.md) | Add files request | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CheckoutBranch

> CheckoutBranch(ctx).Request(request).Execute()

Checkout branch or commit



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-client-go"
)

func main() {
	request := *openapiclient.NewGitCheckoutRequest("Branch_example", "Path_example") // GitCheckoutRequest | Checkout request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.GitAPI.CheckoutBranch(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GitAPI.CheckoutBranch``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCheckoutBranchRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**GitCheckoutRequest**](GitCheckoutRequest.md) | Checkout request | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CloneRepository

> CloneRepository(ctx).Request(request).Execute()

Clone a Git repository



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-client-go"
)

func main() {
	request := *openapiclient.NewGitCloneRequest("Path_example", "Url_example") // GitCloneRequest | Clone repository request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.GitAPI.CloneRepository(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GitAPI.CloneRepository``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCloneRepositoryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**GitCloneRequest**](GitCloneRequest.md) | Clone repository request | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CommitChanges

> GitCommitResponse CommitChanges(ctx).Request(request).Execute()

Commit changes



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-client-go"
)

func main() {
	request := *openapiclient.NewGitCommitRequest("Author_example", "Email_example", "Message_example", "Path_example") // GitCommitRequest | Commit request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GitAPI.CommitChanges(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GitAPI.CommitChanges``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CommitChanges`: GitCommitResponse
	fmt.Fprintf(os.Stdout, "Response from `GitAPI.CommitChanges`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCommitChangesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**GitCommitRequest**](GitCommitRequest.md) | Commit request | 

### Return type

[**GitCommitResponse**](GitCommitResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateBranch

> CreateBranch(ctx).Request(request).Execute()

Create a new branch



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-client-go"
)

func main() {
	request := *openapiclient.NewGitBranchRequest("Name_example", "Path_example") // GitBranchRequest | Create branch request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.GitAPI.CreateBranch(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GitAPI.CreateBranch``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateBranchRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**GitBranchRequest**](GitBranchRequest.md) | Create branch request | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteBranch

> DeleteBranch(ctx).Request(request).Execute()

Delete a branch



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-client-go"
)

func main() {
	request := *openapiclient.NewGitGitDeleteBranchRequest("Name_example", "Path_example") // GitGitDeleteBranchRequest | Delete branch request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.GitAPI.DeleteBranch(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GitAPI.DeleteBranch``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeleteBranchRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**GitGitDeleteBranchRequest**](GitGitDeleteBranchRequest.md) | Delete branch request | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetCommitHistory

> []GitCommitInfo GetCommitHistory(ctx).Path(path).Execute()

Get commit history



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-client-go"
)

func main() {
	path := "path_example" // string | Repository path

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GitAPI.GetCommitHistory(context.Background()).Path(path).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GitAPI.GetCommitHistory``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetCommitHistory`: []GitCommitInfo
	fmt.Fprintf(os.Stdout, "Response from `GitAPI.GetCommitHistory`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetCommitHistoryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **path** | **string** | Repository path | 

### Return type

[**[]GitCommitInfo**](GitCommitInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetStatus

> GitStatus GetStatus(ctx).Path(path).Execute()

Get Git status



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-client-go"
)

func main() {
	path := "path_example" // string | Repository path

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GitAPI.GetStatus(context.Background()).Path(path).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GitAPI.GetStatus``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetStatus`: GitStatus
	fmt.Fprintf(os.Stdout, "Response from `GitAPI.GetStatus`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetStatusRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **path** | **string** | Repository path | 

### Return type

[**GitStatus**](GitStatus.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListBranches

> ListBranchResponse ListBranches(ctx).Path(path).Execute()

List branches



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-client-go"
)

func main() {
	path := "path_example" // string | Repository path

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GitAPI.ListBranches(context.Background()).Path(path).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GitAPI.ListBranches``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListBranches`: ListBranchResponse
	fmt.Fprintf(os.Stdout, "Response from `GitAPI.ListBranches`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListBranchesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **path** | **string** | Repository path | 

### Return type

[**ListBranchResponse**](ListBranchResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## PullChanges

> PullChanges(ctx).Request(request).Execute()

Pull changes from remote



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-client-go"
)

func main() {
	request := *openapiclient.NewGitRepoRequest("Path_example") // GitRepoRequest | Pull request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.GitAPI.PullChanges(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GitAPI.PullChanges``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiPullChangesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**GitRepoRequest**](GitRepoRequest.md) | Pull request | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## PushChanges

> PushChanges(ctx).Request(request).Execute()

Push changes to remote



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-client-go"
)

func main() {
	request := *openapiclient.NewGitRepoRequest("Path_example") // GitRepoRequest | Push request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.GitAPI.PushChanges(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GitAPI.PushChanges``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiPushChangesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**GitRepoRequest**](GitRepoRequest.md) | Push request | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

