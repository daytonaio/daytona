# Daytona.ToolboxApiClient.Api.LspApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|--------|--------------|-------------|
| [**Completions**](LspApi.md#completions) | **POST** /lsp/completions | Get code completions |
| [**DidClose**](LspApi.md#didclose) | **POST** /lsp/did-close | Notify document closed |
| [**DidOpen**](LspApi.md#didopen) | **POST** /lsp/did-open | Notify document opened |
| [**DocumentSymbols**](LspApi.md#documentsymbols) | **GET** /lsp/document-symbols | Get document symbols |
| [**Start**](LspApi.md#start) | **POST** /lsp/start | Start LSP server |
| [**Stop**](LspApi.md#stop) | **POST** /lsp/stop | Stop LSP server |
| [**WorkspaceSymbols**](LspApi.md#workspacesymbols) | **GET** /lsp/workspacesymbols | Get workspace symbols |

<a id="completions"></a>
# **Completions**
> CompletionList Completions (LspCompletionParams request)

Get code completions

Get code completion suggestions from the LSP server

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using Daytona.ToolboxApiClient.Api;
using Daytona.ToolboxApiClient.Client;
using Daytona.ToolboxApiClient.Model;

namespace Example
{
    public class CompletionsExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new LspApi(httpClient, config, httpClientHandler);
            var request = new LspCompletionParams(); // LspCompletionParams | Completion request

            try
            {
                // Get code completions
                CompletionList result = apiInstance.Completions(request);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling LspApi.Completions: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the CompletionsWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get code completions
    ApiResponse<CompletionList> response = apiInstance.CompletionsWithHttpInfo(request);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling LspApi.CompletionsWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **request** | [**LspCompletionParams**](LspCompletionParams.md) | Completion request |  |

### Return type

[**CompletionList**](CompletionList.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="didclose"></a>
# **DidClose**
> void DidClose (LspDocumentRequest request)

Notify document closed

Notify the LSP server that a document has been closed

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using Daytona.ToolboxApiClient.Api;
using Daytona.ToolboxApiClient.Client;
using Daytona.ToolboxApiClient.Model;

namespace Example
{
    public class DidCloseExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new LspApi(httpClient, config, httpClientHandler);
            var request = new LspDocumentRequest(); // LspDocumentRequest | Document request

            try
            {
                // Notify document closed
                apiInstance.DidClose(request);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling LspApi.DidClose: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DidCloseWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Notify document closed
    apiInstance.DidCloseWithHttpInfo(request);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling LspApi.DidCloseWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **request** | [**LspDocumentRequest**](LspDocumentRequest.md) | Document request |  |

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="didopen"></a>
# **DidOpen**
> void DidOpen (LspDocumentRequest request)

Notify document opened

Notify the LSP server that a document has been opened

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using Daytona.ToolboxApiClient.Api;
using Daytona.ToolboxApiClient.Client;
using Daytona.ToolboxApiClient.Model;

namespace Example
{
    public class DidOpenExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new LspApi(httpClient, config, httpClientHandler);
            var request = new LspDocumentRequest(); // LspDocumentRequest | Document request

            try
            {
                // Notify document opened
                apiInstance.DidOpen(request);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling LspApi.DidOpen: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DidOpenWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Notify document opened
    apiInstance.DidOpenWithHttpInfo(request);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling LspApi.DidOpenWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **request** | [**LspDocumentRequest**](LspDocumentRequest.md) | Document request |  |

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="documentsymbols"></a>
# **DocumentSymbols**
> List&lt;LspSymbol&gt; DocumentSymbols (string languageId, string pathToProject, string uri)

Get document symbols

Get symbols (functions, classes, etc.) from a document

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using Daytona.ToolboxApiClient.Api;
using Daytona.ToolboxApiClient.Client;
using Daytona.ToolboxApiClient.Model;

namespace Example
{
    public class DocumentSymbolsExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new LspApi(httpClient, config, httpClientHandler);
            var languageId = "languageId_example";  // string | Language ID (e.g., python, typescript)
            var pathToProject = "pathToProject_example";  // string | Path to project
            var uri = "uri_example";  // string | Document URI

            try
            {
                // Get document symbols
                List<LspSymbol> result = apiInstance.DocumentSymbols(languageId, pathToProject, uri);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling LspApi.DocumentSymbols: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DocumentSymbolsWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get document symbols
    ApiResponse<List<LspSymbol>> response = apiInstance.DocumentSymbolsWithHttpInfo(languageId, pathToProject, uri);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling LspApi.DocumentSymbolsWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **languageId** | **string** | Language ID (e.g., python, typescript) |  |
| **pathToProject** | **string** | Path to project |  |
| **uri** | **string** | Document URI |  |

### Return type

[**List&lt;LspSymbol&gt;**](LspSymbol.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="start"></a>
# **Start**
> void Start (LspServerRequest request)

Start LSP server

Start a Language Server Protocol server for the specified language

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using Daytona.ToolboxApiClient.Api;
using Daytona.ToolboxApiClient.Client;
using Daytona.ToolboxApiClient.Model;

namespace Example
{
    public class StartExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new LspApi(httpClient, config, httpClientHandler);
            var request = new LspServerRequest(); // LspServerRequest | LSP server request

            try
            {
                // Start LSP server
                apiInstance.Start(request);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling LspApi.Start: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the StartWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Start LSP server
    apiInstance.StartWithHttpInfo(request);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling LspApi.StartWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **request** | [**LspServerRequest**](LspServerRequest.md) | LSP server request |  |

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="stop"></a>
# **Stop**
> void Stop (LspServerRequest request)

Stop LSP server

Stop a Language Server Protocol server

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using Daytona.ToolboxApiClient.Api;
using Daytona.ToolboxApiClient.Client;
using Daytona.ToolboxApiClient.Model;

namespace Example
{
    public class StopExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new LspApi(httpClient, config, httpClientHandler);
            var request = new LspServerRequest(); // LspServerRequest | LSP server request

            try
            {
                // Stop LSP server
                apiInstance.Stop(request);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling LspApi.Stop: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the StopWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Stop LSP server
    apiInstance.StopWithHttpInfo(request);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling LspApi.StopWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **request** | [**LspServerRequest**](LspServerRequest.md) | LSP server request |  |

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="workspacesymbols"></a>
# **WorkspaceSymbols**
> List&lt;LspSymbol&gt; WorkspaceSymbols (string query, string languageId, string pathToProject)

Get workspace symbols

Search for symbols across the entire workspace

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using Daytona.ToolboxApiClient.Api;
using Daytona.ToolboxApiClient.Client;
using Daytona.ToolboxApiClient.Model;

namespace Example
{
    public class WorkspaceSymbolsExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new LspApi(httpClient, config, httpClientHandler);
            var query = "query_example";  // string | Search query
            var languageId = "languageId_example";  // string | Language ID (e.g., python, typescript)
            var pathToProject = "pathToProject_example";  // string | Path to project

            try
            {
                // Get workspace symbols
                List<LspSymbol> result = apiInstance.WorkspaceSymbols(query, languageId, pathToProject);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling LspApi.WorkspaceSymbols: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the WorkspaceSymbolsWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get workspace symbols
    ApiResponse<List<LspSymbol>> response = apiInstance.WorkspaceSymbolsWithHttpInfo(query, languageId, pathToProject);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling LspApi.WorkspaceSymbolsWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **query** | **string** | Search query |  |
| **languageId** | **string** | Language ID (e.g., python, typescript) |  |
| **pathToProject** | **string** | Path to project |  |

### Return type

[**List&lt;LspSymbol&gt;**](LspSymbol.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

