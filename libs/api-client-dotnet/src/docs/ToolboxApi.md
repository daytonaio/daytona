# Daytona.ApiClient.Api.ToolboxApi

All URIs are relative to *http://localhost:3000*

| Method | HTTP request | Description |
|--------|--------------|-------------|
| [**ClickMouseDeprecated**](ToolboxApi.md#clickmousedeprecated) | **POST** /toolbox/{sandboxId}/toolbox/computeruse/mouse/click | [DEPRECATED] Click mouse |
| [**CreateFolderDeprecated**](ToolboxApi.md#createfolderdeprecated) | **POST** /toolbox/{sandboxId}/toolbox/files/folder | [DEPRECATED] Create folder |
| [**CreatePTYSessionDeprecated**](ToolboxApi.md#createptysessiondeprecated) | **POST** /toolbox/{sandboxId}/toolbox/process/pty | [DEPRECATED] Create PTY session |
| [**CreateSessionDeprecated**](ToolboxApi.md#createsessiondeprecated) | **POST** /toolbox/{sandboxId}/toolbox/process/session | [DEPRECATED] Create session |
| [**DeleteFileDeprecated**](ToolboxApi.md#deletefiledeprecated) | **DELETE** /toolbox/{sandboxId}/toolbox/files | [DEPRECATED] Delete file |
| [**DeletePTYSessionDeprecated**](ToolboxApi.md#deleteptysessiondeprecated) | **DELETE** /toolbox/{sandboxId}/toolbox/process/pty/{sessionId} | [DEPRECATED] Delete PTY session |
| [**DeleteSessionDeprecated**](ToolboxApi.md#deletesessiondeprecated) | **DELETE** /toolbox/{sandboxId}/toolbox/process/session/{sessionId} | [DEPRECATED] Delete session |
| [**DownloadFileDeprecated**](ToolboxApi.md#downloadfiledeprecated) | **GET** /toolbox/{sandboxId}/toolbox/files/download | [DEPRECATED] Download file |
| [**DownloadFilesDeprecated**](ToolboxApi.md#downloadfilesdeprecated) | **POST** /toolbox/{sandboxId}/toolbox/files/bulk-download | [DEPRECATED] Download multiple files |
| [**DragMouseDeprecated**](ToolboxApi.md#dragmousedeprecated) | **POST** /toolbox/{sandboxId}/toolbox/computeruse/mouse/drag | [DEPRECATED] Drag mouse |
| [**ExecuteCommandDeprecated**](ToolboxApi.md#executecommanddeprecated) | **POST** /toolbox/{sandboxId}/toolbox/process/execute | [DEPRECATED] Execute command |
| [**ExecuteSessionCommandDeprecated**](ToolboxApi.md#executesessioncommanddeprecated) | **POST** /toolbox/{sandboxId}/toolbox/process/session/{sessionId}/exec | [DEPRECATED] Execute command in session |
| [**FindInFilesDeprecated**](ToolboxApi.md#findinfilesdeprecated) | **GET** /toolbox/{sandboxId}/toolbox/files/find | [DEPRECATED] Search for text/pattern in files |
| [**GetComputerUseStatusDeprecated**](ToolboxApi.md#getcomputerusestatusdeprecated) | **GET** /toolbox/{sandboxId}/toolbox/computeruse/status | [DEPRECATED] Get computer use status |
| [**GetDisplayInfoDeprecated**](ToolboxApi.md#getdisplayinfodeprecated) | **GET** /toolbox/{sandboxId}/toolbox/computeruse/display/info | [DEPRECATED] Get display info |
| [**GetFileInfoDeprecated**](ToolboxApi.md#getfileinfodeprecated) | **GET** /toolbox/{sandboxId}/toolbox/files/info | [DEPRECATED] Get file info |
| [**GetMousePositionDeprecated**](ToolboxApi.md#getmousepositiondeprecated) | **GET** /toolbox/{sandboxId}/toolbox/computeruse/mouse/position | [DEPRECATED] Get mouse position |
| [**GetPTYSessionDeprecated**](ToolboxApi.md#getptysessiondeprecated) | **GET** /toolbox/{sandboxId}/toolbox/process/pty/{sessionId} | [DEPRECATED] Get PTY session |
| [**GetProcessErrorsDeprecated**](ToolboxApi.md#getprocesserrorsdeprecated) | **GET** /toolbox/{sandboxId}/toolbox/computeruse/process/{processName}/errors | [DEPRECATED] Get process errors |
| [**GetProcessLogsDeprecated**](ToolboxApi.md#getprocesslogsdeprecated) | **GET** /toolbox/{sandboxId}/toolbox/computeruse/process/{processName}/logs | [DEPRECATED] Get process logs |
| [**GetProcessStatusDeprecated**](ToolboxApi.md#getprocessstatusdeprecated) | **GET** /toolbox/{sandboxId}/toolbox/computeruse/process/{processName}/status | [DEPRECATED] Get process status |
| [**GetProjectDirDeprecated**](ToolboxApi.md#getprojectdirdeprecated) | **GET** /toolbox/{sandboxId}/toolbox/project-dir | [DEPRECATED] Get sandbox project dir |
| [**GetSessionCommandDeprecated**](ToolboxApi.md#getsessioncommanddeprecated) | **GET** /toolbox/{sandboxId}/toolbox/process/session/{sessionId}/command/{commandId} | [DEPRECATED] Get session command |
| [**GetSessionCommandLogsDeprecated**](ToolboxApi.md#getsessioncommandlogsdeprecated) | **GET** /toolbox/{sandboxId}/toolbox/process/session/{sessionId}/command/{commandId}/logs | [DEPRECATED] Get command logs |
| [**GetSessionDeprecated**](ToolboxApi.md#getsessiondeprecated) | **GET** /toolbox/{sandboxId}/toolbox/process/session/{sessionId} | [DEPRECATED] Get session |
| [**GetUserHomeDirDeprecated**](ToolboxApi.md#getuserhomedirdeprecated) | **GET** /toolbox/{sandboxId}/toolbox/user-home-dir | [DEPRECATED] Get sandbox user home dir |
| [**GetWindowsDeprecated**](ToolboxApi.md#getwindowsdeprecated) | **GET** /toolbox/{sandboxId}/toolbox/computeruse/display/windows | [DEPRECATED] Get windows |
| [**GetWorkDirDeprecated**](ToolboxApi.md#getworkdirdeprecated) | **GET** /toolbox/{sandboxId}/toolbox/work-dir | [DEPRECATED] Get sandbox work-dir |
| [**GitAddFilesDeprecated**](ToolboxApi.md#gitaddfilesdeprecated) | **POST** /toolbox/{sandboxId}/toolbox/git/add | [DEPRECATED] Add files |
| [**GitCheckoutBranchDeprecated**](ToolboxApi.md#gitcheckoutbranchdeprecated) | **POST** /toolbox/{sandboxId}/toolbox/git/checkout | [DEPRECATED] Checkout branch |
| [**GitCloneRepositoryDeprecated**](ToolboxApi.md#gitclonerepositorydeprecated) | **POST** /toolbox/{sandboxId}/toolbox/git/clone | [DEPRECATED] Clone repository |
| [**GitCommitChangesDeprecated**](ToolboxApi.md#gitcommitchangesdeprecated) | **POST** /toolbox/{sandboxId}/toolbox/git/commit | [DEPRECATED] Commit changes |
| [**GitCreateBranchDeprecated**](ToolboxApi.md#gitcreatebranchdeprecated) | **POST** /toolbox/{sandboxId}/toolbox/git/branches | [DEPRECATED] Create branch |
| [**GitDeleteBranchDeprecated**](ToolboxApi.md#gitdeletebranchdeprecated) | **DELETE** /toolbox/{sandboxId}/toolbox/git/branches | [DEPRECATED] Delete branch |
| [**GitGetHistoryDeprecated**](ToolboxApi.md#gitgethistorydeprecated) | **GET** /toolbox/{sandboxId}/toolbox/git/history | [DEPRECATED] Get commit history |
| [**GitGetStatusDeprecated**](ToolboxApi.md#gitgetstatusdeprecated) | **GET** /toolbox/{sandboxId}/toolbox/git/status | [DEPRECATED] Get git status |
| [**GitListBranchesDeprecated**](ToolboxApi.md#gitlistbranchesdeprecated) | **GET** /toolbox/{sandboxId}/toolbox/git/branches | [DEPRECATED] Get branch list |
| [**GitPullChangesDeprecated**](ToolboxApi.md#gitpullchangesdeprecated) | **POST** /toolbox/{sandboxId}/toolbox/git/pull | [DEPRECATED] Pull changes |
| [**GitPushChangesDeprecated**](ToolboxApi.md#gitpushchangesdeprecated) | **POST** /toolbox/{sandboxId}/toolbox/git/push | [DEPRECATED] Push changes |
| [**ListFilesDeprecated**](ToolboxApi.md#listfilesdeprecated) | **GET** /toolbox/{sandboxId}/toolbox/files | [DEPRECATED] List files |
| [**ListPTYSessionsDeprecated**](ToolboxApi.md#listptysessionsdeprecated) | **GET** /toolbox/{sandboxId}/toolbox/process/pty | [DEPRECATED] List PTY sessions |
| [**ListSessionsDeprecated**](ToolboxApi.md#listsessionsdeprecated) | **GET** /toolbox/{sandboxId}/toolbox/process/session | [DEPRECATED] List sessions |
| [**LspCompletionsDeprecated**](ToolboxApi.md#lspcompletionsdeprecated) | **POST** /toolbox/{sandboxId}/toolbox/lsp/completions | [DEPRECATED] Get Lsp Completions |
| [**LspDidCloseDeprecated**](ToolboxApi.md#lspdidclosedeprecated) | **POST** /toolbox/{sandboxId}/toolbox/lsp/did-close | [DEPRECATED] Call Lsp DidClose |
| [**LspDidOpenDeprecated**](ToolboxApi.md#lspdidopendeprecated) | **POST** /toolbox/{sandboxId}/toolbox/lsp/did-open | [DEPRECATED] Call Lsp DidOpen |
| [**LspDocumentSymbolsDeprecated**](ToolboxApi.md#lspdocumentsymbolsdeprecated) | **GET** /toolbox/{sandboxId}/toolbox/lsp/document-symbols | [DEPRECATED] Call Lsp DocumentSymbols |
| [**LspStartDeprecated**](ToolboxApi.md#lspstartdeprecated) | **POST** /toolbox/{sandboxId}/toolbox/lsp/start | [DEPRECATED] Start Lsp server |
| [**LspStopDeprecated**](ToolboxApi.md#lspstopdeprecated) | **POST** /toolbox/{sandboxId}/toolbox/lsp/stop | [DEPRECATED] Stop Lsp server |
| [**LspWorkspaceSymbolsDeprecated**](ToolboxApi.md#lspworkspacesymbolsdeprecated) | **GET** /toolbox/{sandboxId}/toolbox/lsp/workspace-symbols | [DEPRECATED] Call Lsp WorkspaceSymbols |
| [**MoveFileDeprecated**](ToolboxApi.md#movefiledeprecated) | **POST** /toolbox/{sandboxId}/toolbox/files/move | [DEPRECATED] Move file |
| [**MoveMouseDeprecated**](ToolboxApi.md#movemousedeprecated) | **POST** /toolbox/{sandboxId}/toolbox/computeruse/mouse/move | [DEPRECATED] Move mouse |
| [**PressHotkeyDeprecated**](ToolboxApi.md#presshotkeydeprecated) | **POST** /toolbox/{sandboxId}/toolbox/computeruse/keyboard/hotkey | [DEPRECATED] Press hotkey |
| [**PressKeyDeprecated**](ToolboxApi.md#presskeydeprecated) | **POST** /toolbox/{sandboxId}/toolbox/computeruse/keyboard/key | [DEPRECATED] Press key |
| [**ReplaceInFilesDeprecated**](ToolboxApi.md#replaceinfilesdeprecated) | **POST** /toolbox/{sandboxId}/toolbox/files/replace | [DEPRECATED] Replace in files |
| [**ResizePTYSessionDeprecated**](ToolboxApi.md#resizeptysessiondeprecated) | **POST** /toolbox/{sandboxId}/toolbox/process/pty/{sessionId}/resize | [DEPRECATED] Resize PTY session |
| [**RestartProcessDeprecated**](ToolboxApi.md#restartprocessdeprecated) | **POST** /toolbox/{sandboxId}/toolbox/computeruse/process/{processName}/restart | [DEPRECATED] Restart process |
| [**ScrollMouseDeprecated**](ToolboxApi.md#scrollmousedeprecated) | **POST** /toolbox/{sandboxId}/toolbox/computeruse/mouse/scroll | [DEPRECATED] Scroll mouse |
| [**SearchFilesDeprecated**](ToolboxApi.md#searchfilesdeprecated) | **GET** /toolbox/{sandboxId}/toolbox/files/search | [DEPRECATED] Search files |
| [**SetFilePermissionsDeprecated**](ToolboxApi.md#setfilepermissionsdeprecated) | **POST** /toolbox/{sandboxId}/toolbox/files/permissions | [DEPRECATED] Set file permissions |
| [**StartComputerUseDeprecated**](ToolboxApi.md#startcomputerusedeprecated) | **POST** /toolbox/{sandboxId}/toolbox/computeruse/start | [DEPRECATED] Start computer use processes |
| [**StopComputerUseDeprecated**](ToolboxApi.md#stopcomputerusedeprecated) | **POST** /toolbox/{sandboxId}/toolbox/computeruse/stop | [DEPRECATED] Stop computer use processes |
| [**TakeCompressedRegionScreenshotDeprecated**](ToolboxApi.md#takecompressedregionscreenshotdeprecated) | **GET** /toolbox/{sandboxId}/toolbox/computeruse/screenshot/region/compressed | [DEPRECATED] Take compressed region screenshot |
| [**TakeCompressedScreenshotDeprecated**](ToolboxApi.md#takecompressedscreenshotdeprecated) | **GET** /toolbox/{sandboxId}/toolbox/computeruse/screenshot/compressed | [DEPRECATED] Take compressed screenshot |
| [**TakeRegionScreenshotDeprecated**](ToolboxApi.md#takeregionscreenshotdeprecated) | **GET** /toolbox/{sandboxId}/toolbox/computeruse/screenshot/region | [DEPRECATED] Take region screenshot |
| [**TakeScreenshotDeprecated**](ToolboxApi.md#takescreenshotdeprecated) | **GET** /toolbox/{sandboxId}/toolbox/computeruse/screenshot | [DEPRECATED] Take screenshot |
| [**TypeTextDeprecated**](ToolboxApi.md#typetextdeprecated) | **POST** /toolbox/{sandboxId}/toolbox/computeruse/keyboard/type | [DEPRECATED] Type text |
| [**UploadFileDeprecated**](ToolboxApi.md#uploadfiledeprecated) | **POST** /toolbox/{sandboxId}/toolbox/files/upload | [DEPRECATED] Upload file |
| [**UploadFilesDeprecated**](ToolboxApi.md#uploadfilesdeprecated) | **POST** /toolbox/{sandboxId}/toolbox/files/bulk-upload | [DEPRECATED] Upload multiple files |

<a id="clickmousedeprecated"></a>
# **ClickMouseDeprecated**
> MouseClickResponse ClickMouseDeprecated (string sandboxId, MouseClickRequest mouseClickRequest, string? xDaytonaOrganizationID = null)

[DEPRECATED] Click mouse

Click mouse at specified coordinates

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
    public class ClickMouseDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var mouseClickRequest = new MouseClickRequest(); // MouseClickRequest | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Click mouse
                MouseClickResponse result = apiInstance.ClickMouseDeprecated(sandboxId, mouseClickRequest, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.ClickMouseDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ClickMouseDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Click mouse
    ApiResponse<MouseClickResponse> response = apiInstance.ClickMouseDeprecatedWithHttpInfo(sandboxId, mouseClickRequest, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.ClickMouseDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **mouseClickRequest** | [**MouseClickRequest**](MouseClickRequest.md) |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**MouseClickResponse**](MouseClickResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Mouse clicked successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="createfolderdeprecated"></a>
# **CreateFolderDeprecated**
> void CreateFolderDeprecated (string sandboxId, string path, string mode, string? xDaytonaOrganizationID = null)

[DEPRECATED] Create folder

Create folder inside sandbox

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
    public class CreateFolderDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var path = "path_example";  // string | 
            var mode = "mode_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Create folder
                apiInstance.CreateFolderDeprecated(sandboxId, path, mode, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.CreateFolderDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the CreateFolderDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Create folder
    apiInstance.CreateFolderDeprecatedWithHttpInfo(sandboxId, path, mode, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.CreateFolderDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **path** | **string** |  |  |
| **mode** | **string** |  |  |
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
| **200** | Folder created successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="createptysessiondeprecated"></a>
# **CreatePTYSessionDeprecated**
> PtyCreateResponse CreatePTYSessionDeprecated (string sandboxId, PtyCreateRequest ptyCreateRequest, string? xDaytonaOrganizationID = null)

[DEPRECATED] Create PTY session

Create a new PTY session in the sandbox

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
    public class CreatePTYSessionDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var ptyCreateRequest = new PtyCreateRequest(); // PtyCreateRequest | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Create PTY session
                PtyCreateResponse result = apiInstance.CreatePTYSessionDeprecated(sandboxId, ptyCreateRequest, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.CreatePTYSessionDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the CreatePTYSessionDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Create PTY session
    ApiResponse<PtyCreateResponse> response = apiInstance.CreatePTYSessionDeprecatedWithHttpInfo(sandboxId, ptyCreateRequest, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.CreatePTYSessionDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **ptyCreateRequest** | [**PtyCreateRequest**](PtyCreateRequest.md) |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**PtyCreateResponse**](PtyCreateResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** | PTY session created successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="createsessiondeprecated"></a>
# **CreateSessionDeprecated**
> void CreateSessionDeprecated (string sandboxId, CreateSessionRequest createSessionRequest, string? xDaytonaOrganizationID = null)

[DEPRECATED] Create session

Create a new session in the sandbox

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
    public class CreateSessionDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var createSessionRequest = new CreateSessionRequest(); // CreateSessionRequest | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Create session
                apiInstance.CreateSessionDeprecated(sandboxId, createSessionRequest, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.CreateSessionDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the CreateSessionDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Create session
    apiInstance.CreateSessionDeprecatedWithHttpInfo(sandboxId, createSessionRequest, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.CreateSessionDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **createSessionRequest** | [**CreateSessionRequest**](CreateSessionRequest.md) |  |  |
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
| **200** |  |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="deletefiledeprecated"></a>
# **DeleteFileDeprecated**
> void DeleteFileDeprecated (string sandboxId, string path, string? xDaytonaOrganizationID = null, bool? recursive = null)

[DEPRECATED] Delete file

Delete file inside sandbox

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
    public class DeleteFileDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var path = "path_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 
            var recursive = true;  // bool? |  (optional) 

            try
            {
                // [DEPRECATED] Delete file
                apiInstance.DeleteFileDeprecated(sandboxId, path, xDaytonaOrganizationID, recursive);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.DeleteFileDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DeleteFileDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Delete file
    apiInstance.DeleteFileDeprecatedWithHttpInfo(sandboxId, path, xDaytonaOrganizationID, recursive);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.DeleteFileDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **path** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |
| **recursive** | **bool?** |  | [optional]  |

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
| **200** | File deleted successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="deleteptysessiondeprecated"></a>
# **DeletePTYSessionDeprecated**
> void DeletePTYSessionDeprecated (string sandboxId, string sessionId, string? xDaytonaOrganizationID = null)

[DEPRECATED] Delete PTY session

Delete a PTY session and terminate the associated process

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
    public class DeletePTYSessionDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var sessionId = "sessionId_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Delete PTY session
                apiInstance.DeletePTYSessionDeprecated(sandboxId, sessionId, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.DeletePTYSessionDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DeletePTYSessionDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Delete PTY session
    apiInstance.DeletePTYSessionDeprecatedWithHttpInfo(sandboxId, sessionId, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.DeletePTYSessionDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **sessionId** | **string** |  |  |
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
| **200** | PTY session deleted successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="deletesessiondeprecated"></a>
# **DeleteSessionDeprecated**
> void DeleteSessionDeprecated (string sandboxId, string sessionId, string? xDaytonaOrganizationID = null)

[DEPRECATED] Delete session

Delete a specific session

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
    public class DeleteSessionDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var sessionId = "sessionId_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Delete session
                apiInstance.DeleteSessionDeprecated(sandboxId, sessionId, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.DeleteSessionDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DeleteSessionDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Delete session
    apiInstance.DeleteSessionDeprecatedWithHttpInfo(sandboxId, sessionId, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.DeleteSessionDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **sessionId** | **string** |  |  |
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
| **200** | Session deleted successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="downloadfiledeprecated"></a>
# **DownloadFileDeprecated**
> FileParameter DownloadFileDeprecated (string sandboxId, string path, string? xDaytonaOrganizationID = null)

[DEPRECATED] Download file

Download file from sandbox

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
    public class DownloadFileDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var path = "path_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Download file
                FileParameter result = apiInstance.DownloadFileDeprecated(sandboxId, path, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.DownloadFileDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DownloadFileDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Download file
    ApiResponse<FileParameter> response = apiInstance.DownloadFileDeprecatedWithHttpInfo(sandboxId, path, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.DownloadFileDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **path** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**FileParameter**](FileParameter.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | File downloaded successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="downloadfilesdeprecated"></a>
# **DownloadFilesDeprecated**
> FileParameter DownloadFilesDeprecated (string sandboxId, DownloadFiles downloadFiles, string? xDaytonaOrganizationID = null)

[DEPRECATED] Download multiple files

Streams back a multipart/form-data bundle of the requested paths

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
    public class DownloadFilesDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var downloadFiles = new DownloadFiles(); // DownloadFiles | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Download multiple files
                FileParameter result = apiInstance.DownloadFilesDeprecated(sandboxId, downloadFiles, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.DownloadFilesDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DownloadFilesDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Download multiple files
    ApiResponse<FileParameter> response = apiInstance.DownloadFilesDeprecatedWithHttpInfo(sandboxId, downloadFiles, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.DownloadFilesDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **downloadFiles** | [**DownloadFiles**](DownloadFiles.md) |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**FileParameter**](FileParameter.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | A multipart/form-data response with each file as a part |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="dragmousedeprecated"></a>
# **DragMouseDeprecated**
> MouseDragResponse DragMouseDeprecated (string sandboxId, MouseDragRequest mouseDragRequest, string? xDaytonaOrganizationID = null)

[DEPRECATED] Drag mouse

Drag mouse from start to end coordinates

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
    public class DragMouseDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var mouseDragRequest = new MouseDragRequest(); // MouseDragRequest | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Drag mouse
                MouseDragResponse result = apiInstance.DragMouseDeprecated(sandboxId, mouseDragRequest, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.DragMouseDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DragMouseDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Drag mouse
    ApiResponse<MouseDragResponse> response = apiInstance.DragMouseDeprecatedWithHttpInfo(sandboxId, mouseDragRequest, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.DragMouseDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **mouseDragRequest** | [**MouseDragRequest**](MouseDragRequest.md) |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**MouseDragResponse**](MouseDragResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Mouse dragged successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="executecommanddeprecated"></a>
# **ExecuteCommandDeprecated**
> ExecuteResponse ExecuteCommandDeprecated (string sandboxId, ExecuteRequest executeRequest, string? xDaytonaOrganizationID = null)

[DEPRECATED] Execute command

Execute command synchronously inside sandbox

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
    public class ExecuteCommandDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var executeRequest = new ExecuteRequest(); // ExecuteRequest | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Execute command
                ExecuteResponse result = apiInstance.ExecuteCommandDeprecated(sandboxId, executeRequest, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.ExecuteCommandDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ExecuteCommandDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Execute command
    ApiResponse<ExecuteResponse> response = apiInstance.ExecuteCommandDeprecatedWithHttpInfo(sandboxId, executeRequest, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.ExecuteCommandDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **executeRequest** | [**ExecuteRequest**](ExecuteRequest.md) |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**ExecuteResponse**](ExecuteResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Command executed successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="executesessioncommanddeprecated"></a>
# **ExecuteSessionCommandDeprecated**
> SessionExecuteResponse ExecuteSessionCommandDeprecated (string sandboxId, string sessionId, SessionExecuteRequest sessionExecuteRequest, string? xDaytonaOrganizationID = null)

[DEPRECATED] Execute command in session

Execute a command in a specific session

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
    public class ExecuteSessionCommandDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var sessionId = "sessionId_example";  // string | 
            var sessionExecuteRequest = new SessionExecuteRequest(); // SessionExecuteRequest | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Execute command in session
                SessionExecuteResponse result = apiInstance.ExecuteSessionCommandDeprecated(sandboxId, sessionId, sessionExecuteRequest, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.ExecuteSessionCommandDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ExecuteSessionCommandDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Execute command in session
    ApiResponse<SessionExecuteResponse> response = apiInstance.ExecuteSessionCommandDeprecatedWithHttpInfo(sandboxId, sessionId, sessionExecuteRequest, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.ExecuteSessionCommandDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **sessionId** | **string** |  |  |
| **sessionExecuteRequest** | [**SessionExecuteRequest**](SessionExecuteRequest.md) |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**SessionExecuteResponse**](SessionExecuteResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Command executed successfully |  -  |
| **202** | Command accepted and is being processed |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="findinfilesdeprecated"></a>
# **FindInFilesDeprecated**
> List&lt;Match&gt; FindInFilesDeprecated (string sandboxId, string path, string pattern, string? xDaytonaOrganizationID = null)

[DEPRECATED] Search for text/pattern in files

Search for text/pattern inside sandbox files

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
    public class FindInFilesDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var path = "path_example";  // string | 
            var pattern = "pattern_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Search for text/pattern in files
                List<Match> result = apiInstance.FindInFilesDeprecated(sandboxId, path, pattern, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.FindInFilesDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the FindInFilesDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Search for text/pattern in files
    ApiResponse<List<Match>> response = apiInstance.FindInFilesDeprecatedWithHttpInfo(sandboxId, path, pattern, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.FindInFilesDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **path** | **string** |  |  |
| **pattern** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**List&lt;Match&gt;**](Match.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Search completed successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getcomputerusestatusdeprecated"></a>
# **GetComputerUseStatusDeprecated**
> ComputerUseStatusResponse GetComputerUseStatusDeprecated (string sandboxId, string? xDaytonaOrganizationID = null)

[DEPRECATED] Get computer use status

Get status of all VNC desktop processes

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
    public class GetComputerUseStatusDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Get computer use status
                ComputerUseStatusResponse result = apiInstance.GetComputerUseStatusDeprecated(sandboxId, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.GetComputerUseStatusDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetComputerUseStatusDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Get computer use status
    ApiResponse<ComputerUseStatusResponse> response = apiInstance.GetComputerUseStatusDeprecatedWithHttpInfo(sandboxId, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.GetComputerUseStatusDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**ComputerUseStatusResponse**](ComputerUseStatusResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Computer use status retrieved successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getdisplayinfodeprecated"></a>
# **GetDisplayInfoDeprecated**
> DisplayInfoResponse GetDisplayInfoDeprecated (string sandboxId, string? xDaytonaOrganizationID = null)

[DEPRECATED] Get display info

Get information about displays

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
    public class GetDisplayInfoDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Get display info
                DisplayInfoResponse result = apiInstance.GetDisplayInfoDeprecated(sandboxId, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.GetDisplayInfoDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetDisplayInfoDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Get display info
    ApiResponse<DisplayInfoResponse> response = apiInstance.GetDisplayInfoDeprecatedWithHttpInfo(sandboxId, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.GetDisplayInfoDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**DisplayInfoResponse**](DisplayInfoResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Display info retrieved successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getfileinfodeprecated"></a>
# **GetFileInfoDeprecated**
> FileInfo GetFileInfoDeprecated (string sandboxId, string path, string? xDaytonaOrganizationID = null)

[DEPRECATED] Get file info

Get file info inside sandbox

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
    public class GetFileInfoDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var path = "path_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Get file info
                FileInfo result = apiInstance.GetFileInfoDeprecated(sandboxId, path, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.GetFileInfoDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetFileInfoDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Get file info
    ApiResponse<FileInfo> response = apiInstance.GetFileInfoDeprecatedWithHttpInfo(sandboxId, path, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.GetFileInfoDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **path** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**FileInfo**](FileInfo.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | File info retrieved successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getmousepositiondeprecated"></a>
# **GetMousePositionDeprecated**
> MousePosition GetMousePositionDeprecated (string sandboxId, string? xDaytonaOrganizationID = null)

[DEPRECATED] Get mouse position

Get current mouse cursor position

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
    public class GetMousePositionDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Get mouse position
                MousePosition result = apiInstance.GetMousePositionDeprecated(sandboxId, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.GetMousePositionDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetMousePositionDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Get mouse position
    ApiResponse<MousePosition> response = apiInstance.GetMousePositionDeprecatedWithHttpInfo(sandboxId, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.GetMousePositionDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**MousePosition**](MousePosition.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Mouse position retrieved successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getptysessiondeprecated"></a>
# **GetPTYSessionDeprecated**
> PtySessionInfo GetPTYSessionDeprecated (string sandboxId, string sessionId, string? xDaytonaOrganizationID = null)

[DEPRECATED] Get PTY session

Get PTY session information by ID

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
    public class GetPTYSessionDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var sessionId = "sessionId_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Get PTY session
                PtySessionInfo result = apiInstance.GetPTYSessionDeprecated(sandboxId, sessionId, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.GetPTYSessionDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetPTYSessionDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Get PTY session
    ApiResponse<PtySessionInfo> response = apiInstance.GetPTYSessionDeprecatedWithHttpInfo(sandboxId, sessionId, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.GetPTYSessionDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **sessionId** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**PtySessionInfo**](PtySessionInfo.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | PTY session retrieved successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getprocesserrorsdeprecated"></a>
# **GetProcessErrorsDeprecated**
> ProcessErrorsResponse GetProcessErrorsDeprecated (string processName, string sandboxId, string? xDaytonaOrganizationID = null)

[DEPRECATED] Get process errors

Get error logs for a specific VNC process

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
    public class GetProcessErrorsDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var processName = "processName_example";  // string | 
            var sandboxId = "sandboxId_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Get process errors
                ProcessErrorsResponse result = apiInstance.GetProcessErrorsDeprecated(processName, sandboxId, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.GetProcessErrorsDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetProcessErrorsDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Get process errors
    ApiResponse<ProcessErrorsResponse> response = apiInstance.GetProcessErrorsDeprecatedWithHttpInfo(processName, sandboxId, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.GetProcessErrorsDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **processName** | **string** |  |  |
| **sandboxId** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**ProcessErrorsResponse**](ProcessErrorsResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Process errors retrieved successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getprocesslogsdeprecated"></a>
# **GetProcessLogsDeprecated**
> ProcessLogsResponse GetProcessLogsDeprecated (string processName, string sandboxId, string? xDaytonaOrganizationID = null)

[DEPRECATED] Get process logs

Get logs for a specific VNC process

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
    public class GetProcessLogsDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var processName = "processName_example";  // string | 
            var sandboxId = "sandboxId_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Get process logs
                ProcessLogsResponse result = apiInstance.GetProcessLogsDeprecated(processName, sandboxId, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.GetProcessLogsDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetProcessLogsDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Get process logs
    ApiResponse<ProcessLogsResponse> response = apiInstance.GetProcessLogsDeprecatedWithHttpInfo(processName, sandboxId, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.GetProcessLogsDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **processName** | **string** |  |  |
| **sandboxId** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**ProcessLogsResponse**](ProcessLogsResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Process logs retrieved successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getprocessstatusdeprecated"></a>
# **GetProcessStatusDeprecated**
> ProcessStatusResponse GetProcessStatusDeprecated (string processName, string sandboxId, string? xDaytonaOrganizationID = null)

[DEPRECATED] Get process status

Get status of a specific VNC process

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
    public class GetProcessStatusDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var processName = "processName_example";  // string | 
            var sandboxId = "sandboxId_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Get process status
                ProcessStatusResponse result = apiInstance.GetProcessStatusDeprecated(processName, sandboxId, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.GetProcessStatusDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetProcessStatusDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Get process status
    ApiResponse<ProcessStatusResponse> response = apiInstance.GetProcessStatusDeprecatedWithHttpInfo(processName, sandboxId, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.GetProcessStatusDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **processName** | **string** |  |  |
| **sandboxId** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**ProcessStatusResponse**](ProcessStatusResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Process status retrieved successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getprojectdirdeprecated"></a>
# **GetProjectDirDeprecated**
> ProjectDirResponse GetProjectDirDeprecated (string sandboxId, string? xDaytonaOrganizationID = null)

[DEPRECATED] Get sandbox project dir

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
    public class GetProjectDirDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Get sandbox project dir
                ProjectDirResponse result = apiInstance.GetProjectDirDeprecated(sandboxId, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.GetProjectDirDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetProjectDirDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Get sandbox project dir
    ApiResponse<ProjectDirResponse> response = apiInstance.GetProjectDirDeprecatedWithHttpInfo(sandboxId, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.GetProjectDirDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**ProjectDirResponse**](ProjectDirResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Project directory retrieved successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getsessioncommanddeprecated"></a>
# **GetSessionCommandDeprecated**
> Command GetSessionCommandDeprecated (string sandboxId, string sessionId, string commandId, string? xDaytonaOrganizationID = null)

[DEPRECATED] Get session command

Get session command by ID

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
    public class GetSessionCommandDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var sessionId = "sessionId_example";  // string | 
            var commandId = "commandId_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Get session command
                Command result = apiInstance.GetSessionCommandDeprecated(sandboxId, sessionId, commandId, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.GetSessionCommandDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetSessionCommandDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Get session command
    ApiResponse<Command> response = apiInstance.GetSessionCommandDeprecatedWithHttpInfo(sandboxId, sessionId, commandId, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.GetSessionCommandDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **sessionId** | **string** |  |  |
| **commandId** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**Command**](Command.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Session command retrieved successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getsessioncommandlogsdeprecated"></a>
# **GetSessionCommandLogsDeprecated**
> string GetSessionCommandLogsDeprecated (string sandboxId, string sessionId, string commandId, string? xDaytonaOrganizationID = null, bool? follow = null)

[DEPRECATED] Get command logs

Get logs for a specific command in a session

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
    public class GetSessionCommandLogsDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var sessionId = "sessionId_example";  // string | 
            var commandId = "commandId_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 
            var follow = true;  // bool? | Whether to stream the logs (optional) 

            try
            {
                // [DEPRECATED] Get command logs
                string result = apiInstance.GetSessionCommandLogsDeprecated(sandboxId, sessionId, commandId, xDaytonaOrganizationID, follow);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.GetSessionCommandLogsDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetSessionCommandLogsDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Get command logs
    ApiResponse<string> response = apiInstance.GetSessionCommandLogsDeprecatedWithHttpInfo(sandboxId, sessionId, commandId, xDaytonaOrganizationID, follow);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.GetSessionCommandLogsDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **sessionId** | **string** |  |  |
| **commandId** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |
| **follow** | **bool?** | Whether to stream the logs | [optional]  |

### Return type

**string**

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: text/plain


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Command log stream marked with stdout and stderr prefixes |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getsessiondeprecated"></a>
# **GetSessionDeprecated**
> Session GetSessionDeprecated (string sandboxId, string sessionId, string? xDaytonaOrganizationID = null)

[DEPRECATED] Get session

Get session by ID

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
    public class GetSessionDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var sessionId = "sessionId_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Get session
                Session result = apiInstance.GetSessionDeprecated(sandboxId, sessionId, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.GetSessionDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetSessionDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Get session
    ApiResponse<Session> response = apiInstance.GetSessionDeprecatedWithHttpInfo(sandboxId, sessionId, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.GetSessionDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **sessionId** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**Session**](Session.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Session retrieved successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getuserhomedirdeprecated"></a>
# **GetUserHomeDirDeprecated**
> UserHomeDirResponse GetUserHomeDirDeprecated (string sandboxId, string? xDaytonaOrganizationID = null)

[DEPRECATED] Get sandbox user home dir

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
    public class GetUserHomeDirDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Get sandbox user home dir
                UserHomeDirResponse result = apiInstance.GetUserHomeDirDeprecated(sandboxId, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.GetUserHomeDirDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetUserHomeDirDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Get sandbox user home dir
    ApiResponse<UserHomeDirResponse> response = apiInstance.GetUserHomeDirDeprecatedWithHttpInfo(sandboxId, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.GetUserHomeDirDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**UserHomeDirResponse**](UserHomeDirResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | User home directory retrieved successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getwindowsdeprecated"></a>
# **GetWindowsDeprecated**
> WindowsResponse GetWindowsDeprecated (string sandboxId, string? xDaytonaOrganizationID = null)

[DEPRECATED] Get windows

Get list of open windows

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
    public class GetWindowsDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Get windows
                WindowsResponse result = apiInstance.GetWindowsDeprecated(sandboxId, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.GetWindowsDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetWindowsDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Get windows
    ApiResponse<WindowsResponse> response = apiInstance.GetWindowsDeprecatedWithHttpInfo(sandboxId, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.GetWindowsDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**WindowsResponse**](WindowsResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Windows list retrieved successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getworkdirdeprecated"></a>
# **GetWorkDirDeprecated**
> WorkDirResponse GetWorkDirDeprecated (string sandboxId, string? xDaytonaOrganizationID = null)

[DEPRECATED] Get sandbox work-dir

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
    public class GetWorkDirDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Get sandbox work-dir
                WorkDirResponse result = apiInstance.GetWorkDirDeprecated(sandboxId, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.GetWorkDirDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetWorkDirDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Get sandbox work-dir
    ApiResponse<WorkDirResponse> response = apiInstance.GetWorkDirDeprecatedWithHttpInfo(sandboxId, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.GetWorkDirDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**WorkDirResponse**](WorkDirResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Work-dir retrieved successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="gitaddfilesdeprecated"></a>
# **GitAddFilesDeprecated**
> void GitAddFilesDeprecated (string sandboxId, GitAddRequest gitAddRequest, string? xDaytonaOrganizationID = null)

[DEPRECATED] Add files

Add files to git commit

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
    public class GitAddFilesDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var gitAddRequest = new GitAddRequest(); // GitAddRequest | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Add files
                apiInstance.GitAddFilesDeprecated(sandboxId, gitAddRequest, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.GitAddFilesDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GitAddFilesDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Add files
    apiInstance.GitAddFilesDeprecatedWithHttpInfo(sandboxId, gitAddRequest, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.GitAddFilesDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **gitAddRequest** | [**GitAddRequest**](GitAddRequest.md) |  |  |
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
| **200** | Files added to git successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="gitcheckoutbranchdeprecated"></a>
# **GitCheckoutBranchDeprecated**
> void GitCheckoutBranchDeprecated (string sandboxId, GitCheckoutRequest gitCheckoutRequest, string? xDaytonaOrganizationID = null)

[DEPRECATED] Checkout branch

Checkout branch or commit in git repository

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
    public class GitCheckoutBranchDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var gitCheckoutRequest = new GitCheckoutRequest(); // GitCheckoutRequest | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Checkout branch
                apiInstance.GitCheckoutBranchDeprecated(sandboxId, gitCheckoutRequest, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.GitCheckoutBranchDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GitCheckoutBranchDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Checkout branch
    apiInstance.GitCheckoutBranchDeprecatedWithHttpInfo(sandboxId, gitCheckoutRequest, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.GitCheckoutBranchDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **gitCheckoutRequest** | [**GitCheckoutRequest**](GitCheckoutRequest.md) |  |  |
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
| **200** | Branch checked out successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="gitclonerepositorydeprecated"></a>
# **GitCloneRepositoryDeprecated**
> void GitCloneRepositoryDeprecated (string sandboxId, GitCloneRequest gitCloneRequest, string? xDaytonaOrganizationID = null)

[DEPRECATED] Clone repository

Clone git repository

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
    public class GitCloneRepositoryDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var gitCloneRequest = new GitCloneRequest(); // GitCloneRequest | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Clone repository
                apiInstance.GitCloneRepositoryDeprecated(sandboxId, gitCloneRequest, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.GitCloneRepositoryDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GitCloneRepositoryDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Clone repository
    apiInstance.GitCloneRepositoryDeprecatedWithHttpInfo(sandboxId, gitCloneRequest, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.GitCloneRepositoryDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **gitCloneRequest** | [**GitCloneRequest**](GitCloneRequest.md) |  |  |
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
| **200** | Repository cloned successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="gitcommitchangesdeprecated"></a>
# **GitCommitChangesDeprecated**
> GitCommitResponse GitCommitChangesDeprecated (string sandboxId, GitCommitRequest gitCommitRequest, string? xDaytonaOrganizationID = null)

[DEPRECATED] Commit changes

Commit changes to git repository

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
    public class GitCommitChangesDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var gitCommitRequest = new GitCommitRequest(); // GitCommitRequest | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Commit changes
                GitCommitResponse result = apiInstance.GitCommitChangesDeprecated(sandboxId, gitCommitRequest, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.GitCommitChangesDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GitCommitChangesDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Commit changes
    ApiResponse<GitCommitResponse> response = apiInstance.GitCommitChangesDeprecatedWithHttpInfo(sandboxId, gitCommitRequest, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.GitCommitChangesDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **gitCommitRequest** | [**GitCommitRequest**](GitCommitRequest.md) |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**GitCommitResponse**](GitCommitResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Changes committed successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="gitcreatebranchdeprecated"></a>
# **GitCreateBranchDeprecated**
> void GitCreateBranchDeprecated (string sandboxId, GitBranchRequest gitBranchRequest, string? xDaytonaOrganizationID = null)

[DEPRECATED] Create branch

Create branch on git repository

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
    public class GitCreateBranchDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var gitBranchRequest = new GitBranchRequest(); // GitBranchRequest | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Create branch
                apiInstance.GitCreateBranchDeprecated(sandboxId, gitBranchRequest, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.GitCreateBranchDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GitCreateBranchDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Create branch
    apiInstance.GitCreateBranchDeprecatedWithHttpInfo(sandboxId, gitBranchRequest, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.GitCreateBranchDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **gitBranchRequest** | [**GitBranchRequest**](GitBranchRequest.md) |  |  |
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
| **200** | Branch created successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="gitdeletebranchdeprecated"></a>
# **GitDeleteBranchDeprecated**
> void GitDeleteBranchDeprecated (string sandboxId, GitDeleteBranchRequest gitDeleteBranchRequest, string? xDaytonaOrganizationID = null)

[DEPRECATED] Delete branch

Delete branch on git repository

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
    public class GitDeleteBranchDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var gitDeleteBranchRequest = new GitDeleteBranchRequest(); // GitDeleteBranchRequest | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Delete branch
                apiInstance.GitDeleteBranchDeprecated(sandboxId, gitDeleteBranchRequest, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.GitDeleteBranchDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GitDeleteBranchDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Delete branch
    apiInstance.GitDeleteBranchDeprecatedWithHttpInfo(sandboxId, gitDeleteBranchRequest, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.GitDeleteBranchDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **gitDeleteBranchRequest** | [**GitDeleteBranchRequest**](GitDeleteBranchRequest.md) |  |  |
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
| **200** | Branch deleted successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="gitgethistorydeprecated"></a>
# **GitGetHistoryDeprecated**
> List&lt;GitCommitInfo&gt; GitGetHistoryDeprecated (string sandboxId, string path, string? xDaytonaOrganizationID = null)

[DEPRECATED] Get commit history

Get commit history from git repository

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
    public class GitGetHistoryDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var path = "path_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Get commit history
                List<GitCommitInfo> result = apiInstance.GitGetHistoryDeprecated(sandboxId, path, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.GitGetHistoryDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GitGetHistoryDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Get commit history
    ApiResponse<List<GitCommitInfo>> response = apiInstance.GitGetHistoryDeprecatedWithHttpInfo(sandboxId, path, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.GitGetHistoryDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **path** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**List&lt;GitCommitInfo&gt;**](GitCommitInfo.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Commit history retrieved successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="gitgetstatusdeprecated"></a>
# **GitGetStatusDeprecated**
> GitStatus GitGetStatusDeprecated (string sandboxId, string path, string? xDaytonaOrganizationID = null)

[DEPRECATED] Get git status

Get status from git repository

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
    public class GitGetStatusDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var path = "path_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Get git status
                GitStatus result = apiInstance.GitGetStatusDeprecated(sandboxId, path, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.GitGetStatusDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GitGetStatusDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Get git status
    ApiResponse<GitStatus> response = apiInstance.GitGetStatusDeprecatedWithHttpInfo(sandboxId, path, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.GitGetStatusDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **path** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**GitStatus**](GitStatus.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Git status retrieved successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="gitlistbranchesdeprecated"></a>
# **GitListBranchesDeprecated**
> ListBranchResponse GitListBranchesDeprecated (string sandboxId, string path, string? xDaytonaOrganizationID = null)

[DEPRECATED] Get branch list

Get branch list from git repository

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
    public class GitListBranchesDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var path = "path_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Get branch list
                ListBranchResponse result = apiInstance.GitListBranchesDeprecated(sandboxId, path, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.GitListBranchesDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GitListBranchesDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Get branch list
    ApiResponse<ListBranchResponse> response = apiInstance.GitListBranchesDeprecatedWithHttpInfo(sandboxId, path, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.GitListBranchesDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **path** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**ListBranchResponse**](ListBranchResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Branch list retrieved successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="gitpullchangesdeprecated"></a>
# **GitPullChangesDeprecated**
> void GitPullChangesDeprecated (string sandboxId, GitRepoRequest gitRepoRequest, string? xDaytonaOrganizationID = null)

[DEPRECATED] Pull changes

Pull changes from remote

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
    public class GitPullChangesDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var gitRepoRequest = new GitRepoRequest(); // GitRepoRequest | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Pull changes
                apiInstance.GitPullChangesDeprecated(sandboxId, gitRepoRequest, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.GitPullChangesDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GitPullChangesDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Pull changes
    apiInstance.GitPullChangesDeprecatedWithHttpInfo(sandboxId, gitRepoRequest, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.GitPullChangesDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **gitRepoRequest** | [**GitRepoRequest**](GitRepoRequest.md) |  |  |
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
| **200** | Changes pulled successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="gitpushchangesdeprecated"></a>
# **GitPushChangesDeprecated**
> void GitPushChangesDeprecated (string sandboxId, GitRepoRequest gitRepoRequest, string? xDaytonaOrganizationID = null)

[DEPRECATED] Push changes

Push changes to remote

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
    public class GitPushChangesDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var gitRepoRequest = new GitRepoRequest(); // GitRepoRequest | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Push changes
                apiInstance.GitPushChangesDeprecated(sandboxId, gitRepoRequest, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.GitPushChangesDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GitPushChangesDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Push changes
    apiInstance.GitPushChangesDeprecatedWithHttpInfo(sandboxId, gitRepoRequest, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.GitPushChangesDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **gitRepoRequest** | [**GitRepoRequest**](GitRepoRequest.md) |  |  |
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
| **200** | Changes pushed successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="listfilesdeprecated"></a>
# **ListFilesDeprecated**
> List&lt;FileInfo&gt; ListFilesDeprecated (string sandboxId, string? xDaytonaOrganizationID = null, string? path = null)

[DEPRECATED] List files

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
    public class ListFilesDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 
            var path = "path_example";  // string? |  (optional) 

            try
            {
                // [DEPRECATED] List files
                List<FileInfo> result = apiInstance.ListFilesDeprecated(sandboxId, xDaytonaOrganizationID, path);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.ListFilesDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ListFilesDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] List files
    ApiResponse<List<FileInfo>> response = apiInstance.ListFilesDeprecatedWithHttpInfo(sandboxId, xDaytonaOrganizationID, path);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.ListFilesDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |
| **path** | **string?** |  | [optional]  |

### Return type

[**List&lt;FileInfo&gt;**](FileInfo.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Files listed successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="listptysessionsdeprecated"></a>
# **ListPTYSessionsDeprecated**
> PtyListResponse ListPTYSessionsDeprecated (string sandboxId, string? xDaytonaOrganizationID = null)

[DEPRECATED] List PTY sessions

List all active PTY sessions in the sandbox

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
    public class ListPTYSessionsDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] List PTY sessions
                PtyListResponse result = apiInstance.ListPTYSessionsDeprecated(sandboxId, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.ListPTYSessionsDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ListPTYSessionsDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] List PTY sessions
    ApiResponse<PtyListResponse> response = apiInstance.ListPTYSessionsDeprecatedWithHttpInfo(sandboxId, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.ListPTYSessionsDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**PtyListResponse**](PtyListResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | PTY sessions retrieved successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="listsessionsdeprecated"></a>
# **ListSessionsDeprecated**
> List&lt;Session&gt; ListSessionsDeprecated (string sandboxId, string? xDaytonaOrganizationID = null)

[DEPRECATED] List sessions

List all active sessions in the sandbox

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
    public class ListSessionsDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] List sessions
                List<Session> result = apiInstance.ListSessionsDeprecated(sandboxId, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.ListSessionsDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ListSessionsDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] List sessions
    ApiResponse<List<Session>> response = apiInstance.ListSessionsDeprecatedWithHttpInfo(sandboxId, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.ListSessionsDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**List&lt;Session&gt;**](Session.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Sessions retrieved successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="lspcompletionsdeprecated"></a>
# **LspCompletionsDeprecated**
> CompletionList LspCompletionsDeprecated (string sandboxId, LspCompletionParams lspCompletionParams, string? xDaytonaOrganizationID = null)

[DEPRECATED] Get Lsp Completions

The Completion request is sent from the client to the server to compute completion items at a given cursor position.

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
    public class LspCompletionsDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var lspCompletionParams = new LspCompletionParams(); // LspCompletionParams | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Get Lsp Completions
                CompletionList result = apiInstance.LspCompletionsDeprecated(sandboxId, lspCompletionParams, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.LspCompletionsDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the LspCompletionsDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Get Lsp Completions
    ApiResponse<CompletionList> response = apiInstance.LspCompletionsDeprecatedWithHttpInfo(sandboxId, lspCompletionParams, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.LspCompletionsDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **lspCompletionParams** | [**LspCompletionParams**](LspCompletionParams.md) |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**CompletionList**](CompletionList.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="lspdidclosedeprecated"></a>
# **LspDidCloseDeprecated**
> void LspDidCloseDeprecated (string sandboxId, LspDocumentRequest lspDocumentRequest, string? xDaytonaOrganizationID = null)

[DEPRECATED] Call Lsp DidClose

The document close notification is sent from the client to the server when the document got closed in the client.

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
    public class LspDidCloseDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var lspDocumentRequest = new LspDocumentRequest(); // LspDocumentRequest | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Call Lsp DidClose
                apiInstance.LspDidCloseDeprecated(sandboxId, lspDocumentRequest, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.LspDidCloseDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the LspDidCloseDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Call Lsp DidClose
    apiInstance.LspDidCloseDeprecatedWithHttpInfo(sandboxId, lspDocumentRequest, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.LspDidCloseDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **lspDocumentRequest** | [**LspDocumentRequest**](LspDocumentRequest.md) |  |  |
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
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="lspdidopendeprecated"></a>
# **LspDidOpenDeprecated**
> void LspDidOpenDeprecated (string sandboxId, LspDocumentRequest lspDocumentRequest, string? xDaytonaOrganizationID = null)

[DEPRECATED] Call Lsp DidOpen

The document open notification is sent from the client to the server to signal newly opened text documents.

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
    public class LspDidOpenDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var lspDocumentRequest = new LspDocumentRequest(); // LspDocumentRequest | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Call Lsp DidOpen
                apiInstance.LspDidOpenDeprecated(sandboxId, lspDocumentRequest, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.LspDidOpenDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the LspDidOpenDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Call Lsp DidOpen
    apiInstance.LspDidOpenDeprecatedWithHttpInfo(sandboxId, lspDocumentRequest, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.LspDidOpenDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **lspDocumentRequest** | [**LspDocumentRequest**](LspDocumentRequest.md) |  |  |
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
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="lspdocumentsymbolsdeprecated"></a>
# **LspDocumentSymbolsDeprecated**
> List&lt;LspSymbol&gt; LspDocumentSymbolsDeprecated (string sandboxId, string languageId, string pathToProject, string uri, string? xDaytonaOrganizationID = null)

[DEPRECATED] Call Lsp DocumentSymbols

The document symbol request is sent from the client to the server.

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
    public class LspDocumentSymbolsDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var languageId = "languageId_example";  // string | 
            var pathToProject = "pathToProject_example";  // string | 
            var uri = "uri_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Call Lsp DocumentSymbols
                List<LspSymbol> result = apiInstance.LspDocumentSymbolsDeprecated(sandboxId, languageId, pathToProject, uri, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.LspDocumentSymbolsDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the LspDocumentSymbolsDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Call Lsp DocumentSymbols
    ApiResponse<List<LspSymbol>> response = apiInstance.LspDocumentSymbolsDeprecatedWithHttpInfo(sandboxId, languageId, pathToProject, uri, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.LspDocumentSymbolsDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **languageId** | **string** |  |  |
| **pathToProject** | **string** |  |  |
| **uri** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**List&lt;LspSymbol&gt;**](LspSymbol.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="lspstartdeprecated"></a>
# **LspStartDeprecated**
> void LspStartDeprecated (string sandboxId, LspServerRequest lspServerRequest, string? xDaytonaOrganizationID = null)

[DEPRECATED] Start Lsp server

Start Lsp server process inside sandbox project

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
    public class LspStartDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var lspServerRequest = new LspServerRequest(); // LspServerRequest | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Start Lsp server
                apiInstance.LspStartDeprecated(sandboxId, lspServerRequest, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.LspStartDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the LspStartDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Start Lsp server
    apiInstance.LspStartDeprecatedWithHttpInfo(sandboxId, lspServerRequest, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.LspStartDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **lspServerRequest** | [**LspServerRequest**](LspServerRequest.md) |  |  |
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
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="lspstopdeprecated"></a>
# **LspStopDeprecated**
> void LspStopDeprecated (string sandboxId, LspServerRequest lspServerRequest, string? xDaytonaOrganizationID = null)

[DEPRECATED] Stop Lsp server

Stop Lsp server process inside sandbox project

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
    public class LspStopDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var lspServerRequest = new LspServerRequest(); // LspServerRequest | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Stop Lsp server
                apiInstance.LspStopDeprecated(sandboxId, lspServerRequest, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.LspStopDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the LspStopDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Stop Lsp server
    apiInstance.LspStopDeprecatedWithHttpInfo(sandboxId, lspServerRequest, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.LspStopDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **lspServerRequest** | [**LspServerRequest**](LspServerRequest.md) |  |  |
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
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="lspworkspacesymbolsdeprecated"></a>
# **LspWorkspaceSymbolsDeprecated**
> List&lt;LspSymbol&gt; LspWorkspaceSymbolsDeprecated (string sandboxId, string languageId, string pathToProject, string query, string? xDaytonaOrganizationID = null)

[DEPRECATED] Call Lsp WorkspaceSymbols

The workspace symbol request is sent from the client to the server to list project-wide symbols matching the query string.

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
    public class LspWorkspaceSymbolsDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var languageId = "languageId_example";  // string | 
            var pathToProject = "pathToProject_example";  // string | 
            var query = "query_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Call Lsp WorkspaceSymbols
                List<LspSymbol> result = apiInstance.LspWorkspaceSymbolsDeprecated(sandboxId, languageId, pathToProject, query, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.LspWorkspaceSymbolsDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the LspWorkspaceSymbolsDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Call Lsp WorkspaceSymbols
    ApiResponse<List<LspSymbol>> response = apiInstance.LspWorkspaceSymbolsDeprecatedWithHttpInfo(sandboxId, languageId, pathToProject, query, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.LspWorkspaceSymbolsDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **languageId** | **string** |  |  |
| **pathToProject** | **string** |  |  |
| **query** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**List&lt;LspSymbol&gt;**](LspSymbol.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="movefiledeprecated"></a>
# **MoveFileDeprecated**
> void MoveFileDeprecated (string sandboxId, string source, string destination, string? xDaytonaOrganizationID = null)

[DEPRECATED] Move file

Move file inside sandbox

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
    public class MoveFileDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var source = "source_example";  // string | 
            var destination = "destination_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Move file
                apiInstance.MoveFileDeprecated(sandboxId, source, destination, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.MoveFileDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the MoveFileDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Move file
    apiInstance.MoveFileDeprecatedWithHttpInfo(sandboxId, source, destination, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.MoveFileDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **source** | **string** |  |  |
| **destination** | **string** |  |  |
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
| **200** | File moved successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="movemousedeprecated"></a>
# **MoveMouseDeprecated**
> MouseMoveResponse MoveMouseDeprecated (string sandboxId, MouseMoveRequest mouseMoveRequest, string? xDaytonaOrganizationID = null)

[DEPRECATED] Move mouse

Move mouse cursor to specified coordinates

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
    public class MoveMouseDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var mouseMoveRequest = new MouseMoveRequest(); // MouseMoveRequest | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Move mouse
                MouseMoveResponse result = apiInstance.MoveMouseDeprecated(sandboxId, mouseMoveRequest, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.MoveMouseDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the MoveMouseDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Move mouse
    ApiResponse<MouseMoveResponse> response = apiInstance.MoveMouseDeprecatedWithHttpInfo(sandboxId, mouseMoveRequest, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.MoveMouseDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **mouseMoveRequest** | [**MouseMoveRequest**](MouseMoveRequest.md) |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**MouseMoveResponse**](MouseMoveResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Mouse moved successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="presshotkeydeprecated"></a>
# **PressHotkeyDeprecated**
> void PressHotkeyDeprecated (string sandboxId, KeyboardHotkeyRequest keyboardHotkeyRequest, string? xDaytonaOrganizationID = null)

[DEPRECATED] Press hotkey

Press a hotkey combination

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
    public class PressHotkeyDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var keyboardHotkeyRequest = new KeyboardHotkeyRequest(); // KeyboardHotkeyRequest | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Press hotkey
                apiInstance.PressHotkeyDeprecated(sandboxId, keyboardHotkeyRequest, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.PressHotkeyDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the PressHotkeyDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Press hotkey
    apiInstance.PressHotkeyDeprecatedWithHttpInfo(sandboxId, keyboardHotkeyRequest, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.PressHotkeyDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **keyboardHotkeyRequest** | [**KeyboardHotkeyRequest**](KeyboardHotkeyRequest.md) |  |  |
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
| **200** | Hotkey pressed successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="presskeydeprecated"></a>
# **PressKeyDeprecated**
> void PressKeyDeprecated (string sandboxId, KeyboardPressRequest keyboardPressRequest, string? xDaytonaOrganizationID = null)

[DEPRECATED] Press key

Press a key with optional modifiers

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
    public class PressKeyDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var keyboardPressRequest = new KeyboardPressRequest(); // KeyboardPressRequest | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Press key
                apiInstance.PressKeyDeprecated(sandboxId, keyboardPressRequest, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.PressKeyDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the PressKeyDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Press key
    apiInstance.PressKeyDeprecatedWithHttpInfo(sandboxId, keyboardPressRequest, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.PressKeyDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **keyboardPressRequest** | [**KeyboardPressRequest**](KeyboardPressRequest.md) |  |  |
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
| **200** | Key pressed successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="replaceinfilesdeprecated"></a>
# **ReplaceInFilesDeprecated**
> List&lt;ReplaceResult&gt; ReplaceInFilesDeprecated (string sandboxId, ReplaceRequest replaceRequest, string? xDaytonaOrganizationID = null)

[DEPRECATED] Replace in files

Replace text/pattern in multiple files inside sandbox

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
    public class ReplaceInFilesDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var replaceRequest = new ReplaceRequest(); // ReplaceRequest | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Replace in files
                List<ReplaceResult> result = apiInstance.ReplaceInFilesDeprecated(sandboxId, replaceRequest, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.ReplaceInFilesDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ReplaceInFilesDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Replace in files
    ApiResponse<List<ReplaceResult>> response = apiInstance.ReplaceInFilesDeprecatedWithHttpInfo(sandboxId, replaceRequest, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.ReplaceInFilesDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **replaceRequest** | [**ReplaceRequest**](ReplaceRequest.md) |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**List&lt;ReplaceResult&gt;**](ReplaceResult.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Text replaced successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="resizeptysessiondeprecated"></a>
# **ResizePTYSessionDeprecated**
> PtySessionInfo ResizePTYSessionDeprecated (string sandboxId, string sessionId, PtyResizeRequest ptyResizeRequest, string? xDaytonaOrganizationID = null)

[DEPRECATED] Resize PTY session

Resize a PTY session

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
    public class ResizePTYSessionDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var sessionId = "sessionId_example";  // string | 
            var ptyResizeRequest = new PtyResizeRequest(); // PtyResizeRequest | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Resize PTY session
                PtySessionInfo result = apiInstance.ResizePTYSessionDeprecated(sandboxId, sessionId, ptyResizeRequest, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.ResizePTYSessionDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ResizePTYSessionDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Resize PTY session
    ApiResponse<PtySessionInfo> response = apiInstance.ResizePTYSessionDeprecatedWithHttpInfo(sandboxId, sessionId, ptyResizeRequest, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.ResizePTYSessionDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **sessionId** | **string** |  |  |
| **ptyResizeRequest** | [**PtyResizeRequest**](PtyResizeRequest.md) |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**PtySessionInfo**](PtySessionInfo.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | PTY session resized successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="restartprocessdeprecated"></a>
# **RestartProcessDeprecated**
> ProcessRestartResponse RestartProcessDeprecated (string processName, string sandboxId, string? xDaytonaOrganizationID = null)

[DEPRECATED] Restart process

Restart a specific VNC process

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
    public class RestartProcessDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var processName = "processName_example";  // string | 
            var sandboxId = "sandboxId_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Restart process
                ProcessRestartResponse result = apiInstance.RestartProcessDeprecated(processName, sandboxId, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.RestartProcessDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the RestartProcessDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Restart process
    ApiResponse<ProcessRestartResponse> response = apiInstance.RestartProcessDeprecatedWithHttpInfo(processName, sandboxId, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.RestartProcessDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **processName** | **string** |  |  |
| **sandboxId** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**ProcessRestartResponse**](ProcessRestartResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Process restarted successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="scrollmousedeprecated"></a>
# **ScrollMouseDeprecated**
> MouseScrollResponse ScrollMouseDeprecated (string sandboxId, MouseScrollRequest mouseScrollRequest, string? xDaytonaOrganizationID = null)

[DEPRECATED] Scroll mouse

Scroll mouse at specified coordinates

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
    public class ScrollMouseDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var mouseScrollRequest = new MouseScrollRequest(); // MouseScrollRequest | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Scroll mouse
                MouseScrollResponse result = apiInstance.ScrollMouseDeprecated(sandboxId, mouseScrollRequest, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.ScrollMouseDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ScrollMouseDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Scroll mouse
    ApiResponse<MouseScrollResponse> response = apiInstance.ScrollMouseDeprecatedWithHttpInfo(sandboxId, mouseScrollRequest, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.ScrollMouseDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **mouseScrollRequest** | [**MouseScrollRequest**](MouseScrollRequest.md) |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**MouseScrollResponse**](MouseScrollResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Mouse scrolled successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="searchfilesdeprecated"></a>
# **SearchFilesDeprecated**
> SearchFilesResponse SearchFilesDeprecated (string sandboxId, string path, string pattern, string? xDaytonaOrganizationID = null)

[DEPRECATED] Search files

Search for files inside sandbox

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
    public class SearchFilesDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var path = "path_example";  // string | 
            var pattern = "pattern_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Search files
                SearchFilesResponse result = apiInstance.SearchFilesDeprecated(sandboxId, path, pattern, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.SearchFilesDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the SearchFilesDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Search files
    ApiResponse<SearchFilesResponse> response = apiInstance.SearchFilesDeprecatedWithHttpInfo(sandboxId, path, pattern, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.SearchFilesDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **path** | **string** |  |  |
| **pattern** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**SearchFilesResponse**](SearchFilesResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Search completed successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="setfilepermissionsdeprecated"></a>
# **SetFilePermissionsDeprecated**
> void SetFilePermissionsDeprecated (string sandboxId, string path, string? xDaytonaOrganizationID = null, string? owner = null, string? group = null, string? mode = null)

[DEPRECATED] Set file permissions

Set file owner/group/permissions inside sandbox

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
    public class SetFilePermissionsDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var path = "path_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 
            var owner = "owner_example";  // string? |  (optional) 
            var group = "group_example";  // string? |  (optional) 
            var mode = "mode_example";  // string? |  (optional) 

            try
            {
                // [DEPRECATED] Set file permissions
                apiInstance.SetFilePermissionsDeprecated(sandboxId, path, xDaytonaOrganizationID, owner, group, mode);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.SetFilePermissionsDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the SetFilePermissionsDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Set file permissions
    apiInstance.SetFilePermissionsDeprecatedWithHttpInfo(sandboxId, path, xDaytonaOrganizationID, owner, group, mode);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.SetFilePermissionsDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **path** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |
| **owner** | **string?** |  | [optional]  |
| **group** | **string?** |  | [optional]  |
| **mode** | **string?** |  | [optional]  |

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
| **200** | File permissions updated successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="startcomputerusedeprecated"></a>
# **StartComputerUseDeprecated**
> ComputerUseStartResponse StartComputerUseDeprecated (string sandboxId, string? xDaytonaOrganizationID = null)

[DEPRECATED] Start computer use processes

Start all VNC desktop processes (Xvfb, xfce4, x11vnc, novnc)

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
    public class StartComputerUseDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Start computer use processes
                ComputerUseStartResponse result = apiInstance.StartComputerUseDeprecated(sandboxId, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.StartComputerUseDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the StartComputerUseDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Start computer use processes
    ApiResponse<ComputerUseStartResponse> response = apiInstance.StartComputerUseDeprecatedWithHttpInfo(sandboxId, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.StartComputerUseDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**ComputerUseStartResponse**](ComputerUseStartResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Computer use processes started successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="stopcomputerusedeprecated"></a>
# **StopComputerUseDeprecated**
> ComputerUseStopResponse StopComputerUseDeprecated (string sandboxId, string? xDaytonaOrganizationID = null)

[DEPRECATED] Stop computer use processes

Stop all VNC desktop processes (Xvfb, xfce4, x11vnc, novnc)

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
    public class StopComputerUseDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Stop computer use processes
                ComputerUseStopResponse result = apiInstance.StopComputerUseDeprecated(sandboxId, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.StopComputerUseDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the StopComputerUseDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Stop computer use processes
    ApiResponse<ComputerUseStopResponse> response = apiInstance.StopComputerUseDeprecatedWithHttpInfo(sandboxId, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.StopComputerUseDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**ComputerUseStopResponse**](ComputerUseStopResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Computer use processes stopped successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="takecompressedregionscreenshotdeprecated"></a>
# **TakeCompressedRegionScreenshotDeprecated**
> CompressedScreenshotResponse TakeCompressedRegionScreenshotDeprecated (string sandboxId, decimal height, decimal width, decimal y, decimal x, string? xDaytonaOrganizationID = null, decimal? scale = null, decimal? quality = null, string? format = null, bool? showCursor = null)

[DEPRECATED] Take compressed region screenshot

Take a compressed screenshot of a specific region

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
    public class TakeCompressedRegionScreenshotDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var height = 8.14D;  // decimal | 
            var width = 8.14D;  // decimal | 
            var y = 8.14D;  // decimal | 
            var x = 8.14D;  // decimal | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 
            var scale = 8.14D;  // decimal? |  (optional) 
            var quality = 8.14D;  // decimal? |  (optional) 
            var format = "format_example";  // string? |  (optional) 
            var showCursor = true;  // bool? |  (optional) 

            try
            {
                // [DEPRECATED] Take compressed region screenshot
                CompressedScreenshotResponse result = apiInstance.TakeCompressedRegionScreenshotDeprecated(sandboxId, height, width, y, x, xDaytonaOrganizationID, scale, quality, format, showCursor);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.TakeCompressedRegionScreenshotDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the TakeCompressedRegionScreenshotDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Take compressed region screenshot
    ApiResponse<CompressedScreenshotResponse> response = apiInstance.TakeCompressedRegionScreenshotDeprecatedWithHttpInfo(sandboxId, height, width, y, x, xDaytonaOrganizationID, scale, quality, format, showCursor);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.TakeCompressedRegionScreenshotDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **height** | **decimal** |  |  |
| **width** | **decimal** |  |  |
| **y** | **decimal** |  |  |
| **x** | **decimal** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |
| **scale** | **decimal?** |  | [optional]  |
| **quality** | **decimal?** |  | [optional]  |
| **format** | **string?** |  | [optional]  |
| **showCursor** | **bool?** |  | [optional]  |

### Return type

[**CompressedScreenshotResponse**](CompressedScreenshotResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Compressed region screenshot taken successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="takecompressedscreenshotdeprecated"></a>
# **TakeCompressedScreenshotDeprecated**
> CompressedScreenshotResponse TakeCompressedScreenshotDeprecated (string sandboxId, string? xDaytonaOrganizationID = null, decimal? scale = null, decimal? quality = null, string? format = null, bool? showCursor = null)

[DEPRECATED] Take compressed screenshot

Take a compressed screenshot with format, quality, and scale options

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
    public class TakeCompressedScreenshotDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 
            var scale = 8.14D;  // decimal? |  (optional) 
            var quality = 8.14D;  // decimal? |  (optional) 
            var format = "format_example";  // string? |  (optional) 
            var showCursor = true;  // bool? |  (optional) 

            try
            {
                // [DEPRECATED] Take compressed screenshot
                CompressedScreenshotResponse result = apiInstance.TakeCompressedScreenshotDeprecated(sandboxId, xDaytonaOrganizationID, scale, quality, format, showCursor);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.TakeCompressedScreenshotDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the TakeCompressedScreenshotDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Take compressed screenshot
    ApiResponse<CompressedScreenshotResponse> response = apiInstance.TakeCompressedScreenshotDeprecatedWithHttpInfo(sandboxId, xDaytonaOrganizationID, scale, quality, format, showCursor);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.TakeCompressedScreenshotDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |
| **scale** | **decimal?** |  | [optional]  |
| **quality** | **decimal?** |  | [optional]  |
| **format** | **string?** |  | [optional]  |
| **showCursor** | **bool?** |  | [optional]  |

### Return type

[**CompressedScreenshotResponse**](CompressedScreenshotResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Compressed screenshot taken successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="takeregionscreenshotdeprecated"></a>
# **TakeRegionScreenshotDeprecated**
> RegionScreenshotResponse TakeRegionScreenshotDeprecated (string sandboxId, decimal height, decimal width, decimal y, decimal x, string? xDaytonaOrganizationID = null, bool? showCursor = null)

[DEPRECATED] Take region screenshot

Take a screenshot of a specific region

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
    public class TakeRegionScreenshotDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var height = 8.14D;  // decimal | 
            var width = 8.14D;  // decimal | 
            var y = 8.14D;  // decimal | 
            var x = 8.14D;  // decimal | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 
            var showCursor = true;  // bool? |  (optional) 

            try
            {
                // [DEPRECATED] Take region screenshot
                RegionScreenshotResponse result = apiInstance.TakeRegionScreenshotDeprecated(sandboxId, height, width, y, x, xDaytonaOrganizationID, showCursor);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.TakeRegionScreenshotDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the TakeRegionScreenshotDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Take region screenshot
    ApiResponse<RegionScreenshotResponse> response = apiInstance.TakeRegionScreenshotDeprecatedWithHttpInfo(sandboxId, height, width, y, x, xDaytonaOrganizationID, showCursor);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.TakeRegionScreenshotDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **height** | **decimal** |  |  |
| **width** | **decimal** |  |  |
| **y** | **decimal** |  |  |
| **x** | **decimal** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |
| **showCursor** | **bool?** |  | [optional]  |

### Return type

[**RegionScreenshotResponse**](RegionScreenshotResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Region screenshot taken successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="takescreenshotdeprecated"></a>
# **TakeScreenshotDeprecated**
> ScreenshotResponse TakeScreenshotDeprecated (string sandboxId, string? xDaytonaOrganizationID = null, bool? showCursor = null)

[DEPRECATED] Take screenshot

Take a screenshot of the entire screen

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
    public class TakeScreenshotDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 
            var showCursor = true;  // bool? |  (optional) 

            try
            {
                // [DEPRECATED] Take screenshot
                ScreenshotResponse result = apiInstance.TakeScreenshotDeprecated(sandboxId, xDaytonaOrganizationID, showCursor);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.TakeScreenshotDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the TakeScreenshotDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Take screenshot
    ApiResponse<ScreenshotResponse> response = apiInstance.TakeScreenshotDeprecatedWithHttpInfo(sandboxId, xDaytonaOrganizationID, showCursor);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.TakeScreenshotDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |
| **showCursor** | **bool?** |  | [optional]  |

### Return type

[**ScreenshotResponse**](ScreenshotResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Screenshot taken successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="typetextdeprecated"></a>
# **TypeTextDeprecated**
> void TypeTextDeprecated (string sandboxId, KeyboardTypeRequest keyboardTypeRequest, string? xDaytonaOrganizationID = null)

[DEPRECATED] Type text

Type text using keyboard

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
    public class TypeTextDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var keyboardTypeRequest = new KeyboardTypeRequest(); // KeyboardTypeRequest | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Type text
                apiInstance.TypeTextDeprecated(sandboxId, keyboardTypeRequest, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.TypeTextDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the TypeTextDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Type text
    apiInstance.TypeTextDeprecatedWithHttpInfo(sandboxId, keyboardTypeRequest, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.TypeTextDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **keyboardTypeRequest** | [**KeyboardTypeRequest**](KeyboardTypeRequest.md) |  |  |
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
| **200** | Text typed successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="uploadfiledeprecated"></a>
# **UploadFileDeprecated**
> void UploadFileDeprecated (string sandboxId, string path, string? xDaytonaOrganizationID = null, FileParameter? file = null)

[DEPRECATED] Upload file

Upload file inside sandbox

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
    public class UploadFileDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var path = "path_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 
            var file = new System.IO.MemoryStream(System.IO.File.ReadAllBytes("/path/to/file.txt"));  // FileParameter? |  (optional) 

            try
            {
                // [DEPRECATED] Upload file
                apiInstance.UploadFileDeprecated(sandboxId, path, xDaytonaOrganizationID, file);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.UploadFileDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UploadFileDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Upload file
    apiInstance.UploadFileDeprecatedWithHttpInfo(sandboxId, path, xDaytonaOrganizationID, file);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.UploadFileDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **path** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |
| **file** | **FileParameter?****FileParameter?** |  | [optional]  |

### Return type

void (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: multipart/form-data
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | File uploaded successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="uploadfilesdeprecated"></a>
# **UploadFilesDeprecated**
> void UploadFilesDeprecated (string sandboxId, string? xDaytonaOrganizationID = null)

[DEPRECATED] Upload multiple files

Upload multiple files inside sandbox

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
    public class UploadFilesDeprecatedExample
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
            var apiInstance = new ToolboxApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // [DEPRECATED] Upload multiple files
                apiInstance.UploadFilesDeprecated(sandboxId, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ToolboxApi.UploadFilesDeprecated: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UploadFilesDeprecatedWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // [DEPRECATED] Upload multiple files
    apiInstance.UploadFilesDeprecatedWithHttpInfo(sandboxId, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ToolboxApi.UploadFilesDeprecatedWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

void (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: multipart/form-data
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Files uploaded successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

