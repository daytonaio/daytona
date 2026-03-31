# Daytona.ApiClient.Api.SnapshotsApi

All URIs are relative to *http://localhost:3000*

| Method | HTTP request | Description |
|--------|--------------|-------------|
| [**ActivateSnapshot**](SnapshotsApi.md#activatesnapshot) | **POST** /snapshots/{id}/activate | Activate a snapshot |
| [**CanCleanupImage**](SnapshotsApi.md#cancleanupimage) | **GET** /snapshots/can-cleanup-image | Check if an image can be cleaned up |
| [**CreateSnapshot**](SnapshotsApi.md#createsnapshot) | **POST** /snapshots | Create a new snapshot |
| [**DeactivateSnapshot**](SnapshotsApi.md#deactivatesnapshot) | **POST** /snapshots/{id}/deactivate | Deactivate a snapshot |
| [**GetAllSnapshots**](SnapshotsApi.md#getallsnapshots) | **GET** /snapshots | List all snapshots |
| [**GetSnapshot**](SnapshotsApi.md#getsnapshot) | **GET** /snapshots/{id} | Get snapshot by ID or name |
| [**GetSnapshotBuildLogs**](SnapshotsApi.md#getsnapshotbuildlogs) | **GET** /snapshots/{id}/build-logs | Get snapshot build logs |
| [**GetSnapshotBuildLogsUrl**](SnapshotsApi.md#getsnapshotbuildlogsurl) | **GET** /snapshots/{id}/build-logs-url | Get snapshot build logs URL |
| [**RemoveSnapshot**](SnapshotsApi.md#removesnapshot) | **DELETE** /snapshots/{id} | Delete snapshot |
| [**SetSnapshotGeneralStatus**](SnapshotsApi.md#setsnapshotgeneralstatus) | **PATCH** /snapshots/{id}/general | Set snapshot general status |

<a id="activatesnapshot"></a>
# **ActivateSnapshot**
> SnapshotDto ActivateSnapshot (string id, string? xDaytonaOrganizationID = null)

Activate a snapshot

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
    public class ActivateSnapshotExample
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
            var apiInstance = new SnapshotsApi(httpClient, config, httpClientHandler);
            var id = "id_example";  // string | Snapshot ID
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Activate a snapshot
                SnapshotDto result = apiInstance.ActivateSnapshot(id, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SnapshotsApi.ActivateSnapshot: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ActivateSnapshotWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Activate a snapshot
    ApiResponse<SnapshotDto> response = apiInstance.ActivateSnapshotWithHttpInfo(id, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SnapshotsApi.ActivateSnapshotWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **id** | **string** | Snapshot ID |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="cancleanupimage"></a>
# **CanCleanupImage**
> bool CanCleanupImage (string imageName, string? xDaytonaOrganizationID = null)

Check if an image can be cleaned up

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
    public class CanCleanupImageExample
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
            var apiInstance = new SnapshotsApi(httpClient, config, httpClientHandler);
            var imageName = "imageName_example";  // string | Image name with tag to check
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Check if an image can be cleaned up
                bool result = apiInstance.CanCleanupImage(imageName, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SnapshotsApi.CanCleanupImage: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the CanCleanupImageWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Check if an image can be cleaned up
    ApiResponse<bool> response = apiInstance.CanCleanupImageWithHttpInfo(imageName, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SnapshotsApi.CanCleanupImageWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **imageName** | **string** | Image name with tag to check |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

**bool**

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Boolean indicating if image can be cleaned up |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="createsnapshot"></a>
# **CreateSnapshot**
> SnapshotDto CreateSnapshot (CreateSnapshot createSnapshot, string? xDaytonaOrganizationID = null)

Create a new snapshot

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
    public class CreateSnapshotExample
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
            var apiInstance = new SnapshotsApi(httpClient, config, httpClientHandler);
            var createSnapshot = new CreateSnapshot(); // CreateSnapshot | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Create a new snapshot
                SnapshotDto result = apiInstance.CreateSnapshot(createSnapshot, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SnapshotsApi.CreateSnapshot: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the CreateSnapshotWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Create a new snapshot
    ApiResponse<SnapshotDto> response = apiInstance.CreateSnapshotWithHttpInfo(createSnapshot, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SnapshotsApi.CreateSnapshotWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **createSnapshot** | [**CreateSnapshot**](CreateSnapshot.md) |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="deactivatesnapshot"></a>
# **DeactivateSnapshot**
> void DeactivateSnapshot (string id, string? xDaytonaOrganizationID = null)

Deactivate a snapshot

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
    public class DeactivateSnapshotExample
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
            var apiInstance = new SnapshotsApi(httpClient, config, httpClientHandler);
            var id = "id_example";  // string | Snapshot ID
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Deactivate a snapshot
                apiInstance.DeactivateSnapshot(id, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SnapshotsApi.DeactivateSnapshot: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DeactivateSnapshotWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Deactivate a snapshot
    apiInstance.DeactivateSnapshotWithHttpInfo(id, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SnapshotsApi.DeactivateSnapshotWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **id** | **string** | Snapshot ID |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

void (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | The snapshot has been successfully deactivated. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getallsnapshots"></a>
# **GetAllSnapshots**
> PaginatedSnapshots GetAllSnapshots (string? xDaytonaOrganizationID = null, decimal? page = null, decimal? limit = null, string? name = null, string? sort = null, string? order = null)

List all snapshots

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
    public class GetAllSnapshotsExample
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
            var apiInstance = new SnapshotsApi(httpClient, config, httpClientHandler);
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 
            var page = 1MD;  // decimal? | Page number of the results (optional)  (default to 1M)
            var limit = 100MD;  // decimal? | Number of results per page (optional)  (default to 100M)
            var name = abc123;  // string? | Filter by partial name match (optional) 
            var sort = "name";  // string? | Field to sort by (optional)  (default to lastUsedAt)
            var order = "asc";  // string? | Direction to sort by (optional)  (default to desc)

            try
            {
                // List all snapshots
                PaginatedSnapshots result = apiInstance.GetAllSnapshots(xDaytonaOrganizationID, page, limit, name, sort, order);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SnapshotsApi.GetAllSnapshots: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetAllSnapshotsWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // List all snapshots
    ApiResponse<PaginatedSnapshots> response = apiInstance.GetAllSnapshotsWithHttpInfo(xDaytonaOrganizationID, page, limit, name, sort, order);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SnapshotsApi.GetAllSnapshotsWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |
| **page** | **decimal?** | Page number of the results | [optional] [default to 1M] |
| **limit** | **decimal?** | Number of results per page | [optional] [default to 100M] |
| **name** | **string?** | Filter by partial name match | [optional]  |
| **sort** | **string?** | Field to sort by | [optional] [default to lastUsedAt] |
| **order** | **string?** | Direction to sort by | [optional] [default to desc] |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getsnapshot"></a>
# **GetSnapshot**
> SnapshotDto GetSnapshot (string id, string? xDaytonaOrganizationID = null)

Get snapshot by ID or name

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
    public class GetSnapshotExample
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
            var apiInstance = new SnapshotsApi(httpClient, config, httpClientHandler);
            var id = "id_example";  // string | Snapshot ID or name
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Get snapshot by ID or name
                SnapshotDto result = apiInstance.GetSnapshot(id, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SnapshotsApi.GetSnapshot: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetSnapshotWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get snapshot by ID or name
    ApiResponse<SnapshotDto> response = apiInstance.GetSnapshotWithHttpInfo(id, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SnapshotsApi.GetSnapshotWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **id** | **string** | Snapshot ID or name |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getsnapshotbuildlogs"></a>
# **GetSnapshotBuildLogs**
> void GetSnapshotBuildLogs (string id, string? xDaytonaOrganizationID = null, bool? follow = null)

Get snapshot build logs

This endpoint is deprecated. Use `getSnapshotBuildLogsUrl` instead.

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
    public class GetSnapshotBuildLogsExample
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
            var apiInstance = new SnapshotsApi(httpClient, config, httpClientHandler);
            var id = "id_example";  // string | Snapshot ID
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 
            var follow = true;  // bool? | Whether to follow the logs stream (optional) 

            try
            {
                // Get snapshot build logs
                apiInstance.GetSnapshotBuildLogs(id, xDaytonaOrganizationID, follow);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SnapshotsApi.GetSnapshotBuildLogs: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetSnapshotBuildLogsWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get snapshot build logs
    apiInstance.GetSnapshotBuildLogsWithHttpInfo(id, xDaytonaOrganizationID, follow);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SnapshotsApi.GetSnapshotBuildLogsWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **id** | **string** | Snapshot ID |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |
| **follow** | **bool?** | Whether to follow the logs stream | [optional]  |

### Return type

void (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** |  |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getsnapshotbuildlogsurl"></a>
# **GetSnapshotBuildLogsUrl**
> Url GetSnapshotBuildLogsUrl (string id, string? xDaytonaOrganizationID = null)

Get snapshot build logs URL

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
    public class GetSnapshotBuildLogsUrlExample
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
            var apiInstance = new SnapshotsApi(httpClient, config, httpClientHandler);
            var id = "id_example";  // string | Snapshot ID
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Get snapshot build logs URL
                Url result = apiInstance.GetSnapshotBuildLogsUrl(id, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SnapshotsApi.GetSnapshotBuildLogsUrl: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetSnapshotBuildLogsUrlWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get snapshot build logs URL
    ApiResponse<Url> response = apiInstance.GetSnapshotBuildLogsUrlWithHttpInfo(id, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SnapshotsApi.GetSnapshotBuildLogsUrlWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **id** | **string** | Snapshot ID |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="removesnapshot"></a>
# **RemoveSnapshot**
> void RemoveSnapshot (string id, string? xDaytonaOrganizationID = null)

Delete snapshot

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
    public class RemoveSnapshotExample
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
            var apiInstance = new SnapshotsApi(httpClient, config, httpClientHandler);
            var id = "id_example";  // string | Snapshot ID
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Delete snapshot
                apiInstance.RemoveSnapshot(id, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SnapshotsApi.RemoveSnapshot: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the RemoveSnapshotWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Delete snapshot
    apiInstance.RemoveSnapshotWithHttpInfo(id, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SnapshotsApi.RemoveSnapshotWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **id** | **string** | Snapshot ID |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

void (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Snapshot has been deleted |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="setsnapshotgeneralstatus"></a>
# **SetSnapshotGeneralStatus**
> SnapshotDto SetSnapshotGeneralStatus (string id, SetSnapshotGeneralStatusDto setSnapshotGeneralStatusDto, string? xDaytonaOrganizationID = null)

Set snapshot general status

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
    public class SetSnapshotGeneralStatusExample
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
            var apiInstance = new SnapshotsApi(httpClient, config, httpClientHandler);
            var id = "id_example";  // string | Snapshot ID
            var setSnapshotGeneralStatusDto = new SetSnapshotGeneralStatusDto(); // SetSnapshotGeneralStatusDto | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Set snapshot general status
                SnapshotDto result = apiInstance.SetSnapshotGeneralStatus(id, setSnapshotGeneralStatusDto, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SnapshotsApi.SetSnapshotGeneralStatus: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the SetSnapshotGeneralStatusWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Set snapshot general status
    ApiResponse<SnapshotDto> response = apiInstance.SetSnapshotGeneralStatusWithHttpInfo(id, setSnapshotGeneralStatusDto, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SnapshotsApi.SetSnapshotGeneralStatusWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **id** | **string** | Snapshot ID |  |
| **setSnapshotGeneralStatusDto** | [**SetSnapshotGeneralStatusDto**](SetSnapshotGeneralStatusDto.md) |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

