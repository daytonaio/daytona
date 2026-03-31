# Daytona.ToolboxApiClient.Api.InterpreterApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|--------|--------------|-------------|
| [**CreateInterpreterContext**](InterpreterApi.md#createinterpretercontext) | **POST** /process/interpreter/context | Create a new interpreter context |
| [**DeleteInterpreterContext**](InterpreterApi.md#deleteinterpretercontext) | **DELETE** /process/interpreter/context/{id} | Delete an interpreter context |
| [**ExecuteInterpreterCode**](InterpreterApi.md#executeinterpretercode) | **GET** /process/interpreter/execute | Execute code in an interpreter context |
| [**ListInterpreterContexts**](InterpreterApi.md#listinterpretercontexts) | **GET** /process/interpreter/context | List all user-created interpreter contexts |

<a id="createinterpretercontext"></a>
# **CreateInterpreterContext**
> InterpreterContext CreateInterpreterContext (CreateContextRequest request)

Create a new interpreter context

Creates a new isolated interpreter context with optional working directory and language

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
    public class CreateInterpreterContextExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new InterpreterApi(httpClient, config, httpClientHandler);
            var request = new CreateContextRequest(); // CreateContextRequest | Context configuration

            try
            {
                // Create a new interpreter context
                InterpreterContext result = apiInstance.CreateInterpreterContext(request);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling InterpreterApi.CreateInterpreterContext: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the CreateInterpreterContextWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Create a new interpreter context
    ApiResponse<InterpreterContext> response = apiInstance.CreateInterpreterContextWithHttpInfo(request);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling InterpreterApi.CreateInterpreterContextWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **request** | [**CreateContextRequest**](CreateContextRequest.md) | Context configuration |  |

### Return type

[**InterpreterContext**](InterpreterContext.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **400** | Bad Request |  -  |
| **500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="deleteinterpretercontext"></a>
# **DeleteInterpreterContext**
> Dictionary&lt;string, string&gt; DeleteInterpreterContext (string id)

Delete an interpreter context

Deletes an interpreter context and shuts down its worker process

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
    public class DeleteInterpreterContextExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new InterpreterApi(httpClient, config, httpClientHandler);
            var id = "id_example";  // string | Context ID

            try
            {
                // Delete an interpreter context
                Dictionary<string, string> result = apiInstance.DeleteInterpreterContext(id);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling InterpreterApi.DeleteInterpreterContext: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DeleteInterpreterContextWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Delete an interpreter context
    ApiResponse<Dictionary<string, string>> response = apiInstance.DeleteInterpreterContextWithHttpInfo(id);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling InterpreterApi.DeleteInterpreterContextWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **id** | **string** | Context ID |  |

### Return type

**Dictionary<string, string>**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **400** | Bad Request |  -  |
| **404** | Not Found |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="executeinterpretercode"></a>
# **ExecuteInterpreterCode**
> void ExecuteInterpreterCode ()

Execute code in an interpreter context

Executes code in a specified context (or default context if not specified) via WebSocket streaming

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
    public class ExecuteInterpreterCodeExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new InterpreterApi(httpClient, config, httpClientHandler);

            try
            {
                // Execute code in an interpreter context
                apiInstance.ExecuteInterpreterCode();
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling InterpreterApi.ExecuteInterpreterCode: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ExecuteInterpreterCodeWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Execute code in an interpreter context
    apiInstance.ExecuteInterpreterCodeWithHttpInfo();
}
catch (ApiException e)
{
    Debug.Print("Exception when calling InterpreterApi.ExecuteInterpreterCodeWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters
This endpoint does not need any parameter.
### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **101** | Switching Protocols |  * Connection - Upgrade <br>  * Upgrade - websocket <br>  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="listinterpretercontexts"></a>
# **ListInterpreterContexts**
> ListContextsResponse ListInterpreterContexts ()

List all user-created interpreter contexts

Returns information about all user-created interpreter contexts (excludes default context)

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
    public class ListInterpreterContextsExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new InterpreterApi(httpClient, config, httpClientHandler);

            try
            {
                // List all user-created interpreter contexts
                ListContextsResponse result = apiInstance.ListInterpreterContexts();
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling InterpreterApi.ListInterpreterContexts: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ListInterpreterContextsWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // List all user-created interpreter contexts
    ApiResponse<ListContextsResponse> response = apiInstance.ListInterpreterContextsWithHttpInfo();
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling InterpreterApi.ListInterpreterContextsWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters
This endpoint does not need any parameter.
### Return type

[**ListContextsResponse**](ListContextsResponse.md)

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

