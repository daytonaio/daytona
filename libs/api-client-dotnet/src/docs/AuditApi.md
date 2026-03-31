# Daytona.ApiClient.Api.AuditApi

All URIs are relative to *http://localhost:3000*

| Method | HTTP request | Description |
|--------|--------------|-------------|
| [**GetAllAuditLogs**](AuditApi.md#getallauditlogs) | **GET** /audit | Get all audit logs |
| [**GetOrganizationAuditLogs**](AuditApi.md#getorganizationauditlogs) | **GET** /audit/organizations/{organizationId} | Get audit logs for organization |

<a id="getallauditlogs"></a>
# **GetAllAuditLogs**
> PaginatedAuditLogs GetAllAuditLogs (decimal? page = null, decimal? limit = null, DateTime? from = null, DateTime? to = null, string? nextToken = null)

Get all audit logs

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
    public class GetAllAuditLogsExample
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
            var apiInstance = new AuditApi(httpClient, config, httpClientHandler);
            var page = 1MD;  // decimal? | Page number of the results (optional)  (default to 1M)
            var limit = 100MD;  // decimal? | Number of results per page (optional)  (default to 100M)
            var from = DateTime.Parse("2013-10-20T19:20:30+01:00");  // DateTime? | From date (ISO 8601 format) (optional) 
            var to = DateTime.Parse("2013-10-20T19:20:30+01:00");  // DateTime? | To date (ISO 8601 format) (optional) 
            var nextToken = "nextToken_example";  // string? | Token for cursor-based pagination. When provided, takes precedence over page parameter. (optional) 

            try
            {
                // Get all audit logs
                PaginatedAuditLogs result = apiInstance.GetAllAuditLogs(page, limit, from, to, nextToken);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling AuditApi.GetAllAuditLogs: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetAllAuditLogsWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get all audit logs
    ApiResponse<PaginatedAuditLogs> response = apiInstance.GetAllAuditLogsWithHttpInfo(page, limit, from, to, nextToken);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling AuditApi.GetAllAuditLogsWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **page** | **decimal?** | Page number of the results | [optional] [default to 1M] |
| **limit** | **decimal?** | Number of results per page | [optional] [default to 100M] |
| **from** | **DateTime?** | From date (ISO 8601 format) | [optional]  |
| **to** | **DateTime?** | To date (ISO 8601 format) | [optional]  |
| **nextToken** | **string?** | Token for cursor-based pagination. When provided, takes precedence over page parameter. | [optional]  |

### Return type

[**PaginatedAuditLogs**](PaginatedAuditLogs.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Paginated list of all audit logs |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getorganizationauditlogs"></a>
# **GetOrganizationAuditLogs**
> PaginatedAuditLogs GetOrganizationAuditLogs (string organizationId, decimal? page = null, decimal? limit = null, DateTime? from = null, DateTime? to = null, string? nextToken = null)

Get audit logs for organization

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
    public class GetOrganizationAuditLogsExample
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
            var apiInstance = new AuditApi(httpClient, config, httpClientHandler);
            var organizationId = "organizationId_example";  // string | Organization ID
            var page = 1MD;  // decimal? | Page number of the results (optional)  (default to 1M)
            var limit = 100MD;  // decimal? | Number of results per page (optional)  (default to 100M)
            var from = DateTime.Parse("2013-10-20T19:20:30+01:00");  // DateTime? | From date (ISO 8601 format) (optional) 
            var to = DateTime.Parse("2013-10-20T19:20:30+01:00");  // DateTime? | To date (ISO 8601 format) (optional) 
            var nextToken = "nextToken_example";  // string? | Token for cursor-based pagination. When provided, takes precedence over page parameter. (optional) 

            try
            {
                // Get audit logs for organization
                PaginatedAuditLogs result = apiInstance.GetOrganizationAuditLogs(organizationId, page, limit, from, to, nextToken);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling AuditApi.GetOrganizationAuditLogs: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetOrganizationAuditLogsWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get audit logs for organization
    ApiResponse<PaginatedAuditLogs> response = apiInstance.GetOrganizationAuditLogsWithHttpInfo(organizationId, page, limit, from, to, nextToken);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling AuditApi.GetOrganizationAuditLogsWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **organizationId** | **string** | Organization ID |  |
| **page** | **decimal?** | Page number of the results | [optional] [default to 1M] |
| **limit** | **decimal?** | Number of results per page | [optional] [default to 100M] |
| **from** | **DateTime?** | From date (ISO 8601 format) | [optional]  |
| **to** | **DateTime?** | To date (ISO 8601 format) | [optional]  |
| **nextToken** | **string?** | Token for cursor-based pagination. When provided, takes precedence over page parameter. | [optional]  |

### Return type

[**PaginatedAuditLogs**](PaginatedAuditLogs.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Paginated list of organization audit logs |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

