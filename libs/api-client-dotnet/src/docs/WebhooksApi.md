# Daytona.ApiClient.Api.WebhooksApi

All URIs are relative to *http://localhost:3000*

| Method | HTTP request | Description |
|--------|--------------|-------------|
| [**WebhookControllerGetAppPortalAccess**](WebhooksApi.md#webhookcontrollergetappportalaccess) | **POST** /webhooks/organizations/{organizationId}/app-portal-access | Get Svix Consumer App Portal access for an organization |
| [**WebhookControllerGetInitializationStatus**](WebhooksApi.md#webhookcontrollergetinitializationstatus) | **GET** /webhooks/organizations/{organizationId}/initialization-status | Get webhook initialization status for an organization |
| [**WebhookControllerGetMessageAttempts**](WebhooksApi.md#webhookcontrollergetmessageattempts) | **GET** /webhooks/organizations/{organizationId}/messages/{messageId}/attempts | Get delivery attempts for a webhook message |
| [**WebhookControllerGetStatus**](WebhooksApi.md#webhookcontrollergetstatus) | **GET** /webhooks/status | Get webhook service status |
| [**WebhookControllerInitializeWebhooks**](WebhooksApi.md#webhookcontrollerinitializewebhooks) | **POST** /webhooks/organizations/{organizationId}/initialize | Initialize webhooks for an organization |
| [**WebhookControllerSendWebhook**](WebhooksApi.md#webhookcontrollersendwebhook) | **POST** /webhooks/organizations/{organizationId}/send | Send a webhook message to an organization |

<a id="webhookcontrollergetappportalaccess"></a>
# **WebhookControllerGetAppPortalAccess**
> WebhookAppPortalAccess WebhookControllerGetAppPortalAccess (string organizationId, string? xDaytonaOrganizationID = null)

Get Svix Consumer App Portal access for an organization

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using Daytona.ApiClient.Api;
using Daytona.ApiClient.Client;
using Daytona.ApiClient.Model;

namespace Example
{
    public class WebhookControllerGetAppPortalAccessExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:3000";
            // Configure Bearer token for authorization: bearer
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new WebhooksApi(httpClient, config, httpClientHandler);
            var organizationId = "organizationId_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Get Svix Consumer App Portal access for an organization
                WebhookAppPortalAccess result = apiInstance.WebhookControllerGetAppPortalAccess(organizationId, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling WebhooksApi.WebhookControllerGetAppPortalAccess: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the WebhookControllerGetAppPortalAccessWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get Svix Consumer App Portal access for an organization
    ApiResponse<WebhookAppPortalAccess> response = apiInstance.WebhookControllerGetAppPortalAccessWithHttpInfo(organizationId, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling WebhooksApi.WebhookControllerGetAppPortalAccessWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **organizationId** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="webhookcontrollergetinitializationstatus"></a>
# **WebhookControllerGetInitializationStatus**
> WebhookInitializationStatus WebhookControllerGetInitializationStatus (string organizationId, string? xDaytonaOrganizationID = null)

Get webhook initialization status for an organization

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using Daytona.ApiClient.Api;
using Daytona.ApiClient.Client;
using Daytona.ApiClient.Model;

namespace Example
{
    public class WebhookControllerGetInitializationStatusExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:3000";
            // Configure Bearer token for authorization: bearer
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new WebhooksApi(httpClient, config, httpClientHandler);
            var organizationId = "organizationId_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Get webhook initialization status for an organization
                WebhookInitializationStatus result = apiInstance.WebhookControllerGetInitializationStatus(organizationId, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling WebhooksApi.WebhookControllerGetInitializationStatus: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the WebhookControllerGetInitializationStatusWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get webhook initialization status for an organization
    ApiResponse<WebhookInitializationStatus> response = apiInstance.WebhookControllerGetInitializationStatusWithHttpInfo(organizationId, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling WebhooksApi.WebhookControllerGetInitializationStatusWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **organizationId** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="webhookcontrollergetmessageattempts"></a>
# **WebhookControllerGetMessageAttempts**
> List&lt;Object&gt; WebhookControllerGetMessageAttempts (string organizationId, string messageId, string? xDaytonaOrganizationID = null)

Get delivery attempts for a webhook message

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using Daytona.ApiClient.Api;
using Daytona.ApiClient.Client;
using Daytona.ApiClient.Model;

namespace Example
{
    public class WebhookControllerGetMessageAttemptsExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:3000";
            // Configure Bearer token for authorization: bearer
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new WebhooksApi(httpClient, config, httpClientHandler);
            var organizationId = "organizationId_example";  // string | 
            var messageId = "messageId_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Get delivery attempts for a webhook message
                List<Object> result = apiInstance.WebhookControllerGetMessageAttempts(organizationId, messageId, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling WebhooksApi.WebhookControllerGetMessageAttempts: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the WebhookControllerGetMessageAttemptsWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get delivery attempts for a webhook message
    ApiResponse<List<Object>> response = apiInstance.WebhookControllerGetMessageAttemptsWithHttpInfo(organizationId, messageId, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling WebhooksApi.WebhookControllerGetMessageAttemptsWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **organizationId** | **string** |  |  |
| **messageId** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

**List<Object>**

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | List of delivery attempts |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="webhookcontrollergetstatus"></a>
# **WebhookControllerGetStatus**
> WebhookControllerGetStatus200Response WebhookControllerGetStatus (string? xDaytonaOrganizationID = null)

Get webhook service status

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using Daytona.ApiClient.Api;
using Daytona.ApiClient.Client;
using Daytona.ApiClient.Model;

namespace Example
{
    public class WebhookControllerGetStatusExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:3000";
            // Configure Bearer token for authorization: bearer
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new WebhooksApi(httpClient, config, httpClientHandler);
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Get webhook service status
                WebhookControllerGetStatus200Response result = apiInstance.WebhookControllerGetStatus(xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling WebhooksApi.WebhookControllerGetStatus: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the WebhookControllerGetStatusWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get webhook service status
    ApiResponse<WebhookControllerGetStatus200Response> response = apiInstance.WebhookControllerGetStatusWithHttpInfo(xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling WebhooksApi.WebhookControllerGetStatusWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="webhookcontrollerinitializewebhooks"></a>
# **WebhookControllerInitializeWebhooks**
> void WebhookControllerInitializeWebhooks (string organizationId, string? xDaytonaOrganizationID = null)

Initialize webhooks for an organization

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using Daytona.ApiClient.Api;
using Daytona.ApiClient.Client;
using Daytona.ApiClient.Model;

namespace Example
{
    public class WebhookControllerInitializeWebhooksExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:3000";
            // Configure Bearer token for authorization: bearer
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new WebhooksApi(httpClient, config, httpClientHandler);
            var organizationId = "organizationId_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Initialize webhooks for an organization
                apiInstance.WebhookControllerInitializeWebhooks(organizationId, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling WebhooksApi.WebhookControllerInitializeWebhooks: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the WebhookControllerInitializeWebhooksWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Initialize webhooks for an organization
    apiInstance.WebhookControllerInitializeWebhooksWithHttpInfo(organizationId, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling WebhooksApi.WebhookControllerInitializeWebhooksWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **organizationId** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

void (empty response body)

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="webhookcontrollersendwebhook"></a>
# **WebhookControllerSendWebhook**
> void WebhookControllerSendWebhook (string organizationId, SendWebhookDto sendWebhookDto, string? xDaytonaOrganizationID = null)

Send a webhook message to an organization

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using Daytona.ApiClient.Api;
using Daytona.ApiClient.Client;
using Daytona.ApiClient.Model;

namespace Example
{
    public class WebhookControllerSendWebhookExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:3000";
            // Configure Bearer token for authorization: bearer
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new WebhooksApi(httpClient, config, httpClientHandler);
            var organizationId = "organizationId_example";  // string | 
            var sendWebhookDto = new SendWebhookDto(); // SendWebhookDto | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Send a webhook message to an organization
                apiInstance.WebhookControllerSendWebhook(organizationId, sendWebhookDto, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling WebhooksApi.WebhookControllerSendWebhook: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the WebhookControllerSendWebhookWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Send a webhook message to an organization
    apiInstance.WebhookControllerSendWebhookWithHttpInfo(organizationId, sendWebhookDto, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling WebhooksApi.WebhookControllerSendWebhookWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **organizationId** | **string** |  |  |
| **sendWebhookDto** | [**SendWebhookDto**](SendWebhookDto.md) |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

void (empty response body)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Webhook message sent successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

