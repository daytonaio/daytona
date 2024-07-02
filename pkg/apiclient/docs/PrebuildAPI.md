# \PrebuildAPI

All URIs are relative to *http://localhost:3986*

Method | HTTP request | Description
------------- | ------------- | -------------
[**RegisterPrebuildWebhook**](PrebuildAPI.md#RegisterPrebuildWebhook) | **Post** /prebuild/register-webhook | RegisterPrebuildWebhook
[**WebhookEvent**](PrebuildAPI.md#WebhookEvent) | **Post** /prebuild/webhook-event | WebhookEvent



## RegisterPrebuildWebhook

> RegisterPrebuildWebhook(ctx).PrebuildWebhook(prebuildWebhook).Execute()

RegisterPrebuildWebhook



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
	prebuildWebhook := *openapiclient.NewRegisterPrebuildWebhookRequest() // RegisterPrebuildWebhookRequest | Register prebuild webhook

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PrebuildAPI.RegisterPrebuildWebhook(context.Background()).PrebuildWebhook(prebuildWebhook).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PrebuildAPI.RegisterPrebuildWebhook``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiRegisterPrebuildWebhookRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **prebuildWebhook** | [**RegisterPrebuildWebhookRequest**](RegisterPrebuildWebhookRequest.md) | Register prebuild webhook | 

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


## WebhookEvent

> WebhookEvent(ctx).Workspace(workspace).Execute()

WebhookEvent



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
	workspace := map[string]interface{}{ ... } // map[string]interface{} | Webhook event

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PrebuildAPI.WebhookEvent(context.Background()).Workspace(workspace).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PrebuildAPI.WebhookEvent``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiWebhookEventRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **workspace** | **map[string]interface{}** | Webhook event | 

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

