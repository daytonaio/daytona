# \SampleAPI

All URIs are relative to *http://localhost:3986*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ListSamples**](SampleAPI.md#ListSamples) | **Get** /sample | List samples



## ListSamples

> []Sample ListSamples(ctx).Execute()

List samples



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
	resp, r, err := apiClient.SampleAPI.ListSamples(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SampleAPI.ListSamples``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListSamples`: []Sample
	fmt.Fprintf(os.Stdout, "Response from `SampleAPI.ListSamples`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiListSamplesRequest struct via the builder pattern


### Return type

[**[]Sample**](Sample.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

