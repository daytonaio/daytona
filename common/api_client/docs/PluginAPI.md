# \PluginAPI

All URIs are relative to *http://localhost:3000*

Method | HTTP request | Description
------------- | ------------- | -------------
[**InstallProvisionerPlugin**](PluginAPI.md#InstallProvisionerPlugin) | **Post** /plugin/install/provisioner | Install a provisioner plugin



## InstallProvisionerPlugin

> InstallProvisionerPlugin(ctx).Plugin(plugin).Execute()

Install a provisioner plugin



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID/api_client"
)

func main() {
	plugin := *openapiclient.NewInstallPluginRequest() // InstallPluginRequest | Plugin to install

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PluginAPI.InstallProvisionerPlugin(context.Background()).Plugin(plugin).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PluginAPI.InstallProvisionerPlugin``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiInstallProvisionerPluginRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **plugin** | [**InstallPluginRequest**](InstallPluginRequest.md) | Plugin to install | 

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

