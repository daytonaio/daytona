# \LspAPI

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**Completions**](LspAPI.md#Completions) | **Post** /lsp/completions | Get code completions
[**DidClose**](LspAPI.md#DidClose) | **Post** /lsp/did-close | Notify document closed
[**DidOpen**](LspAPI.md#DidOpen) | **Post** /lsp/did-open | Notify document opened
[**DocumentSymbols**](LspAPI.md#DocumentSymbols) | **Get** /lsp/document-symbols | Get document symbols
[**Start**](LspAPI.md#Start) | **Post** /lsp/start | Start LSP server
[**Stop**](LspAPI.md#Stop) | **Post** /lsp/stop | Stop LSP server
[**WorkspaceSymbols**](LspAPI.md#WorkspaceSymbols) | **Get** /lsp/workspacesymbols | Get workspace symbols



## Completions

> CompletionList Completions(ctx).Request(request).Execute()

Get code completions



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	request := *openapiclient.NewLspCompletionParams("LanguageId_example", "PathToProject_example", *openapiclient.NewLspPosition(int32(123), int32(123)), "Uri_example") // LspCompletionParams | Completion request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.LspAPI.Completions(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `LspAPI.Completions``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `Completions`: CompletionList
	fmt.Fprintf(os.Stdout, "Response from `LspAPI.Completions`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCompletionsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**LspCompletionParams**](LspCompletionParams.md) | Completion request | 

### Return type

[**CompletionList**](CompletionList.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DidClose

> DidClose(ctx).Request(request).Execute()

Notify document closed



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	request := *openapiclient.NewLspDocumentRequest("LanguageId_example", "PathToProject_example", "Uri_example") // LspDocumentRequest | Document request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.LspAPI.DidClose(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `LspAPI.DidClose``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDidCloseRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**LspDocumentRequest**](LspDocumentRequest.md) | Document request | 

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


## DidOpen

> DidOpen(ctx).Request(request).Execute()

Notify document opened



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	request := *openapiclient.NewLspDocumentRequest("LanguageId_example", "PathToProject_example", "Uri_example") // LspDocumentRequest | Document request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.LspAPI.DidOpen(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `LspAPI.DidOpen``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDidOpenRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**LspDocumentRequest**](LspDocumentRequest.md) | Document request | 

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


## DocumentSymbols

> []LspSymbol DocumentSymbols(ctx).LanguageId(languageId).PathToProject(pathToProject).Uri(uri).Execute()

Get document symbols



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	languageId := "languageId_example" // string | Language ID (e.g., python, typescript)
	pathToProject := "pathToProject_example" // string | Path to project
	uri := "uri_example" // string | Document URI

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.LspAPI.DocumentSymbols(context.Background()).LanguageId(languageId).PathToProject(pathToProject).Uri(uri).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `LspAPI.DocumentSymbols``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DocumentSymbols`: []LspSymbol
	fmt.Fprintf(os.Stdout, "Response from `LspAPI.DocumentSymbols`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDocumentSymbolsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **languageId** | **string** | Language ID (e.g., python, typescript) | 
 **pathToProject** | **string** | Path to project | 
 **uri** | **string** | Document URI | 

### Return type

[**[]LspSymbol**](LspSymbol.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## Start

> Start(ctx).Request(request).Execute()

Start LSP server



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	request := *openapiclient.NewLspServerRequest("LanguageId_example", "PathToProject_example") // LspServerRequest | LSP server request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.LspAPI.Start(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `LspAPI.Start``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiStartRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**LspServerRequest**](LspServerRequest.md) | LSP server request | 

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


## Stop

> Stop(ctx).Request(request).Execute()

Stop LSP server



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	request := *openapiclient.NewLspServerRequest("LanguageId_example", "PathToProject_example") // LspServerRequest | LSP server request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.LspAPI.Stop(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `LspAPI.Stop``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiStopRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**LspServerRequest**](LspServerRequest.md) | LSP server request | 

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


## WorkspaceSymbols

> []LspSymbol WorkspaceSymbols(ctx).Query(query).LanguageId(languageId).PathToProject(pathToProject).Execute()

Get workspace symbols



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	query := "query_example" // string | Search query
	languageId := "languageId_example" // string | Language ID (e.g., python, typescript)
	pathToProject := "pathToProject_example" // string | Path to project

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.LspAPI.WorkspaceSymbols(context.Background()).Query(query).LanguageId(languageId).PathToProject(pathToProject).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `LspAPI.WorkspaceSymbols``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `WorkspaceSymbols`: []LspSymbol
	fmt.Fprintf(os.Stdout, "Response from `LspAPI.WorkspaceSymbols`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiWorkspaceSymbolsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **query** | **string** | Search query | 
 **languageId** | **string** | Language ID (e.g., python, typescript) | 
 **pathToProject** | **string** | Path to project | 

### Return type

[**[]LspSymbol**](LspSymbol.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

