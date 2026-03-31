# FileSystemApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**createFolder**](FileSystemApi.md#createFolder) | **POST** /files/folder | Create a folder |
| [**deleteFile**](FileSystemApi.md#deleteFile) | **DELETE** /files | Delete a file or directory |
| [**downloadFile**](FileSystemApi.md#downloadFile) | **GET** /files/download | Download a file |
| [**downloadFiles**](FileSystemApi.md#downloadFiles) | **POST** /files/bulk-download | Download multiple files |
| [**findInFiles**](FileSystemApi.md#findInFiles) | **GET** /files/find | Find text in files |
| [**getFileInfo**](FileSystemApi.md#getFileInfo) | **GET** /files/info | Get file information |
| [**listFiles**](FileSystemApi.md#listFiles) | **GET** /files | List files and directories |
| [**moveFile**](FileSystemApi.md#moveFile) | **POST** /files/move | Move or rename file/directory |
| [**replaceInFiles**](FileSystemApi.md#replaceInFiles) | **POST** /files/replace | Replace text in files |
| [**searchFiles**](FileSystemApi.md#searchFiles) | **GET** /files/search | Search files by pattern |
| [**setFilePermissions**](FileSystemApi.md#setFilePermissions) | **POST** /files/permissions | Set file permissions |
| [**uploadFile**](FileSystemApi.md#uploadFile) | **POST** /files/upload | Upload a file |
| [**uploadFiles**](FileSystemApi.md#uploadFiles) | **POST** /files/bulk-upload | Upload multiple files |


<a id="createFolder"></a>
# **createFolder**
> createFolder(path, mode)

Create a folder

Create a folder with the specified path and optional permissions

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.FileSystemApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    FileSystemApi apiInstance = new FileSystemApi(defaultClient);
    String path = "path_example"; // String | Folder path to create
    String mode = "mode_example"; // String | Octal permission mode (default: 0755)
    try {
      apiInstance.createFolder(path, mode);
    } catch (ApiException e) {
      System.err.println("Exception when calling FileSystemApi#createFolder");
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
| **path** | **String**| Folder path to create | |
| **mode** | **String**| Octal permission mode (default: 0755) | |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** | Created |  -  |

<a id="deleteFile"></a>
# **deleteFile**
> deleteFile(path, recursive)

Delete a file or directory

Delete a file or directory at the specified path

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.FileSystemApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    FileSystemApi apiInstance = new FileSystemApi(defaultClient);
    String path = "path_example"; // String | File or directory path to delete
    Boolean recursive = true; // Boolean | Enable recursive deletion for directories
    try {
      apiInstance.deleteFile(path, recursive);
    } catch (ApiException e) {
      System.err.println("Exception when calling FileSystemApi#deleteFile");
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
| **path** | **String**| File or directory path to delete | |
| **recursive** | **Boolean**| Enable recursive deletion for directories | [optional] |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | No Content |  -  |

<a id="downloadFile"></a>
# **downloadFile**
> File downloadFile(path)

Download a file

Download a file by providing its path

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.FileSystemApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    FileSystemApi apiInstance = new FileSystemApi(defaultClient);
    String path = "path_example"; // String | File path to download
    try {
      File result = apiInstance.downloadFile(path);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling FileSystemApi#downloadFile");
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
| **path** | **String**| File path to download | |

### Return type

[**File**](File.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/octet-stream

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="downloadFiles"></a>
# **downloadFiles**
> Map&lt;String, Object&gt; downloadFiles(downloadFiles)

Download multiple files

Download multiple files by providing their paths

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.FileSystemApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    FileSystemApi apiInstance = new FileSystemApi(defaultClient);
    FilesDownloadRequest downloadFiles = new FilesDownloadRequest(); // FilesDownloadRequest | Paths of files to download
    try {
      Map<String, Object> result = apiInstance.downloadFiles(downloadFiles);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling FileSystemApi#downloadFiles");
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
| **downloadFiles** | [**FilesDownloadRequest**](FilesDownloadRequest.md)| Paths of files to download | |

### Return type

**Map&lt;String, Object&gt;**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: multipart/form-data

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="findInFiles"></a>
# **findInFiles**
> List&lt;Match&gt; findInFiles(path, pattern)

Find text in files

Search for text pattern within files in a directory

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.FileSystemApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    FileSystemApi apiInstance = new FileSystemApi(defaultClient);
    String path = "path_example"; // String | Directory path to search in
    String pattern = "pattern_example"; // String | Text pattern to search for
    try {
      List<Match> result = apiInstance.findInFiles(path, pattern);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling FileSystemApi#findInFiles");
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
| **path** | **String**| Directory path to search in | |
| **pattern** | **String**| Text pattern to search for | |

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

<a id="getFileInfo"></a>
# **getFileInfo**
> FileInfo getFileInfo(path)

Get file information

Get detailed information about a file or directory

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.FileSystemApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    FileSystemApi apiInstance = new FileSystemApi(defaultClient);
    String path = "path_example"; // String | File or directory path
    try {
      FileInfo result = apiInstance.getFileInfo(path);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling FileSystemApi#getFileInfo");
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
| **path** | **String**| File or directory path | |

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

<a id="listFiles"></a>
# **listFiles**
> List&lt;FileInfo&gt; listFiles(path)

List files and directories

List files and directories in the specified path

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.FileSystemApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    FileSystemApi apiInstance = new FileSystemApi(defaultClient);
    String path = "path_example"; // String | Directory path to list (defaults to working directory)
    try {
      List<FileInfo> result = apiInstance.listFiles(path);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling FileSystemApi#listFiles");
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
| **path** | **String**| Directory path to list (defaults to working directory) | [optional] |

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

<a id="moveFile"></a>
# **moveFile**
> moveFile(source, destination)

Move or rename file/directory

Move or rename a file or directory from source to destination

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.FileSystemApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    FileSystemApi apiInstance = new FileSystemApi(defaultClient);
    String source = "source_example"; // String | Source file or directory path
    String destination = "destination_example"; // String | Destination file or directory path
    try {
      apiInstance.moveFile(source, destination);
    } catch (ApiException e) {
      System.err.println("Exception when calling FileSystemApi#moveFile");
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
| **source** | **String**| Source file or directory path | |
| **destination** | **String**| Destination file or directory path | |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="replaceInFiles"></a>
# **replaceInFiles**
> List&lt;ReplaceResult&gt; replaceInFiles(request)

Replace text in files

Replace text pattern with new value in multiple files

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.FileSystemApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    FileSystemApi apiInstance = new FileSystemApi(defaultClient);
    ReplaceRequest request = new ReplaceRequest(); // ReplaceRequest | Replace request
    try {
      List<ReplaceResult> result = apiInstance.replaceInFiles(request);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling FileSystemApi#replaceInFiles");
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
| **request** | [**ReplaceRequest**](ReplaceRequest.md)| Replace request | |

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

<a id="searchFiles"></a>
# **searchFiles**
> SearchFilesResponse searchFiles(path, pattern)

Search files by pattern

Search for files matching a specific pattern in a directory

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.FileSystemApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    FileSystemApi apiInstance = new FileSystemApi(defaultClient);
    String path = "path_example"; // String | Directory path to search in
    String pattern = "pattern_example"; // String | File pattern to match (e.g., *.txt, *.go)
    try {
      SearchFilesResponse result = apiInstance.searchFiles(path, pattern);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling FileSystemApi#searchFiles");
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
| **path** | **String**| Directory path to search in | |
| **pattern** | **String**| File pattern to match (e.g., *.txt, *.go) | |

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

<a id="setFilePermissions"></a>
# **setFilePermissions**
> setFilePermissions(path, owner, group, mode)

Set file permissions

Set file permissions, ownership, and group for a file or directory

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.FileSystemApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    FileSystemApi apiInstance = new FileSystemApi(defaultClient);
    String path = "path_example"; // String | File or directory path
    String owner = "owner_example"; // String | Owner (username or UID)
    String group = "group_example"; // String | Group (group name or GID)
    String mode = "mode_example"; // String | File mode in octal format (e.g., 0755)
    try {
      apiInstance.setFilePermissions(path, owner, group, mode);
    } catch (ApiException e) {
      System.err.println("Exception when calling FileSystemApi#setFilePermissions");
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
| **path** | **String**| File or directory path | |
| **owner** | **String**| Owner (username or UID) | [optional] |
| **group** | **String**| Group (group name or GID) | [optional] |
| **mode** | **String**| File mode in octal format (e.g., 0755) | [optional] |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="uploadFile"></a>
# **uploadFile**
> Map&lt;String, Object&gt; uploadFile(path, _file)

Upload a file

Upload a file to the specified path

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.FileSystemApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    FileSystemApi apiInstance = new FileSystemApi(defaultClient);
    String path = "path_example"; // String | Destination path for the uploaded file
    File _file = new File("/path/to/file"); // File | File to upload
    try {
      Map<String, Object> result = apiInstance.uploadFile(path, _file);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling FileSystemApi#uploadFile");
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
| **path** | **String**| Destination path for the uploaded file | |
| **_file** | **File**| File to upload | |

### Return type

**Map&lt;String, Object&gt;**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: multipart/form-data
 - **Accept**: */*

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="uploadFiles"></a>
# **uploadFiles**
> uploadFiles()

Upload multiple files

Upload multiple files with their destination paths

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.FileSystemApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    FileSystemApi apiInstance = new FileSystemApi(defaultClient);
    try {
      apiInstance.uploadFiles();
    } catch (ApiException e) {
      System.err.println("Exception when calling FileSystemApi#uploadFiles");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

