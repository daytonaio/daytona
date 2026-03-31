# HealthApi

All URIs are relative to *http://localhost:3000*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**healthControllerCheck**](HealthApi.md#healthControllerCheck) | **GET** /health/ready |  |
| [**healthControllerLive**](HealthApi.md#healthControllerLive) | **GET** /health |  |


<a id="healthControllerCheck"></a>
# **healthControllerCheck**
> HealthControllerCheck200Response healthControllerCheck()



### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.HealthApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");

    HealthApi apiInstance = new HealthApi(defaultClient);
    try {
      HealthControllerCheck200Response result = apiInstance.healthControllerCheck();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling HealthApi#healthControllerCheck");
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

[**HealthControllerCheck200Response**](HealthControllerCheck200Response.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The Health Check is successful |  -  |
| **503** | The Health Check is not successful |  -  |

<a id="healthControllerLive"></a>
# **healthControllerLive**
> healthControllerLive()



### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.HealthApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");

    HealthApi apiInstance = new HealthApi(defaultClient);
    try {
      apiInstance.healthControllerLive();
    } catch (ApiException e) {
      System.err.println("Exception when calling HealthApi#healthControllerLive");
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
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** |  |  -  |

