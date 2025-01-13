# \GitProviderAPI

All URIs are relative to *http://localhost:3986*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeleteGitProvider**](GitProviderAPI.md#DeleteGitProvider) | **Delete** /gitprovider/{gitProviderId} | Delete Git provider
[**FindGitProvider**](GitProviderAPI.md#FindGitProvider) | **Get** /gitprovider/{gitProviderId} | Find Git provider
[**FindGitProviderIdForUrl**](GitProviderAPI.md#FindGitProviderIdForUrl) | **Get** /gitprovider/id-for-url/{url} | Find Git provider ID
[**GetGitContext**](GitProviderAPI.md#GetGitContext) | **Post** /gitprovider/context | Get Git context
[**GetGitUser**](GitProviderAPI.md#GetGitUser) | **Get** /gitprovider/{gitProviderId}/user | Get Git context
[**GetNamespaces**](GitProviderAPI.md#GetNamespaces) | **Get** /gitprovider/{gitProviderId}/namespaces | Get Git namespaces
[**GetRepoBranches**](GitProviderAPI.md#GetRepoBranches) | **Get** /gitprovider/{gitProviderId}/{namespaceId}/{repositoryId}/branches | Get Git repository branches
[**GetRepoPRs**](GitProviderAPI.md#GetRepoPRs) | **Get** /gitprovider/{gitProviderId}/{namespaceId}/{repositoryId}/pull-requests | Get Git repository PRs
[**GetRepositories**](GitProviderAPI.md#GetRepositories) | **Get** /gitprovider/{gitProviderId}/{namespaceId}/repositories | Get Git repositories
[**GetUrlFromRepository**](GitProviderAPI.md#GetUrlFromRepository) | **Post** /gitprovider/context/url | Get URL from Git repository
[**ListGitProviders**](GitProviderAPI.md#ListGitProviders) | **Get** /gitprovider | List Git providers
[**ListGitProvidersForUrl**](GitProviderAPI.md#ListGitProvidersForUrl) | **Get** /gitprovider/for-url/{url} | List Git providers for url
[**SaveGitProvider**](GitProviderAPI.md#SaveGitProvider) | **Put** /gitprovider | Save Git provider



## DeleteGitProvider

> DeleteGitProvider(ctx, gitProviderId).Execute()

Delete Git provider



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
	gitProviderId := "gitProviderId_example" // string | Git provider

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.GitProviderAPI.DeleteGitProvider(context.Background(), gitProviderId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GitProviderAPI.DeleteGitProvider``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**gitProviderId** | **string** | Git provider | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteGitProviderRequest struct via the builder pattern


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


## FindGitProvider

> GitProvider FindGitProvider(ctx, gitProviderId).Execute()

Find Git provider



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
	gitProviderId := "gitProviderId_example" // string | ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GitProviderAPI.FindGitProvider(context.Background(), gitProviderId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GitProviderAPI.FindGitProvider``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `FindGitProvider`: GitProvider
	fmt.Fprintf(os.Stdout, "Response from `GitProviderAPI.FindGitProvider`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**gitProviderId** | **string** | ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiFindGitProviderRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**GitProvider**](GitProvider.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: text/plain

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## FindGitProviderIdForUrl

> string FindGitProviderIdForUrl(ctx, url).Execute()

Find Git provider ID



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
	url := "url_example" // string | Url

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GitProviderAPI.FindGitProviderIdForUrl(context.Background(), url).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GitProviderAPI.FindGitProviderIdForUrl``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `FindGitProviderIdForUrl`: string
	fmt.Fprintf(os.Stdout, "Response from `GitProviderAPI.FindGitProviderIdForUrl`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**url** | **string** | Url | 

### Other Parameters

Other parameters are passed through a pointer to a apiFindGitProviderIdForUrlRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

**string**

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: text/plain

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetGitContext

> GitRepository GetGitContext(ctx).Repository(repository).Execute()

Get Git context



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
	repository := *openapiclient.NewGetRepositoryContext("Url_example") // GetRepositoryContext | Get repository context

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GitProviderAPI.GetGitContext(context.Background()).Repository(repository).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GitProviderAPI.GetGitContext``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetGitContext`: GitRepository
	fmt.Fprintf(os.Stdout, "Response from `GitProviderAPI.GetGitContext`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetGitContextRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **repository** | [**GetRepositoryContext**](GetRepositoryContext.md) | Get repository context | 

### Return type

[**GitRepository**](GitRepository.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetGitUser

> GitUser GetGitUser(ctx, gitProviderId).Execute()

Get Git context



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
	gitProviderId := "gitProviderId_example" // string | Git Provider Id

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GitProviderAPI.GetGitUser(context.Background(), gitProviderId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GitProviderAPI.GetGitUser``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetGitUser`: GitUser
	fmt.Fprintf(os.Stdout, "Response from `GitProviderAPI.GetGitUser`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**gitProviderId** | **string** | Git Provider Id | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetGitUserRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**GitUser**](GitUser.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetNamespaces

> []GitNamespace GetNamespaces(ctx, gitProviderId).Page(page).PerPage(perPage).Execute()

Get Git namespaces



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
	gitProviderId := "gitProviderId_example" // string | Git provider
	page := int32(56) // int32 | Page number (optional)
	perPage := int32(56) // int32 | Number of items per page (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GitProviderAPI.GetNamespaces(context.Background(), gitProviderId).Page(page).PerPage(perPage).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GitProviderAPI.GetNamespaces``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNamespaces`: []GitNamespace
	fmt.Fprintf(os.Stdout, "Response from `GitProviderAPI.GetNamespaces`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**gitProviderId** | **string** | Git provider | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetNamespacesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **page** | **int32** | Page number | 
 **perPage** | **int32** | Number of items per page | 

### Return type

[**[]GitNamespace**](GitNamespace.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetRepoBranches

> []GitBranch GetRepoBranches(ctx, gitProviderId, namespaceId, repositoryId).Page(page).PerPage(perPage).Execute()

Get Git repository branches



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
	gitProviderId := "gitProviderId_example" // string | Git provider
	namespaceId := "namespaceId_example" // string | Namespace
	repositoryId := "repositoryId_example" // string | Repository
	page := int32(56) // int32 | Page number (optional)
	perPage := int32(56) // int32 | Number of items per page (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GitProviderAPI.GetRepoBranches(context.Background(), gitProviderId, namespaceId, repositoryId).Page(page).PerPage(perPage).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GitProviderAPI.GetRepoBranches``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetRepoBranches`: []GitBranch
	fmt.Fprintf(os.Stdout, "Response from `GitProviderAPI.GetRepoBranches`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**gitProviderId** | **string** | Git provider | 
**namespaceId** | **string** | Namespace | 
**repositoryId** | **string** | Repository | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetRepoBranchesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **page** | **int32** | Page number | 
 **perPage** | **int32** | Number of items per page | 

### Return type

[**[]GitBranch**](GitBranch.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetRepoPRs

> []GitPullRequest GetRepoPRs(ctx, gitProviderId, namespaceId, repositoryId).Page(page).PerPage(perPage).Execute()

Get Git repository PRs



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
	gitProviderId := "gitProviderId_example" // string | Git provider
	namespaceId := "namespaceId_example" // string | Namespace
	repositoryId := "repositoryId_example" // string | Repository
	page := int32(56) // int32 | Page number (optional)
	perPage := int32(56) // int32 | Number of items per page (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GitProviderAPI.GetRepoPRs(context.Background(), gitProviderId, namespaceId, repositoryId).Page(page).PerPage(perPage).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GitProviderAPI.GetRepoPRs``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetRepoPRs`: []GitPullRequest
	fmt.Fprintf(os.Stdout, "Response from `GitProviderAPI.GetRepoPRs`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**gitProviderId** | **string** | Git provider | 
**namespaceId** | **string** | Namespace | 
**repositoryId** | **string** | Repository | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetRepoPRsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **page** | **int32** | Page number | 
 **perPage** | **int32** | Number of items per page | 

### Return type

[**[]GitPullRequest**](GitPullRequest.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetRepositories

> []GitRepository GetRepositories(ctx, gitProviderId, namespaceId).Page(page).PerPage(perPage).Execute()

Get Git repositories



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
	gitProviderId := "gitProviderId_example" // string | Git provider
	namespaceId := "namespaceId_example" // string | Namespace
	page := int32(56) // int32 | Page number (optional)
	perPage := int32(56) // int32 | Number of items per page (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GitProviderAPI.GetRepositories(context.Background(), gitProviderId, namespaceId).Page(page).PerPage(perPage).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GitProviderAPI.GetRepositories``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetRepositories`: []GitRepository
	fmt.Fprintf(os.Stdout, "Response from `GitProviderAPI.GetRepositories`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**gitProviderId** | **string** | Git provider | 
**namespaceId** | **string** | Namespace | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetRepositoriesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **page** | **int32** | Page number | 
 **perPage** | **int32** | Number of items per page | 

### Return type

[**[]GitRepository**](GitRepository.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetUrlFromRepository

> RepositoryUrl GetUrlFromRepository(ctx).Repository(repository).Execute()

Get URL from Git repository



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
	repository := *openapiclient.NewGitRepository("Branch_example", "Id_example", "Name_example", "Owner_example", "Sha_example", "Source_example", "Url_example") // GitRepository | Git repository

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GitProviderAPI.GetUrlFromRepository(context.Background()).Repository(repository).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GitProviderAPI.GetUrlFromRepository``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetUrlFromRepository`: RepositoryUrl
	fmt.Fprintf(os.Stdout, "Response from `GitProviderAPI.GetUrlFromRepository`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetUrlFromRepositoryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **repository** | [**GitRepository**](GitRepository.md) | Git repository | 

### Return type

[**RepositoryUrl**](RepositoryUrl.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListGitProviders

> []GitProvider ListGitProviders(ctx).Execute()

List Git providers



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
	resp, r, err := apiClient.GitProviderAPI.ListGitProviders(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GitProviderAPI.ListGitProviders``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListGitProviders`: []GitProvider
	fmt.Fprintf(os.Stdout, "Response from `GitProviderAPI.ListGitProviders`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiListGitProvidersRequest struct via the builder pattern


### Return type

[**[]GitProvider**](GitProvider.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListGitProvidersForUrl

> []GitProvider ListGitProvidersForUrl(ctx, url).Execute()

List Git providers for url



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
	url := "url_example" // string | Url

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GitProviderAPI.ListGitProvidersForUrl(context.Background(), url).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GitProviderAPI.ListGitProvidersForUrl``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListGitProvidersForUrl`: []GitProvider
	fmt.Fprintf(os.Stdout, "Response from `GitProviderAPI.ListGitProvidersForUrl`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**url** | **string** | Url | 

### Other Parameters

Other parameters are passed through a pointer to a apiListGitProvidersForUrlRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]GitProvider**](GitProvider.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SaveGitProvider

> SaveGitProvider(ctx).GitProviderConfig(gitProviderConfig).Execute()

Save Git provider



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
	gitProviderConfig := *openapiclient.NewSetGitProviderConfig("ProviderId_example", "Token_example") // SetGitProviderConfig | Git provider

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.GitProviderAPI.SaveGitProvider(context.Background()).GitProviderConfig(gitProviderConfig).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GitProviderAPI.SaveGitProvider``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSaveGitProviderRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **gitProviderConfig** | [**SetGitProviderConfig**](SetGitProviderConfig.md) | Git provider | 

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

