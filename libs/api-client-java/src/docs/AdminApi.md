# AdminApi

All URIs are relative to *http://localhost:3000*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**adminCreateRunner**](AdminApi.md#adminCreateRunner) | **POST** /admin/runners | Create runner |
| [**adminDeleteRunner**](AdminApi.md#adminDeleteRunner) | **DELETE** /admin/runners/{id} | Delete runner |
| [**adminGetRunnerById**](AdminApi.md#adminGetRunnerById) | **GET** /admin/runners/{id} | Get runner by ID |
| [**adminListRunners**](AdminApi.md#adminListRunners) | **GET** /admin/runners | List all runners |
| [**adminRecoverSandbox**](AdminApi.md#adminRecoverSandbox) | **POST** /admin/sandbox/{sandboxId}/recover | Recover sandbox from error state as an admin |
| [**adminUpdateRunnerScheduling**](AdminApi.md#adminUpdateRunnerScheduling) | **PATCH** /admin/runners/{id}/scheduling | Update runner scheduling status |


<a id="adminCreateRunner"></a>
# **adminCreateRunner**
> CreateRunnerResponse adminCreateRunner(adminCreateRunner)

Create runner

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.AdminApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    AdminApi apiInstance = new AdminApi(defaultClient);
    AdminCreateRunner adminCreateRunner = new AdminCreateRunner(); // AdminCreateRunner | 
    try {
      CreateRunnerResponse result = apiInstance.adminCreateRunner(adminCreateRunner);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling AdminApi#adminCreateRunner");
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
| **adminCreateRunner** | [**AdminCreateRunner**](AdminCreateRunner.md)|  | |

### Return type

[**CreateRunnerResponse**](CreateRunnerResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** |  |  -  |

<a id="adminDeleteRunner"></a>
# **adminDeleteRunner**
> adminDeleteRunner(id)

Delete runner

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.AdminApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    AdminApi apiInstance = new AdminApi(defaultClient);
    String id = "id_example"; // String | Runner ID
    try {
      apiInstance.adminDeleteRunner(id);
    } catch (ApiException e) {
      System.err.println("Exception when calling AdminApi#adminDeleteRunner");
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
| **id** | **String**| Runner ID | |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** |  |  -  |

<a id="adminGetRunnerById"></a>
# **adminGetRunnerById**
> RunnerFull adminGetRunnerById(id)

Get runner by ID

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.AdminApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    AdminApi apiInstance = new AdminApi(defaultClient);
    String id = "id_example"; // String | Runner ID
    try {
      RunnerFull result = apiInstance.adminGetRunnerById(id);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling AdminApi#adminGetRunnerById");
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
| **id** | **String**| Runner ID | |

### Return type

[**RunnerFull**](RunnerFull.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** |  |  -  |

<a id="adminListRunners"></a>
# **adminListRunners**
> List&lt;RunnerFull&gt; adminListRunners(regionId)

List all runners

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.AdminApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    AdminApi apiInstance = new AdminApi(defaultClient);
    String regionId = "regionId_example"; // String | Filter runners by region ID
    try {
      List<RunnerFull> result = apiInstance.adminListRunners(regionId);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling AdminApi#adminListRunners");
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
| **regionId** | **String**| Filter runners by region ID | [optional] |

### Return type

[**List&lt;RunnerFull&gt;**](RunnerFull.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** |  |  -  |

<a id="adminRecoverSandbox"></a>
# **adminRecoverSandbox**
> Sandbox adminRecoverSandbox(sandboxId)

Recover sandbox from error state as an admin

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.AdminApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    AdminApi apiInstance = new AdminApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | ID of the sandbox
    try {
      Sandbox result = apiInstance.adminRecoverSandbox(sandboxId);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling AdminApi#adminRecoverSandbox");
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
| **sandboxId** | **String**| ID of the sandbox | |

### Return type

[**Sandbox**](Sandbox.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Recovery initiated |  -  |

<a id="adminUpdateRunnerScheduling"></a>
# **adminUpdateRunnerScheduling**
> adminUpdateRunnerScheduling(id)

Update runner scheduling status

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.AdminApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    AdminApi apiInstance = new AdminApi(defaultClient);
    String id = "id_example"; // String | 
    try {
      apiInstance.adminUpdateRunnerScheduling(id);
    } catch (ApiException e) {
      System.err.println("Exception when calling AdminApi#adminUpdateRunnerScheduling");
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
| **id** | **String**|  | |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** |  |  -  |

