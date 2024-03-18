# \GitProviderAPI

All URIs are relative to *http://localhost:3000*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetGitUserData**](GitProviderAPI.md#GetGitUserData) | **Get** /gitprovider/{gitProviderId}/user-data | Get Git context



## GetGitUserData

> GitUserData GetGitUserData(ctx, gitProviderId).Execute()

Get Git context



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID/serverapiclient"
)

func main() {
	gitProviderId := "gitProviderId_example" // string | Git Provider Id

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GitProviderAPI.GetGitUserData(context.Background(), gitProviderId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GitProviderAPI.GetGitUserData``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetGitUserData`: GitUserData
	fmt.Fprintf(os.Stdout, "Response from `GitProviderAPI.GetGitUserData`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**gitProviderId** | **string** | Git Provider Id | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetGitUserDataRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**GitUserData**](GitUserData.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

