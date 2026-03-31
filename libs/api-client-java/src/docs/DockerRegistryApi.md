# DockerRegistryApi

All URIs are relative to *http://localhost:3000*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**createRegistry**](DockerRegistryApi.md#createRegistry) | **POST** /docker-registry | Create registry |
| [**deleteRegistry**](DockerRegistryApi.md#deleteRegistry) | **DELETE** /docker-registry/{id} | Delete registry |
| [**getRegistry**](DockerRegistryApi.md#getRegistry) | **GET** /docker-registry/{id} | Get registry |
| [**getTransientPushAccess**](DockerRegistryApi.md#getTransientPushAccess) | **GET** /docker-registry/registry-push-access | Get temporary registry access for pushing snapshots |
| [**listRegistries**](DockerRegistryApi.md#listRegistries) | **GET** /docker-registry | List registries |
| [**setDefaultRegistry**](DockerRegistryApi.md#setDefaultRegistry) | **POST** /docker-registry/{id}/set-default | Set default registry |
| [**updateRegistry**](DockerRegistryApi.md#updateRegistry) | **PATCH** /docker-registry/{id} | Update registry |


<a id="createRegistry"></a>
# **createRegistry**
> DockerRegistry createRegistry(createDockerRegistry, xDaytonaOrganizationID)

Create registry

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.DockerRegistryApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    DockerRegistryApi apiInstance = new DockerRegistryApi(defaultClient);
    CreateDockerRegistry createDockerRegistry = new CreateDockerRegistry(); // CreateDockerRegistry | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      DockerRegistry result = apiInstance.createRegistry(createDockerRegistry, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DockerRegistryApi#createRegistry");
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
| **createDockerRegistry** | [**CreateDockerRegistry**](CreateDockerRegistry.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

[**DockerRegistry**](DockerRegistry.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** | The docker registry has been successfully created. |  -  |

<a id="deleteRegistry"></a>
# **deleteRegistry**
> deleteRegistry(id, xDaytonaOrganizationID)

Delete registry

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.DockerRegistryApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    DockerRegistryApi apiInstance = new DockerRegistryApi(defaultClient);
    String id = "id_example"; // String | ID of the docker registry
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.deleteRegistry(id, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling DockerRegistryApi#deleteRegistry");
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
| **id** | **String**| ID of the docker registry | |
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
| **204** | The docker registry has been successfully deleted. |  -  |

<a id="getRegistry"></a>
# **getRegistry**
> DockerRegistry getRegistry(id, xDaytonaOrganizationID)

Get registry

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.DockerRegistryApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    DockerRegistryApi apiInstance = new DockerRegistryApi(defaultClient);
    String id = "id_example"; // String | ID of the docker registry
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      DockerRegistry result = apiInstance.getRegistry(id, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DockerRegistryApi#getRegistry");
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
| **id** | **String**| ID of the docker registry | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

[**DockerRegistry**](DockerRegistry.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The docker registry |  -  |

<a id="getTransientPushAccess"></a>
# **getTransientPushAccess**
> RegistryPushAccessDto getTransientPushAccess(xDaytonaOrganizationID, regionId)

Get temporary registry access for pushing snapshots

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.DockerRegistryApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    DockerRegistryApi apiInstance = new DockerRegistryApi(defaultClient);
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    String regionId = "regionId_example"; // String | ID of the region where the snapshot will be available (defaults to organization default region)
    try {
      RegistryPushAccessDto result = apiInstance.getTransientPushAccess(xDaytonaOrganizationID, regionId);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DockerRegistryApi#getTransientPushAccess");
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
| **regionId** | **String**| ID of the region where the snapshot will be available (defaults to organization default region) | [optional] |

### Return type

[**RegistryPushAccessDto**](RegistryPushAccessDto.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Temporary registry access has been generated |  -  |

<a id="listRegistries"></a>
# **listRegistries**
> List&lt;DockerRegistry&gt; listRegistries(xDaytonaOrganizationID)

List registries

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.DockerRegistryApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    DockerRegistryApi apiInstance = new DockerRegistryApi(defaultClient);
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      List<DockerRegistry> result = apiInstance.listRegistries(xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DockerRegistryApi#listRegistries");
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

### Return type

[**List&lt;DockerRegistry&gt;**](DockerRegistry.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | List of all docker registries |  -  |

<a id="setDefaultRegistry"></a>
# **setDefaultRegistry**
> DockerRegistry setDefaultRegistry(id, xDaytonaOrganizationID)

Set default registry

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.DockerRegistryApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    DockerRegistryApi apiInstance = new DockerRegistryApi(defaultClient);
    String id = "id_example"; // String | ID of the docker registry
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      DockerRegistry result = apiInstance.setDefaultRegistry(id, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DockerRegistryApi#setDefaultRegistry");
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
| **id** | **String**| ID of the docker registry | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

[**DockerRegistry**](DockerRegistry.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The docker registry has been set as default. |  -  |

<a id="updateRegistry"></a>
# **updateRegistry**
> DockerRegistry updateRegistry(id, updateDockerRegistry, xDaytonaOrganizationID)

Update registry

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.DockerRegistryApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    DockerRegistryApi apiInstance = new DockerRegistryApi(defaultClient);
    String id = "id_example"; // String | ID of the docker registry
    UpdateDockerRegistry updateDockerRegistry = new UpdateDockerRegistry(); // UpdateDockerRegistry | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      DockerRegistry result = apiInstance.updateRegistry(id, updateDockerRegistry, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DockerRegistryApi#updateRegistry");
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
| **id** | **String**| ID of the docker registry | |
| **updateDockerRegistry** | [**UpdateDockerRegistry**](UpdateDockerRegistry.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

[**DockerRegistry**](DockerRegistry.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The docker registry has been successfully updated. |  -  |

