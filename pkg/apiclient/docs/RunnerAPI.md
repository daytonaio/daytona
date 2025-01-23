# \RunnerAPI

All URIs are relative to *http://localhost:3986*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateRunner**](RunnerAPI.md#CreateRunner) | **Post** /runner | Create a runner
[**DeleteRunner**](RunnerAPI.md#DeleteRunner) | **Delete** /runner/{runnerId} | Delete runner
[**FindRunner**](RunnerAPI.md#FindRunner) | **Get** /runner/{runnerId} | Find a runner
[**ListRunnerJobs**](RunnerAPI.md#ListRunnerJobs) | **Get** /runner/{runnerId}/jobs | List runner jobs
[**ListRunners**](RunnerAPI.md#ListRunners) | **Get** /runner | List runners
[**UpdateJobState**](RunnerAPI.md#UpdateJobState) | **Post** /runner/{runnerId}/jobs/{jobId}/state | Update job state
[**UpdateRunnerMetadata**](RunnerAPI.md#UpdateRunnerMetadata) | **Post** /runner/{runnerId}/metadata | Update runner metadata



## CreateRunner

> CreateRunnerResultDTO CreateRunner(ctx).Runner(runner).Execute()

Create a runner



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
	runner := *openapiclient.NewCreateRunnerDTO("Id_example", "Name_example") // CreateRunnerDTO | Runner

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RunnerAPI.CreateRunner(context.Background()).Runner(runner).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RunnerAPI.CreateRunner``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateRunner`: CreateRunnerResultDTO
	fmt.Fprintf(os.Stdout, "Response from `RunnerAPI.CreateRunner`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateRunnerRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **runner** | [**CreateRunnerDTO**](CreateRunnerDTO.md) | Runner | 

### Return type

[**CreateRunnerResultDTO**](CreateRunnerResultDTO.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteRunner

> DeleteRunner(ctx, runnerId).Execute()

Delete runner



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
	r, err := apiClient.RunnerAPI.DeleteRunner(context.Background(), runnerId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RunnerAPI.DeleteRunner``: %v\n", err)
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

Other parameters are passed through a pointer to a apiDeleteRunnerRequest struct via the builder pattern


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


## FindRunner

> RunnerDTO FindRunner(ctx, runnerId).Execute()

Find a runner



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
	resp, r, err := apiClient.RunnerAPI.FindRunner(context.Background(), runnerId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RunnerAPI.FindRunner``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `FindRunner`: RunnerDTO
	fmt.Fprintf(os.Stdout, "Response from `RunnerAPI.FindRunner`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**runnerId** | **string** | Runner ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiFindRunnerRequest struct via the builder pattern


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


## UpdateRunnerMetadata

> UpdateRunnerMetadata(ctx, runnerId).RunnerMetadata(runnerMetadata).Execute()

Update runner metadata



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
	runnerMetadata := *openapiclient.NewUpdateRunnerMetadataDTO([]openapiclient.ProviderInfo{*openapiclient.NewProviderInfo("Name_example", "RunnerId_example", "RunnerName_example", map[string]TargetConfigProperty{"key": *openapiclient.NewTargetConfigProperty()}, "Version_example")}, int32(123)) // UpdateRunnerMetadataDTO | Runner Metadata

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.RunnerAPI.UpdateRunnerMetadata(context.Background(), runnerId).RunnerMetadata(runnerMetadata).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RunnerAPI.UpdateRunnerMetadata``: %v\n", err)
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

Other parameters are passed through a pointer to a apiUpdateRunnerMetadataRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **runnerMetadata** | [**UpdateRunnerMetadataDTO**](UpdateRunnerMetadataDTO.md) | Runner Metadata | 

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

