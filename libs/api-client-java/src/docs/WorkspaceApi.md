# WorkspaceApi

All URIs are relative to *http://localhost:3000*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**archiveWorkspaceDeprecated**](WorkspaceApi.md#archiveWorkspaceDeprecated) | **POST** /workspace/{workspaceId}/archive | [DEPRECATED] Archive workspace |
| [**createBackupWorkspaceDeprecated**](WorkspaceApi.md#createBackupWorkspaceDeprecated) | **POST** /workspace/{workspaceId}/backup | [DEPRECATED] Create workspace backup |
| [**createWorkspaceDeprecated**](WorkspaceApi.md#createWorkspaceDeprecated) | **POST** /workspace | [DEPRECATED] Create a new workspace |
| [**deleteWorkspaceDeprecated**](WorkspaceApi.md#deleteWorkspaceDeprecated) | **DELETE** /workspace/{workspaceId} | [DEPRECATED] Delete workspace |
| [**getBuildLogsWorkspaceDeprecated**](WorkspaceApi.md#getBuildLogsWorkspaceDeprecated) | **GET** /workspace/{workspaceId}/build-logs | [DEPRECATED] Get build logs |
| [**getPortPreviewUrlWorkspaceDeprecated**](WorkspaceApi.md#getPortPreviewUrlWorkspaceDeprecated) | **GET** /workspace/{workspaceId}/ports/{port}/preview-url | [DEPRECATED] Get preview URL for a workspace port |
| [**getWorkspaceDeprecated**](WorkspaceApi.md#getWorkspaceDeprecated) | **GET** /workspace/{workspaceId} | [DEPRECATED] Get workspace details |
| [**listWorkspacesDeprecated**](WorkspaceApi.md#listWorkspacesDeprecated) | **GET** /workspace | [DEPRECATED] List all workspaces |
| [**replaceLabelsWorkspaceDeprecated**](WorkspaceApi.md#replaceLabelsWorkspaceDeprecated) | **PUT** /workspace/{workspaceId}/labels | [DEPRECATED] Replace workspace labels |
| [**setAutoArchiveIntervalWorkspaceDeprecated**](WorkspaceApi.md#setAutoArchiveIntervalWorkspaceDeprecated) | **POST** /workspace/{workspaceId}/autoarchive/{interval} | [DEPRECATED] Set workspace auto-archive interval |
| [**setAutostopIntervalWorkspaceDeprecated**](WorkspaceApi.md#setAutostopIntervalWorkspaceDeprecated) | **POST** /workspace/{workspaceId}/autostop/{interval} | [DEPRECATED] Set workspace auto-stop interval |
| [**startWorkspaceDeprecated**](WorkspaceApi.md#startWorkspaceDeprecated) | **POST** /workspace/{workspaceId}/start | [DEPRECATED] Start workspace |
| [**stopWorkspaceDeprecated**](WorkspaceApi.md#stopWorkspaceDeprecated) | **POST** /workspace/{workspaceId}/stop | [DEPRECATED] Stop workspace |
| [**updatePublicStatusWorkspaceDeprecated**](WorkspaceApi.md#updatePublicStatusWorkspaceDeprecated) | **POST** /workspace/{workspaceId}/public/{isPublic} | [DEPRECATED] Update public status |


<a id="archiveWorkspaceDeprecated"></a>
# **archiveWorkspaceDeprecated**
> archiveWorkspaceDeprecated(workspaceId, xDaytonaOrganizationID)

[DEPRECATED] Archive workspace

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.WorkspaceApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    WorkspaceApi apiInstance = new WorkspaceApi(defaultClient);
    String workspaceId = "workspaceId_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.archiveWorkspaceDeprecated(workspaceId, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling WorkspaceApi#archiveWorkspaceDeprecated");
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
| **workspaceId** | **String**|  | |
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
| **200** | Workspace has been archived |  -  |

<a id="createBackupWorkspaceDeprecated"></a>
# **createBackupWorkspaceDeprecated**
> Workspace createBackupWorkspaceDeprecated(workspaceId, xDaytonaOrganizationID)

[DEPRECATED] Create workspace backup

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.WorkspaceApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    WorkspaceApi apiInstance = new WorkspaceApi(defaultClient);
    String workspaceId = "workspaceId_example"; // String | ID of the workspace
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      Workspace result = apiInstance.createBackupWorkspaceDeprecated(workspaceId, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling WorkspaceApi#createBackupWorkspaceDeprecated");
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
| **workspaceId** | **String**| ID of the workspace | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

[**Workspace**](Workspace.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Workspace backup has been initiated |  -  |

<a id="createWorkspaceDeprecated"></a>
# **createWorkspaceDeprecated**
> Workspace createWorkspaceDeprecated(createWorkspace, xDaytonaOrganizationID)

[DEPRECATED] Create a new workspace

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.WorkspaceApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    WorkspaceApi apiInstance = new WorkspaceApi(defaultClient);
    CreateWorkspace createWorkspace = new CreateWorkspace(); // CreateWorkspace | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      Workspace result = apiInstance.createWorkspaceDeprecated(createWorkspace, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling WorkspaceApi#createWorkspaceDeprecated");
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
| **createWorkspace** | [**CreateWorkspace**](CreateWorkspace.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

[**Workspace**](Workspace.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The workspace has been successfully created. |  -  |

<a id="deleteWorkspaceDeprecated"></a>
# **deleteWorkspaceDeprecated**
> deleteWorkspaceDeprecated(workspaceId, force, xDaytonaOrganizationID)

[DEPRECATED] Delete workspace

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.WorkspaceApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    WorkspaceApi apiInstance = new WorkspaceApi(defaultClient);
    String workspaceId = "workspaceId_example"; // String | ID of the workspace
    Boolean force = true; // Boolean | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.deleteWorkspaceDeprecated(workspaceId, force, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling WorkspaceApi#deleteWorkspaceDeprecated");
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
| **workspaceId** | **String**| ID of the workspace | |
| **force** | **Boolean**|  | |
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
| **200** | Workspace has been deleted |  -  |

<a id="getBuildLogsWorkspaceDeprecated"></a>
# **getBuildLogsWorkspaceDeprecated**
> getBuildLogsWorkspaceDeprecated(workspaceId, xDaytonaOrganizationID, follow)

[DEPRECATED] Get build logs

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.WorkspaceApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    WorkspaceApi apiInstance = new WorkspaceApi(defaultClient);
    String workspaceId = "workspaceId_example"; // String | ID of the workspace
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    Boolean follow = true; // Boolean | Whether to follow the logs stream
    try {
      apiInstance.getBuildLogsWorkspaceDeprecated(workspaceId, xDaytonaOrganizationID, follow);
    } catch (ApiException e) {
      System.err.println("Exception when calling WorkspaceApi#getBuildLogsWorkspaceDeprecated");
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
| **workspaceId** | **String**| ID of the workspace | |
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

<a id="getPortPreviewUrlWorkspaceDeprecated"></a>
# **getPortPreviewUrlWorkspaceDeprecated**
> WorkspacePortPreviewUrl getPortPreviewUrlWorkspaceDeprecated(workspaceId, port, xDaytonaOrganizationID)

[DEPRECATED] Get preview URL for a workspace port

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.WorkspaceApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    WorkspaceApi apiInstance = new WorkspaceApi(defaultClient);
    String workspaceId = "workspaceId_example"; // String | ID of the workspace
    BigDecimal port = new BigDecimal(78); // BigDecimal | Port number to get preview URL for
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      WorkspacePortPreviewUrl result = apiInstance.getPortPreviewUrlWorkspaceDeprecated(workspaceId, port, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling WorkspaceApi#getPortPreviewUrlWorkspaceDeprecated");
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
| **workspaceId** | **String**| ID of the workspace | |
| **port** | **BigDecimal**| Port number to get preview URL for | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

[**WorkspacePortPreviewUrl**](WorkspacePortPreviewUrl.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Preview URL for the specified port |  -  |

<a id="getWorkspaceDeprecated"></a>
# **getWorkspaceDeprecated**
> Workspace getWorkspaceDeprecated(workspaceId, xDaytonaOrganizationID, verbose)

[DEPRECATED] Get workspace details

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.WorkspaceApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    WorkspaceApi apiInstance = new WorkspaceApi(defaultClient);
    String workspaceId = "workspaceId_example"; // String | ID of the workspace
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    Boolean verbose = true; // Boolean | Include verbose output
    try {
      Workspace result = apiInstance.getWorkspaceDeprecated(workspaceId, xDaytonaOrganizationID, verbose);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling WorkspaceApi#getWorkspaceDeprecated");
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
| **workspaceId** | **String**| ID of the workspace | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |
| **verbose** | **Boolean**| Include verbose output | [optional] |

### Return type

[**Workspace**](Workspace.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Workspace details |  -  |

<a id="listWorkspacesDeprecated"></a>
# **listWorkspacesDeprecated**
> List&lt;Workspace&gt; listWorkspacesDeprecated(xDaytonaOrganizationID, verbose, labels)

[DEPRECATED] List all workspaces

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.WorkspaceApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    WorkspaceApi apiInstance = new WorkspaceApi(defaultClient);
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    Boolean verbose = true; // Boolean | Include verbose output
    String labels = "{\"label1\": \"value1\", \"label2\": \"value2\"}"; // String | JSON encoded labels to filter by
    try {
      List<Workspace> result = apiInstance.listWorkspacesDeprecated(xDaytonaOrganizationID, verbose, labels);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling WorkspaceApi#listWorkspacesDeprecated");
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

### Return type

[**List&lt;Workspace&gt;**](Workspace.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | List of all workspacees |  -  |

<a id="replaceLabelsWorkspaceDeprecated"></a>
# **replaceLabelsWorkspaceDeprecated**
> SandboxLabels replaceLabelsWorkspaceDeprecated(workspaceId, sandboxLabels, xDaytonaOrganizationID)

[DEPRECATED] Replace workspace labels

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.WorkspaceApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    WorkspaceApi apiInstance = new WorkspaceApi(defaultClient);
    String workspaceId = "workspaceId_example"; // String | ID of the workspace
    SandboxLabels sandboxLabels = new SandboxLabels(); // SandboxLabels | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      SandboxLabels result = apiInstance.replaceLabelsWorkspaceDeprecated(workspaceId, sandboxLabels, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling WorkspaceApi#replaceLabelsWorkspaceDeprecated");
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
| **workspaceId** | **String**| ID of the workspace | |
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

<a id="setAutoArchiveIntervalWorkspaceDeprecated"></a>
# **setAutoArchiveIntervalWorkspaceDeprecated**
> setAutoArchiveIntervalWorkspaceDeprecated(workspaceId, interval, xDaytonaOrganizationID)

[DEPRECATED] Set workspace auto-archive interval

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.WorkspaceApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    WorkspaceApi apiInstance = new WorkspaceApi(defaultClient);
    String workspaceId = "workspaceId_example"; // String | ID of the workspace
    BigDecimal interval = new BigDecimal(78); // BigDecimal | Auto-archive interval in minutes (0 means the maximum interval will be used)
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.setAutoArchiveIntervalWorkspaceDeprecated(workspaceId, interval, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling WorkspaceApi#setAutoArchiveIntervalWorkspaceDeprecated");
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
| **workspaceId** | **String**| ID of the workspace | |
| **interval** | **BigDecimal**| Auto-archive interval in minutes (0 means the maximum interval will be used) | |
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
| **200** | Auto-archive interval has been set |  -  |

<a id="setAutostopIntervalWorkspaceDeprecated"></a>
# **setAutostopIntervalWorkspaceDeprecated**
> setAutostopIntervalWorkspaceDeprecated(workspaceId, interval, xDaytonaOrganizationID)

[DEPRECATED] Set workspace auto-stop interval

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.WorkspaceApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    WorkspaceApi apiInstance = new WorkspaceApi(defaultClient);
    String workspaceId = "workspaceId_example"; // String | ID of the workspace
    BigDecimal interval = new BigDecimal(78); // BigDecimal | Auto-stop interval in minutes (0 to disable)
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.setAutostopIntervalWorkspaceDeprecated(workspaceId, interval, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling WorkspaceApi#setAutostopIntervalWorkspaceDeprecated");
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
| **workspaceId** | **String**| ID of the workspace | |
| **interval** | **BigDecimal**| Auto-stop interval in minutes (0 to disable) | |
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
| **200** | Auto-stop interval has been set |  -  |

<a id="startWorkspaceDeprecated"></a>
# **startWorkspaceDeprecated**
> startWorkspaceDeprecated(workspaceId, xDaytonaOrganizationID)

[DEPRECATED] Start workspace

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.WorkspaceApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    WorkspaceApi apiInstance = new WorkspaceApi(defaultClient);
    String workspaceId = "workspaceId_example"; // String | ID of the workspace
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.startWorkspaceDeprecated(workspaceId, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling WorkspaceApi#startWorkspaceDeprecated");
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
| **workspaceId** | **String**| ID of the workspace | |
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
| **200** | Workspace has been started |  -  |

<a id="stopWorkspaceDeprecated"></a>
# **stopWorkspaceDeprecated**
> stopWorkspaceDeprecated(workspaceId, xDaytonaOrganizationID)

[DEPRECATED] Stop workspace

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.WorkspaceApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    WorkspaceApi apiInstance = new WorkspaceApi(defaultClient);
    String workspaceId = "workspaceId_example"; // String | ID of the workspace
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.stopWorkspaceDeprecated(workspaceId, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling WorkspaceApi#stopWorkspaceDeprecated");
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
| **workspaceId** | **String**| ID of the workspace | |
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
| **200** | Workspace has been stopped |  -  |

<a id="updatePublicStatusWorkspaceDeprecated"></a>
# **updatePublicStatusWorkspaceDeprecated**
> updatePublicStatusWorkspaceDeprecated(workspaceId, isPublic, xDaytonaOrganizationID)

[DEPRECATED] Update public status

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.WorkspaceApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    WorkspaceApi apiInstance = new WorkspaceApi(defaultClient);
    String workspaceId = "workspaceId_example"; // String | ID of the workspace
    Boolean isPublic = true; // Boolean | Public status to set
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.updatePublicStatusWorkspaceDeprecated(workspaceId, isPublic, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling WorkspaceApi#updatePublicStatusWorkspaceDeprecated");
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
| **workspaceId** | **String**| ID of the workspace | |
| **isPublic** | **Boolean**| Public status to set | |
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
| **201** |  |  -  |

