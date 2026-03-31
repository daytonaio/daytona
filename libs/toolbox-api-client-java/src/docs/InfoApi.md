# InfoApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**getUserHomeDir**](InfoApi.md#getUserHomeDir) | **GET** /user-home-dir | Get user home directory |
| [**getVersion**](InfoApi.md#getVersion) | **GET** /version | Get version |
| [**getWorkDir**](InfoApi.md#getWorkDir) | **GET** /work-dir | Get working directory |


<a id="getUserHomeDir"></a>
# **getUserHomeDir**
> UserHomeDirResponse getUserHomeDir()

Get user home directory

Get the current user home directory path.

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.InfoApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    InfoApi apiInstance = new InfoApi(defaultClient);
    try {
      UserHomeDirResponse result = apiInstance.getUserHomeDir();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling InfoApi#getUserHomeDir");
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

[**UserHomeDirResponse**](UserHomeDirResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="getVersion"></a>
# **getVersion**
> Map&lt;String, String&gt; getVersion()

Get version

Get the current daemon version

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.InfoApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    InfoApi apiInstance = new InfoApi(defaultClient);
    try {
      Map<String, String> result = apiInstance.getVersion();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling InfoApi#getVersion");
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

<a id="getWorkDir"></a>
# **getWorkDir**
> WorkDirResponse getWorkDir()

Get working directory

Get the current working directory path. This is default directory used for running commands.

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.InfoApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    InfoApi apiInstance = new InfoApi(defaultClient);
    try {
      WorkDirResponse result = apiInstance.getWorkDir();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling InfoApi#getWorkDir");
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

[**WorkDirResponse**](WorkDirResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

