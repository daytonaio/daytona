# \TargetAPI

All URIs are relative to *http://localhost:3986*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateTarget**](TargetAPI.md#CreateTarget) | **Post** /target | Create a target
[**DeleteTarget**](TargetAPI.md#DeleteTarget) | **Delete** /target/{targetId} | Delete target
[**FindTarget**](TargetAPI.md#FindTarget) | **Get** /target/{targetId} | Find target
[**HandleSuccessfulCreation**](TargetAPI.md#HandleSuccessfulCreation) | **Post** /target/{targetId}/handle-successful-creation | Handles successful creation of the target
[**ListTargets**](TargetAPI.md#ListTargets) | **Get** /target | List targets
[**RestartTarget**](TargetAPI.md#RestartTarget) | **Post** /target/{targetId}/restart | Restart target
[**SetDefaultTarget**](TargetAPI.md#SetDefaultTarget) | **Patch** /target/{targetId}/set-default | Set target to be used by default
[**StartTarget**](TargetAPI.md#StartTarget) | **Post** /target/{targetId}/start | Start target
[**StopTarget**](TargetAPI.md#StopTarget) | **Post** /target/{targetId}/stop | Stop target
[**UpdateTargetMetadata**](TargetAPI.md#UpdateTargetMetadata) | **Post** /target/{targetId}/metadata | Update target metadata
[**UpdateTargetProviderMetadata**](TargetAPI.md#UpdateTargetProviderMetadata) | **Post** /target/{targetId}/provider-metadata | Update target provider metadata



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
	target := *openapiclient.NewCreateTargetDTO("Id_example", "Name_example", "TargetConfigId_example") // CreateTargetDTO | Create target

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


## DeleteTarget

> DeleteTarget(ctx, targetId).Force(force).Execute()

Delete target



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
	r, err := apiClient.TargetAPI.DeleteTarget(context.Background(), targetId).Force(force).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TargetAPI.DeleteTarget``: %v\n", err)
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

Other parameters are passed through a pointer to a apiDeleteTargetRequest struct via the builder pattern


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


## FindTarget

> TargetDTO FindTarget(ctx, targetId).ShowOptions(showOptions).Execute()

Find target



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
	showOptions := true // bool | Show target config options (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TargetAPI.FindTarget(context.Background(), targetId).ShowOptions(showOptions).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TargetAPI.FindTarget``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `FindTarget`: TargetDTO
	fmt.Fprintf(os.Stdout, "Response from `TargetAPI.FindTarget`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**targetId** | **string** | Target ID or Name | 

### Other Parameters

Other parameters are passed through a pointer to a apiFindTargetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **showOptions** | **bool** | Show target config options | 

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


## HandleSuccessfulCreation

> HandleSuccessfulCreation(ctx, targetId).Execute()

Handles successful creation of the target



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
	targetId := "targetId_example" // string | Target ID or name

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.TargetAPI.HandleSuccessfulCreation(context.Background(), targetId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TargetAPI.HandleSuccessfulCreation``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**targetId** | **string** | Target ID or name | 

### Other Parameters

Other parameters are passed through a pointer to a apiHandleSuccessfulCreationRequest struct via the builder pattern


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


## ListTargets

> []TargetDTO ListTargets(ctx).ShowOptions(showOptions).Execute()

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
	showOptions := true // bool | Show target config options (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TargetAPI.ListTargets(context.Background()).ShowOptions(showOptions).Execute()
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
 **showOptions** | **bool** | Show target config options | 

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


## RestartTarget

> RestartTarget(ctx, targetId).Execute()

Restart target



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
	r, err := apiClient.TargetAPI.RestartTarget(context.Background(), targetId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TargetAPI.RestartTarget``: %v\n", err)
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

Other parameters are passed through a pointer to a apiRestartTargetRequest struct via the builder pattern


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


## SetDefaultTarget

> SetDefaultTarget(ctx, targetId).Execute()

Set target to be used by default



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
	targetId := "targetId_example" // string | Target ID or name

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.TargetAPI.SetDefaultTarget(context.Background(), targetId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TargetAPI.SetDefaultTarget``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**targetId** | **string** | Target ID or name | 

### Other Parameters

Other parameters are passed through a pointer to a apiSetDefaultTargetRequest struct via the builder pattern


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


## UpdateTargetMetadata

> UpdateTargetMetadata(ctx, targetId).TargetMetadata(targetMetadata).Execute()

Update target metadata



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
	targetMetadata := *openapiclient.NewUpdateTargetMetadataDTO(int32(123)) // UpdateTargetMetadataDTO | Target Metadata

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.TargetAPI.UpdateTargetMetadata(context.Background(), targetId).TargetMetadata(targetMetadata).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TargetAPI.UpdateTargetMetadata``: %v\n", err)
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

Other parameters are passed through a pointer to a apiUpdateTargetMetadataRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **targetMetadata** | [**UpdateTargetMetadataDTO**](UpdateTargetMetadataDTO.md) | Target Metadata | 

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


## UpdateTargetProviderMetadata

> UpdateTargetProviderMetadata(ctx, targetId).Metadata(metadata).Execute()

Update target provider metadata



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
	metadata := *openapiclient.NewUpdateTargetProviderMetadataDTO("Metadata_example") // UpdateTargetProviderMetadataDTO | Provider metadata

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.TargetAPI.UpdateTargetProviderMetadata(context.Background(), targetId).Metadata(metadata).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TargetAPI.UpdateTargetProviderMetadata``: %v\n", err)
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

Other parameters are passed through a pointer to a apiUpdateTargetProviderMetadataRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **metadata** | [**UpdateTargetProviderMetadataDTO**](UpdateTargetProviderMetadataDTO.md) | Provider metadata | 

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

