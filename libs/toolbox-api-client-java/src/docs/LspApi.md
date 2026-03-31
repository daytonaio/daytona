# LspApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**completions**](LspApi.md#completions) | **POST** /lsp/completions | Get code completions |
| [**didClose**](LspApi.md#didClose) | **POST** /lsp/did-close | Notify document closed |
| [**didOpen**](LspApi.md#didOpen) | **POST** /lsp/did-open | Notify document opened |
| [**documentSymbols**](LspApi.md#documentSymbols) | **GET** /lsp/document-symbols | Get document symbols |
| [**start**](LspApi.md#start) | **POST** /lsp/start | Start LSP server |
| [**stop**](LspApi.md#stop) | **POST** /lsp/stop | Stop LSP server |
| [**workspaceSymbols**](LspApi.md#workspaceSymbols) | **GET** /lsp/workspacesymbols | Get workspace symbols |


<a id="completions"></a>
# **completions**
> CompletionList completions(request)

Get code completions

Get code completion suggestions from the LSP server

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.LspApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    LspApi apiInstance = new LspApi(defaultClient);
    LspCompletionParams request = new LspCompletionParams(); // LspCompletionParams | Completion request
    try {
      CompletionList result = apiInstance.completions(request);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling LspApi#completions");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **request** | [**LspCompletionParams**](LspCompletionParams.md)| Completion request | |

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

<a id="didClose"></a>
# **didClose**
> didClose(request)

Notify document closed

Notify the LSP server that a document has been closed

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.LspApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    LspApi apiInstance = new LspApi(defaultClient);
    LspDocumentRequest request = new LspDocumentRequest(); // LspDocumentRequest | Document request
    try {
      apiInstance.didClose(request);
    } catch (ApiException e) {
      System.err.println("Exception when calling LspApi#didClose");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **request** | [**LspDocumentRequest**](LspDocumentRequest.md)| Document request | |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="didOpen"></a>
# **didOpen**
> didOpen(request)

Notify document opened

Notify the LSP server that a document has been opened

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.LspApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    LspApi apiInstance = new LspApi(defaultClient);
    LspDocumentRequest request = new LspDocumentRequest(); // LspDocumentRequest | Document request
    try {
      apiInstance.didOpen(request);
    } catch (ApiException e) {
      System.err.println("Exception when calling LspApi#didOpen");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **request** | [**LspDocumentRequest**](LspDocumentRequest.md)| Document request | |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="documentSymbols"></a>
# **documentSymbols**
> List&lt;LspSymbol&gt; documentSymbols(languageId, pathToProject, uri)

Get document symbols

Get symbols (functions, classes, etc.) from a document

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.LspApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    LspApi apiInstance = new LspApi(defaultClient);
    String languageId = "languageId_example"; // String | Language ID (e.g., python, typescript)
    String pathToProject = "pathToProject_example"; // String | Path to project
    String uri = "uri_example"; // String | Document URI
    try {
      List<LspSymbol> result = apiInstance.documentSymbols(languageId, pathToProject, uri);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling LspApi#documentSymbols");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **languageId** | **String**| Language ID (e.g., python, typescript) | |
| **pathToProject** | **String**| Path to project | |
| **uri** | **String**| Document URI | |

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

<a id="start"></a>
# **start**
> start(request)

Start LSP server

Start a Language Server Protocol server for the specified language

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.LspApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    LspApi apiInstance = new LspApi(defaultClient);
    LspServerRequest request = new LspServerRequest(); // LspServerRequest | LSP server request
    try {
      apiInstance.start(request);
    } catch (ApiException e) {
      System.err.println("Exception when calling LspApi#start");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **request** | [**LspServerRequest**](LspServerRequest.md)| LSP server request | |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="stop"></a>
# **stop**
> stop(request)

Stop LSP server

Stop a Language Server Protocol server

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.LspApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    LspApi apiInstance = new LspApi(defaultClient);
    LspServerRequest request = new LspServerRequest(); // LspServerRequest | LSP server request
    try {
      apiInstance.stop(request);
    } catch (ApiException e) {
      System.err.println("Exception when calling LspApi#stop");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **request** | [**LspServerRequest**](LspServerRequest.md)| LSP server request | |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="workspaceSymbols"></a>
# **workspaceSymbols**
> List&lt;LspSymbol&gt; workspaceSymbols(query, languageId, pathToProject)

Get workspace symbols

Search for symbols across the entire workspace

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.LspApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    LspApi apiInstance = new LspApi(defaultClient);
    String query = "query_example"; // String | Search query
    String languageId = "languageId_example"; // String | Language ID (e.g., python, typescript)
    String pathToProject = "pathToProject_example"; // String | Path to project
    try {
      List<LspSymbol> result = apiInstance.workspaceSymbols(query, languageId, pathToProject);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling LspApi#workspaceSymbols");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **query** | **String**| Search query | |
| **languageId** | **String**| Language ID (e.g., python, typescript) | |
| **pathToProject** | **String**| Path to project | |

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

