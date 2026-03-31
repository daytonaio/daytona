# SnapshotsApi

All URIs are relative to *http://localhost:3000*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**activateSnapshot**](SnapshotsApi.md#activateSnapshot) | **POST** /snapshots/{id}/activate | Activate a snapshot |
| [**canCleanupImage**](SnapshotsApi.md#canCleanupImage) | **GET** /snapshots/can-cleanup-image | Check if an image can be cleaned up |
| [**createSnapshot**](SnapshotsApi.md#createSnapshot) | **POST** /snapshots | Create a new snapshot |
| [**deactivateSnapshot**](SnapshotsApi.md#deactivateSnapshot) | **POST** /snapshots/{id}/deactivate | Deactivate a snapshot |
| [**getAllSnapshots**](SnapshotsApi.md#getAllSnapshots) | **GET** /snapshots | List all snapshots |
| [**getSnapshot**](SnapshotsApi.md#getSnapshot) | **GET** /snapshots/{id} | Get snapshot by ID or name |
| [**getSnapshotBuildLogs**](SnapshotsApi.md#getSnapshotBuildLogs) | **GET** /snapshots/{id}/build-logs | Get snapshot build logs |
| [**getSnapshotBuildLogsUrl**](SnapshotsApi.md#getSnapshotBuildLogsUrl) | **GET** /snapshots/{id}/build-logs-url | Get snapshot build logs URL |
| [**removeSnapshot**](SnapshotsApi.md#removeSnapshot) | **DELETE** /snapshots/{id} | Delete snapshot |
| [**setSnapshotGeneralStatus**](SnapshotsApi.md#setSnapshotGeneralStatus) | **PATCH** /snapshots/{id}/general | Set snapshot general status |


<a id="activateSnapshot"></a>
# **activateSnapshot**
> SnapshotDto activateSnapshot(id, xDaytonaOrganizationID)

Activate a snapshot

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SnapshotsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SnapshotsApi apiInstance = new SnapshotsApi(defaultClient);
    String id = "id_example"; // String | Snapshot ID
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      SnapshotDto result = apiInstance.activateSnapshot(id, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SnapshotsApi#activateSnapshot");
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
| **id** | **String**| Snapshot ID | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

[**SnapshotDto**](SnapshotDto.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The snapshot has been successfully activated. |  -  |
| **400** | Bad request - Snapshot is already active, not in inactive state, or has associated snapshot runners |  -  |
| **404** | Snapshot not found |  -  |

<a id="canCleanupImage"></a>
# **canCleanupImage**
> Boolean canCleanupImage(imageName, xDaytonaOrganizationID)

Check if an image can be cleaned up

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SnapshotsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SnapshotsApi apiInstance = new SnapshotsApi(defaultClient);
    String imageName = "imageName_example"; // String | Image name with tag to check
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      Boolean result = apiInstance.canCleanupImage(imageName, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SnapshotsApi#canCleanupImage");
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
| **imageName** | **String**| Image name with tag to check | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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
| **200** | Boolean indicating if image can be cleaned up |  -  |

<a id="createSnapshot"></a>
# **createSnapshot**
> SnapshotDto createSnapshot(createSnapshot, xDaytonaOrganizationID)

Create a new snapshot

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SnapshotsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SnapshotsApi apiInstance = new SnapshotsApi(defaultClient);
    CreateSnapshot createSnapshot = new CreateSnapshot(); // CreateSnapshot | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      SnapshotDto result = apiInstance.createSnapshot(createSnapshot, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SnapshotsApi#createSnapshot");
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
| **createSnapshot** | [**CreateSnapshot**](CreateSnapshot.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

[**SnapshotDto**](SnapshotDto.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The snapshot has been successfully created. |  -  |
| **400** | Bad request - Snapshots with tag \&quot;:latest\&quot; are not allowed |  -  |

<a id="deactivateSnapshot"></a>
# **deactivateSnapshot**
> deactivateSnapshot(id, xDaytonaOrganizationID)

Deactivate a snapshot

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SnapshotsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SnapshotsApi apiInstance = new SnapshotsApi(defaultClient);
    String id = "id_example"; // String | Snapshot ID
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.deactivateSnapshot(id, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling SnapshotsApi#deactivateSnapshot");
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
| **id** | **String**| Snapshot ID | |
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
| **204** | The snapshot has been successfully deactivated. |  -  |

<a id="getAllSnapshots"></a>
# **getAllSnapshots**
> PaginatedSnapshots getAllSnapshots(xDaytonaOrganizationID, page, limit, name, sort, order)

List all snapshots

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SnapshotsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SnapshotsApi apiInstance = new SnapshotsApi(defaultClient);
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    BigDecimal page = new BigDecimal("1"); // BigDecimal | Page number of the results
    BigDecimal limit = new BigDecimal("100"); // BigDecimal | Number of results per page
    String name = "abc123"; // String | Filter by partial name match
    String sort = "name"; // String | Field to sort by
    String order = "asc"; // String | Direction to sort by
    try {
      PaginatedSnapshots result = apiInstance.getAllSnapshots(xDaytonaOrganizationID, page, limit, name, sort, order);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SnapshotsApi#getAllSnapshots");
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
| **name** | **String**| Filter by partial name match | [optional] |
| **sort** | **String**| Field to sort by | [optional] [default to lastUsedAt] [enum: name, state, lastUsedAt, createdAt] |
| **order** | **String**| Direction to sort by | [optional] [default to desc] [enum: asc, desc] |

### Return type

[**PaginatedSnapshots**](PaginatedSnapshots.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Paginated list of all snapshots |  -  |

<a id="getSnapshot"></a>
# **getSnapshot**
> SnapshotDto getSnapshot(id, xDaytonaOrganizationID)

Get snapshot by ID or name

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SnapshotsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SnapshotsApi apiInstance = new SnapshotsApi(defaultClient);
    String id = "id_example"; // String | Snapshot ID or name
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      SnapshotDto result = apiInstance.getSnapshot(id, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SnapshotsApi#getSnapshot");
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
| **id** | **String**| Snapshot ID or name | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

[**SnapshotDto**](SnapshotDto.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The snapshot |  -  |
| **404** | Snapshot not found |  -  |

<a id="getSnapshotBuildLogs"></a>
# **getSnapshotBuildLogs**
> getSnapshotBuildLogs(id, xDaytonaOrganizationID, follow)

Get snapshot build logs

This endpoint is deprecated. Use &#x60;getSnapshotBuildLogsUrl&#x60; instead.

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SnapshotsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SnapshotsApi apiInstance = new SnapshotsApi(defaultClient);
    String id = "id_example"; // String | Snapshot ID
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    Boolean follow = true; // Boolean | Whether to follow the logs stream
    try {
      apiInstance.getSnapshotBuildLogs(id, xDaytonaOrganizationID, follow);
    } catch (ApiException e) {
      System.err.println("Exception when calling SnapshotsApi#getSnapshotBuildLogs");
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
| **id** | **String**| Snapshot ID | |
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
| **200** |  |  -  |

<a id="getSnapshotBuildLogsUrl"></a>
# **getSnapshotBuildLogsUrl**
> Url getSnapshotBuildLogsUrl(id, xDaytonaOrganizationID)

Get snapshot build logs URL

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SnapshotsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SnapshotsApi apiInstance = new SnapshotsApi(defaultClient);
    String id = "id_example"; // String | Snapshot ID
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      Url result = apiInstance.getSnapshotBuildLogsUrl(id, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SnapshotsApi#getSnapshotBuildLogsUrl");
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
| **id** | **String**| Snapshot ID | |
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
| **200** | The snapshot build logs URL |  -  |

<a id="removeSnapshot"></a>
# **removeSnapshot**
> removeSnapshot(id, xDaytonaOrganizationID)

Delete snapshot

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SnapshotsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SnapshotsApi apiInstance = new SnapshotsApi(defaultClient);
    String id = "id_example"; // String | Snapshot ID
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.removeSnapshot(id, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling SnapshotsApi#removeSnapshot");
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
| **id** | **String**| Snapshot ID | |
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
| **200** | Snapshot has been deleted |  -  |

<a id="setSnapshotGeneralStatus"></a>
# **setSnapshotGeneralStatus**
> SnapshotDto setSnapshotGeneralStatus(id, setSnapshotGeneralStatusDto, xDaytonaOrganizationID)

Set snapshot general status

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.SnapshotsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    SnapshotsApi apiInstance = new SnapshotsApi(defaultClient);
    String id = "id_example"; // String | Snapshot ID
    SetSnapshotGeneralStatusDto setSnapshotGeneralStatusDto = new SetSnapshotGeneralStatusDto(); // SetSnapshotGeneralStatusDto | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      SnapshotDto result = apiInstance.setSnapshotGeneralStatus(id, setSnapshotGeneralStatusDto, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling SnapshotsApi#setSnapshotGeneralStatus");
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
| **id** | **String**| Snapshot ID | |
| **setSnapshotGeneralStatusDto** | [**SetSnapshotGeneralStatusDto**](SetSnapshotGeneralStatusDto.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

[**SnapshotDto**](SnapshotDto.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Snapshot general status has been set |  -  |

