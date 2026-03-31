# InterpreterApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**createInterpreterContext**](InterpreterApi.md#createInterpreterContext) | **POST** /process/interpreter/context | Create a new interpreter context |
| [**deleteInterpreterContext**](InterpreterApi.md#deleteInterpreterContext) | **DELETE** /process/interpreter/context/{id} | Delete an interpreter context |
| [**executeInterpreterCode**](InterpreterApi.md#executeInterpreterCode) | **GET** /process/interpreter/execute | Execute code in an interpreter context |
| [**listInterpreterContexts**](InterpreterApi.md#listInterpreterContexts) | **GET** /process/interpreter/context | List all user-created interpreter contexts |


<a id="createInterpreterContext"></a>
# **createInterpreterContext**
> InterpreterContext createInterpreterContext(request)

Create a new interpreter context

Creates a new isolated interpreter context with optional working directory and language

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.InterpreterApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    InterpreterApi apiInstance = new InterpreterApi(defaultClient);
    CreateContextRequest request = new CreateContextRequest(); // CreateContextRequest | Context configuration
    try {
      InterpreterContext result = apiInstance.createInterpreterContext(request);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling InterpreterApi#createInterpreterContext");
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
| **request** | [**CreateContextRequest**](CreateContextRequest.md)| Context configuration | |

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

<a id="deleteInterpreterContext"></a>
# **deleteInterpreterContext**
> Map&lt;String, String&gt; deleteInterpreterContext(id)

Delete an interpreter context

Deletes an interpreter context and shuts down its worker process

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.InterpreterApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    InterpreterApi apiInstance = new InterpreterApi(defaultClient);
    String id = "id_example"; // String | Context ID
    try {
      Map<String, String> result = apiInstance.deleteInterpreterContext(id);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling InterpreterApi#deleteInterpreterContext");
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
| **id** | **String**| Context ID | |

### Return type

**Map&lt;String, String&gt;**

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

<a id="executeInterpreterCode"></a>
# **executeInterpreterCode**
> executeInterpreterCode()

Execute code in an interpreter context

Executes code in a specified context (or default context if not specified) via WebSocket streaming

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.InterpreterApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    InterpreterApi apiInstance = new InterpreterApi(defaultClient);
    try {
      apiInstance.executeInterpreterCode();
    } catch (ApiException e) {
      System.err.println("Exception when calling InterpreterApi#executeInterpreterCode");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **101** | Switching Protocols |  * Connection - Upgrade <br>  * Upgrade - websocket <br>  |

<a id="listInterpreterContexts"></a>
# **listInterpreterContexts**
> ListContextsResponse listInterpreterContexts()

List all user-created interpreter contexts

Returns information about all user-created interpreter contexts (excludes default context)

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.InterpreterApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    InterpreterApi apiInstance = new InterpreterApi(defaultClient);
    try {
      ListContextsResponse result = apiInstance.listInterpreterContexts();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling InterpreterApi#listInterpreterContexts");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
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

