# Daytona.ApiClient.Api.WorkspaceApi

All URIs are relative to *http://localhost:3000*

| Method | HTTP request | Description |
|--------|--------------|-------------|
| [**ArchiveWorkspaceDeprecated**](WorkspaceApi.md#archiveworkspacedeprecated) | **POST** /workspace/{workspaceId}/archive | [DEPRECATED] Archive workspace |
| [**CreateBackupWorkspaceDeprecated**](WorkspaceApi.md#createbackupworkspacedeprecated) | **POST** /workspace/{workspaceId}/backup | [DEPRECATED] Create workspace backup |
| [**CreateWorkspaceDeprecated**](WorkspaceApi.md#createworkspacedeprecated) | **POST** /workspace | [DEPRECATED] Create a new workspace |
| [**DeleteWorkspaceDeprecated**](WorkspaceApi.md#deleteworkspacedeprecated) | **DELETE** /workspace/{workspaceId} | [DEPRECATED] Delete workspace |
| [**GetBuildLogsWorkspaceDeprecated**](WorkspaceApi.md#getbuildlogsworkspacedeprecated) | **GET** /workspace/{workspaceId}/build-logs | [DEPRECATED] Get build logs |
| [**GetPortPreviewUrlWorkspaceDeprecated**](WorkspaceApi.md#getportpreviewurlworkspacedeprecated) | **GET** /workspace/{workspaceId}/ports/{port}/preview-url | [DEPRECATED] Get preview URL for a workspace port |
| [**GetWorkspaceDeprecated**](WorkspaceApi.md#getworkspacedeprecated) | **GET** /workspace/{workspaceId} | [DEPRECATED] Get workspace details |
| [**ListWorkspacesDeprecated**](WorkspaceApi.md#listworkspacesdeprecated) | **GET** /workspace | [DEPRECATED] List all workspaces |
| [**ReplaceLabelsWorkspaceDeprecated**](WorkspaceApi.md#replacelabelsworkspacedeprecated) | **PUT** /workspace/{workspaceId}/labels | [DEPRECATED] Replace workspace labels |
| [**SetAutoArchiveIntervalWorkspaceDeprecated**](WorkspaceApi.md#setautoarchiveintervalworkspacedeprecated) | **POST** /workspace/{workspaceId}/autoarchive/{interval} | [DEPRECATED] Set workspace auto-archive interval |
| [**SetAutostopIntervalWorkspaceDeprecated**](WorkspaceApi.md#setautostopintervalworkspacedeprecated) | **POST** /workspace/{workspaceId}/autostop/{interval} | [DEPRECATED] Set workspace auto-stop interval |
| [**StartWorkspaceDeprecated**](WorkspaceApi.md#startworkspacedeprecated) | **POST** /workspace/{workspaceId}/start | [DEPRECATED] Start workspace |
| [**StopWorkspaceDeprecated**](WorkspaceApi.md#stopworkspacedeprecated) | **POST** /workspace/{workspaceId}/stop | [DEPRECATED] Stop workspace |
| [**UpdatePublicStatusWorkspaceDeprecated**](WorkspaceApi.md#updatepublicstatusworkspacedeprecated) | **POST** /workspace/{workspaceId}/public/{isPublic} | [DEPRECATED] Update public status |

<a id="archiveworkspacedeprecated"></a>
# **ArchiveWorkspaceDeprecated**
> void ArchiveWorkspaceDeprecated (string workspaceId, string? xDaytonaOrganizationID = null)

[DEPRECATED] Archive workspace

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
    public class ArchiveWorkspaceDeprecatedExample
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
            var apiInstance = new WorkspaceApi(httpClient, config, httpClientHandler);
            var workspaceId = "workspaceId_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Archive workspace
                apiInstance.ArchiveWorkspaceDeprecated(workspaceId, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling WorkspaceApi.ArchiveWorkspaceDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ArchiveWorkspaceDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Archive workspace
    apiInstance.ArchiveWorkspaceDeprecatedWithHttpInfo(workspaceId, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling WorkspaceApi.ArchiveWorkspaceDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **workspaceId** | **string** |  |  |
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
| **200** | Workspace has been archived |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="createbackupworkspacedeprecated"></a>
# **CreateBackupWorkspaceDeprecated**
> Workspace CreateBackupWorkspaceDeprecated (string workspaceId, string? xDaytonaOrganizationID = null)

[DEPRECATED] Create workspace backup

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
    public class CreateBackupWorkspaceDeprecatedExample
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
            var apiInstance = new WorkspaceApi(httpClient, config, httpClientHandler);
            var workspaceId = "workspaceId_example";  // string | ID of the workspace
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Create workspace backup
                Workspace result = apiInstance.CreateBackupWorkspaceDeprecated(workspaceId, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling WorkspaceApi.CreateBackupWorkspaceDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the CreateBackupWorkspaceDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Create workspace backup
    ApiResponse<Workspace> response = apiInstance.CreateBackupWorkspaceDeprecatedWithHttpInfo(workspaceId, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling WorkspaceApi.CreateBackupWorkspaceDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **workspaceId** | **string** | ID of the workspace |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="createworkspacedeprecated"></a>
# **CreateWorkspaceDeprecated**
> Workspace CreateWorkspaceDeprecated (CreateWorkspace createWorkspace, string? xDaytonaOrganizationID = null)

[DEPRECATED] Create a new workspace

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
    public class CreateWorkspaceDeprecatedExample
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
            var apiInstance = new WorkspaceApi(httpClient, config, httpClientHandler);
            var createWorkspace = new CreateWorkspace(); // CreateWorkspace | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Create a new workspace
                Workspace result = apiInstance.CreateWorkspaceDeprecated(createWorkspace, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling WorkspaceApi.CreateWorkspaceDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the CreateWorkspaceDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Create a new workspace
    ApiResponse<Workspace> response = apiInstance.CreateWorkspaceDeprecatedWithHttpInfo(createWorkspace, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling WorkspaceApi.CreateWorkspaceDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **createWorkspace** | [**CreateWorkspace**](CreateWorkspace.md) |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="deleteworkspacedeprecated"></a>
# **DeleteWorkspaceDeprecated**
> void DeleteWorkspaceDeprecated (string workspaceId, bool force, string? xDaytonaOrganizationID = null)

[DEPRECATED] Delete workspace

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
    public class DeleteWorkspaceDeprecatedExample
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
            var apiInstance = new WorkspaceApi(httpClient, config, httpClientHandler);
            var workspaceId = "workspaceId_example";  // string | ID of the workspace
            var force = true;  // bool | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Delete workspace
                apiInstance.DeleteWorkspaceDeprecated(workspaceId, force, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling WorkspaceApi.DeleteWorkspaceDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DeleteWorkspaceDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Delete workspace
    apiInstance.DeleteWorkspaceDeprecatedWithHttpInfo(workspaceId, force, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling WorkspaceApi.DeleteWorkspaceDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **workspaceId** | **string** | ID of the workspace |  |
| **force** | **bool** |  |  |
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
| **200** | Workspace has been deleted |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getbuildlogsworkspacedeprecated"></a>
# **GetBuildLogsWorkspaceDeprecated**
> void GetBuildLogsWorkspaceDeprecated (string workspaceId, string? xDaytonaOrganizationID = null, bool? follow = null)

[DEPRECATED] Get build logs

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
    public class GetBuildLogsWorkspaceDeprecatedExample
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
            var apiInstance = new WorkspaceApi(httpClient, config, httpClientHandler);
            var workspaceId = "workspaceId_example";  // string | ID of the workspace
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 
            var follow = true;  // bool? | Whether to follow the logs stream (optional) 

            try
            {
                // [DEPRECATED] Get build logs
                apiInstance.GetBuildLogsWorkspaceDeprecated(workspaceId, xDaytonaOrganizationID, follow);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling WorkspaceApi.GetBuildLogsWorkspaceDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetBuildLogsWorkspaceDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Get build logs
    apiInstance.GetBuildLogsWorkspaceDeprecatedWithHttpInfo(workspaceId, xDaytonaOrganizationID, follow);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling WorkspaceApi.GetBuildLogsWorkspaceDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **workspaceId** | **string** | ID of the workspace |  |
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
| **200** | Build logs stream |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getportpreviewurlworkspacedeprecated"></a>
# **GetPortPreviewUrlWorkspaceDeprecated**
> WorkspacePortPreviewUrl GetPortPreviewUrlWorkspaceDeprecated (string workspaceId, decimal port, string? xDaytonaOrganizationID = null)

[DEPRECATED] Get preview URL for a workspace port

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
    public class GetPortPreviewUrlWorkspaceDeprecatedExample
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
            var apiInstance = new WorkspaceApi(httpClient, config, httpClientHandler);
            var workspaceId = "workspaceId_example";  // string | ID of the workspace
            var port = 8.14D;  // decimal | Port number to get preview URL for
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Get preview URL for a workspace port
                WorkspacePortPreviewUrl result = apiInstance.GetPortPreviewUrlWorkspaceDeprecated(workspaceId, port, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling WorkspaceApi.GetPortPreviewUrlWorkspaceDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetPortPreviewUrlWorkspaceDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Get preview URL for a workspace port
    ApiResponse<WorkspacePortPreviewUrl> response = apiInstance.GetPortPreviewUrlWorkspaceDeprecatedWithHttpInfo(workspaceId, port, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling WorkspaceApi.GetPortPreviewUrlWorkspaceDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **workspaceId** | **string** | ID of the workspace |  |
| **port** | **decimal** | Port number to get preview URL for |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getworkspacedeprecated"></a>
# **GetWorkspaceDeprecated**
> Workspace GetWorkspaceDeprecated (string workspaceId, string? xDaytonaOrganizationID = null, bool? verbose = null)

[DEPRECATED] Get workspace details

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
    public class GetWorkspaceDeprecatedExample
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
            var apiInstance = new WorkspaceApi(httpClient, config, httpClientHandler);
            var workspaceId = "workspaceId_example";  // string | ID of the workspace
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 
            var verbose = true;  // bool? | Include verbose output (optional) 

            try
            {
                // [DEPRECATED] Get workspace details
                Workspace result = apiInstance.GetWorkspaceDeprecated(workspaceId, xDaytonaOrganizationID, verbose);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling WorkspaceApi.GetWorkspaceDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetWorkspaceDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Get workspace details
    ApiResponse<Workspace> response = apiInstance.GetWorkspaceDeprecatedWithHttpInfo(workspaceId, xDaytonaOrganizationID, verbose);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling WorkspaceApi.GetWorkspaceDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **workspaceId** | **string** | ID of the workspace |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |
| **verbose** | **bool?** | Include verbose output | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="listworkspacesdeprecated"></a>
# **ListWorkspacesDeprecated**
> List&lt;Workspace&gt; ListWorkspacesDeprecated (string? xDaytonaOrganizationID = null, bool? verbose = null, string? labels = null)

[DEPRECATED] List all workspaces

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
    public class ListWorkspacesDeprecatedExample
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
            var apiInstance = new WorkspaceApi(httpClient, config, httpClientHandler);
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 
            var verbose = true;  // bool? | Include verbose output (optional) 
            var labels = {"label1": "value1", "label2": "value2"};  // string? | JSON encoded labels to filter by (optional) 

            try
            {
                // [DEPRECATED] List all workspaces
                List<Workspace> result = apiInstance.ListWorkspacesDeprecated(xDaytonaOrganizationID, verbose, labels);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling WorkspaceApi.ListWorkspacesDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ListWorkspacesDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] List all workspaces
    ApiResponse<List<Workspace>> response = apiInstance.ListWorkspacesDeprecatedWithHttpInfo(xDaytonaOrganizationID, verbose, labels);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling WorkspaceApi.ListWorkspacesDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |
| **verbose** | **bool?** | Include verbose output | [optional]  |
| **labels** | **string?** | JSON encoded labels to filter by | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="replacelabelsworkspacedeprecated"></a>
# **ReplaceLabelsWorkspaceDeprecated**
> SandboxLabels ReplaceLabelsWorkspaceDeprecated (string workspaceId, SandboxLabels sandboxLabels, string? xDaytonaOrganizationID = null)

[DEPRECATED] Replace workspace labels

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
    public class ReplaceLabelsWorkspaceDeprecatedExample
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
            var apiInstance = new WorkspaceApi(httpClient, config, httpClientHandler);
            var workspaceId = "workspaceId_example";  // string | ID of the workspace
            var sandboxLabels = new SandboxLabels(); // SandboxLabels | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Replace workspace labels
                SandboxLabels result = apiInstance.ReplaceLabelsWorkspaceDeprecated(workspaceId, sandboxLabels, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling WorkspaceApi.ReplaceLabelsWorkspaceDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ReplaceLabelsWorkspaceDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Replace workspace labels
    ApiResponse<SandboxLabels> response = apiInstance.ReplaceLabelsWorkspaceDeprecatedWithHttpInfo(workspaceId, sandboxLabels, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling WorkspaceApi.ReplaceLabelsWorkspaceDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **workspaceId** | **string** | ID of the workspace |  |
| **sandboxLabels** | [**SandboxLabels**](SandboxLabels.md) |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="setautoarchiveintervalworkspacedeprecated"></a>
# **SetAutoArchiveIntervalWorkspaceDeprecated**
> void SetAutoArchiveIntervalWorkspaceDeprecated (string workspaceId, decimal interval, string? xDaytonaOrganizationID = null)

[DEPRECATED] Set workspace auto-archive interval

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
    public class SetAutoArchiveIntervalWorkspaceDeprecatedExample
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
            var apiInstance = new WorkspaceApi(httpClient, config, httpClientHandler);
            var workspaceId = "workspaceId_example";  // string | ID of the workspace
            var interval = 8.14D;  // decimal | Auto-archive interval in minutes (0 means the maximum interval will be used)
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Set workspace auto-archive interval
                apiInstance.SetAutoArchiveIntervalWorkspaceDeprecated(workspaceId, interval, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling WorkspaceApi.SetAutoArchiveIntervalWorkspaceDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the SetAutoArchiveIntervalWorkspaceDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Set workspace auto-archive interval
    apiInstance.SetAutoArchiveIntervalWorkspaceDeprecatedWithHttpInfo(workspaceId, interval, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling WorkspaceApi.SetAutoArchiveIntervalWorkspaceDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **workspaceId** | **string** | ID of the workspace |  |
| **interval** | **decimal** | Auto-archive interval in minutes (0 means the maximum interval will be used) |  |
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
| **200** | Auto-archive interval has been set |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="setautostopintervalworkspacedeprecated"></a>
# **SetAutostopIntervalWorkspaceDeprecated**
> void SetAutostopIntervalWorkspaceDeprecated (string workspaceId, decimal interval, string? xDaytonaOrganizationID = null)

[DEPRECATED] Set workspace auto-stop interval

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
    public class SetAutostopIntervalWorkspaceDeprecatedExample
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
            var apiInstance = new WorkspaceApi(httpClient, config, httpClientHandler);
            var workspaceId = "workspaceId_example";  // string | ID of the workspace
            var interval = 8.14D;  // decimal | Auto-stop interval in minutes (0 to disable)
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Set workspace auto-stop interval
                apiInstance.SetAutostopIntervalWorkspaceDeprecated(workspaceId, interval, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling WorkspaceApi.SetAutostopIntervalWorkspaceDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the SetAutostopIntervalWorkspaceDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Set workspace auto-stop interval
    apiInstance.SetAutostopIntervalWorkspaceDeprecatedWithHttpInfo(workspaceId, interval, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling WorkspaceApi.SetAutostopIntervalWorkspaceDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **workspaceId** | **string** | ID of the workspace |  |
| **interval** | **decimal** | Auto-stop interval in minutes (0 to disable) |  |
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
| **200** | Auto-stop interval has been set |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="startworkspacedeprecated"></a>
# **StartWorkspaceDeprecated**
> void StartWorkspaceDeprecated (string workspaceId, string? xDaytonaOrganizationID = null)

[DEPRECATED] Start workspace

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
    public class StartWorkspaceDeprecatedExample
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
            var apiInstance = new WorkspaceApi(httpClient, config, httpClientHandler);
            var workspaceId = "workspaceId_example";  // string | ID of the workspace
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Start workspace
                apiInstance.StartWorkspaceDeprecated(workspaceId, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling WorkspaceApi.StartWorkspaceDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the StartWorkspaceDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Start workspace
    apiInstance.StartWorkspaceDeprecatedWithHttpInfo(workspaceId, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling WorkspaceApi.StartWorkspaceDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **workspaceId** | **string** | ID of the workspace |  |
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
| **200** | Workspace has been started |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="stopworkspacedeprecated"></a>
# **StopWorkspaceDeprecated**
> void StopWorkspaceDeprecated (string workspaceId, string? xDaytonaOrganizationID = null)

[DEPRECATED] Stop workspace

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
    public class StopWorkspaceDeprecatedExample
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
            var apiInstance = new WorkspaceApi(httpClient, config, httpClientHandler);
            var workspaceId = "workspaceId_example";  // string | ID of the workspace
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Stop workspace
                apiInstance.StopWorkspaceDeprecated(workspaceId, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling WorkspaceApi.StopWorkspaceDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the StopWorkspaceDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Stop workspace
    apiInstance.StopWorkspaceDeprecatedWithHttpInfo(workspaceId, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling WorkspaceApi.StopWorkspaceDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **workspaceId** | **string** | ID of the workspace |  |
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
| **200** | Workspace has been stopped |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="updatepublicstatusworkspacedeprecated"></a>
# **UpdatePublicStatusWorkspaceDeprecated**
> void UpdatePublicStatusWorkspaceDeprecated (string workspaceId, bool isPublic, string? xDaytonaOrganizationID = null)

[DEPRECATED] Update public status

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
    public class UpdatePublicStatusWorkspaceDeprecatedExample
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
            var apiInstance = new WorkspaceApi(httpClient, config, httpClientHandler);
            var workspaceId = "workspaceId_example";  // string | ID of the workspace
            var isPublic = true;  // bool | Public status to set
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Update public status
                apiInstance.UpdatePublicStatusWorkspaceDeprecated(workspaceId, isPublic, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling WorkspaceApi.UpdatePublicStatusWorkspaceDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UpdatePublicStatusWorkspaceDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Update public status
    apiInstance.UpdatePublicStatusWorkspaceDeprecatedWithHttpInfo(workspaceId, isPublic, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling WorkspaceApi.UpdatePublicStatusWorkspaceDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **workspaceId** | **string** | ID of the workspace |  |
| **isPublic** | **bool** | Public status to set |  |
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
| **201** |  |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

