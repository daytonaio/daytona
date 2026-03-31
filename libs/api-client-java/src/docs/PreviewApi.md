# PreviewApi

All URIs are relative to *http://localhost:3000*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**getSandboxIdFromSignedPreviewUrlToken**](PreviewApi.md#getSandboxIdFromSignedPreviewUrlToken) | **GET** /preview/{signedPreviewToken}/{port}/sandbox-id | Get sandbox ID from signed preview URL token |
| [**hasSandboxAccess**](PreviewApi.md#hasSandboxAccess) | **GET** /preview/{sandboxId}/access | Check if user has access to the sandbox |
| [**isSandboxPublic**](PreviewApi.md#isSandboxPublic) | **GET** /preview/{sandboxId}/public | Check if sandbox is public |
| [**isValidAuthToken**](PreviewApi.md#isValidAuthToken) | **GET** /preview/{sandboxId}/validate/{authToken} | Check if sandbox auth token is valid |


<a id="getSandboxIdFromSignedPreviewUrlToken"></a>
# **getSandboxIdFromSignedPreviewUrlToken**
> String getSandboxIdFromSignedPreviewUrlToken(signedPreviewToken, port)

Get sandbox ID from signed preview URL token

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.PreviewApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");

    PreviewApi apiInstance = new PreviewApi(defaultClient);
    String signedPreviewToken = "signedPreviewToken_example"; // String | Signed preview URL token
    BigDecimal port = new BigDecimal(78); // BigDecimal | Port number to get sandbox ID from signed preview URL token
    try {
      String result = apiInstance.getSandboxIdFromSignedPreviewUrlToken(signedPreviewToken, port);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling PreviewApi#getSandboxIdFromSignedPreviewUrlToken");
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
| **signedPreviewToken** | **String**| Signed preview URL token | |
| **port** | **BigDecimal**| Port number to get sandbox ID from signed preview URL token | |

### Return type

**String**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Sandbox ID from signed preview URL token |  -  |

<a id="hasSandboxAccess"></a>
# **hasSandboxAccess**
> Boolean hasSandboxAccess(sandboxId)

Check if user has access to the sandbox

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.PreviewApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    PreviewApi apiInstance = new PreviewApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    try {
      Boolean result = apiInstance.hasSandboxAccess(sandboxId);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling PreviewApi#hasSandboxAccess");
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
| **sandboxId** | **String**|  | |

### Return type

**Boolean**

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | User access status to the sandbox |  -  |

<a id="isSandboxPublic"></a>
# **isSandboxPublic**
> Boolean isSandboxPublic(sandboxId)

Check if sandbox is public

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.PreviewApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");

    PreviewApi apiInstance = new PreviewApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | ID of the sandbox
    try {
      Boolean result = apiInstance.isSandboxPublic(sandboxId);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling PreviewApi#isSandboxPublic");
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

**Boolean**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Public status of the sandbox |  -  |

<a id="isValidAuthToken"></a>
# **isValidAuthToken**
> Boolean isValidAuthToken(sandboxId, authToken)

Check if sandbox auth token is valid

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.PreviewApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");

    PreviewApi apiInstance = new PreviewApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | ID of the sandbox
    String authToken = "authToken_example"; // String | Auth token of the sandbox
    try {
      Boolean result = apiInstance.isValidAuthToken(sandboxId, authToken);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling PreviewApi#isValidAuthToken");
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
| **authToken** | **String**| Auth token of the sandbox | |

### Return type

**Boolean**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Sandbox auth token validation status |  -  |

