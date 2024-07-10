# \PrebuildAPI

All URIs are relative to *http://localhost:3986*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeletePrebuild**](PrebuildAPI.md#DeletePrebuild) | **Delete** /prebuild | Delete prebuild
[**ListPrebuilds**](PrebuildAPI.md#ListPrebuilds) | **Get** /prebuild/list | List prebuilds
[**SetPrebuild**](PrebuildAPI.md#SetPrebuild) | **Get** /prebuild | Upsert prebuild



## DeletePrebuild

> DeletePrebuild(ctx).Prebuild(prebuild).Execute()

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
	prebuild := *openapiclient.NewPrebuildConfig() // PrebuildConfig | Prebuild

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PrebuildAPI.DeletePrebuild(context.Background()).Prebuild(prebuild).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PrebuildAPI.DeletePrebuild``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeletePrebuildRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **prebuild** | [**PrebuildConfig**](PrebuildConfig.md) | Prebuild | 

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


## ListPrebuilds

> ListPrebuilds(ctx).Prebuild(prebuild).Execute()

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
	prebuild := *openapiclient.NewPrebuildPrebuildFilter() // PrebuildPrebuildFilter | Prebuild Filter

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PrebuildAPI.ListPrebuilds(context.Background()).Prebuild(prebuild).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PrebuildAPI.ListPrebuilds``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListPrebuildsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **prebuild** | [**PrebuildPrebuildFilter**](PrebuildPrebuildFilter.md) | Prebuild Filter | 

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


## SetPrebuild

> SetPrebuild(ctx).Prebuild(prebuild).Execute()

Upsert prebuild



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
	prebuild := *openapiclient.NewPrebuildConfig() // PrebuildConfig | Prebuild

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PrebuildAPI.SetPrebuild(context.Background()).Prebuild(prebuild).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PrebuildAPI.SetPrebuild``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSetPrebuildRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **prebuild** | [**PrebuildConfig**](PrebuildConfig.md) | Prebuild | 

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

