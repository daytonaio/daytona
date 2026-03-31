# ServerApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**initialize**](ServerApi.md#initialize) | **POST** /init | Initialize toolbox server |


<a id="initialize"></a>
# **initialize**
> Map&lt;String, String&gt; initialize(request)

Initialize toolbox server

Set the auth token and initialize telemetry for the toolbox server

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ServerApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ServerApi apiInstance = new ServerApi(defaultClient);
    InitializeRequest request = new InitializeRequest(); // InitializeRequest | Initialization request
    try {
      Map<String, String> result = apiInstance.initialize(request);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ServerApi#initialize");
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
| **request** | [**InitializeRequest**](InitializeRequest.md)| Initialization request | |

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

