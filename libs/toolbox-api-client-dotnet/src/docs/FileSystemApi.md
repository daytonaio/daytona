# Daytona.ToolboxApiClient.Api.FileSystemApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|--------|--------------|-------------|
| [**CreateFolder**](FileSystemApi.md#createfolder) | **POST** /files/folder | Create a folder |
| [**DeleteFile**](FileSystemApi.md#deletefile) | **DELETE** /files | Delete a file or directory |
| [**DownloadFile**](FileSystemApi.md#downloadfile) | **GET** /files/download | Download a file |
| [**DownloadFiles**](FileSystemApi.md#downloadfiles) | **POST** /files/bulk-download | Download multiple files |
| [**FindInFiles**](FileSystemApi.md#findinfiles) | **GET** /files/find | Find text in files |
| [**GetFileInfo**](FileSystemApi.md#getfileinfo) | **GET** /files/info | Get file information |
| [**ListFiles**](FileSystemApi.md#listfiles) | **GET** /files | List files and directories |
| [**MoveFile**](FileSystemApi.md#movefile) | **POST** /files/move | Move or rename file/directory |
| [**ReplaceInFiles**](FileSystemApi.md#replaceinfiles) | **POST** /files/replace | Replace text in files |
| [**SearchFiles**](FileSystemApi.md#searchfiles) | **GET** /files/search | Search files by pattern |
| [**SetFilePermissions**](FileSystemApi.md#setfilepermissions) | **POST** /files/permissions | Set file permissions |
| [**UploadFile**](FileSystemApi.md#uploadfile) | **POST** /files/upload | Upload a file |
| [**UploadFiles**](FileSystemApi.md#uploadfiles) | **POST** /files/bulk-upload | Upload multiple files |

<a id="createfolder"></a>
# **CreateFolder**
> void CreateFolder (string path, string mode)

Create a folder

Create a folder with the specified path and optional permissions

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using Daytona.ToolboxApiClient.Api;
using Daytona.ToolboxApiClient.Client;
using Daytona.ToolboxApiClient.Model;

namespace Example
{
    public class CreateFolderExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new FileSystemApi(httpClient, config, httpClientHandler);
            var path = "path_example";  // string | Folder path to create
            var mode = "mode_example";  // string | Octal permission mode (default: 0755)

            try
            {
                // Create a folder
                apiInstance.CreateFolder(path, mode);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling FileSystemApi.CreateFolder: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the CreateFolderWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Create a folder
    apiInstance.CreateFolderWithHttpInfo(path, mode);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling FileSystemApi.CreateFolderWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **path** | **string** | Folder path to create |  |
| **mode** | **string** | Octal permission mode (default: 0755) |  |

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** | Created |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="deletefile"></a>
# **DeleteFile**
> void DeleteFile (string path, bool? recursive = null)

Delete a file or directory

Delete a file or directory at the specified path

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using Daytona.ToolboxApiClient.Api;
using Daytona.ToolboxApiClient.Client;
using Daytona.ToolboxApiClient.Model;

namespace Example
{
    public class DeleteFileExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new FileSystemApi(httpClient, config, httpClientHandler);
            var path = "path_example";  // string | File or directory path to delete
            var recursive = true;  // bool? | Enable recursive deletion for directories (optional) 

            try
            {
                // Delete a file or directory
                apiInstance.DeleteFile(path, recursive);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling FileSystemApi.DeleteFile: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DeleteFileWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Delete a file or directory
    apiInstance.DeleteFileWithHttpInfo(path, recursive);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling FileSystemApi.DeleteFileWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **path** | **string** | File or directory path to delete |  |
| **recursive** | **bool?** | Enable recursive deletion for directories | [optional]  |

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | No Content |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="downloadfile"></a>
# **DownloadFile**
> FileParameter DownloadFile (string path)

Download a file

Download a file by providing its path

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using Daytona.ToolboxApiClient.Api;
using Daytona.ToolboxApiClient.Client;
using Daytona.ToolboxApiClient.Model;

namespace Example
{
    public class DownloadFileExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new FileSystemApi(httpClient, config, httpClientHandler);
            var path = "path_example";  // string | File path to download

            try
            {
                // Download a file
                FileParameter result = apiInstance.DownloadFile(path);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling FileSystemApi.DownloadFile: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DownloadFileWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Download a file
    ApiResponse<FileParameter> response = apiInstance.DownloadFileWithHttpInfo(path);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling FileSystemApi.DownloadFileWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **path** | **string** | File path to download |  |

### Return type

[**FileParameter**](FileParameter.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/octet-stream


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="downloadfiles"></a>
# **DownloadFiles**
> Dictionary&lt;string, Object&gt; DownloadFiles (FilesDownloadRequest downloadFiles)

Download multiple files

Download multiple files by providing their paths

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using Daytona.ToolboxApiClient.Api;
using Daytona.ToolboxApiClient.Client;
using Daytona.ToolboxApiClient.Model;

namespace Example
{
    public class DownloadFilesExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new FileSystemApi(httpClient, config, httpClientHandler);
            var downloadFiles = new FilesDownloadRequest(); // FilesDownloadRequest | Paths of files to download

            try
            {
                // Download multiple files
                Dictionary<string, Object> result = apiInstance.DownloadFiles(downloadFiles);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling FileSystemApi.DownloadFiles: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DownloadFilesWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Download multiple files
    ApiResponse<Dictionary<string, Object>> response = apiInstance.DownloadFilesWithHttpInfo(downloadFiles);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling FileSystemApi.DownloadFilesWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **downloadFiles** | [**FilesDownloadRequest**](FilesDownloadRequest.md) | Paths of files to download |  |

### Return type

**Dictionary<string, Object>**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: multipart/form-data


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="findinfiles"></a>
# **FindInFiles**
> List&lt;Match&gt; FindInFiles (string path, string pattern)

Find text in files

Search for text pattern within files in a directory

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using Daytona.ToolboxApiClient.Api;
using Daytona.ToolboxApiClient.Client;
using Daytona.ToolboxApiClient.Model;

namespace Example
{
    public class FindInFilesExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new FileSystemApi(httpClient, config, httpClientHandler);
            var path = "path_example";  // string | Directory path to search in
            var pattern = "pattern_example";  // string | Text pattern to search for

            try
            {
                // Find text in files
                List<Match> result = apiInstance.FindInFiles(path, pattern);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling FileSystemApi.FindInFiles: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the FindInFilesWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Find text in files
    ApiResponse<List<Match>> response = apiInstance.FindInFilesWithHttpInfo(path, pattern);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling FileSystemApi.FindInFilesWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **path** | **string** | Directory path to search in |  |
| **pattern** | **string** | Text pattern to search for |  |

### Return type

[**List&lt;Match&gt;**](Match.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getfileinfo"></a>
# **GetFileInfo**
> FileInfo GetFileInfo (string path)

Get file information

Get detailed information about a file or directory

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using Daytona.ToolboxApiClient.Api;
using Daytona.ToolboxApiClient.Client;
using Daytona.ToolboxApiClient.Model;

namespace Example
{
    public class GetFileInfoExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new FileSystemApi(httpClient, config, httpClientHandler);
            var path = "path_example";  // string | File or directory path

            try
            {
                // Get file information
                FileInfo result = apiInstance.GetFileInfo(path);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling FileSystemApi.GetFileInfo: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetFileInfoWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get file information
    ApiResponse<FileInfo> response = apiInstance.GetFileInfoWithHttpInfo(path);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling FileSystemApi.GetFileInfoWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **path** | **string** | File or directory path |  |

### Return type

[**FileInfo**](FileInfo.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="listfiles"></a>
# **ListFiles**
> List&lt;FileInfo&gt; ListFiles (string? path = null)

List files and directories

List files and directories in the specified path

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using Daytona.ToolboxApiClient.Api;
using Daytona.ToolboxApiClient.Client;
using Daytona.ToolboxApiClient.Model;

namespace Example
{
    public class ListFilesExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new FileSystemApi(httpClient, config, httpClientHandler);
            var path = "path_example";  // string? | Directory path to list (defaults to working directory) (optional) 

            try
            {
                // List files and directories
                List<FileInfo> result = apiInstance.ListFiles(path);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling FileSystemApi.ListFiles: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ListFilesWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // List files and directories
    ApiResponse<List<FileInfo>> response = apiInstance.ListFilesWithHttpInfo(path);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling FileSystemApi.ListFilesWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **path** | **string?** | Directory path to list (defaults to working directory) | [optional]  |

### Return type

[**List&lt;FileInfo&gt;**](FileInfo.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="movefile"></a>
# **MoveFile**
> void MoveFile (string source, string destination)

Move or rename file/directory

Move or rename a file or directory from source to destination

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using Daytona.ToolboxApiClient.Api;
using Daytona.ToolboxApiClient.Client;
using Daytona.ToolboxApiClient.Model;

namespace Example
{
    public class MoveFileExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new FileSystemApi(httpClient, config, httpClientHandler);
            var source = "source_example";  // string | Source file or directory path
            var destination = "destination_example";  // string | Destination file or directory path

            try
            {
                // Move or rename file/directory
                apiInstance.MoveFile(source, destination);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling FileSystemApi.MoveFile: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the MoveFileWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Move or rename file/directory
    apiInstance.MoveFileWithHttpInfo(source, destination);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling FileSystemApi.MoveFileWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **source** | **string** | Source file or directory path |  |
| **destination** | **string** | Destination file or directory path |  |

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="replaceinfiles"></a>
# **ReplaceInFiles**
> List&lt;ReplaceResult&gt; ReplaceInFiles (ReplaceRequest request)

Replace text in files

Replace text pattern with new value in multiple files

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using Daytona.ToolboxApiClient.Api;
using Daytona.ToolboxApiClient.Client;
using Daytona.ToolboxApiClient.Model;

namespace Example
{
    public class ReplaceInFilesExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new FileSystemApi(httpClient, config, httpClientHandler);
            var request = new ReplaceRequest(); // ReplaceRequest | Replace request

            try
            {
                // Replace text in files
                List<ReplaceResult> result = apiInstance.ReplaceInFiles(request);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling FileSystemApi.ReplaceInFiles: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ReplaceInFilesWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Replace text in files
    ApiResponse<List<ReplaceResult>> response = apiInstance.ReplaceInFilesWithHttpInfo(request);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling FileSystemApi.ReplaceInFilesWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **request** | [**ReplaceRequest**](ReplaceRequest.md) | Replace request |  |

### Return type

[**List&lt;ReplaceResult&gt;**](ReplaceResult.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="searchfiles"></a>
# **SearchFiles**
> SearchFilesResponse SearchFiles (string path, string pattern)

Search files by pattern

Search for files matching a specific pattern in a directory

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using Daytona.ToolboxApiClient.Api;
using Daytona.ToolboxApiClient.Client;
using Daytona.ToolboxApiClient.Model;

namespace Example
{
    public class SearchFilesExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new FileSystemApi(httpClient, config, httpClientHandler);
            var path = "path_example";  // string | Directory path to search in
            var pattern = "pattern_example";  // string | File pattern to match (e.g., *.txt, *.go)

            try
            {
                // Search files by pattern
                SearchFilesResponse result = apiInstance.SearchFiles(path, pattern);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling FileSystemApi.SearchFiles: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the SearchFilesWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Search files by pattern
    ApiResponse<SearchFilesResponse> response = apiInstance.SearchFilesWithHttpInfo(path, pattern);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling FileSystemApi.SearchFilesWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **path** | **string** | Directory path to search in |  |
| **pattern** | **string** | File pattern to match (e.g., *.txt, *.go) |  |

### Return type

[**SearchFilesResponse**](SearchFilesResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="setfilepermissions"></a>
# **SetFilePermissions**
> void SetFilePermissions (string path, string? owner = null, string? group = null, string? mode = null)

Set file permissions

Set file permissions, ownership, and group for a file or directory

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using Daytona.ToolboxApiClient.Api;
using Daytona.ToolboxApiClient.Client;
using Daytona.ToolboxApiClient.Model;

namespace Example
{
    public class SetFilePermissionsExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new FileSystemApi(httpClient, config, httpClientHandler);
            var path = "path_example";  // string | File or directory path
            var owner = "owner_example";  // string? | Owner (username or UID) (optional) 
            var group = "group_example";  // string? | Group (group name or GID) (optional) 
            var mode = "mode_example";  // string? | File mode in octal format (e.g., 0755) (optional) 

            try
            {
                // Set file permissions
                apiInstance.SetFilePermissions(path, owner, group, mode);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling FileSystemApi.SetFilePermissions: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the SetFilePermissionsWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Set file permissions
    apiInstance.SetFilePermissionsWithHttpInfo(path, owner, group, mode);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling FileSystemApi.SetFilePermissionsWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **path** | **string** | File or directory path |  |
| **owner** | **string?** | Owner (username or UID) | [optional]  |
| **group** | **string?** | Group (group name or GID) | [optional]  |
| **mode** | **string?** | File mode in octal format (e.g., 0755) | [optional]  |

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="uploadfile"></a>
# **UploadFile**
> Dictionary&lt;string, Object&gt; UploadFile (string path, FileParameter file)

Upload a file

Upload a file to the specified path

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using Daytona.ToolboxApiClient.Api;
using Daytona.ToolboxApiClient.Client;
using Daytona.ToolboxApiClient.Model;

namespace Example
{
    public class UploadFileExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new FileSystemApi(httpClient, config, httpClientHandler);
            var path = "path_example";  // string | Destination path for the uploaded file
            var file = new System.IO.MemoryStream(System.IO.File.ReadAllBytes("/path/to/file.txt"));  // FileParameter | File to upload

            try
            {
                // Upload a file
                Dictionary<string, Object> result = apiInstance.UploadFile(path, file);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling FileSystemApi.UploadFile: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UploadFileWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Upload a file
    ApiResponse<Dictionary<string, Object>> response = apiInstance.UploadFileWithHttpInfo(path, file);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling FileSystemApi.UploadFileWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **path** | **string** | Destination path for the uploaded file |  |
| **file** | **FileParameter****FileParameter** | File to upload |  |

### Return type

**Dictionary<string, Object>**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: multipart/form-data
 - **Accept**: */*


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="uploadfiles"></a>
# **UploadFiles**
> void UploadFiles ()

Upload multiple files

Upload multiple files with their destination paths

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using Daytona.ToolboxApiClient.Api;
using Daytona.ToolboxApiClient.Client;
using Daytona.ToolboxApiClient.Model;

namespace Example
{
    public class UploadFilesExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new FileSystemApi(httpClient, config, httpClientHandler);

            try
            {
                // Upload multiple files
                apiInstance.UploadFiles();
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling FileSystemApi.UploadFiles: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UploadFilesWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Upload multiple files
    apiInstance.UploadFilesWithHttpInfo();
}
catch (ApiException e)
{
    Debug.Print("Exception when calling FileSystemApi.UploadFilesWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters
This endpoint does not need any parameter.
### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

