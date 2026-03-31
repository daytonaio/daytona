# Daytona.ApiClient.Api.SandboxApi

All URIs are relative to *http://localhost:3000*

| Method | HTTP request | Description |
|--------|--------------|-------------|
| [**ArchiveSandbox**](SandboxApi.md#archivesandbox) | **POST** /sandbox/{sandboxIdOrName}/archive | Archive sandbox |
| [**CreateBackup**](SandboxApi.md#createbackup) | **POST** /sandbox/{sandboxIdOrName}/backup | Create sandbox backup |
| [**CreateSandbox**](SandboxApi.md#createsandbox) | **POST** /sandbox | Create a new sandbox |
| [**CreateSshAccess**](SandboxApi.md#createsshaccess) | **POST** /sandbox/{sandboxIdOrName}/ssh-access | Create SSH access for sandbox |
| [**DeleteSandbox**](SandboxApi.md#deletesandbox) | **DELETE** /sandbox/{sandboxIdOrName} | Delete sandbox |
| [**ExpireSignedPortPreviewUrl**](SandboxApi.md#expiresignedportpreviewurl) | **POST** /sandbox/{sandboxIdOrName}/ports/{port}/signed-preview-url/{token}/expire | Expire signed preview URL for a sandbox port |
| [**GetBuildLogs**](SandboxApi.md#getbuildlogs) | **GET** /sandbox/{sandboxIdOrName}/build-logs | Get build logs |
| [**GetBuildLogsUrl**](SandboxApi.md#getbuildlogsurl) | **GET** /sandbox/{sandboxIdOrName}/build-logs-url | Get build logs URL |
| [**GetPortPreviewUrl**](SandboxApi.md#getportpreviewurl) | **GET** /sandbox/{sandboxIdOrName}/ports/{port}/preview-url | Get preview URL for a sandbox port |
| [**GetSandbox**](SandboxApi.md#getsandbox) | **GET** /sandbox/{sandboxIdOrName} | Get sandbox details |
| [**GetSandboxLogs**](SandboxApi.md#getsandboxlogs) | **GET** /sandbox/{sandboxId}/telemetry/logs | Get sandbox logs |
| [**GetSandboxMetrics**](SandboxApi.md#getsandboxmetrics) | **GET** /sandbox/{sandboxId}/telemetry/metrics | Get sandbox metrics |
| [**GetSandboxTraceSpans**](SandboxApi.md#getsandboxtracespans) | **GET** /sandbox/{sandboxId}/telemetry/traces/{traceId} | Get trace spans |
| [**GetSandboxTraces**](SandboxApi.md#getsandboxtraces) | **GET** /sandbox/{sandboxId}/telemetry/traces | Get sandbox traces |
| [**GetSandboxesForRunner**](SandboxApi.md#getsandboxesforrunner) | **GET** /sandbox/for-runner | Get sandboxes for the authenticated runner |
| [**GetSignedPortPreviewUrl**](SandboxApi.md#getsignedportpreviewurl) | **GET** /sandbox/{sandboxIdOrName}/ports/{port}/signed-preview-url | Get signed preview URL for a sandbox port |
| [**GetToolboxProxyUrl**](SandboxApi.md#gettoolboxproxyurl) | **GET** /sandbox/{sandboxId}/toolbox-proxy-url | Get toolbox proxy URL for a sandbox |
| [**ListSandboxes**](SandboxApi.md#listsandboxes) | **GET** /sandbox | List all sandboxes |
| [**ListSandboxesPaginated**](SandboxApi.md#listsandboxespaginated) | **GET** /sandbox/paginated | List all sandboxes paginated |
| [**RecoverSandbox**](SandboxApi.md#recoversandbox) | **POST** /sandbox/{sandboxIdOrName}/recover | Recover sandbox from error state |
| [**ReplaceLabels**](SandboxApi.md#replacelabels) | **PUT** /sandbox/{sandboxIdOrName}/labels | Replace sandbox labels |
| [**ResizeSandbox**](SandboxApi.md#resizesandbox) | **POST** /sandbox/{sandboxIdOrName}/resize | Resize sandbox resources |
| [**RevokeSshAccess**](SandboxApi.md#revokesshaccess) | **DELETE** /sandbox/{sandboxIdOrName}/ssh-access | Revoke SSH access for sandbox |
| [**SetAutoArchiveInterval**](SandboxApi.md#setautoarchiveinterval) | **POST** /sandbox/{sandboxIdOrName}/autoarchive/{interval} | Set sandbox auto-archive interval |
| [**SetAutoDeleteInterval**](SandboxApi.md#setautodeleteinterval) | **POST** /sandbox/{sandboxIdOrName}/autodelete/{interval} | Set sandbox auto-delete interval |
| [**SetAutostopInterval**](SandboxApi.md#setautostopinterval) | **POST** /sandbox/{sandboxIdOrName}/autostop/{interval} | Set sandbox auto-stop interval |
| [**StartSandbox**](SandboxApi.md#startsandbox) | **POST** /sandbox/{sandboxIdOrName}/start | Start sandbox |
| [**StopSandbox**](SandboxApi.md#stopsandbox) | **POST** /sandbox/{sandboxIdOrName}/stop | Stop sandbox |
| [**UpdateLastActivity**](SandboxApi.md#updatelastactivity) | **POST** /sandbox/{sandboxId}/last-activity | Update sandbox last activity |
| [**UpdatePublicStatus**](SandboxApi.md#updatepublicstatus) | **POST** /sandbox/{sandboxIdOrName}/public/{isPublic} | Update public status |
| [**UpdateSandboxState**](SandboxApi.md#updatesandboxstate) | **PUT** /sandbox/{sandboxId}/state | Update sandbox state |
| [**ValidateSshAccess**](SandboxApi.md#validatesshaccess) | **GET** /sandbox/ssh-access/validate | Validate SSH access for sandbox |

<a id="archivesandbox"></a>
# **ArchiveSandbox**
> Sandbox ArchiveSandbox (string sandboxIdOrName, string? xDaytonaOrganizationID = null)

Archive sandbox

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
    public class ArchiveSandboxExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var sandboxIdOrName = "sandboxIdOrName_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Archive sandbox
                Sandbox result = apiInstance.ArchiveSandbox(sandboxIdOrName, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.ArchiveSandbox: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ArchiveSandboxWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Archive sandbox
    ApiResponse<Sandbox> response = apiInstance.ArchiveSandboxWithHttpInfo(sandboxIdOrName, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.ArchiveSandboxWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxIdOrName** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="createbackup"></a>
# **CreateBackup**
> Sandbox CreateBackup (string sandboxIdOrName, string? xDaytonaOrganizationID = null)

Create sandbox backup

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
    public class CreateBackupExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var sandboxIdOrName = "sandboxIdOrName_example";  // string | ID or name of the sandbox
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Create sandbox backup
                Sandbox result = apiInstance.CreateBackup(sandboxIdOrName, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.CreateBackup: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the CreateBackupWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Create sandbox backup
    ApiResponse<Sandbox> response = apiInstance.CreateBackupWithHttpInfo(sandboxIdOrName, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.CreateBackupWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxIdOrName** | **string** | ID or name of the sandbox |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="createsandbox"></a>
# **CreateSandbox**
> Sandbox CreateSandbox (CreateSandbox createSandbox, string? xDaytonaOrganizationID = null)

Create a new sandbox

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
    public class CreateSandboxExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var createSandbox = new CreateSandbox(); // CreateSandbox | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Create a new sandbox
                Sandbox result = apiInstance.CreateSandbox(createSandbox, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.CreateSandbox: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the CreateSandboxWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Create a new sandbox
    ApiResponse<Sandbox> response = apiInstance.CreateSandboxWithHttpInfo(createSandbox, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.CreateSandboxWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **createSandbox** | [**CreateSandbox**](CreateSandbox.md) |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="createsshaccess"></a>
# **CreateSshAccess**
> SshAccessDto CreateSshAccess (string sandboxIdOrName, string? xDaytonaOrganizationID = null, decimal? expiresInMinutes = null)

Create SSH access for sandbox

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
    public class CreateSshAccessExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var sandboxIdOrName = "sandboxIdOrName_example";  // string | ID or name of the sandbox
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 
            var expiresInMinutes = 8.14D;  // decimal? | Expiration time in minutes (default: 60) (optional) 

            try
            {
                // Create SSH access for sandbox
                SshAccessDto result = apiInstance.CreateSshAccess(sandboxIdOrName, xDaytonaOrganizationID, expiresInMinutes);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.CreateSshAccess: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the CreateSshAccessWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Create SSH access for sandbox
    ApiResponse<SshAccessDto> response = apiInstance.CreateSshAccessWithHttpInfo(sandboxIdOrName, xDaytonaOrganizationID, expiresInMinutes);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.CreateSshAccessWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxIdOrName** | **string** | ID or name of the sandbox |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |
| **expiresInMinutes** | **decimal?** | Expiration time in minutes (default: 60) | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="deletesandbox"></a>
# **DeleteSandbox**
> Sandbox DeleteSandbox (string sandboxIdOrName, string? xDaytonaOrganizationID = null)

Delete sandbox

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
    public class DeleteSandboxExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var sandboxIdOrName = "sandboxIdOrName_example";  // string | ID or name of the sandbox
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Delete sandbox
                Sandbox result = apiInstance.DeleteSandbox(sandboxIdOrName, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.DeleteSandbox: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DeleteSandboxWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Delete sandbox
    ApiResponse<Sandbox> response = apiInstance.DeleteSandboxWithHttpInfo(sandboxIdOrName, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.DeleteSandboxWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxIdOrName** | **string** | ID or name of the sandbox |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="expiresignedportpreviewurl"></a>
# **ExpireSignedPortPreviewUrl**
> void ExpireSignedPortPreviewUrl (string sandboxIdOrName, int port, string token, string? xDaytonaOrganizationID = null)

Expire signed preview URL for a sandbox port

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
    public class ExpireSignedPortPreviewUrlExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var sandboxIdOrName = "sandboxIdOrName_example";  // string | ID or name of the sandbox
            var port = 56;  // int | Port number to expire signed preview URL for
            var token = "token_example";  // string | Token to expire signed preview URL for
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Expire signed preview URL for a sandbox port
                apiInstance.ExpireSignedPortPreviewUrl(sandboxIdOrName, port, token, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.ExpireSignedPortPreviewUrl: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ExpireSignedPortPreviewUrlWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Expire signed preview URL for a sandbox port
    apiInstance.ExpireSignedPortPreviewUrlWithHttpInfo(sandboxIdOrName, port, token, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.ExpireSignedPortPreviewUrlWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxIdOrName** | **string** | ID or name of the sandbox |  |
| **port** | **int** | Port number to expire signed preview URL for |  |
| **token** | **string** | Token to expire signed preview URL for |  |
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
| **200** | Signed preview URL has been expired |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getbuildlogs"></a>
# **GetBuildLogs**
> void GetBuildLogs (string sandboxIdOrName, string? xDaytonaOrganizationID = null, bool? follow = null)

Get build logs

This endpoint is deprecated. Use `getBuildLogsUrl` instead.

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
    public class GetBuildLogsExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var sandboxIdOrName = "sandboxIdOrName_example";  // string | ID or name of the sandbox
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 
            var follow = true;  // bool? | Whether to follow the logs stream (optional) 

            try
            {
                // Get build logs
                apiInstance.GetBuildLogs(sandboxIdOrName, xDaytonaOrganizationID, follow);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.GetBuildLogs: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetBuildLogsWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get build logs
    apiInstance.GetBuildLogsWithHttpInfo(sandboxIdOrName, xDaytonaOrganizationID, follow);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.GetBuildLogsWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxIdOrName** | **string** | ID or name of the sandbox |  |
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

<a id="getbuildlogsurl"></a>
# **GetBuildLogsUrl**
> Url GetBuildLogsUrl (string sandboxIdOrName, string? xDaytonaOrganizationID = null)

Get build logs URL

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
    public class GetBuildLogsUrlExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var sandboxIdOrName = "sandboxIdOrName_example";  // string | ID or name of the sandbox
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Get build logs URL
                Url result = apiInstance.GetBuildLogsUrl(sandboxIdOrName, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.GetBuildLogsUrl: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetBuildLogsUrlWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get build logs URL
    ApiResponse<Url> response = apiInstance.GetBuildLogsUrlWithHttpInfo(sandboxIdOrName, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.GetBuildLogsUrlWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxIdOrName** | **string** | ID or name of the sandbox |  |
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
| **200** | Build logs URL |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getportpreviewurl"></a>
# **GetPortPreviewUrl**
> PortPreviewUrl GetPortPreviewUrl (string sandboxIdOrName, decimal port, string? xDaytonaOrganizationID = null)

Get preview URL for a sandbox port

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
    public class GetPortPreviewUrlExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var sandboxIdOrName = "sandboxIdOrName_example";  // string | ID or name of the sandbox
            var port = 8.14D;  // decimal | Port number to get preview URL for
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Get preview URL for a sandbox port
                PortPreviewUrl result = apiInstance.GetPortPreviewUrl(sandboxIdOrName, port, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.GetPortPreviewUrl: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetPortPreviewUrlWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get preview URL for a sandbox port
    ApiResponse<PortPreviewUrl> response = apiInstance.GetPortPreviewUrlWithHttpInfo(sandboxIdOrName, port, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.GetPortPreviewUrlWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxIdOrName** | **string** | ID or name of the sandbox |  |
| **port** | **decimal** | Port number to get preview URL for |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getsandbox"></a>
# **GetSandbox**
> Sandbox GetSandbox (string sandboxIdOrName, string? xDaytonaOrganizationID = null, bool? verbose = null)

Get sandbox details

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
    public class GetSandboxExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var sandboxIdOrName = "sandboxIdOrName_example";  // string | ID or name of the sandbox
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 
            var verbose = true;  // bool? | Include verbose output (optional) 

            try
            {
                // Get sandbox details
                Sandbox result = apiInstance.GetSandbox(sandboxIdOrName, xDaytonaOrganizationID, verbose);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.GetSandbox: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetSandboxWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get sandbox details
    ApiResponse<Sandbox> response = apiInstance.GetSandboxWithHttpInfo(sandboxIdOrName, xDaytonaOrganizationID, verbose);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.GetSandboxWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxIdOrName** | **string** | ID or name of the sandbox |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |
| **verbose** | **bool?** | Include verbose output | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getsandboxlogs"></a>
# **GetSandboxLogs**
> PaginatedLogs GetSandboxLogs (string sandboxId, DateTime from, DateTime to, string? xDaytonaOrganizationID = null, decimal? page = null, decimal? limit = null, List<string>? severities = null, string? search = null)

Get sandbox logs

Retrieve OTEL logs for a sandbox within a time range

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
    public class GetSandboxLogsExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | ID of the sandbox
            var from = DateTime.Parse("2013-10-20T19:20:30+01:00");  // DateTime | Start of time range (ISO 8601)
            var to = DateTime.Parse("2013-10-20T19:20:30+01:00");  // DateTime | End of time range (ISO 8601)
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 
            var page = 1MD;  // decimal? | Page number (1-indexed) (optional)  (default to 1M)
            var limit = 100MD;  // decimal? | Number of items per page (optional)  (default to 100M)
            var severities = new List<string>?(); // List<string>? | Filter by severity levels (DEBUG, INFO, WARN, ERROR) (optional) 
            var search = "search_example";  // string? | Search in log body (optional) 

            try
            {
                // Get sandbox logs
                PaginatedLogs result = apiInstance.GetSandboxLogs(sandboxId, from, to, xDaytonaOrganizationID, page, limit, severities, search);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.GetSandboxLogs: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetSandboxLogsWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get sandbox logs
    ApiResponse<PaginatedLogs> response = apiInstance.GetSandboxLogsWithHttpInfo(sandboxId, from, to, xDaytonaOrganizationID, page, limit, severities, search);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.GetSandboxLogsWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** | ID of the sandbox |  |
| **from** | **DateTime** | Start of time range (ISO 8601) |  |
| **to** | **DateTime** | End of time range (ISO 8601) |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |
| **page** | **decimal?** | Page number (1-indexed) | [optional] [default to 1M] |
| **limit** | **decimal?** | Number of items per page | [optional] [default to 100M] |
| **severities** | [**List&lt;string&gt;?**](string.md) | Filter by severity levels (DEBUG, INFO, WARN, ERROR) | [optional]  |
| **search** | **string?** | Search in log body | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getsandboxmetrics"></a>
# **GetSandboxMetrics**
> MetricsResponse GetSandboxMetrics (string sandboxId, DateTime from, DateTime to, string? xDaytonaOrganizationID = null, List<string>? metricNames = null)

Get sandbox metrics

Retrieve OTEL metrics for a sandbox within a time range

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
    public class GetSandboxMetricsExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | ID of the sandbox
            var from = DateTime.Parse("2013-10-20T19:20:30+01:00");  // DateTime | Start of time range (ISO 8601)
            var to = DateTime.Parse("2013-10-20T19:20:30+01:00");  // DateTime | End of time range (ISO 8601)
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 
            var metricNames = new List<string>?(); // List<string>? | Filter by metric names (optional) 

            try
            {
                // Get sandbox metrics
                MetricsResponse result = apiInstance.GetSandboxMetrics(sandboxId, from, to, xDaytonaOrganizationID, metricNames);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.GetSandboxMetrics: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetSandboxMetricsWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get sandbox metrics
    ApiResponse<MetricsResponse> response = apiInstance.GetSandboxMetricsWithHttpInfo(sandboxId, from, to, xDaytonaOrganizationID, metricNames);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.GetSandboxMetricsWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** | ID of the sandbox |  |
| **from** | **DateTime** | Start of time range (ISO 8601) |  |
| **to** | **DateTime** | End of time range (ISO 8601) |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |
| **metricNames** | [**List&lt;string&gt;?**](string.md) | Filter by metric names | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getsandboxtracespans"></a>
# **GetSandboxTraceSpans**
> List&lt;TraceSpan&gt; GetSandboxTraceSpans (string sandboxId, string traceId, string? xDaytonaOrganizationID = null)

Get trace spans

Retrieve all spans for a specific trace

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
    public class GetSandboxTraceSpansExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | ID of the sandbox
            var traceId = "traceId_example";  // string | ID of the trace
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Get trace spans
                List<TraceSpan> result = apiInstance.GetSandboxTraceSpans(sandboxId, traceId, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.GetSandboxTraceSpans: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetSandboxTraceSpansWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get trace spans
    ApiResponse<List<TraceSpan>> response = apiInstance.GetSandboxTraceSpansWithHttpInfo(sandboxId, traceId, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.GetSandboxTraceSpansWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** | ID of the sandbox |  |
| **traceId** | **string** | ID of the trace |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getsandboxtraces"></a>
# **GetSandboxTraces**
> PaginatedTraces GetSandboxTraces (string sandboxId, DateTime from, DateTime to, string? xDaytonaOrganizationID = null, decimal? page = null, decimal? limit = null)

Get sandbox traces

Retrieve OTEL traces for a sandbox within a time range

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
    public class GetSandboxTracesExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | ID of the sandbox
            var from = DateTime.Parse("2013-10-20T19:20:30+01:00");  // DateTime | Start of time range (ISO 8601)
            var to = DateTime.Parse("2013-10-20T19:20:30+01:00");  // DateTime | End of time range (ISO 8601)
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 
            var page = 1MD;  // decimal? | Page number (1-indexed) (optional)  (default to 1M)
            var limit = 100MD;  // decimal? | Number of items per page (optional)  (default to 100M)

            try
            {
                // Get sandbox traces
                PaginatedTraces result = apiInstance.GetSandboxTraces(sandboxId, from, to, xDaytonaOrganizationID, page, limit);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.GetSandboxTraces: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetSandboxTracesWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get sandbox traces
    ApiResponse<PaginatedTraces> response = apiInstance.GetSandboxTracesWithHttpInfo(sandboxId, from, to, xDaytonaOrganizationID, page, limit);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.GetSandboxTracesWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** | ID of the sandbox |  |
| **from** | **DateTime** | Start of time range (ISO 8601) |  |
| **to** | **DateTime** | End of time range (ISO 8601) |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |
| **page** | **decimal?** | Page number (1-indexed) | [optional] [default to 1M] |
| **limit** | **decimal?** | Number of items per page | [optional] [default to 100M] |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getsandboxesforrunner"></a>
# **GetSandboxesForRunner**
> List&lt;Sandbox&gt; GetSandboxesForRunner (string? xDaytonaOrganizationID = null, string? states = null, bool? skipReconcilingSandboxes = null)

Get sandboxes for the authenticated runner

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
    public class GetSandboxesForRunnerExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 
            var states = "states_example";  // string? | Comma-separated list of sandbox states to filter by (optional) 
            var skipReconcilingSandboxes = true;  // bool? | Skip sandboxes where state differs from desired state (optional) 

            try
            {
                // Get sandboxes for the authenticated runner
                List<Sandbox> result = apiInstance.GetSandboxesForRunner(xDaytonaOrganizationID, states, skipReconcilingSandboxes);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.GetSandboxesForRunner: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetSandboxesForRunnerWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get sandboxes for the authenticated runner
    ApiResponse<List<Sandbox>> response = apiInstance.GetSandboxesForRunnerWithHttpInfo(xDaytonaOrganizationID, states, skipReconcilingSandboxes);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.GetSandboxesForRunnerWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |
| **states** | **string?** | Comma-separated list of sandbox states to filter by | [optional]  |
| **skipReconcilingSandboxes** | **bool?** | Skip sandboxes where state differs from desired state | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getsignedportpreviewurl"></a>
# **GetSignedPortPreviewUrl**
> SignedPortPreviewUrl GetSignedPortPreviewUrl (string sandboxIdOrName, int port, string? xDaytonaOrganizationID = null, int? expiresInSeconds = null)

Get signed preview URL for a sandbox port

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
    public class GetSignedPortPreviewUrlExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var sandboxIdOrName = "sandboxIdOrName_example";  // string | ID or name of the sandbox
            var port = 56;  // int | Port number to get signed preview URL for
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 
            var expiresInSeconds = 56;  // int? | Expiration time in seconds (default: 60 seconds) (optional) 

            try
            {
                // Get signed preview URL for a sandbox port
                SignedPortPreviewUrl result = apiInstance.GetSignedPortPreviewUrl(sandboxIdOrName, port, xDaytonaOrganizationID, expiresInSeconds);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.GetSignedPortPreviewUrl: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetSignedPortPreviewUrlWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get signed preview URL for a sandbox port
    ApiResponse<SignedPortPreviewUrl> response = apiInstance.GetSignedPortPreviewUrlWithHttpInfo(sandboxIdOrName, port, xDaytonaOrganizationID, expiresInSeconds);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.GetSignedPortPreviewUrlWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxIdOrName** | **string** | ID or name of the sandbox |  |
| **port** | **int** | Port number to get signed preview URL for |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |
| **expiresInSeconds** | **int?** | Expiration time in seconds (default: 60 seconds) | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="gettoolboxproxyurl"></a>
# **GetToolboxProxyUrl**
> ToolboxProxyUrl GetToolboxProxyUrl (string sandboxId, string? xDaytonaOrganizationID = null)

Get toolbox proxy URL for a sandbox

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
    public class GetToolboxProxyUrlExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | ID of the sandbox
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Get toolbox proxy URL for a sandbox
                ToolboxProxyUrl result = apiInstance.GetToolboxProxyUrl(sandboxId, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.GetToolboxProxyUrl: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetToolboxProxyUrlWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get toolbox proxy URL for a sandbox
    ApiResponse<ToolboxProxyUrl> response = apiInstance.GetToolboxProxyUrlWithHttpInfo(sandboxId, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.GetToolboxProxyUrlWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** | ID of the sandbox |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="listsandboxes"></a>
# **ListSandboxes**
> List&lt;Sandbox&gt; ListSandboxes (string? xDaytonaOrganizationID = null, bool? verbose = null, string? labels = null, bool? includeErroredDeleted = null)

List all sandboxes

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
    public class ListSandboxesExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 
            var verbose = true;  // bool? | Include verbose output (optional) 
            var labels = {"label1": "value1", "label2": "value2"};  // string? | JSON encoded labels to filter by (optional) 
            var includeErroredDeleted = true;  // bool? | Include errored and deleted sandboxes (optional) 

            try
            {
                // List all sandboxes
                List<Sandbox> result = apiInstance.ListSandboxes(xDaytonaOrganizationID, verbose, labels, includeErroredDeleted);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.ListSandboxes: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ListSandboxesWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // List all sandboxes
    ApiResponse<List<Sandbox>> response = apiInstance.ListSandboxesWithHttpInfo(xDaytonaOrganizationID, verbose, labels, includeErroredDeleted);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.ListSandboxesWithHttpInfo: " + e.Message);
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
| **includeErroredDeleted** | **bool?** | Include errored and deleted sandboxes | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="listsandboxespaginated"></a>
# **ListSandboxesPaginated**
> PaginatedSandboxes ListSandboxesPaginated (string? xDaytonaOrganizationID = null, decimal? page = null, decimal? limit = null, string? id = null, string? name = null, string? labels = null, bool? includeErroredDeleted = null, List<string>? states = null, List<string>? snapshots = null, List<string>? regions = null, decimal? minCpu = null, decimal? maxCpu = null, decimal? minMemoryGiB = null, decimal? maxMemoryGiB = null, decimal? minDiskGiB = null, decimal? maxDiskGiB = null, DateTime? lastEventAfter = null, DateTime? lastEventBefore = null, string? sort = null, string? order = null)

List all sandboxes paginated

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
    public class ListSandboxesPaginatedExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 
            var page = 1MD;  // decimal? | Page number of the results (optional)  (default to 1M)
            var limit = 100MD;  // decimal? | Number of results per page (optional)  (default to 100M)
            var id = abc123;  // string? | Filter by partial ID match (optional) 
            var name = abc123;  // string? | Filter by partial name match (optional) 
            var labels = {"label1": "value1", "label2": "value2"};  // string? | JSON encoded labels to filter by (optional) 
            var includeErroredDeleted = false;  // bool? | Include results with errored state and deleted desired state (optional)  (default to false)
            var states = new List<string>?(); // List<string>? | List of states to filter by (optional) 
            var snapshots = new List<string>?(); // List<string>? | List of snapshot names to filter by (optional) 
            var regions = new List<string>?(); // List<string>? | List of regions to filter by (optional) 
            var minCpu = 8.14D;  // decimal? | Minimum CPU (optional) 
            var maxCpu = 8.14D;  // decimal? | Maximum CPU (optional) 
            var minMemoryGiB = 8.14D;  // decimal? | Minimum memory in GiB (optional) 
            var maxMemoryGiB = 8.14D;  // decimal? | Maximum memory in GiB (optional) 
            var minDiskGiB = 8.14D;  // decimal? | Minimum disk space in GiB (optional) 
            var maxDiskGiB = 8.14D;  // decimal? | Maximum disk space in GiB (optional) 
            var lastEventAfter = 2024-01-01T00:00Z;  // DateTime? | Include items with last event after this timestamp (optional) 
            var lastEventBefore = 2024-12-31T23:59:59Z;  // DateTime? | Include items with last event before this timestamp (optional) 
            var sort = "id";  // string? | Field to sort by (optional)  (default to createdAt)
            var order = "asc";  // string? | Direction to sort by (optional)  (default to desc)

            try
            {
                // List all sandboxes paginated
                PaginatedSandboxes result = apiInstance.ListSandboxesPaginated(xDaytonaOrganizationID, page, limit, id, name, labels, includeErroredDeleted, states, snapshots, regions, minCpu, maxCpu, minMemoryGiB, maxMemoryGiB, minDiskGiB, maxDiskGiB, lastEventAfter, lastEventBefore, sort, order);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.ListSandboxesPaginated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ListSandboxesPaginatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // List all sandboxes paginated
    ApiResponse<PaginatedSandboxes> response = apiInstance.ListSandboxesPaginatedWithHttpInfo(xDaytonaOrganizationID, page, limit, id, name, labels, includeErroredDeleted, states, snapshots, regions, minCpu, maxCpu, minMemoryGiB, maxMemoryGiB, minDiskGiB, maxDiskGiB, lastEventAfter, lastEventBefore, sort, order);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.ListSandboxesPaginatedWithHttpInfo: " + e.Message);
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
| **id** | **string?** | Filter by partial ID match | [optional]  |
| **name** | **string?** | Filter by partial name match | [optional]  |
| **labels** | **string?** | JSON encoded labels to filter by | [optional]  |
| **includeErroredDeleted** | **bool?** | Include results with errored state and deleted desired state | [optional] [default to false] |
| **states** | [**List&lt;string&gt;?**](string.md) | List of states to filter by | [optional]  |
| **snapshots** | [**List&lt;string&gt;?**](string.md) | List of snapshot names to filter by | [optional]  |
| **regions** | [**List&lt;string&gt;?**](string.md) | List of regions to filter by | [optional]  |
| **minCpu** | **decimal?** | Minimum CPU | [optional]  |
| **maxCpu** | **decimal?** | Maximum CPU | [optional]  |
| **minMemoryGiB** | **decimal?** | Minimum memory in GiB | [optional]  |
| **maxMemoryGiB** | **decimal?** | Maximum memory in GiB | [optional]  |
| **minDiskGiB** | **decimal?** | Minimum disk space in GiB | [optional]  |
| **maxDiskGiB** | **decimal?** | Maximum disk space in GiB | [optional]  |
| **lastEventAfter** | **DateTime?** | Include items with last event after this timestamp | [optional]  |
| **lastEventBefore** | **DateTime?** | Include items with last event before this timestamp | [optional]  |
| **sort** | **string?** | Field to sort by | [optional] [default to createdAt] |
| **order** | **string?** | Direction to sort by | [optional] [default to desc] |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="recoversandbox"></a>
# **RecoverSandbox**
> Sandbox RecoverSandbox (string sandboxIdOrName, string? xDaytonaOrganizationID = null)

Recover sandbox from error state

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
    public class RecoverSandboxExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var sandboxIdOrName = "sandboxIdOrName_example";  // string | ID or name of the sandbox
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Recover sandbox from error state
                Sandbox result = apiInstance.RecoverSandbox(sandboxIdOrName, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.RecoverSandbox: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the RecoverSandboxWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Recover sandbox from error state
    ApiResponse<Sandbox> response = apiInstance.RecoverSandboxWithHttpInfo(sandboxIdOrName, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.RecoverSandboxWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxIdOrName** | **string** | ID or name of the sandbox |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="replacelabels"></a>
# **ReplaceLabels**
> SandboxLabels ReplaceLabels (string sandboxIdOrName, SandboxLabels sandboxLabels, string? xDaytonaOrganizationID = null)

Replace sandbox labels

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
    public class ReplaceLabelsExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var sandboxIdOrName = "sandboxIdOrName_example";  // string | ID or name of the sandbox
            var sandboxLabels = new SandboxLabels(); // SandboxLabels | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Replace sandbox labels
                SandboxLabels result = apiInstance.ReplaceLabels(sandboxIdOrName, sandboxLabels, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.ReplaceLabels: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ReplaceLabelsWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Replace sandbox labels
    ApiResponse<SandboxLabels> response = apiInstance.ReplaceLabelsWithHttpInfo(sandboxIdOrName, sandboxLabels, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.ReplaceLabelsWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxIdOrName** | **string** | ID or name of the sandbox |  |
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

<a id="resizesandbox"></a>
# **ResizeSandbox**
> Sandbox ResizeSandbox (string sandboxIdOrName, ResizeSandbox resizeSandbox, string? xDaytonaOrganizationID = null)

Resize sandbox resources

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
    public class ResizeSandboxExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var sandboxIdOrName = "sandboxIdOrName_example";  // string | ID or name of the sandbox
            var resizeSandbox = new ResizeSandbox(); // ResizeSandbox | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Resize sandbox resources
                Sandbox result = apiInstance.ResizeSandbox(sandboxIdOrName, resizeSandbox, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.ResizeSandbox: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ResizeSandboxWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Resize sandbox resources
    ApiResponse<Sandbox> response = apiInstance.ResizeSandboxWithHttpInfo(sandboxIdOrName, resizeSandbox, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.ResizeSandboxWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxIdOrName** | **string** | ID or name of the sandbox |  |
| **resizeSandbox** | [**ResizeSandbox**](ResizeSandbox.md) |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="revokesshaccess"></a>
# **RevokeSshAccess**
> Sandbox RevokeSshAccess (string sandboxIdOrName, string? xDaytonaOrganizationID = null, string? token = null)

Revoke SSH access for sandbox

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
    public class RevokeSshAccessExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var sandboxIdOrName = "sandboxIdOrName_example";  // string | ID or name of the sandbox
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 
            var token = "token_example";  // string? | SSH access token to revoke. If not provided, all SSH access for the sandbox will be revoked. (optional) 

            try
            {
                // Revoke SSH access for sandbox
                Sandbox result = apiInstance.RevokeSshAccess(sandboxIdOrName, xDaytonaOrganizationID, token);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.RevokeSshAccess: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the RevokeSshAccessWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Revoke SSH access for sandbox
    ApiResponse<Sandbox> response = apiInstance.RevokeSshAccessWithHttpInfo(sandboxIdOrName, xDaytonaOrganizationID, token);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.RevokeSshAccessWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxIdOrName** | **string** | ID or name of the sandbox |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |
| **token** | **string?** | SSH access token to revoke. If not provided, all SSH access for the sandbox will be revoked. | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="setautoarchiveinterval"></a>
# **SetAutoArchiveInterval**
> Sandbox SetAutoArchiveInterval (string sandboxIdOrName, decimal interval, string? xDaytonaOrganizationID = null)

Set sandbox auto-archive interval

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
    public class SetAutoArchiveIntervalExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var sandboxIdOrName = "sandboxIdOrName_example";  // string | ID or name of the sandbox
            var interval = 8.14D;  // decimal | Auto-archive interval in minutes (0 means the maximum interval will be used)
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Set sandbox auto-archive interval
                Sandbox result = apiInstance.SetAutoArchiveInterval(sandboxIdOrName, interval, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.SetAutoArchiveInterval: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the SetAutoArchiveIntervalWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Set sandbox auto-archive interval
    ApiResponse<Sandbox> response = apiInstance.SetAutoArchiveIntervalWithHttpInfo(sandboxIdOrName, interval, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.SetAutoArchiveIntervalWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxIdOrName** | **string** | ID or name of the sandbox |  |
| **interval** | **decimal** | Auto-archive interval in minutes (0 means the maximum interval will be used) |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="setautodeleteinterval"></a>
# **SetAutoDeleteInterval**
> Sandbox SetAutoDeleteInterval (string sandboxIdOrName, decimal interval, string? xDaytonaOrganizationID = null)

Set sandbox auto-delete interval

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
    public class SetAutoDeleteIntervalExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var sandboxIdOrName = "sandboxIdOrName_example";  // string | ID or name of the sandbox
            var interval = 8.14D;  // decimal | Auto-delete interval in minutes (negative value means disabled, 0 means delete immediately upon stopping)
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Set sandbox auto-delete interval
                Sandbox result = apiInstance.SetAutoDeleteInterval(sandboxIdOrName, interval, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.SetAutoDeleteInterval: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the SetAutoDeleteIntervalWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Set sandbox auto-delete interval
    ApiResponse<Sandbox> response = apiInstance.SetAutoDeleteIntervalWithHttpInfo(sandboxIdOrName, interval, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.SetAutoDeleteIntervalWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxIdOrName** | **string** | ID or name of the sandbox |  |
| **interval** | **decimal** | Auto-delete interval in minutes (negative value means disabled, 0 means delete immediately upon stopping) |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="setautostopinterval"></a>
# **SetAutostopInterval**
> Sandbox SetAutostopInterval (string sandboxIdOrName, decimal interval, string? xDaytonaOrganizationID = null)

Set sandbox auto-stop interval

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
    public class SetAutostopIntervalExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var sandboxIdOrName = "sandboxIdOrName_example";  // string | ID or name of the sandbox
            var interval = 8.14D;  // decimal | Auto-stop interval in minutes (0 to disable)
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Set sandbox auto-stop interval
                Sandbox result = apiInstance.SetAutostopInterval(sandboxIdOrName, interval, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.SetAutostopInterval: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the SetAutostopIntervalWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Set sandbox auto-stop interval
    ApiResponse<Sandbox> response = apiInstance.SetAutostopIntervalWithHttpInfo(sandboxIdOrName, interval, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.SetAutostopIntervalWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxIdOrName** | **string** | ID or name of the sandbox |  |
| **interval** | **decimal** | Auto-stop interval in minutes (0 to disable) |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="startsandbox"></a>
# **StartSandbox**
> Sandbox StartSandbox (string sandboxIdOrName, string? xDaytonaOrganizationID = null)

Start sandbox

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
    public class StartSandboxExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var sandboxIdOrName = "sandboxIdOrName_example";  // string | ID or name of the sandbox
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Start sandbox
                Sandbox result = apiInstance.StartSandbox(sandboxIdOrName, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.StartSandbox: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the StartSandboxWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Start sandbox
    ApiResponse<Sandbox> response = apiInstance.StartSandboxWithHttpInfo(sandboxIdOrName, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.StartSandboxWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxIdOrName** | **string** | ID or name of the sandbox |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="stopsandbox"></a>
# **StopSandbox**
> Sandbox StopSandbox (string sandboxIdOrName, string? xDaytonaOrganizationID = null, bool? force = null)

Stop sandbox

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
    public class StopSandboxExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var sandboxIdOrName = "sandboxIdOrName_example";  // string | ID or name of the sandbox
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 
            var force = true;  // bool? | Force stop the sandbox using SIGKILL instead of SIGTERM (optional) 

            try
            {
                // Stop sandbox
                Sandbox result = apiInstance.StopSandbox(sandboxIdOrName, xDaytonaOrganizationID, force);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.StopSandbox: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the StopSandboxWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Stop sandbox
    ApiResponse<Sandbox> response = apiInstance.StopSandboxWithHttpInfo(sandboxIdOrName, xDaytonaOrganizationID, force);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.StopSandboxWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxIdOrName** | **string** | ID or name of the sandbox |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |
| **force** | **bool?** | Force stop the sandbox using SIGKILL instead of SIGTERM | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="updatelastactivity"></a>
# **UpdateLastActivity**
> void UpdateLastActivity (string sandboxId, string? xDaytonaOrganizationID = null)

Update sandbox last activity

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
    public class UpdateLastActivityExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | ID of the sandbox
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Update sandbox last activity
                apiInstance.UpdateLastActivity(sandboxId, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.UpdateLastActivity: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UpdateLastActivityWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Update sandbox last activity
    apiInstance.UpdateLastActivityWithHttpInfo(sandboxId, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.UpdateLastActivityWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** | ID of the sandbox |  |
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
| **201** | Last activity has been updated |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="updatepublicstatus"></a>
# **UpdatePublicStatus**
> Sandbox UpdatePublicStatus (string sandboxIdOrName, bool isPublic, string? xDaytonaOrganizationID = null)

Update public status

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
    public class UpdatePublicStatusExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var sandboxIdOrName = "sandboxIdOrName_example";  // string | ID or name of the sandbox
            var isPublic = true;  // bool | Public status to set
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Update public status
                Sandbox result = apiInstance.UpdatePublicStatus(sandboxIdOrName, isPublic, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.UpdatePublicStatus: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UpdatePublicStatusWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Update public status
    ApiResponse<Sandbox> response = apiInstance.UpdatePublicStatusWithHttpInfo(sandboxIdOrName, isPublic, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.UpdatePublicStatusWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxIdOrName** | **string** | ID or name of the sandbox |  |
| **isPublic** | **bool** | Public status to set |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="updatesandboxstate"></a>
# **UpdateSandboxState**
> void UpdateSandboxState (string sandboxId, UpdateSandboxStateDto updateSandboxStateDto, string? xDaytonaOrganizationID = null)

Update sandbox state

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
    public class UpdateSandboxStateExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | ID of the sandbox
            var updateSandboxStateDto = new UpdateSandboxStateDto(); // UpdateSandboxStateDto | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Update sandbox state
                apiInstance.UpdateSandboxState(sandboxId, updateSandboxStateDto, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.UpdateSandboxState: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UpdateSandboxStateWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Update sandbox state
    apiInstance.UpdateSandboxStateWithHttpInfo(sandboxId, updateSandboxStateDto, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.UpdateSandboxStateWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** | ID of the sandbox |  |
| **updateSandboxStateDto** | [**UpdateSandboxStateDto**](UpdateSandboxStateDto.md) |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

void (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Sandbox state has been successfully updated |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="validatesshaccess"></a>
# **ValidateSshAccess**
> SshAccessValidationDto ValidateSshAccess (string token, string? xDaytonaOrganizationID = null)

Validate SSH access for sandbox

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
    public class ValidateSshAccessExample
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
            var apiInstance = new SandboxApi(httpClient, config, httpClientHandler);
            var token = "token_example";  // string | SSH access token to validate
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Validate SSH access for sandbox
                SshAccessValidationDto result = apiInstance.ValidateSshAccess(token, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling SandboxApi.ValidateSshAccess: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ValidateSshAccessWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Validate SSH access for sandbox
    ApiResponse<SshAccessValidationDto> response = apiInstance.ValidateSshAccessWithHttpInfo(token, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling SandboxApi.ValidateSshAccessWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **token** | **string** | SSH access token to validate |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

