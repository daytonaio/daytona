# ConfigApi

All URIs are relative to *http://localhost:3000*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**configControllerGetConfig**](ConfigApi.md#configControllerGetConfig) | **GET** /config | Get config |


<a id="configControllerGetConfig"></a>
# **configControllerGetConfig**
> DaytonaConfiguration configControllerGetConfig()

Get config

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ConfigApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");

    ConfigApi apiInstance = new ConfigApi(defaultClient);
    try {
      DaytonaConfiguration result = apiInstance.configControllerGetConfig();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ConfigApi#configControllerGetConfig");
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

[**DaytonaConfiguration**](DaytonaConfiguration.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Daytona configuration |  -  |

