# PortApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**getPorts**](PortApi.md#getPorts) | **GET** /port | Get active ports |
| [**isPortInUse**](PortApi.md#isPortInUse) | **GET** /port/{port}/in-use | Check if port is in use |


<a id="getPorts"></a>
# **getPorts**
> PortList getPorts()

Get active ports

Get a list of all currently active ports

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.PortApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    PortApi apiInstance = new PortApi(defaultClient);
    try {
      PortList result = apiInstance.getPorts();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling PortApi#getPorts");
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

[**PortList**](PortList.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="isPortInUse"></a>
# **isPortInUse**
> IsPortInUseResponse isPortInUse(port)

Check if port is in use

Check if a specific port is currently in use

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.PortApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    PortApi apiInstance = new PortApi(defaultClient);
    Integer port = 56; // Integer | Port number (3000-9999)
    try {
      IsPortInUseResponse result = apiInstance.isPortInUse(port);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling PortApi#isPortInUse");
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
| **port** | **Integer**| Port number (3000-9999) | |

### Return type

[**IsPortInUseResponse**](IsPortInUseResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

