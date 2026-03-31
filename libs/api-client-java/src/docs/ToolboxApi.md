# ToolboxApi

All URIs are relative to *http://localhost:3000*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**clickMouseDeprecated**](ToolboxApi.md#clickMouseDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/computeruse/mouse/click | [DEPRECATED] Click mouse |
| [**createFolderDeprecated**](ToolboxApi.md#createFolderDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/files/folder | [DEPRECATED] Create folder |
| [**createPTYSessionDeprecated**](ToolboxApi.md#createPTYSessionDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/process/pty | [DEPRECATED] Create PTY session |
| [**createSessionDeprecated**](ToolboxApi.md#createSessionDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/process/session | [DEPRECATED] Create session |
| [**deleteFileDeprecated**](ToolboxApi.md#deleteFileDeprecated) | **DELETE** /toolbox/{sandboxId}/toolbox/files | [DEPRECATED] Delete file |
| [**deletePTYSessionDeprecated**](ToolboxApi.md#deletePTYSessionDeprecated) | **DELETE** /toolbox/{sandboxId}/toolbox/process/pty/{sessionId} | [DEPRECATED] Delete PTY session |
| [**deleteSessionDeprecated**](ToolboxApi.md#deleteSessionDeprecated) | **DELETE** /toolbox/{sandboxId}/toolbox/process/session/{sessionId} | [DEPRECATED] Delete session |
| [**downloadFileDeprecated**](ToolboxApi.md#downloadFileDeprecated) | **GET** /toolbox/{sandboxId}/toolbox/files/download | [DEPRECATED] Download file |
| [**downloadFilesDeprecated**](ToolboxApi.md#downloadFilesDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/files/bulk-download | [DEPRECATED] Download multiple files |
| [**dragMouseDeprecated**](ToolboxApi.md#dragMouseDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/computeruse/mouse/drag | [DEPRECATED] Drag mouse |
| [**executeCommandDeprecated**](ToolboxApi.md#executeCommandDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/process/execute | [DEPRECATED] Execute command |
| [**executeSessionCommandDeprecated**](ToolboxApi.md#executeSessionCommandDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/process/session/{sessionId}/exec | [DEPRECATED] Execute command in session |
| [**findInFilesDeprecated**](ToolboxApi.md#findInFilesDeprecated) | **GET** /toolbox/{sandboxId}/toolbox/files/find | [DEPRECATED] Search for text/pattern in files |
| [**getComputerUseStatusDeprecated**](ToolboxApi.md#getComputerUseStatusDeprecated) | **GET** /toolbox/{sandboxId}/toolbox/computeruse/status | [DEPRECATED] Get computer use status |
| [**getDisplayInfoDeprecated**](ToolboxApi.md#getDisplayInfoDeprecated) | **GET** /toolbox/{sandboxId}/toolbox/computeruse/display/info | [DEPRECATED] Get display info |
| [**getFileInfoDeprecated**](ToolboxApi.md#getFileInfoDeprecated) | **GET** /toolbox/{sandboxId}/toolbox/files/info | [DEPRECATED] Get file info |
| [**getMousePositionDeprecated**](ToolboxApi.md#getMousePositionDeprecated) | **GET** /toolbox/{sandboxId}/toolbox/computeruse/mouse/position | [DEPRECATED] Get mouse position |
| [**getPTYSessionDeprecated**](ToolboxApi.md#getPTYSessionDeprecated) | **GET** /toolbox/{sandboxId}/toolbox/process/pty/{sessionId} | [DEPRECATED] Get PTY session |
| [**getProcessErrorsDeprecated**](ToolboxApi.md#getProcessErrorsDeprecated) | **GET** /toolbox/{sandboxId}/toolbox/computeruse/process/{processName}/errors | [DEPRECATED] Get process errors |
| [**getProcessLogsDeprecated**](ToolboxApi.md#getProcessLogsDeprecated) | **GET** /toolbox/{sandboxId}/toolbox/computeruse/process/{processName}/logs | [DEPRECATED] Get process logs |
| [**getProcessStatusDeprecated**](ToolboxApi.md#getProcessStatusDeprecated) | **GET** /toolbox/{sandboxId}/toolbox/computeruse/process/{processName}/status | [DEPRECATED] Get process status |
| [**getProjectDirDeprecated**](ToolboxApi.md#getProjectDirDeprecated) | **GET** /toolbox/{sandboxId}/toolbox/project-dir | [DEPRECATED] Get sandbox project dir |
| [**getSessionCommandDeprecated**](ToolboxApi.md#getSessionCommandDeprecated) | **GET** /toolbox/{sandboxId}/toolbox/process/session/{sessionId}/command/{commandId} | [DEPRECATED] Get session command |
| [**getSessionCommandLogsDeprecated**](ToolboxApi.md#getSessionCommandLogsDeprecated) | **GET** /toolbox/{sandboxId}/toolbox/process/session/{sessionId}/command/{commandId}/logs | [DEPRECATED] Get command logs |
| [**getSessionDeprecated**](ToolboxApi.md#getSessionDeprecated) | **GET** /toolbox/{sandboxId}/toolbox/process/session/{sessionId} | [DEPRECATED] Get session |
| [**getUserHomeDirDeprecated**](ToolboxApi.md#getUserHomeDirDeprecated) | **GET** /toolbox/{sandboxId}/toolbox/user-home-dir | [DEPRECATED] Get sandbox user home dir |
| [**getWindowsDeprecated**](ToolboxApi.md#getWindowsDeprecated) | **GET** /toolbox/{sandboxId}/toolbox/computeruse/display/windows | [DEPRECATED] Get windows |
| [**getWorkDirDeprecated**](ToolboxApi.md#getWorkDirDeprecated) | **GET** /toolbox/{sandboxId}/toolbox/work-dir | [DEPRECATED] Get sandbox work-dir |
| [**gitAddFilesDeprecated**](ToolboxApi.md#gitAddFilesDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/git/add | [DEPRECATED] Add files |
| [**gitCheckoutBranchDeprecated**](ToolboxApi.md#gitCheckoutBranchDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/git/checkout | [DEPRECATED] Checkout branch |
| [**gitCloneRepositoryDeprecated**](ToolboxApi.md#gitCloneRepositoryDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/git/clone | [DEPRECATED] Clone repository |
| [**gitCommitChangesDeprecated**](ToolboxApi.md#gitCommitChangesDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/git/commit | [DEPRECATED] Commit changes |
| [**gitCreateBranchDeprecated**](ToolboxApi.md#gitCreateBranchDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/git/branches | [DEPRECATED] Create branch |
| [**gitDeleteBranchDeprecated**](ToolboxApi.md#gitDeleteBranchDeprecated) | **DELETE** /toolbox/{sandboxId}/toolbox/git/branches | [DEPRECATED] Delete branch |
| [**gitGetHistoryDeprecated**](ToolboxApi.md#gitGetHistoryDeprecated) | **GET** /toolbox/{sandboxId}/toolbox/git/history | [DEPRECATED] Get commit history |
| [**gitGetStatusDeprecated**](ToolboxApi.md#gitGetStatusDeprecated) | **GET** /toolbox/{sandboxId}/toolbox/git/status | [DEPRECATED] Get git status |
| [**gitListBranchesDeprecated**](ToolboxApi.md#gitListBranchesDeprecated) | **GET** /toolbox/{sandboxId}/toolbox/git/branches | [DEPRECATED] Get branch list |
| [**gitPullChangesDeprecated**](ToolboxApi.md#gitPullChangesDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/git/pull | [DEPRECATED] Pull changes |
| [**gitPushChangesDeprecated**](ToolboxApi.md#gitPushChangesDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/git/push | [DEPRECATED] Push changes |
| [**listFilesDeprecated**](ToolboxApi.md#listFilesDeprecated) | **GET** /toolbox/{sandboxId}/toolbox/files | [DEPRECATED] List files |
| [**listPTYSessionsDeprecated**](ToolboxApi.md#listPTYSessionsDeprecated) | **GET** /toolbox/{sandboxId}/toolbox/process/pty | [DEPRECATED] List PTY sessions |
| [**listSessionsDeprecated**](ToolboxApi.md#listSessionsDeprecated) | **GET** /toolbox/{sandboxId}/toolbox/process/session | [DEPRECATED] List sessions |
| [**lspCompletionsDeprecated**](ToolboxApi.md#lspCompletionsDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/lsp/completions | [DEPRECATED] Get Lsp Completions |
| [**lspDidCloseDeprecated**](ToolboxApi.md#lspDidCloseDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/lsp/did-close | [DEPRECATED] Call Lsp DidClose |
| [**lspDidOpenDeprecated**](ToolboxApi.md#lspDidOpenDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/lsp/did-open | [DEPRECATED] Call Lsp DidOpen |
| [**lspDocumentSymbolsDeprecated**](ToolboxApi.md#lspDocumentSymbolsDeprecated) | **GET** /toolbox/{sandboxId}/toolbox/lsp/document-symbols | [DEPRECATED] Call Lsp DocumentSymbols |
| [**lspStartDeprecated**](ToolboxApi.md#lspStartDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/lsp/start | [DEPRECATED] Start Lsp server |
| [**lspStopDeprecated**](ToolboxApi.md#lspStopDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/lsp/stop | [DEPRECATED] Stop Lsp server |
| [**lspWorkspaceSymbolsDeprecated**](ToolboxApi.md#lspWorkspaceSymbolsDeprecated) | **GET** /toolbox/{sandboxId}/toolbox/lsp/workspace-symbols | [DEPRECATED] Call Lsp WorkspaceSymbols |
| [**moveFileDeprecated**](ToolboxApi.md#moveFileDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/files/move | [DEPRECATED] Move file |
| [**moveMouseDeprecated**](ToolboxApi.md#moveMouseDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/computeruse/mouse/move | [DEPRECATED] Move mouse |
| [**pressHotkeyDeprecated**](ToolboxApi.md#pressHotkeyDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/computeruse/keyboard/hotkey | [DEPRECATED] Press hotkey |
| [**pressKeyDeprecated**](ToolboxApi.md#pressKeyDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/computeruse/keyboard/key | [DEPRECATED] Press key |
| [**replaceInFilesDeprecated**](ToolboxApi.md#replaceInFilesDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/files/replace | [DEPRECATED] Replace in files |
| [**resizePTYSessionDeprecated**](ToolboxApi.md#resizePTYSessionDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/process/pty/{sessionId}/resize | [DEPRECATED] Resize PTY session |
| [**restartProcessDeprecated**](ToolboxApi.md#restartProcessDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/computeruse/process/{processName}/restart | [DEPRECATED] Restart process |
| [**scrollMouseDeprecated**](ToolboxApi.md#scrollMouseDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/computeruse/mouse/scroll | [DEPRECATED] Scroll mouse |
| [**searchFilesDeprecated**](ToolboxApi.md#searchFilesDeprecated) | **GET** /toolbox/{sandboxId}/toolbox/files/search | [DEPRECATED] Search files |
| [**setFilePermissionsDeprecated**](ToolboxApi.md#setFilePermissionsDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/files/permissions | [DEPRECATED] Set file permissions |
| [**startComputerUseDeprecated**](ToolboxApi.md#startComputerUseDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/computeruse/start | [DEPRECATED] Start computer use processes |
| [**stopComputerUseDeprecated**](ToolboxApi.md#stopComputerUseDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/computeruse/stop | [DEPRECATED] Stop computer use processes |
| [**takeCompressedRegionScreenshotDeprecated**](ToolboxApi.md#takeCompressedRegionScreenshotDeprecated) | **GET** /toolbox/{sandboxId}/toolbox/computeruse/screenshot/region/compressed | [DEPRECATED] Take compressed region screenshot |
| [**takeCompressedScreenshotDeprecated**](ToolboxApi.md#takeCompressedScreenshotDeprecated) | **GET** /toolbox/{sandboxId}/toolbox/computeruse/screenshot/compressed | [DEPRECATED] Take compressed screenshot |
| [**takeRegionScreenshotDeprecated**](ToolboxApi.md#takeRegionScreenshotDeprecated) | **GET** /toolbox/{sandboxId}/toolbox/computeruse/screenshot/region | [DEPRECATED] Take region screenshot |
| [**takeScreenshotDeprecated**](ToolboxApi.md#takeScreenshotDeprecated) | **GET** /toolbox/{sandboxId}/toolbox/computeruse/screenshot | [DEPRECATED] Take screenshot |
| [**typeTextDeprecated**](ToolboxApi.md#typeTextDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/computeruse/keyboard/type | [DEPRECATED] Type text |
| [**uploadFileDeprecated**](ToolboxApi.md#uploadFileDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/files/upload | [DEPRECATED] Upload file |
| [**uploadFilesDeprecated**](ToolboxApi.md#uploadFilesDeprecated) | **POST** /toolbox/{sandboxId}/toolbox/files/bulk-upload | [DEPRECATED] Upload multiple files |


<a id="clickMouseDeprecated"></a>
# **clickMouseDeprecated**
> MouseClickResponse clickMouseDeprecated(sandboxId, mouseClickRequest, xDaytonaOrganizationID)

[DEPRECATED] Click mouse

Click mouse at specified coordinates

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    MouseClickRequest mouseClickRequest = new MouseClickRequest(); // MouseClickRequest | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      MouseClickResponse result = apiInstance.clickMouseDeprecated(sandboxId, mouseClickRequest, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#clickMouseDeprecated");
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
| **sandboxId** | **String**|  | |
| **mouseClickRequest** | [**MouseClickRequest**](MouseClickRequest.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="createFolderDeprecated"></a>
# **createFolderDeprecated**
> createFolderDeprecated(sandboxId, path, mode, xDaytonaOrganizationID)

[DEPRECATED] Create folder

Create folder inside sandbox

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String path = "path_example"; // String | 
    String mode = "mode_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.createFolderDeprecated(sandboxId, path, mode, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#createFolderDeprecated");
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
| **sandboxId** | **String**|  | |
| **path** | **String**|  | |
| **mode** | **String**|  | |
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
| **200** | Folder created successfully |  -  |

<a id="createPTYSessionDeprecated"></a>
# **createPTYSessionDeprecated**
> PtyCreateResponse createPTYSessionDeprecated(sandboxId, ptyCreateRequest, xDaytonaOrganizationID)

[DEPRECATED] Create PTY session

Create a new PTY session in the sandbox

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    PtyCreateRequest ptyCreateRequest = new PtyCreateRequest(); // PtyCreateRequest | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      PtyCreateResponse result = apiInstance.createPTYSessionDeprecated(sandboxId, ptyCreateRequest, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#createPTYSessionDeprecated");
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
| **sandboxId** | **String**|  | |
| **ptyCreateRequest** | [**PtyCreateRequest**](PtyCreateRequest.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="createSessionDeprecated"></a>
# **createSessionDeprecated**
> createSessionDeprecated(sandboxId, createSessionRequest, xDaytonaOrganizationID)

[DEPRECATED] Create session

Create a new session in the sandbox

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    CreateSessionRequest createSessionRequest = new CreateSessionRequest(); // CreateSessionRequest | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.createSessionDeprecated(sandboxId, createSessionRequest, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#createSessionDeprecated");
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
| **sandboxId** | **String**|  | |
| **createSessionRequest** | [**CreateSessionRequest**](CreateSessionRequest.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** |  |  -  |

<a id="deleteFileDeprecated"></a>
# **deleteFileDeprecated**
> deleteFileDeprecated(sandboxId, path, xDaytonaOrganizationID, recursive)

[DEPRECATED] Delete file

Delete file inside sandbox

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String path = "path_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    Boolean recursive = true; // Boolean | 
    try {
      apiInstance.deleteFileDeprecated(sandboxId, path, xDaytonaOrganizationID, recursive);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#deleteFileDeprecated");
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
| **sandboxId** | **String**|  | |
| **path** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |
| **recursive** | **Boolean**|  | [optional] |

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
| **200** | File deleted successfully |  -  |

<a id="deletePTYSessionDeprecated"></a>
# **deletePTYSessionDeprecated**
> deletePTYSessionDeprecated(sandboxId, sessionId, xDaytonaOrganizationID)

[DEPRECATED] Delete PTY session

Delete a PTY session and terminate the associated process

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String sessionId = "sessionId_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.deletePTYSessionDeprecated(sandboxId, sessionId, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#deletePTYSessionDeprecated");
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
| **sandboxId** | **String**|  | |
| **sessionId** | **String**|  | |
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
| **200** | PTY session deleted successfully |  -  |

<a id="deleteSessionDeprecated"></a>
# **deleteSessionDeprecated**
> deleteSessionDeprecated(sandboxId, sessionId, xDaytonaOrganizationID)

[DEPRECATED] Delete session

Delete a specific session

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String sessionId = "sessionId_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.deleteSessionDeprecated(sandboxId, sessionId, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#deleteSessionDeprecated");
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
| **sandboxId** | **String**|  | |
| **sessionId** | **String**|  | |
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
| **200** | Session deleted successfully |  -  |

<a id="downloadFileDeprecated"></a>
# **downloadFileDeprecated**
> File downloadFileDeprecated(sandboxId, path, xDaytonaOrganizationID)

[DEPRECATED] Download file

Download file from sandbox

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String path = "path_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      File result = apiInstance.downloadFileDeprecated(sandboxId, path, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#downloadFileDeprecated");
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
| **sandboxId** | **String**|  | |
| **path** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

[**File**](File.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | File downloaded successfully |  -  |

<a id="downloadFilesDeprecated"></a>
# **downloadFilesDeprecated**
> File downloadFilesDeprecated(sandboxId, downloadFiles, xDaytonaOrganizationID)

[DEPRECATED] Download multiple files

Streams back a multipart/form-data bundle of the requested paths

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    DownloadFiles downloadFiles = new DownloadFiles(); // DownloadFiles | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      File result = apiInstance.downloadFilesDeprecated(sandboxId, downloadFiles, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#downloadFilesDeprecated");
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
| **sandboxId** | **String**|  | |
| **downloadFiles** | [**DownloadFiles**](DownloadFiles.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

[**File**](File.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | A multipart/form-data response with each file as a part |  -  |

<a id="dragMouseDeprecated"></a>
# **dragMouseDeprecated**
> MouseDragResponse dragMouseDeprecated(sandboxId, mouseDragRequest, xDaytonaOrganizationID)

[DEPRECATED] Drag mouse

Drag mouse from start to end coordinates

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    MouseDragRequest mouseDragRequest = new MouseDragRequest(); // MouseDragRequest | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      MouseDragResponse result = apiInstance.dragMouseDeprecated(sandboxId, mouseDragRequest, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#dragMouseDeprecated");
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
| **sandboxId** | **String**|  | |
| **mouseDragRequest** | [**MouseDragRequest**](MouseDragRequest.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="executeCommandDeprecated"></a>
# **executeCommandDeprecated**
> ExecuteResponse executeCommandDeprecated(sandboxId, executeRequest, xDaytonaOrganizationID)

[DEPRECATED] Execute command

Execute command synchronously inside sandbox

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    ExecuteRequest executeRequest = new ExecuteRequest(); // ExecuteRequest | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      ExecuteResponse result = apiInstance.executeCommandDeprecated(sandboxId, executeRequest, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#executeCommandDeprecated");
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
| **sandboxId** | **String**|  | |
| **executeRequest** | [**ExecuteRequest**](ExecuteRequest.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="executeSessionCommandDeprecated"></a>
# **executeSessionCommandDeprecated**
> SessionExecuteResponse executeSessionCommandDeprecated(sandboxId, sessionId, sessionExecuteRequest, xDaytonaOrganizationID)

[DEPRECATED] Execute command in session

Execute a command in a specific session

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String sessionId = "sessionId_example"; // String | 
    SessionExecuteRequest sessionExecuteRequest = new SessionExecuteRequest(); // SessionExecuteRequest | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      SessionExecuteResponse result = apiInstance.executeSessionCommandDeprecated(sandboxId, sessionId, sessionExecuteRequest, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#executeSessionCommandDeprecated");
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
| **sandboxId** | **String**|  | |
| **sessionId** | **String**|  | |
| **sessionExecuteRequest** | [**SessionExecuteRequest**](SessionExecuteRequest.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="findInFilesDeprecated"></a>
# **findInFilesDeprecated**
> List&lt;Match&gt; findInFilesDeprecated(sandboxId, path, pattern, xDaytonaOrganizationID)

[DEPRECATED] Search for text/pattern in files

Search for text/pattern inside sandbox files

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String path = "path_example"; // String | 
    String pattern = "pattern_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      List<Match> result = apiInstance.findInFilesDeprecated(sandboxId, path, pattern, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#findInFilesDeprecated");
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
| **sandboxId** | **String**|  | |
| **path** | **String**|  | |
| **pattern** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="getComputerUseStatusDeprecated"></a>
# **getComputerUseStatusDeprecated**
> ComputerUseStatusResponse getComputerUseStatusDeprecated(sandboxId, xDaytonaOrganizationID)

[DEPRECATED] Get computer use status

Get status of all VNC desktop processes

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      ComputerUseStatusResponse result = apiInstance.getComputerUseStatusDeprecated(sandboxId, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#getComputerUseStatusDeprecated");
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
| **sandboxId** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="getDisplayInfoDeprecated"></a>
# **getDisplayInfoDeprecated**
> DisplayInfoResponse getDisplayInfoDeprecated(sandboxId, xDaytonaOrganizationID)

[DEPRECATED] Get display info

Get information about displays

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      DisplayInfoResponse result = apiInstance.getDisplayInfoDeprecated(sandboxId, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#getDisplayInfoDeprecated");
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
| **sandboxId** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="getFileInfoDeprecated"></a>
# **getFileInfoDeprecated**
> FileInfo getFileInfoDeprecated(sandboxId, path, xDaytonaOrganizationID)

[DEPRECATED] Get file info

Get file info inside sandbox

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String path = "path_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      FileInfo result = apiInstance.getFileInfoDeprecated(sandboxId, path, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#getFileInfoDeprecated");
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
| **sandboxId** | **String**|  | |
| **path** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="getMousePositionDeprecated"></a>
# **getMousePositionDeprecated**
> MousePosition getMousePositionDeprecated(sandboxId, xDaytonaOrganizationID)

[DEPRECATED] Get mouse position

Get current mouse cursor position

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      MousePosition result = apiInstance.getMousePositionDeprecated(sandboxId, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#getMousePositionDeprecated");
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
| **sandboxId** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="getPTYSessionDeprecated"></a>
# **getPTYSessionDeprecated**
> PtySessionInfo getPTYSessionDeprecated(sandboxId, sessionId, xDaytonaOrganizationID)

[DEPRECATED] Get PTY session

Get PTY session information by ID

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String sessionId = "sessionId_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      PtySessionInfo result = apiInstance.getPTYSessionDeprecated(sandboxId, sessionId, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#getPTYSessionDeprecated");
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
| **sandboxId** | **String**|  | |
| **sessionId** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="getProcessErrorsDeprecated"></a>
# **getProcessErrorsDeprecated**
> ProcessErrorsResponse getProcessErrorsDeprecated(processName, sandboxId, xDaytonaOrganizationID)

[DEPRECATED] Get process errors

Get error logs for a specific VNC process

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String processName = "processName_example"; // String | 
    String sandboxId = "sandboxId_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      ProcessErrorsResponse result = apiInstance.getProcessErrorsDeprecated(processName, sandboxId, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#getProcessErrorsDeprecated");
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
| **processName** | **String**|  | |
| **sandboxId** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="getProcessLogsDeprecated"></a>
# **getProcessLogsDeprecated**
> ProcessLogsResponse getProcessLogsDeprecated(processName, sandboxId, xDaytonaOrganizationID)

[DEPRECATED] Get process logs

Get logs for a specific VNC process

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String processName = "processName_example"; // String | 
    String sandboxId = "sandboxId_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      ProcessLogsResponse result = apiInstance.getProcessLogsDeprecated(processName, sandboxId, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#getProcessLogsDeprecated");
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
| **processName** | **String**|  | |
| **sandboxId** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="getProcessStatusDeprecated"></a>
# **getProcessStatusDeprecated**
> ProcessStatusResponse getProcessStatusDeprecated(processName, sandboxId, xDaytonaOrganizationID)

[DEPRECATED] Get process status

Get status of a specific VNC process

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String processName = "processName_example"; // String | 
    String sandboxId = "sandboxId_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      ProcessStatusResponse result = apiInstance.getProcessStatusDeprecated(processName, sandboxId, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#getProcessStatusDeprecated");
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
| **processName** | **String**|  | |
| **sandboxId** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="getProjectDirDeprecated"></a>
# **getProjectDirDeprecated**
> ProjectDirResponse getProjectDirDeprecated(sandboxId, xDaytonaOrganizationID)

[DEPRECATED] Get sandbox project dir

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      ProjectDirResponse result = apiInstance.getProjectDirDeprecated(sandboxId, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#getProjectDirDeprecated");
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
| **sandboxId** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="getSessionCommandDeprecated"></a>
# **getSessionCommandDeprecated**
> Command getSessionCommandDeprecated(sandboxId, sessionId, commandId, xDaytonaOrganizationID)

[DEPRECATED] Get session command

Get session command by ID

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String sessionId = "sessionId_example"; // String | 
    String commandId = "commandId_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      Command result = apiInstance.getSessionCommandDeprecated(sandboxId, sessionId, commandId, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#getSessionCommandDeprecated");
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
| **sandboxId** | **String**|  | |
| **sessionId** | **String**|  | |
| **commandId** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="getSessionCommandLogsDeprecated"></a>
# **getSessionCommandLogsDeprecated**
> String getSessionCommandLogsDeprecated(sandboxId, sessionId, commandId, xDaytonaOrganizationID, follow)

[DEPRECATED] Get command logs

Get logs for a specific command in a session

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String sessionId = "sessionId_example"; // String | 
    String commandId = "commandId_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    Boolean follow = true; // Boolean | Whether to stream the logs
    try {
      String result = apiInstance.getSessionCommandLogsDeprecated(sandboxId, sessionId, commandId, xDaytonaOrganizationID, follow);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#getSessionCommandLogsDeprecated");
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
| **sandboxId** | **String**|  | |
| **sessionId** | **String**|  | |
| **commandId** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |
| **follow** | **Boolean**| Whether to stream the logs | [optional] |

### Return type

**String**

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: text/plain

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Command log stream marked with stdout and stderr prefixes |  -  |

<a id="getSessionDeprecated"></a>
# **getSessionDeprecated**
> Session getSessionDeprecated(sandboxId, sessionId, xDaytonaOrganizationID)

[DEPRECATED] Get session

Get session by ID

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String sessionId = "sessionId_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      Session result = apiInstance.getSessionDeprecated(sandboxId, sessionId, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#getSessionDeprecated");
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
| **sandboxId** | **String**|  | |
| **sessionId** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="getUserHomeDirDeprecated"></a>
# **getUserHomeDirDeprecated**
> UserHomeDirResponse getUserHomeDirDeprecated(sandboxId, xDaytonaOrganizationID)

[DEPRECATED] Get sandbox user home dir

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      UserHomeDirResponse result = apiInstance.getUserHomeDirDeprecated(sandboxId, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#getUserHomeDirDeprecated");
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
| **sandboxId** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="getWindowsDeprecated"></a>
# **getWindowsDeprecated**
> WindowsResponse getWindowsDeprecated(sandboxId, xDaytonaOrganizationID)

[DEPRECATED] Get windows

Get list of open windows

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      WindowsResponse result = apiInstance.getWindowsDeprecated(sandboxId, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#getWindowsDeprecated");
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
| **sandboxId** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="getWorkDirDeprecated"></a>
# **getWorkDirDeprecated**
> WorkDirResponse getWorkDirDeprecated(sandboxId, xDaytonaOrganizationID)

[DEPRECATED] Get sandbox work-dir

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      WorkDirResponse result = apiInstance.getWorkDirDeprecated(sandboxId, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#getWorkDirDeprecated");
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
| **sandboxId** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="gitAddFilesDeprecated"></a>
# **gitAddFilesDeprecated**
> gitAddFilesDeprecated(sandboxId, gitAddRequest, xDaytonaOrganizationID)

[DEPRECATED] Add files

Add files to git commit

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    GitAddRequest gitAddRequest = new GitAddRequest(); // GitAddRequest | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.gitAddFilesDeprecated(sandboxId, gitAddRequest, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#gitAddFilesDeprecated");
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
| **sandboxId** | **String**|  | |
| **gitAddRequest** | [**GitAddRequest**](GitAddRequest.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Files added to git successfully |  -  |

<a id="gitCheckoutBranchDeprecated"></a>
# **gitCheckoutBranchDeprecated**
> gitCheckoutBranchDeprecated(sandboxId, gitCheckoutRequest, xDaytonaOrganizationID)

[DEPRECATED] Checkout branch

Checkout branch or commit in git repository

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    GitCheckoutRequest gitCheckoutRequest = new GitCheckoutRequest(); // GitCheckoutRequest | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.gitCheckoutBranchDeprecated(sandboxId, gitCheckoutRequest, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#gitCheckoutBranchDeprecated");
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
| **sandboxId** | **String**|  | |
| **gitCheckoutRequest** | [**GitCheckoutRequest**](GitCheckoutRequest.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Branch checked out successfully |  -  |

<a id="gitCloneRepositoryDeprecated"></a>
# **gitCloneRepositoryDeprecated**
> gitCloneRepositoryDeprecated(sandboxId, gitCloneRequest, xDaytonaOrganizationID)

[DEPRECATED] Clone repository

Clone git repository

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    GitCloneRequest gitCloneRequest = new GitCloneRequest(); // GitCloneRequest | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.gitCloneRepositoryDeprecated(sandboxId, gitCloneRequest, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#gitCloneRepositoryDeprecated");
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
| **sandboxId** | **String**|  | |
| **gitCloneRequest** | [**GitCloneRequest**](GitCloneRequest.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Repository cloned successfully |  -  |

<a id="gitCommitChangesDeprecated"></a>
# **gitCommitChangesDeprecated**
> GitCommitResponse gitCommitChangesDeprecated(sandboxId, gitCommitRequest, xDaytonaOrganizationID)

[DEPRECATED] Commit changes

Commit changes to git repository

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    GitCommitRequest gitCommitRequest = new GitCommitRequest(); // GitCommitRequest | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      GitCommitResponse result = apiInstance.gitCommitChangesDeprecated(sandboxId, gitCommitRequest, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#gitCommitChangesDeprecated");
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
| **sandboxId** | **String**|  | |
| **gitCommitRequest** | [**GitCommitRequest**](GitCommitRequest.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="gitCreateBranchDeprecated"></a>
# **gitCreateBranchDeprecated**
> gitCreateBranchDeprecated(sandboxId, gitBranchRequest, xDaytonaOrganizationID)

[DEPRECATED] Create branch

Create branch on git repository

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    GitBranchRequest gitBranchRequest = new GitBranchRequest(); // GitBranchRequest | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.gitCreateBranchDeprecated(sandboxId, gitBranchRequest, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#gitCreateBranchDeprecated");
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
| **sandboxId** | **String**|  | |
| **gitBranchRequest** | [**GitBranchRequest**](GitBranchRequest.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Branch created successfully |  -  |

<a id="gitDeleteBranchDeprecated"></a>
# **gitDeleteBranchDeprecated**
> gitDeleteBranchDeprecated(sandboxId, gitDeleteBranchRequest, xDaytonaOrganizationID)

[DEPRECATED] Delete branch

Delete branch on git repository

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    GitDeleteBranchRequest gitDeleteBranchRequest = new GitDeleteBranchRequest(); // GitDeleteBranchRequest | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.gitDeleteBranchDeprecated(sandboxId, gitDeleteBranchRequest, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#gitDeleteBranchDeprecated");
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
| **sandboxId** | **String**|  | |
| **gitDeleteBranchRequest** | [**GitDeleteBranchRequest**](GitDeleteBranchRequest.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Branch deleted successfully |  -  |

<a id="gitGetHistoryDeprecated"></a>
# **gitGetHistoryDeprecated**
> List&lt;GitCommitInfo&gt; gitGetHistoryDeprecated(sandboxId, path, xDaytonaOrganizationID)

[DEPRECATED] Get commit history

Get commit history from git repository

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String path = "path_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      List<GitCommitInfo> result = apiInstance.gitGetHistoryDeprecated(sandboxId, path, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#gitGetHistoryDeprecated");
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
| **sandboxId** | **String**|  | |
| **path** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="gitGetStatusDeprecated"></a>
# **gitGetStatusDeprecated**
> GitStatus gitGetStatusDeprecated(sandboxId, path, xDaytonaOrganizationID)

[DEPRECATED] Get git status

Get status from git repository

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String path = "path_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      GitStatus result = apiInstance.gitGetStatusDeprecated(sandboxId, path, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#gitGetStatusDeprecated");
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
| **sandboxId** | **String**|  | |
| **path** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="gitListBranchesDeprecated"></a>
# **gitListBranchesDeprecated**
> ListBranchResponse gitListBranchesDeprecated(sandboxId, path, xDaytonaOrganizationID)

[DEPRECATED] Get branch list

Get branch list from git repository

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String path = "path_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      ListBranchResponse result = apiInstance.gitListBranchesDeprecated(sandboxId, path, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#gitListBranchesDeprecated");
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
| **sandboxId** | **String**|  | |
| **path** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="gitPullChangesDeprecated"></a>
# **gitPullChangesDeprecated**
> gitPullChangesDeprecated(sandboxId, gitRepoRequest, xDaytonaOrganizationID)

[DEPRECATED] Pull changes

Pull changes from remote

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    GitRepoRequest gitRepoRequest = new GitRepoRequest(); // GitRepoRequest | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.gitPullChangesDeprecated(sandboxId, gitRepoRequest, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#gitPullChangesDeprecated");
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
| **sandboxId** | **String**|  | |
| **gitRepoRequest** | [**GitRepoRequest**](GitRepoRequest.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Changes pulled successfully |  -  |

<a id="gitPushChangesDeprecated"></a>
# **gitPushChangesDeprecated**
> gitPushChangesDeprecated(sandboxId, gitRepoRequest, xDaytonaOrganizationID)

[DEPRECATED] Push changes

Push changes to remote

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    GitRepoRequest gitRepoRequest = new GitRepoRequest(); // GitRepoRequest | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.gitPushChangesDeprecated(sandboxId, gitRepoRequest, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#gitPushChangesDeprecated");
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
| **sandboxId** | **String**|  | |
| **gitRepoRequest** | [**GitRepoRequest**](GitRepoRequest.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Changes pushed successfully |  -  |

<a id="listFilesDeprecated"></a>
# **listFilesDeprecated**
> List&lt;FileInfo&gt; listFilesDeprecated(sandboxId, xDaytonaOrganizationID, path)

[DEPRECATED] List files

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    String path = "path_example"; // String | 
    try {
      List<FileInfo> result = apiInstance.listFilesDeprecated(sandboxId, xDaytonaOrganizationID, path);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#listFilesDeprecated");
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
| **sandboxId** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |
| **path** | **String**|  | [optional] |

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

<a id="listPTYSessionsDeprecated"></a>
# **listPTYSessionsDeprecated**
> PtyListResponse listPTYSessionsDeprecated(sandboxId, xDaytonaOrganizationID)

[DEPRECATED] List PTY sessions

List all active PTY sessions in the sandbox

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      PtyListResponse result = apiInstance.listPTYSessionsDeprecated(sandboxId, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#listPTYSessionsDeprecated");
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
| **sandboxId** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="listSessionsDeprecated"></a>
# **listSessionsDeprecated**
> List&lt;Session&gt; listSessionsDeprecated(sandboxId, xDaytonaOrganizationID)

[DEPRECATED] List sessions

List all active sessions in the sandbox

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      List<Session> result = apiInstance.listSessionsDeprecated(sandboxId, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#listSessionsDeprecated");
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
| **sandboxId** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="lspCompletionsDeprecated"></a>
# **lspCompletionsDeprecated**
> CompletionList lspCompletionsDeprecated(sandboxId, lspCompletionParams, xDaytonaOrganizationID)

[DEPRECATED] Get Lsp Completions

The Completion request is sent from the client to the server to compute completion items at a given cursor position.

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    LspCompletionParams lspCompletionParams = new LspCompletionParams(); // LspCompletionParams | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      CompletionList result = apiInstance.lspCompletionsDeprecated(sandboxId, lspCompletionParams, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#lspCompletionsDeprecated");
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
| **sandboxId** | **String**|  | |
| **lspCompletionParams** | [**LspCompletionParams**](LspCompletionParams.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="lspDidCloseDeprecated"></a>
# **lspDidCloseDeprecated**
> lspDidCloseDeprecated(sandboxId, lspDocumentRequest, xDaytonaOrganizationID)

[DEPRECATED] Call Lsp DidClose

The document close notification is sent from the client to the server when the document got closed in the client.

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    LspDocumentRequest lspDocumentRequest = new LspDocumentRequest(); // LspDocumentRequest | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.lspDidCloseDeprecated(sandboxId, lspDocumentRequest, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#lspDidCloseDeprecated");
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
| **sandboxId** | **String**|  | |
| **lspDocumentRequest** | [**LspDocumentRequest**](LspDocumentRequest.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="lspDidOpenDeprecated"></a>
# **lspDidOpenDeprecated**
> lspDidOpenDeprecated(sandboxId, lspDocumentRequest, xDaytonaOrganizationID)

[DEPRECATED] Call Lsp DidOpen

The document open notification is sent from the client to the server to signal newly opened text documents.

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    LspDocumentRequest lspDocumentRequest = new LspDocumentRequest(); // LspDocumentRequest | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.lspDidOpenDeprecated(sandboxId, lspDocumentRequest, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#lspDidOpenDeprecated");
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
| **sandboxId** | **String**|  | |
| **lspDocumentRequest** | [**LspDocumentRequest**](LspDocumentRequest.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="lspDocumentSymbolsDeprecated"></a>
# **lspDocumentSymbolsDeprecated**
> List&lt;LspSymbol&gt; lspDocumentSymbolsDeprecated(sandboxId, languageId, pathToProject, uri, xDaytonaOrganizationID)

[DEPRECATED] Call Lsp DocumentSymbols

The document symbol request is sent from the client to the server.

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String languageId = "languageId_example"; // String | 
    String pathToProject = "pathToProject_example"; // String | 
    String uri = "uri_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      List<LspSymbol> result = apiInstance.lspDocumentSymbolsDeprecated(sandboxId, languageId, pathToProject, uri, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#lspDocumentSymbolsDeprecated");
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
| **sandboxId** | **String**|  | |
| **languageId** | **String**|  | |
| **pathToProject** | **String**|  | |
| **uri** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="lspStartDeprecated"></a>
# **lspStartDeprecated**
> lspStartDeprecated(sandboxId, lspServerRequest, xDaytonaOrganizationID)

[DEPRECATED] Start Lsp server

Start Lsp server process inside sandbox project

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    LspServerRequest lspServerRequest = new LspServerRequest(); // LspServerRequest | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.lspStartDeprecated(sandboxId, lspServerRequest, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#lspStartDeprecated");
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
| **sandboxId** | **String**|  | |
| **lspServerRequest** | [**LspServerRequest**](LspServerRequest.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="lspStopDeprecated"></a>
# **lspStopDeprecated**
> lspStopDeprecated(sandboxId, lspServerRequest, xDaytonaOrganizationID)

[DEPRECATED] Stop Lsp server

Stop Lsp server process inside sandbox project

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    LspServerRequest lspServerRequest = new LspServerRequest(); // LspServerRequest | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.lspStopDeprecated(sandboxId, lspServerRequest, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#lspStopDeprecated");
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
| **sandboxId** | **String**|  | |
| **lspServerRequest** | [**LspServerRequest**](LspServerRequest.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="lspWorkspaceSymbolsDeprecated"></a>
# **lspWorkspaceSymbolsDeprecated**
> List&lt;LspSymbol&gt; lspWorkspaceSymbolsDeprecated(sandboxId, languageId, pathToProject, query, xDaytonaOrganizationID)

[DEPRECATED] Call Lsp WorkspaceSymbols

The workspace symbol request is sent from the client to the server to list project-wide symbols matching the query string.

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String languageId = "languageId_example"; // String | 
    String pathToProject = "pathToProject_example"; // String | 
    String query = "query_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      List<LspSymbol> result = apiInstance.lspWorkspaceSymbolsDeprecated(sandboxId, languageId, pathToProject, query, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#lspWorkspaceSymbolsDeprecated");
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
| **sandboxId** | **String**|  | |
| **languageId** | **String**|  | |
| **pathToProject** | **String**|  | |
| **query** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="moveFileDeprecated"></a>
# **moveFileDeprecated**
> moveFileDeprecated(sandboxId, source, destination, xDaytonaOrganizationID)

[DEPRECATED] Move file

Move file inside sandbox

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String source = "source_example"; // String | 
    String destination = "destination_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.moveFileDeprecated(sandboxId, source, destination, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#moveFileDeprecated");
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
| **sandboxId** | **String**|  | |
| **source** | **String**|  | |
| **destination** | **String**|  | |
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
| **200** | File moved successfully |  -  |

<a id="moveMouseDeprecated"></a>
# **moveMouseDeprecated**
> MouseMoveResponse moveMouseDeprecated(sandboxId, mouseMoveRequest, xDaytonaOrganizationID)

[DEPRECATED] Move mouse

Move mouse cursor to specified coordinates

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    MouseMoveRequest mouseMoveRequest = new MouseMoveRequest(); // MouseMoveRequest | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      MouseMoveResponse result = apiInstance.moveMouseDeprecated(sandboxId, mouseMoveRequest, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#moveMouseDeprecated");
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
| **sandboxId** | **String**|  | |
| **mouseMoveRequest** | [**MouseMoveRequest**](MouseMoveRequest.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="pressHotkeyDeprecated"></a>
# **pressHotkeyDeprecated**
> pressHotkeyDeprecated(sandboxId, keyboardHotkeyRequest, xDaytonaOrganizationID)

[DEPRECATED] Press hotkey

Press a hotkey combination

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    KeyboardHotkeyRequest keyboardHotkeyRequest = new KeyboardHotkeyRequest(); // KeyboardHotkeyRequest | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.pressHotkeyDeprecated(sandboxId, keyboardHotkeyRequest, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#pressHotkeyDeprecated");
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
| **sandboxId** | **String**|  | |
| **keyboardHotkeyRequest** | [**KeyboardHotkeyRequest**](KeyboardHotkeyRequest.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Hotkey pressed successfully |  -  |

<a id="pressKeyDeprecated"></a>
# **pressKeyDeprecated**
> pressKeyDeprecated(sandboxId, keyboardPressRequest, xDaytonaOrganizationID)

[DEPRECATED] Press key

Press a key with optional modifiers

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    KeyboardPressRequest keyboardPressRequest = new KeyboardPressRequest(); // KeyboardPressRequest | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.pressKeyDeprecated(sandboxId, keyboardPressRequest, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#pressKeyDeprecated");
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
| **sandboxId** | **String**|  | |
| **keyboardPressRequest** | [**KeyboardPressRequest**](KeyboardPressRequest.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Key pressed successfully |  -  |

<a id="replaceInFilesDeprecated"></a>
# **replaceInFilesDeprecated**
> List&lt;ReplaceResult&gt; replaceInFilesDeprecated(sandboxId, replaceRequest, xDaytonaOrganizationID)

[DEPRECATED] Replace in files

Replace text/pattern in multiple files inside sandbox

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    ReplaceRequest replaceRequest = new ReplaceRequest(); // ReplaceRequest | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      List<ReplaceResult> result = apiInstance.replaceInFilesDeprecated(sandboxId, replaceRequest, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#replaceInFilesDeprecated");
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
| **sandboxId** | **String**|  | |
| **replaceRequest** | [**ReplaceRequest**](ReplaceRequest.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="resizePTYSessionDeprecated"></a>
# **resizePTYSessionDeprecated**
> PtySessionInfo resizePTYSessionDeprecated(sandboxId, sessionId, ptyResizeRequest, xDaytonaOrganizationID)

[DEPRECATED] Resize PTY session

Resize a PTY session

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String sessionId = "sessionId_example"; // String | 
    PtyResizeRequest ptyResizeRequest = new PtyResizeRequest(); // PtyResizeRequest | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      PtySessionInfo result = apiInstance.resizePTYSessionDeprecated(sandboxId, sessionId, ptyResizeRequest, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#resizePTYSessionDeprecated");
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
| **sandboxId** | **String**|  | |
| **sessionId** | **String**|  | |
| **ptyResizeRequest** | [**PtyResizeRequest**](PtyResizeRequest.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="restartProcessDeprecated"></a>
# **restartProcessDeprecated**
> ProcessRestartResponse restartProcessDeprecated(processName, sandboxId, xDaytonaOrganizationID)

[DEPRECATED] Restart process

Restart a specific VNC process

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String processName = "processName_example"; // String | 
    String sandboxId = "sandboxId_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      ProcessRestartResponse result = apiInstance.restartProcessDeprecated(processName, sandboxId, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#restartProcessDeprecated");
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
| **processName** | **String**|  | |
| **sandboxId** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="scrollMouseDeprecated"></a>
# **scrollMouseDeprecated**
> MouseScrollResponse scrollMouseDeprecated(sandboxId, mouseScrollRequest, xDaytonaOrganizationID)

[DEPRECATED] Scroll mouse

Scroll mouse at specified coordinates

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    MouseScrollRequest mouseScrollRequest = new MouseScrollRequest(); // MouseScrollRequest | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      MouseScrollResponse result = apiInstance.scrollMouseDeprecated(sandboxId, mouseScrollRequest, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#scrollMouseDeprecated");
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
| **sandboxId** | **String**|  | |
| **mouseScrollRequest** | [**MouseScrollRequest**](MouseScrollRequest.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="searchFilesDeprecated"></a>
# **searchFilesDeprecated**
> SearchFilesResponse searchFilesDeprecated(sandboxId, path, pattern, xDaytonaOrganizationID)

[DEPRECATED] Search files

Search for files inside sandbox

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String path = "path_example"; // String | 
    String pattern = "pattern_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      SearchFilesResponse result = apiInstance.searchFilesDeprecated(sandboxId, path, pattern, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#searchFilesDeprecated");
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
| **sandboxId** | **String**|  | |
| **path** | **String**|  | |
| **pattern** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="setFilePermissionsDeprecated"></a>
# **setFilePermissionsDeprecated**
> setFilePermissionsDeprecated(sandboxId, path, xDaytonaOrganizationID, owner, group, mode)

[DEPRECATED] Set file permissions

Set file owner/group/permissions inside sandbox

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String path = "path_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    String owner = "owner_example"; // String | 
    String group = "group_example"; // String | 
    String mode = "mode_example"; // String | 
    try {
      apiInstance.setFilePermissionsDeprecated(sandboxId, path, xDaytonaOrganizationID, owner, group, mode);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#setFilePermissionsDeprecated");
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
| **sandboxId** | **String**|  | |
| **path** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |
| **owner** | **String**|  | [optional] |
| **group** | **String**|  | [optional] |
| **mode** | **String**|  | [optional] |

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
| **200** | File permissions updated successfully |  -  |

<a id="startComputerUseDeprecated"></a>
# **startComputerUseDeprecated**
> ComputerUseStartResponse startComputerUseDeprecated(sandboxId, xDaytonaOrganizationID)

[DEPRECATED] Start computer use processes

Start all VNC desktop processes (Xvfb, xfce4, x11vnc, novnc)

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      ComputerUseStartResponse result = apiInstance.startComputerUseDeprecated(sandboxId, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#startComputerUseDeprecated");
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
| **sandboxId** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="stopComputerUseDeprecated"></a>
# **stopComputerUseDeprecated**
> ComputerUseStopResponse stopComputerUseDeprecated(sandboxId, xDaytonaOrganizationID)

[DEPRECATED] Stop computer use processes

Stop all VNC desktop processes (Xvfb, xfce4, x11vnc, novnc)

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      ComputerUseStopResponse result = apiInstance.stopComputerUseDeprecated(sandboxId, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#stopComputerUseDeprecated");
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
| **sandboxId** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="takeCompressedRegionScreenshotDeprecated"></a>
# **takeCompressedRegionScreenshotDeprecated**
> CompressedScreenshotResponse takeCompressedRegionScreenshotDeprecated(sandboxId, height, width, y, x, xDaytonaOrganizationID, scale, quality, format, showCursor)

[DEPRECATED] Take compressed region screenshot

Take a compressed screenshot of a specific region

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    BigDecimal height = new BigDecimal(78); // BigDecimal | 
    BigDecimal width = new BigDecimal(78); // BigDecimal | 
    BigDecimal y = new BigDecimal(78); // BigDecimal | 
    BigDecimal x = new BigDecimal(78); // BigDecimal | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    BigDecimal scale = new BigDecimal(78); // BigDecimal | 
    BigDecimal quality = new BigDecimal(78); // BigDecimal | 
    String format = "format_example"; // String | 
    Boolean showCursor = true; // Boolean | 
    try {
      CompressedScreenshotResponse result = apiInstance.takeCompressedRegionScreenshotDeprecated(sandboxId, height, width, y, x, xDaytonaOrganizationID, scale, quality, format, showCursor);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#takeCompressedRegionScreenshotDeprecated");
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
| **sandboxId** | **String**|  | |
| **height** | **BigDecimal**|  | |
| **width** | **BigDecimal**|  | |
| **y** | **BigDecimal**|  | |
| **x** | **BigDecimal**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |
| **scale** | **BigDecimal**|  | [optional] |
| **quality** | **BigDecimal**|  | [optional] |
| **format** | **String**|  | [optional] |
| **showCursor** | **Boolean**|  | [optional] |

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

<a id="takeCompressedScreenshotDeprecated"></a>
# **takeCompressedScreenshotDeprecated**
> CompressedScreenshotResponse takeCompressedScreenshotDeprecated(sandboxId, xDaytonaOrganizationID, scale, quality, format, showCursor)

[DEPRECATED] Take compressed screenshot

Take a compressed screenshot with format, quality, and scale options

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    BigDecimal scale = new BigDecimal(78); // BigDecimal | 
    BigDecimal quality = new BigDecimal(78); // BigDecimal | 
    String format = "format_example"; // String | 
    Boolean showCursor = true; // Boolean | 
    try {
      CompressedScreenshotResponse result = apiInstance.takeCompressedScreenshotDeprecated(sandboxId, xDaytonaOrganizationID, scale, quality, format, showCursor);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#takeCompressedScreenshotDeprecated");
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
| **sandboxId** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |
| **scale** | **BigDecimal**|  | [optional] |
| **quality** | **BigDecimal**|  | [optional] |
| **format** | **String**|  | [optional] |
| **showCursor** | **Boolean**|  | [optional] |

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

<a id="takeRegionScreenshotDeprecated"></a>
# **takeRegionScreenshotDeprecated**
> RegionScreenshotResponse takeRegionScreenshotDeprecated(sandboxId, height, width, y, x, xDaytonaOrganizationID, showCursor)

[DEPRECATED] Take region screenshot

Take a screenshot of a specific region

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    BigDecimal height = new BigDecimal(78); // BigDecimal | 
    BigDecimal width = new BigDecimal(78); // BigDecimal | 
    BigDecimal y = new BigDecimal(78); // BigDecimal | 
    BigDecimal x = new BigDecimal(78); // BigDecimal | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    Boolean showCursor = true; // Boolean | 
    try {
      RegionScreenshotResponse result = apiInstance.takeRegionScreenshotDeprecated(sandboxId, height, width, y, x, xDaytonaOrganizationID, showCursor);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#takeRegionScreenshotDeprecated");
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
| **sandboxId** | **String**|  | |
| **height** | **BigDecimal**|  | |
| **width** | **BigDecimal**|  | |
| **y** | **BigDecimal**|  | |
| **x** | **BigDecimal**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |
| **showCursor** | **Boolean**|  | [optional] |

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

<a id="takeScreenshotDeprecated"></a>
# **takeScreenshotDeprecated**
> ScreenshotResponse takeScreenshotDeprecated(sandboxId, xDaytonaOrganizationID, showCursor)

[DEPRECATED] Take screenshot

Take a screenshot of the entire screen

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    Boolean showCursor = true; // Boolean | 
    try {
      ScreenshotResponse result = apiInstance.takeScreenshotDeprecated(sandboxId, xDaytonaOrganizationID, showCursor);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#takeScreenshotDeprecated");
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
| **sandboxId** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |
| **showCursor** | **Boolean**|  | [optional] |

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

<a id="typeTextDeprecated"></a>
# **typeTextDeprecated**
> typeTextDeprecated(sandboxId, keyboardTypeRequest, xDaytonaOrganizationID)

[DEPRECATED] Type text

Type text using keyboard

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    KeyboardTypeRequest keyboardTypeRequest = new KeyboardTypeRequest(); // KeyboardTypeRequest | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.typeTextDeprecated(sandboxId, keyboardTypeRequest, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#typeTextDeprecated");
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
| **sandboxId** | **String**|  | |
| **keyboardTypeRequest** | [**KeyboardTypeRequest**](KeyboardTypeRequest.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Text typed successfully |  -  |

<a id="uploadFileDeprecated"></a>
# **uploadFileDeprecated**
> uploadFileDeprecated(sandboxId, path, xDaytonaOrganizationID, _file)

[DEPRECATED] Upload file

Upload file inside sandbox

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String path = "path_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    File _file = new File("/path/to/file"); // File | 
    try {
      apiInstance.uploadFileDeprecated(sandboxId, path, xDaytonaOrganizationID, _file);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#uploadFileDeprecated");
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
| **sandboxId** | **String**|  | |
| **path** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |
| **_file** | **File**|  | [optional] |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: multipart/form-data
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | File uploaded successfully |  -  |

<a id="uploadFilesDeprecated"></a>
# **uploadFilesDeprecated**
> uploadFilesDeprecated(sandboxId, xDaytonaOrganizationID)

[DEPRECATED] Upload multiple files

Upload multiple files inside sandbox

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ToolboxApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ToolboxApi apiInstance = new ToolboxApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.uploadFilesDeprecated(sandboxId, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling ToolboxApi#uploadFilesDeprecated");
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
| **sandboxId** | **String**|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: multipart/form-data
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Files uploaded successfully |  -  |

