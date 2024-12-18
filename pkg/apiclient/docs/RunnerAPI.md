# \RunnerAPI

All URIs are relative to *http://localhost:3986*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetRunner**](RunnerAPI.md#GetRunner) | **Get** /runner/{runnerId} | Get a runner
[**ListRunnerJobs**](RunnerAPI.md#ListRunnerJobs) | **Get** /runner/{runnerId}/jobs | List runner jobs
[**ListRunners**](RunnerAPI.md#ListRunners) | **Get** /runner | List runners
[**RegisterRunner**](RunnerAPI.md#RegisterRunner) | **Post** /runner | Register a runner
[**RemoveRunner**](RunnerAPI.md#RemoveRunner) | **Delete** /runner/{runnerId} | Remove runner
[**SetRunnerMetadata**](RunnerAPI.md#SetRunnerMetadata) | **Post** /runner/{runnerId}/metadata | Set runner metadata
[**UpdateJobState**](RunnerAPI.md#UpdateJobState) | **Post** /runner/{runnerId}/jobs/{jobId}/state | Update job state



## GetRunner

> RunnerDTO GetRunner(ctx, runnerId).Execute()

Get a runner



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
	runnerId := "runnerId_example" // string | Runner ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RunnerAPI.GetRunner(context.Background(), runnerId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RunnerAPI.GetRunner``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetRunner`: RunnerDTO
	fmt.Fprintf(os.Stdout, "Response from `RunnerAPI.GetRunner`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**runnerId** | **string** | Runner ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetRunnerRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**RunnerDTO**](RunnerDTO.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListRunnerJobs

> []Job ListRunnerJobs(ctx, runnerId).Execute()

List runner jobs



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
	runnerId := "runnerId_example" // string | Runner ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RunnerAPI.ListRunnerJobs(context.Background(), runnerId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RunnerAPI.ListRunnerJobs``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListRunnerJobs`: []Job
	fmt.Fprintf(os.Stdout, "Response from `RunnerAPI.ListRunnerJobs`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**runnerId** | **string** | Runner ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiListRunnerJobsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]Job**](Job.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListRunners

> []RunnerDTO ListRunners(ctx).Execute()

List runners



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
	resp, r, err := apiClient.RunnerAPI.ListRunners(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RunnerAPI.ListRunners``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListRunners`: []RunnerDTO
	fmt.Fprintf(os.Stdout, "Response from `RunnerAPI.ListRunners`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiListRunnersRequest struct via the builder pattern


### Return type

[**[]RunnerDTO**](RunnerDTO.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RegisterRunner

> RegisterRunnerResultDTO RegisterRunner(ctx).Runner(runner).Execute()

Register a runner



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
	runner := *openapiclient.NewRegisterRunnerDTO("Id_example", "Name_example") // RegisterRunnerDTO | Register runner

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RunnerAPI.RegisterRunner(context.Background()).Runner(runner).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RunnerAPI.RegisterRunner``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `RegisterRunner`: RegisterRunnerResultDTO
	fmt.Fprintf(os.Stdout, "Response from `RunnerAPI.RegisterRunner`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiRegisterRunnerRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **runner** | [**RegisterRunnerDTO**](RegisterRunnerDTO.md) | Register runner | 

### Return type

[**RegisterRunnerResultDTO**](RegisterRunnerResultDTO.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RemoveRunner

> RemoveRunner(ctx, runnerId).Execute()

Remove runner



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
	runnerId := "runnerId_example" // string | Runner ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.RunnerAPI.RemoveRunner(context.Background(), runnerId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RunnerAPI.RemoveRunner``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**runnerId** | **string** | Runner ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiRemoveRunnerRequest struct via the builder pattern


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


## SetRunnerMetadata

> SetRunnerMetadata(ctx, runnerId).SetMetadata(setMetadata).Execute()

Set runner metadata



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
	runnerId := "runnerId_example" // string | Runner ID
	setMetadata := *openapiclient.NewSetRunnerMetadata([]openapiclient.ProviderInfo{*openapiclient.NewProviderInfo("Name_example", "RunnerId_example", map[string]TargetConfigProperty{"key": *openapiclient.NewTargetConfigProperty()}, "Version_example")}, int32(123)) // SetRunnerMetadata | Set Metadata

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.RunnerAPI.SetRunnerMetadata(context.Background(), runnerId).SetMetadata(setMetadata).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RunnerAPI.SetRunnerMetadata``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**runnerId** | **string** | Runner ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiSetRunnerMetadataRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **setMetadata** | [**SetRunnerMetadata**](SetRunnerMetadata.md) | Set Metadata | 

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


## UpdateJobState

> UpdateJobState(ctx, runnerId, jobId).UpdateJobState(updateJobState).Execute()

Update job state



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
	runnerId := "runnerId_example" // string | Runner ID
	jobId := "jobId_example" // string | Job ID
	updateJobState := *openapiclient.NewUpdateJobState(openapiclient.JobState("pending")) // UpdateJobState | Update job state

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.RunnerAPI.UpdateJobState(context.Background(), runnerId, jobId).UpdateJobState(updateJobState).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RunnerAPI.UpdateJobState``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**runnerId** | **string** | Runner ID | 
**jobId** | **string** | Job ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateJobStateRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **updateJobState** | [**UpdateJobState**](UpdateJobState.md) | Update job state | 

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

