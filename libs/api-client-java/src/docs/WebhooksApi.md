# WebhooksApi

All URIs are relative to *http://localhost:3000*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**webhookControllerGetAppPortalAccess**](WebhooksApi.md#webhookControllerGetAppPortalAccess) | **POST** /webhooks/organizations/{organizationId}/app-portal-access | Get Svix Consumer App Portal access for an organization |
| [**webhookControllerGetInitializationStatus**](WebhooksApi.md#webhookControllerGetInitializationStatus) | **GET** /webhooks/organizations/{organizationId}/initialization-status | Get webhook initialization status for an organization |
| [**webhookControllerGetMessageAttempts**](WebhooksApi.md#webhookControllerGetMessageAttempts) | **GET** /webhooks/organizations/{organizationId}/messages/{messageId}/attempts | Get delivery attempts for a webhook message |
| [**webhookControllerGetStatus**](WebhooksApi.md#webhookControllerGetStatus) | **GET** /webhooks/status | Get webhook service status |
| [**webhookControllerInitializeWebhooks**](WebhooksApi.md#webhookControllerInitializeWebhooks) | **POST** /webhooks/organizations/{organizationId}/initialize | Initialize webhooks for an organization |
| [**webhookControllerSendWebhook**](WebhooksApi.md#webhookControllerSendWebhook) | **POST** /webhooks/organizations/{organizationId}/send | Send a webhook message to an organization |


<a id="webhookControllerGetAppPortalAccess"></a>
# **webhookControllerGetAppPortalAccess**
> WebhookAppPortalAccess webhookControllerGetAppPortalAccess(organizationId, xDaytonaOrganizationID)

Get Svix Consumer App Portal access for an organization

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.WebhooksApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");

    WebhooksApi apiInstance = new WebhooksApi(defaultClient);
    String organizationId = "organizationId_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      WebhookAppPortalAccess result = apiInstance.webhookControllerGetAppPortalAccess(organizationId, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling WebhooksApi#webhookControllerGetAppPortalAccess");
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
| **organizationId** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

[**WebhookAppPortalAccess**](WebhookAppPortalAccess.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | App Portal access generated successfully |  -  |

<a id="webhookControllerGetInitializationStatus"></a>
# **webhookControllerGetInitializationStatus**
> WebhookInitializationStatus webhookControllerGetInitializationStatus(organizationId, xDaytonaOrganizationID)

Get webhook initialization status for an organization

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.WebhooksApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");

    WebhooksApi apiInstance = new WebhooksApi(defaultClient);
    String organizationId = "organizationId_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      WebhookInitializationStatus result = apiInstance.webhookControllerGetInitializationStatus(organizationId, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling WebhooksApi#webhookControllerGetInitializationStatus");
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
| **organizationId** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

[**WebhookInitializationStatus**](WebhookInitializationStatus.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Webhook initialization status |  -  |
| **404** | Webhook initialization status not found |  -  |

<a id="webhookControllerGetMessageAttempts"></a>
# **webhookControllerGetMessageAttempts**
> List&lt;Object&gt; webhookControllerGetMessageAttempts(organizationId, messageId, xDaytonaOrganizationID)

Get delivery attempts for a webhook message

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.WebhooksApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");

    WebhooksApi apiInstance = new WebhooksApi(defaultClient);
    String organizationId = "organizationId_example"; // String | 
    String messageId = "messageId_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      List<Object> result = apiInstance.webhookControllerGetMessageAttempts(organizationId, messageId, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling WebhooksApi#webhookControllerGetMessageAttempts");
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
| **organizationId** | **String**|  | |
| **messageId** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

**List&lt;Object&gt;**

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | List of delivery attempts |  -  |

<a id="webhookControllerGetStatus"></a>
# **webhookControllerGetStatus**
> WebhookControllerGetStatus200Response webhookControllerGetStatus(xDaytonaOrganizationID)

Get webhook service status

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.WebhooksApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");

    WebhooksApi apiInstance = new WebhooksApi(defaultClient);
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      WebhookControllerGetStatus200Response result = apiInstance.webhookControllerGetStatus(xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling WebhooksApi#webhookControllerGetStatus");
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

[**WebhookControllerGetStatus200Response**](WebhookControllerGetStatus200Response.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Webhook service status |  -  |

<a id="webhookControllerInitializeWebhooks"></a>
# **webhookControllerInitializeWebhooks**
> webhookControllerInitializeWebhooks(organizationId, xDaytonaOrganizationID)

Initialize webhooks for an organization

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.WebhooksApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");

    WebhooksApi apiInstance = new WebhooksApi(defaultClient);
    String organizationId = "organizationId_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.webhookControllerInitializeWebhooks(organizationId, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling WebhooksApi#webhookControllerInitializeWebhooks");
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
| **organizationId** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** | Webhooks initialized successfully |  -  |
| **403** | User does not have access to this organization |  -  |
| **404** | Organization not found |  -  |

<a id="webhookControllerSendWebhook"></a>
# **webhookControllerSendWebhook**
> webhookControllerSendWebhook(organizationId, sendWebhookDto, xDaytonaOrganizationID)

Send a webhook message to an organization

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.WebhooksApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");

    WebhooksApi apiInstance = new WebhooksApi(defaultClient);
    String organizationId = "organizationId_example"; // String | 
    SendWebhookDto sendWebhookDto = new SendWebhookDto(); // SendWebhookDto | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.webhookControllerSendWebhook(organizationId, sendWebhookDto, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling WebhooksApi#webhookControllerSendWebhook");
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
| **organizationId** | **String**|  | |
| **sendWebhookDto** | [**SendWebhookDto**](SendWebhookDto.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Webhook message sent successfully |  -  |

