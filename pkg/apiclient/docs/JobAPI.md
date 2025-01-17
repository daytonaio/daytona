# \JobAPI

All URIs are relative to *http://localhost:3986*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ListJobs**](JobAPI.md#ListJobs) | **Get** /job | List jobs



## ListJobs

> []Job ListJobs(ctx).States(states).Actions(actions).ResourceId(resourceId).ResourceType(resourceType).Execute()

List jobs



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
	states := []string{"Inner_example"} // []string | Job States (optional)
	actions := []string{"Inner_example"} // []string | Job Actions (optional)
	resourceId := "resourceId_example" // string | Resource ID (optional)
	resourceType := "resourceType_example" // string | Resource Type (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.JobAPI.ListJobs(context.Background()).States(states).Actions(actions).ResourceId(resourceId).ResourceType(resourceType).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `JobAPI.ListJobs``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListJobs`: []Job
	fmt.Fprintf(os.Stdout, "Response from `JobAPI.ListJobs`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListJobsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **states** | **[]string** | Job States | 
 **actions** | **[]string** | Job Actions | 
 **resourceId** | **string** | Resource ID | 
 **resourceType** | **string** | Resource Type | 

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

