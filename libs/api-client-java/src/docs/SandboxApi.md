# SandboxApi

All URIs are relative to *http://localhost:3000*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**archiveSandbox**](SandboxApi.md#archiveSandbox) | **POST** /sandbox/{sandboxIdOrName}/archive | Archive sandbox |
| [**createBackup**](SandboxApi.md#createBackup) | **POST** /sandbox/{sandboxIdOrName}/backup | Create sandbox backup |
| [**createSandbox**](SandboxApi.md#createSandbox) | **POST** /sandbox | Create a new sandbox |
| [**createSshAccess**](SandboxApi.md#createSshAccess) | **POST** /sandbox/{sandboxIdOrName}/ssh-access | Create SSH access for sandbox |
| [**deleteSandbox**](SandboxApi.md#deleteSandbox) | **DELETE** /sandbox/{sandboxIdOrName} | Delete sandbox |
| [**expireSignedPortPreviewUrl**](SandboxApi.md#expireSignedPortPreviewUrl) | **POST** /sandbox/{sandboxIdOrName}/ports/{port}/signed-preview-url/{token}/expire | Expire signed preview URL for a sandbox port |
| [**getBuildLogs**](SandboxApi.md#getBuildLogs) | **GET** /sandbox/{sandboxIdOrName}/build-logs | Get build logs |
| [**getBuildLogsUrl**](SandboxApi.md#getBuildLogsUrl) | **GET** /sandbox/{sandboxIdOrName}/build-logs-url | Get build logs URL |
| [**getPortPreviewUrl**](SandboxApi.md#getPortPreviewUrl) | **GET** /sandbox/{sandboxIdOrName}/ports/{port}/preview-url | Get preview URL for a sandbox port |
| [**getSandbox**](SandboxApi.md#getSandbox) | **GET** /sandbox/{sandboxIdOrName} | Get sandbox details |
| [**getSandboxLogs**](SandboxApi.md#getSandboxLogs) | **GET** /sandbox/{sandboxId}/telemetry/logs | Get sandbox logs |
| [**getSandboxMetrics**](SandboxApi.md#getSandboxMetrics) | **GET** /sandbox/{sandboxId}/telemetry/metrics | Get sandbox metrics |
| [**getSandboxTraceSpans**](SandboxApi.md#getSandboxTraceSpans) | **GET** /sandbox/{sandboxId}/telemetry/traces/{traceId} | Get trace spans |
| [**getSandboxTraces**](SandboxApi.md#getSandboxTraces) | **GET** /sandbox/{sandboxId}/telemetry/traces | Get sandbox traces |
| [**getSandboxesForRunner**](SandboxApi.md#getSandboxesForRunner) | **GET** /sandbox/for-runner | Get sandboxes for the authenticated runner |
| [**getSignedPortPreviewUrl**](SandboxApi.md#getSignedPortPreviewUrl) | **GET** /sandbox/{sandboxIdOrName}/ports/{port}/signed-preview-url | Get signed preview URL for a sandbox port |
| [**getToolboxProxyUrl**](SandboxApi.md#getToolboxProxyUrl) | **GET** /sandbox/{sandboxId}/toolbox-proxy-url | Get toolbox proxy URL for a sandbox |
| [**listSandboxes**](SandboxApi.md#listSandboxes) | **GET** /sandbox | List all sandboxes |
| [**listSandboxesPaginated**](SandboxApi.md#listSandboxesPaginated) | **GET** /sandbox/paginated | List all sandboxes paginated |
| [**recoverSandbox**](SandboxApi.md#recoverSandbox) | **POST** /sandbox/{sandboxIdOrName}/recover | Recover sandbox from error state |
| [**replaceLabels**](SandboxApi.md#replaceLabels) | **PUT** /sandbox/{sandboxIdOrName}/labels | Replace sandbox labels |
| [**resizeSandbox**](SandboxApi.md#resizeSandbox) | **POST** /sandbox/{sandboxIdOrName}/resize | Resize sandbox resources |
| [**revokeSshAccess**](SandboxApi.md#revokeSshAccess) | **DELETE** /sandbox/{sandboxIdOrName}/ssh-access | Revoke SSH access for sandbox |
| [**setAutoArchiveInterval**](SandboxApi.md#setAutoArchiveInterval) | **POST** /sandbox/{sandboxIdOrName}/autoarchive/{interval} | Set sandbox auto-archive interval |
| [**setAutoDeleteInterval**](SandboxApi.md#setAutoDeleteInterval) | **POST** /sandbox/{sandboxIdOrName}/autodelete/{interval} | Set sandbox auto-delete interval |
| [**setAutostopInterval**](SandboxApi.md#setAutostopInterval) | **POST** /sandbox/{sandboxIdOrName}/autostop/{interval} | Set sandbox auto-stop interval |
| [**startSandbox**](SandboxApi.md#startSandbox) | **POST** /sandbox/{sandboxIdOrName}/start | Start sandbox |
| [**stopSandbox**](SandboxApi.md#stopSandbox) | **POST** /sandbox/{sandboxIdOrName}/stop | Stop sandbox |
| [**updateLastActivity**](SandboxApi.md#updateLastActivity) | **POST** /sandbox/{sandboxId}/last-activity | Update sandbox last activity |
| [**updatePublicStatus**](SandboxApi.md#updatePublicStatus) | **POST** /sandbox/{sandboxIdOrName}/public/{isPublic} | Update public status |
| [**updateSandboxState**](SandboxApi.md#updateSandboxState) | **PUT** /sandbox/{sandboxId}/state | Update sandbox state |
| [**validateSshAccess**](SandboxApi.md#validateSshAccess) | **GET** /sandbox/ssh-access/validate | Validate SSH access for sandbox |


<a id="archiveSandbox"></a>
# **archiveSandbox**
> Sandbox archiveSandbox(sandboxIdOrName, xDaytonaOrganizationID)

Archive sandbox

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String sandboxIdOrName = "sandboxIdOrName_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      Sandbox result = apiInstance.archiveSandbox(sandboxIdOrName, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#archiveSandbox");
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
| **sandboxIdOrName** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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
| **200** | Sandbox has been archived |  -  |

<a id="createBackup"></a>
# **createBackup**
> Sandbox createBackup(sandboxIdOrName, xDaytonaOrganizationID)

Create sandbox backup

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String sandboxIdOrName = "sandboxIdOrName_example"; // String | ID or name of the sandbox
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      Sandbox result = apiInstance.createBackup(sandboxIdOrName, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#createBackup");
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
| **sandboxIdOrName** | **String**| ID or name of the sandbox | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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
| **200** | Sandbox backup has been initiated |  -  |

<a id="createSandbox"></a>
# **createSandbox**
> Sandbox createSandbox(createSandbox, xDaytonaOrganizationID)

Create a new sandbox

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    CreateSandbox createSandbox = new CreateSandbox(); // CreateSandbox | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      Sandbox result = apiInstance.createSandbox(createSandbox, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#createSandbox");
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
| **createSandbox** | [**CreateSandbox**](CreateSandbox.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

[**Sandbox**](Sandbox.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The sandbox has been successfully created. |  -  |

<a id="createSshAccess"></a>
# **createSshAccess**
> SshAccessDto createSshAccess(sandboxIdOrName, xDaytonaOrganizationID, expiresInMinutes)

Create SSH access for sandbox

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String sandboxIdOrName = "sandboxIdOrName_example"; // String | ID or name of the sandbox
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    BigDecimal expiresInMinutes = new BigDecimal(78); // BigDecimal | Expiration time in minutes (default: 60)
    try {
      SshAccessDto result = apiInstance.createSshAccess(sandboxIdOrName, xDaytonaOrganizationID, expiresInMinutes);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#createSshAccess");
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
| **sandboxIdOrName** | **String**| ID or name of the sandbox | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |
| **expiresInMinutes** | **BigDecimal**| Expiration time in minutes (default: 60) | [optional] |

### Return type

[**SshAccessDto**](SshAccessDto.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | SSH access has been created |  -  |

<a id="deleteSandbox"></a>
# **deleteSandbox**
> Sandbox deleteSandbox(sandboxIdOrName, xDaytonaOrganizationID)

Delete sandbox

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String sandboxIdOrName = "sandboxIdOrName_example"; // String | ID or name of the sandbox
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      Sandbox result = apiInstance.deleteSandbox(sandboxIdOrName, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#deleteSandbox");
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
| **sandboxIdOrName** | **String**| ID or name of the sandbox | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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
| **200** | Sandbox has been deleted |  -  |

<a id="expireSignedPortPreviewUrl"></a>
# **expireSignedPortPreviewUrl**
> expireSignedPortPreviewUrl(sandboxIdOrName, port, token, xDaytonaOrganizationID)

Expire signed preview URL for a sandbox port

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String sandboxIdOrName = "sandboxIdOrName_example"; // String | ID or name of the sandbox
    Integer port = 56; // Integer | Port number to expire signed preview URL for
    String token = "token_example"; // String | Token to expire signed preview URL for
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.expireSignedPortPreviewUrl(sandboxIdOrName, port, token, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#expireSignedPortPreviewUrl");
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
| **sandboxIdOrName** | **String**| ID or name of the sandbox | |
| **port** | **Integer**| Port number to expire signed preview URL for | |
| **token** | **String**| Token to expire signed preview URL for | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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
| **200** | Signed preview URL has been expired |  -  |

<a id="getBuildLogs"></a>
# **getBuildLogs**
> getBuildLogs(sandboxIdOrName, xDaytonaOrganizationID, follow)

Get build logs

This endpoint is deprecated. Use &#x60;getBuildLogsUrl&#x60; instead.

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String sandboxIdOrName = "sandboxIdOrName_example"; // String | ID or name of the sandbox
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    Boolean follow = true; // Boolean | Whether to follow the logs stream
    try {
      apiInstance.getBuildLogs(sandboxIdOrName, xDaytonaOrganizationID, follow);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#getBuildLogs");
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
| **sandboxIdOrName** | **String**| ID or name of the sandbox | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |
| **follow** | **Boolean**| Whether to follow the logs stream | [optional] |

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
| **200** | Build logs stream |  -  |

<a id="getBuildLogsUrl"></a>
# **getBuildLogsUrl**
> Url getBuildLogsUrl(sandboxIdOrName, xDaytonaOrganizationID)

Get build logs URL

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String sandboxIdOrName = "sandboxIdOrName_example"; // String | ID or name of the sandbox
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      Url result = apiInstance.getBuildLogsUrl(sandboxIdOrName, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#getBuildLogsUrl");
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
| **sandboxIdOrName** | **String**| ID or name of the sandbox | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

[**Url**](Url.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Build logs URL |  -  |

<a id="getPortPreviewUrl"></a>
# **getPortPreviewUrl**
> PortPreviewUrl getPortPreviewUrl(sandboxIdOrName, port, xDaytonaOrganizationID)

Get preview URL for a sandbox port

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String sandboxIdOrName = "sandboxIdOrName_example"; // String | ID or name of the sandbox
    BigDecimal port = new BigDecimal(78); // BigDecimal | Port number to get preview URL for
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      PortPreviewUrl result = apiInstance.getPortPreviewUrl(sandboxIdOrName, port, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#getPortPreviewUrl");
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
| **sandboxIdOrName** | **String**| ID or name of the sandbox | |
| **port** | **BigDecimal**| Port number to get preview URL for | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

[**PortPreviewUrl**](PortPreviewUrl.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Preview URL for the specified port |  -  |

<a id="getSandbox"></a>
# **getSandbox**
> Sandbox getSandbox(sandboxIdOrName, xDaytonaOrganizationID, verbose)

Get sandbox details

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String sandboxIdOrName = "sandboxIdOrName_example"; // String | ID or name of the sandbox
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    Boolean verbose = true; // Boolean | Include verbose output
    try {
      Sandbox result = apiInstance.getSandbox(sandboxIdOrName, xDaytonaOrganizationID, verbose);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#getSandbox");
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
| **sandboxIdOrName** | **String**| ID or name of the sandbox | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |
| **verbose** | **Boolean**| Include verbose output | [optional] |

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
| **200** | Sandbox details |  -  |

<a id="getSandboxLogs"></a>
# **getSandboxLogs**
> PaginatedLogs getSandboxLogs(sandboxId, from, to, xDaytonaOrganizationID, page, limit, severities, search)

Get sandbox logs

Retrieve OTEL logs for a sandbox within a time range

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | ID of the sandbox
    OffsetDateTime from = OffsetDateTime.now(); // OffsetDateTime | Start of time range (ISO 8601)
    OffsetDateTime to = OffsetDateTime.now(); // OffsetDateTime | End of time range (ISO 8601)
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    BigDecimal page = new BigDecimal("1"); // BigDecimal | Page number (1-indexed)
    BigDecimal limit = new BigDecimal("100"); // BigDecimal | Number of items per page
    List<String> severities = Arrays.asList(); // List<String> | Filter by severity levels (DEBUG, INFO, WARN, ERROR)
    String search = "search_example"; // String | Search in log body
    try {
      PaginatedLogs result = apiInstance.getSandboxLogs(sandboxId, from, to, xDaytonaOrganizationID, page, limit, severities, search);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#getSandboxLogs");
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
| **from** | **OffsetDateTime**| Start of time range (ISO 8601) | |
| **to** | **OffsetDateTime**| End of time range (ISO 8601) | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |
| **page** | **BigDecimal**| Page number (1-indexed) | [optional] [default to 1] |
| **limit** | **BigDecimal**| Number of items per page | [optional] [default to 100] |
| **severities** | [**List&lt;String&gt;**](String.md)| Filter by severity levels (DEBUG, INFO, WARN, ERROR) | [optional] |
| **search** | **String**| Search in log body | [optional] |

### Return type

[**PaginatedLogs**](PaginatedLogs.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Paginated list of log entries |  -  |

<a id="getSandboxMetrics"></a>
# **getSandboxMetrics**
> MetricsResponse getSandboxMetrics(sandboxId, from, to, xDaytonaOrganizationID, metricNames)

Get sandbox metrics

Retrieve OTEL metrics for a sandbox within a time range

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | ID of the sandbox
    OffsetDateTime from = OffsetDateTime.now(); // OffsetDateTime | Start of time range (ISO 8601)
    OffsetDateTime to = OffsetDateTime.now(); // OffsetDateTime | End of time range (ISO 8601)
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    List<String> metricNames = Arrays.asList(); // List<String> | Filter by metric names
    try {
      MetricsResponse result = apiInstance.getSandboxMetrics(sandboxId, from, to, xDaytonaOrganizationID, metricNames);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#getSandboxMetrics");
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
| **from** | **OffsetDateTime**| Start of time range (ISO 8601) | |
| **to** | **OffsetDateTime**| End of time range (ISO 8601) | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |
| **metricNames** | [**List&lt;String&gt;**](String.md)| Filter by metric names | [optional] |

### Return type

[**MetricsResponse**](MetricsResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Metrics time series data |  -  |

<a id="getSandboxTraceSpans"></a>
# **getSandboxTraceSpans**
> List&lt;TraceSpan&gt; getSandboxTraceSpans(sandboxId, traceId, xDaytonaOrganizationID)

Get trace spans

Retrieve all spans for a specific trace

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | ID of the sandbox
    String traceId = "traceId_example"; // String | ID of the trace
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      List<TraceSpan> result = apiInstance.getSandboxTraceSpans(sandboxId, traceId, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#getSandboxTraceSpans");
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
| **traceId** | **String**| ID of the trace | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

[**List&lt;TraceSpan&gt;**](TraceSpan.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | List of spans in the trace |  -  |

<a id="getSandboxTraces"></a>
# **getSandboxTraces**
> PaginatedTraces getSandboxTraces(sandboxId, from, to, xDaytonaOrganizationID, page, limit)

Get sandbox traces

Retrieve OTEL traces for a sandbox within a time range

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | ID of the sandbox
    OffsetDateTime from = OffsetDateTime.now(); // OffsetDateTime | Start of time range (ISO 8601)
    OffsetDateTime to = OffsetDateTime.now(); // OffsetDateTime | End of time range (ISO 8601)
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    BigDecimal page = new BigDecimal("1"); // BigDecimal | Page number (1-indexed)
    BigDecimal limit = new BigDecimal("100"); // BigDecimal | Number of items per page
    try {
      PaginatedTraces result = apiInstance.getSandboxTraces(sandboxId, from, to, xDaytonaOrganizationID, page, limit);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#getSandboxTraces");
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
| **from** | **OffsetDateTime**| Start of time range (ISO 8601) | |
| **to** | **OffsetDateTime**| End of time range (ISO 8601) | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |
| **page** | **BigDecimal**| Page number (1-indexed) | [optional] [default to 1] |
| **limit** | **BigDecimal**| Number of items per page | [optional] [default to 100] |

### Return type

[**PaginatedTraces**](PaginatedTraces.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Paginated list of trace summaries |  -  |

<a id="getSandboxesForRunner"></a>
# **getSandboxesForRunner**
> List&lt;Sandbox&gt; getSandboxesForRunner(xDaytonaOrganizationID, states, skipReconcilingSandboxes)

Get sandboxes for the authenticated runner

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    String states = "states_example"; // String | Comma-separated list of sandbox states to filter by
    Boolean skipReconcilingSandboxes = true; // Boolean | Skip sandboxes where state differs from desired state
    try {
      List<Sandbox> result = apiInstance.getSandboxesForRunner(xDaytonaOrganizationID, states, skipReconcilingSandboxes);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#getSandboxesForRunner");
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
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |
| **states** | **String**| Comma-separated list of sandbox states to filter by | [optional] |
| **skipReconcilingSandboxes** | **Boolean**| Skip sandboxes where state differs from desired state | [optional] |

### Return type

[**List&lt;Sandbox&gt;**](Sandbox.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | List of sandboxes for the authenticated runner |  -  |

<a id="getSignedPortPreviewUrl"></a>
# **getSignedPortPreviewUrl**
> SignedPortPreviewUrl getSignedPortPreviewUrl(sandboxIdOrName, port, xDaytonaOrganizationID, expiresInSeconds)

Get signed preview URL for a sandbox port

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String sandboxIdOrName = "sandboxIdOrName_example"; // String | ID or name of the sandbox
    Integer port = 56; // Integer | Port number to get signed preview URL for
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    Integer expiresInSeconds = 56; // Integer | Expiration time in seconds (default: 60 seconds)
    try {
      SignedPortPreviewUrl result = apiInstance.getSignedPortPreviewUrl(sandboxIdOrName, port, xDaytonaOrganizationID, expiresInSeconds);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#getSignedPortPreviewUrl");
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
| **sandboxIdOrName** | **String**| ID or name of the sandbox | |
| **port** | **Integer**| Port number to get signed preview URL for | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |
| **expiresInSeconds** | **Integer**| Expiration time in seconds (default: 60 seconds) | [optional] |

### Return type

[**SignedPortPreviewUrl**](SignedPortPreviewUrl.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Signed preview URL for the specified port |  -  |

<a id="getToolboxProxyUrl"></a>
# **getToolboxProxyUrl**
> ToolboxProxyUrl getToolboxProxyUrl(sandboxId, xDaytonaOrganizationID)

Get toolbox proxy URL for a sandbox

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | ID of the sandbox
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      ToolboxProxyUrl result = apiInstance.getToolboxProxyUrl(sandboxId, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#getToolboxProxyUrl");
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
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

[**ToolboxProxyUrl**](ToolboxProxyUrl.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Toolbox proxy URL for the specified sandbox |  -  |

<a id="listSandboxes"></a>
# **listSandboxes**
> List&lt;Sandbox&gt; listSandboxes(xDaytonaOrganizationID, verbose, labels, includeErroredDeleted)

List all sandboxes

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    Boolean verbose = true; // Boolean | Include verbose output
    String labels = "{\"label1\": \"value1\", \"label2\": \"value2\"}"; // String | JSON encoded labels to filter by
    Boolean includeErroredDeleted = true; // Boolean | Include errored and deleted sandboxes
    try {
      List<Sandbox> result = apiInstance.listSandboxes(xDaytonaOrganizationID, verbose, labels, includeErroredDeleted);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#listSandboxes");
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
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |
| **verbose** | **Boolean**| Include verbose output | [optional] |
| **labels** | **String**| JSON encoded labels to filter by | [optional] |
| **includeErroredDeleted** | **Boolean**| Include errored and deleted sandboxes | [optional] |

### Return type

[**List&lt;Sandbox&gt;**](Sandbox.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | List of all sandboxes |  -  |

<a id="listSandboxesPaginated"></a>
# **listSandboxesPaginated**
> PaginatedSandboxes listSandboxesPaginated(xDaytonaOrganizationID, page, limit, id, name, labels, includeErroredDeleted, states, snapshots, regions, minCpu, maxCpu, minMemoryGiB, maxMemoryGiB, minDiskGiB, maxDiskGiB, lastEventAfter, lastEventBefore, sort, order)

List all sandboxes paginated

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    BigDecimal page = new BigDecimal("1"); // BigDecimal | Page number of the results
    BigDecimal limit = new BigDecimal("100"); // BigDecimal | Number of results per page
    String id = "abc123"; // String | Filter by partial ID match
    String name = "abc123"; // String | Filter by partial name match
    String labels = "{\"label1\": \"value1\", \"label2\": \"value2\"}"; // String | JSON encoded labels to filter by
    Boolean includeErroredDeleted = false; // Boolean | Include results with errored state and deleted desired state
    List<String> states = Arrays.asList(); // List<String> | List of states to filter by
    List<String> snapshots = Arrays.asList(); // List<String> | List of snapshot names to filter by
    List<String> regions = Arrays.asList(); // List<String> | List of regions to filter by
    BigDecimal minCpu = new BigDecimal(78); // BigDecimal | Minimum CPU
    BigDecimal maxCpu = new BigDecimal(78); // BigDecimal | Maximum CPU
    BigDecimal minMemoryGiB = new BigDecimal(78); // BigDecimal | Minimum memory in GiB
    BigDecimal maxMemoryGiB = new BigDecimal(78); // BigDecimal | Maximum memory in GiB
    BigDecimal minDiskGiB = new BigDecimal(78); // BigDecimal | Minimum disk space in GiB
    BigDecimal maxDiskGiB = new BigDecimal(78); // BigDecimal | Maximum disk space in GiB
    OffsetDateTime lastEventAfter = OffsetDateTime.parse("2024-01-01T00:00Z"); // OffsetDateTime | Include items with last event after this timestamp
    OffsetDateTime lastEventBefore = OffsetDateTime.parse("2024-12-31T23:59:59Z"); // OffsetDateTime | Include items with last event before this timestamp
    String sort = "id"; // String | Field to sort by
    String order = "asc"; // String | Direction to sort by
    try {
      PaginatedSandboxes result = apiInstance.listSandboxesPaginated(xDaytonaOrganizationID, page, limit, id, name, labels, includeErroredDeleted, states, snapshots, regions, minCpu, maxCpu, minMemoryGiB, maxMemoryGiB, minDiskGiB, maxDiskGiB, lastEventAfter, lastEventBefore, sort, order);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#listSandboxesPaginated");
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
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |
| **page** | **BigDecimal**| Page number of the results | [optional] [default to 1] |
| **limit** | **BigDecimal**| Number of results per page | [optional] [default to 100] |
| **id** | **String**| Filter by partial ID match | [optional] |
| **name** | **String**| Filter by partial name match | [optional] |
| **labels** | **String**| JSON encoded labels to filter by | [optional] |
| **includeErroredDeleted** | **Boolean**| Include results with errored state and deleted desired state | [optional] [default to false] |
| **states** | [**List&lt;String&gt;**](String.md)| List of states to filter by | [optional] [enum: creating, restoring, destroying, started, stopped, starting, stopping, error, build_failed, pending_build, building_snapshot, unknown, pulling_snapshot, archived, archiving, resizing] |
| **snapshots** | [**List&lt;String&gt;**](String.md)| List of snapshot names to filter by | [optional] |
| **regions** | [**List&lt;String&gt;**](String.md)| List of regions to filter by | [optional] |
| **minCpu** | **BigDecimal**| Minimum CPU | [optional] |
| **maxCpu** | **BigDecimal**| Maximum CPU | [optional] |
| **minMemoryGiB** | **BigDecimal**| Minimum memory in GiB | [optional] |
| **maxMemoryGiB** | **BigDecimal**| Maximum memory in GiB | [optional] |
| **minDiskGiB** | **BigDecimal**| Minimum disk space in GiB | [optional] |
| **maxDiskGiB** | **BigDecimal**| Maximum disk space in GiB | [optional] |
| **lastEventAfter** | **OffsetDateTime**| Include items with last event after this timestamp | [optional] |
| **lastEventBefore** | **OffsetDateTime**| Include items with last event before this timestamp | [optional] |
| **sort** | **String**| Field to sort by | [optional] [default to createdAt] [enum: id, name, state, snapshot, region, updatedAt, createdAt] |
| **order** | **String**| Direction to sort by | [optional] [default to desc] [enum: asc, desc] |

### Return type

[**PaginatedSandboxes**](PaginatedSandboxes.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Paginated list of all sandboxes |  -  |

<a id="recoverSandbox"></a>
# **recoverSandbox**
> Sandbox recoverSandbox(sandboxIdOrName, xDaytonaOrganizationID)

Recover sandbox from error state

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String sandboxIdOrName = "sandboxIdOrName_example"; // String | ID or name of the sandbox
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      Sandbox result = apiInstance.recoverSandbox(sandboxIdOrName, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#recoverSandbox");
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
| **sandboxIdOrName** | **String**| ID or name of the sandbox | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="replaceLabels"></a>
# **replaceLabels**
> SandboxLabels replaceLabels(sandboxIdOrName, sandboxLabels, xDaytonaOrganizationID)

Replace sandbox labels

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String sandboxIdOrName = "sandboxIdOrName_example"; // String | ID or name of the sandbox
    SandboxLabels sandboxLabels = new SandboxLabels(); // SandboxLabels | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      SandboxLabels result = apiInstance.replaceLabels(sandboxIdOrName, sandboxLabels, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#replaceLabels");
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
| **sandboxIdOrName** | **String**| ID or name of the sandbox | |
| **sandboxLabels** | [**SandboxLabels**](SandboxLabels.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

[**SandboxLabels**](SandboxLabels.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Labels have been successfully replaced |  -  |

<a id="resizeSandbox"></a>
# **resizeSandbox**
> Sandbox resizeSandbox(sandboxIdOrName, resizeSandbox, xDaytonaOrganizationID)

Resize sandbox resources

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String sandboxIdOrName = "sandboxIdOrName_example"; // String | ID or name of the sandbox
    ResizeSandbox resizeSandbox = new ResizeSandbox(); // ResizeSandbox | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      Sandbox result = apiInstance.resizeSandbox(sandboxIdOrName, resizeSandbox, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#resizeSandbox");
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
| **sandboxIdOrName** | **String**| ID or name of the sandbox | |
| **resizeSandbox** | [**ResizeSandbox**](ResizeSandbox.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

[**Sandbox**](Sandbox.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Sandbox has been resized |  -  |

<a id="revokeSshAccess"></a>
# **revokeSshAccess**
> Sandbox revokeSshAccess(sandboxIdOrName, xDaytonaOrganizationID, token)

Revoke SSH access for sandbox

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String sandboxIdOrName = "sandboxIdOrName_example"; // String | ID or name of the sandbox
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    String token = "token_example"; // String | SSH access token to revoke. If not provided, all SSH access for the sandbox will be revoked.
    try {
      Sandbox result = apiInstance.revokeSshAccess(sandboxIdOrName, xDaytonaOrganizationID, token);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#revokeSshAccess");
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
| **sandboxIdOrName** | **String**| ID or name of the sandbox | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |
| **token** | **String**| SSH access token to revoke. If not provided, all SSH access for the sandbox will be revoked. | [optional] |

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
| **200** | SSH access has been revoked |  -  |

<a id="setAutoArchiveInterval"></a>
# **setAutoArchiveInterval**
> Sandbox setAutoArchiveInterval(sandboxIdOrName, interval, xDaytonaOrganizationID)

Set sandbox auto-archive interval

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String sandboxIdOrName = "sandboxIdOrName_example"; // String | ID or name of the sandbox
    BigDecimal interval = new BigDecimal(78); // BigDecimal | Auto-archive interval in minutes (0 means the maximum interval will be used)
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      Sandbox result = apiInstance.setAutoArchiveInterval(sandboxIdOrName, interval, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#setAutoArchiveInterval");
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
| **sandboxIdOrName** | **String**| ID or name of the sandbox | |
| **interval** | **BigDecimal**| Auto-archive interval in minutes (0 means the maximum interval will be used) | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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
| **200** | Auto-archive interval has been set |  -  |

<a id="setAutoDeleteInterval"></a>
# **setAutoDeleteInterval**
> Sandbox setAutoDeleteInterval(sandboxIdOrName, interval, xDaytonaOrganizationID)

Set sandbox auto-delete interval

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String sandboxIdOrName = "sandboxIdOrName_example"; // String | ID or name of the sandbox
    BigDecimal interval = new BigDecimal(78); // BigDecimal | Auto-delete interval in minutes (negative value means disabled, 0 means delete immediately upon stopping)
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      Sandbox result = apiInstance.setAutoDeleteInterval(sandboxIdOrName, interval, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#setAutoDeleteInterval");
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
| **sandboxIdOrName** | **String**| ID or name of the sandbox | |
| **interval** | **BigDecimal**| Auto-delete interval in minutes (negative value means disabled, 0 means delete immediately upon stopping) | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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
| **200** | Auto-delete interval has been set |  -  |

<a id="setAutostopInterval"></a>
# **setAutostopInterval**
> Sandbox setAutostopInterval(sandboxIdOrName, interval, xDaytonaOrganizationID)

Set sandbox auto-stop interval

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String sandboxIdOrName = "sandboxIdOrName_example"; // String | ID or name of the sandbox
    BigDecimal interval = new BigDecimal(78); // BigDecimal | Auto-stop interval in minutes (0 to disable)
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      Sandbox result = apiInstance.setAutostopInterval(sandboxIdOrName, interval, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#setAutostopInterval");
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
| **sandboxIdOrName** | **String**| ID or name of the sandbox | |
| **interval** | **BigDecimal**| Auto-stop interval in minutes (0 to disable) | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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
| **200** | Auto-stop interval has been set |  -  |

<a id="startSandbox"></a>
# **startSandbox**
> Sandbox startSandbox(sandboxIdOrName, xDaytonaOrganizationID)

Start sandbox

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String sandboxIdOrName = "sandboxIdOrName_example"; // String | ID or name of the sandbox
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      Sandbox result = apiInstance.startSandbox(sandboxIdOrName, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#startSandbox");
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
| **sandboxIdOrName** | **String**| ID or name of the sandbox | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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
| **200** | Sandbox has been started or is being restored from archived state |  -  |

<a id="stopSandbox"></a>
# **stopSandbox**
> Sandbox stopSandbox(sandboxIdOrName, xDaytonaOrganizationID, force)

Stop sandbox

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String sandboxIdOrName = "sandboxIdOrName_example"; // String | ID or name of the sandbox
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    Boolean force = true; // Boolean | Force stop the sandbox using SIGKILL instead of SIGTERM
    try {
      Sandbox result = apiInstance.stopSandbox(sandboxIdOrName, xDaytonaOrganizationID, force);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#stopSandbox");
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
| **sandboxIdOrName** | **String**| ID or name of the sandbox | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |
| **force** | **Boolean**| Force stop the sandbox using SIGKILL instead of SIGTERM | [optional] |

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
| **200** | Sandbox has been stopped |  -  |

<a id="updateLastActivity"></a>
# **updateLastActivity**
> updateLastActivity(sandboxId, xDaytonaOrganizationID)

Update sandbox last activity

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | ID of the sandbox
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.updateLastActivity(sandboxId, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#updateLastActivity");
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
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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
| **201** | Last activity has been updated |  -  |

<a id="updatePublicStatus"></a>
# **updatePublicStatus**
> Sandbox updatePublicStatus(sandboxIdOrName, isPublic, xDaytonaOrganizationID)

Update public status

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String sandboxIdOrName = "sandboxIdOrName_example"; // String | ID or name of the sandbox
    Boolean isPublic = true; // Boolean | Public status to set
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      Sandbox result = apiInstance.updatePublicStatus(sandboxIdOrName, isPublic, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#updatePublicStatus");
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
| **sandboxIdOrName** | **String**| ID or name of the sandbox | |
| **isPublic** | **Boolean**| Public status to set | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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
| **200** | Public status has been successfully updated |  -  |

<a id="updateSandboxState"></a>
# **updateSandboxState**
> updateSandboxState(sandboxId, updateSandboxStateDto, xDaytonaOrganizationID)

Update sandbox state

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | ID of the sandbox
    UpdateSandboxStateDto updateSandboxStateDto = new UpdateSandboxStateDto(); // UpdateSandboxStateDto | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.updateSandboxState(sandboxId, updateSandboxStateDto, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#updateSandboxState");
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
| **updateSandboxStateDto** | [**UpdateSandboxStateDto**](UpdateSandboxStateDto.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Sandbox state has been successfully updated |  -  |

<a id="validateSshAccess"></a>
# **validateSshAccess**
> SshAccessValidationDto validateSshAccess(token, xDaytonaOrganizationID)

Validate SSH access for sandbox

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SandboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SandboxApi apiInstance = new SandboxApi(defaultClient);
    String token = "token_example"; // String | SSH access token to validate
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      SshAccessValidationDto result = apiInstance.validateSshAccess(token, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SandboxApi#validateSshAccess");
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
| **token** | **String**| SSH access token to validate | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

[**SshAccessValidationDto**](SshAccessValidationDto.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | SSH access validation result |  -  |

