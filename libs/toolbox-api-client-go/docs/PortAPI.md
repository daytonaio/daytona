# \PortAPI

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetPorts**](PortAPI.md#GetPorts) | **Get** /port | Get active ports
[**IsPortInUse**](PortAPI.md#IsPortInUse) | **Get** /port/{port}/in-use | Check if port is in use



## GetPorts

> PortList GetPorts(ctx).Execute()

Get active ports



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

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PortAPI.GetPorts(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PortAPI.GetPorts``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetPorts`: PortList
	fmt.Fprintf(os.Stdout, "Response from `PortAPI.GetPorts`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetPortsRequest struct via the builder pattern


### Return type

[**PortList**](PortList.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## IsPortInUse

> IsPortInUseResponse IsPortInUse(ctx, port).Execute()

Check if port is in use



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
	port := int32(56) // int32 | Port number (3000-9999)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PortAPI.IsPortInUse(context.Background(), port).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PortAPI.IsPortInUse``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `IsPortInUse`: IsPortInUseResponse
	fmt.Fprintf(os.Stdout, "Response from `PortAPI.IsPortInUse`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**port** | **int32** | Port number (3000-9999) | 

### Other Parameters

Other parameters are passed through a pointer to a apiIsPortInUseRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**IsPortInUseResponse**](IsPortInUseResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

