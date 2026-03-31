# Daytona.ToolboxApiClient.Api.GitApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|--------|--------------|-------------|
| [**AddFiles**](GitApi.md#addfiles) | **POST** /git/add | Add files to Git staging |
| [**CheckoutBranch**](GitApi.md#checkoutbranch) | **POST** /git/checkout | Checkout branch or commit |
| [**CloneRepository**](GitApi.md#clonerepository) | **POST** /git/clone | Clone a Git repository |
| [**CommitChanges**](GitApi.md#commitchanges) | **POST** /git/commit | Commit changes |
| [**CreateBranch**](GitApi.md#createbranch) | **POST** /git/branches | Create a new branch |
| [**DeleteBranch**](GitApi.md#deletebranch) | **DELETE** /git/branches | Delete a branch |
| [**GetCommitHistory**](GitApi.md#getcommithistory) | **GET** /git/history | Get commit history |
| [**GetStatus**](GitApi.md#getstatus) | **GET** /git/status | Get Git status |
| [**ListBranches**](GitApi.md#listbranches) | **GET** /git/branches | List branches |
| [**PullChanges**](GitApi.md#pullchanges) | **POST** /git/pull | Pull changes from remote |
| [**PushChanges**](GitApi.md#pushchanges) | **POST** /git/push | Push changes to remote |

<a id="addfiles"></a>
# **AddFiles**
> void AddFiles (GitAddRequest request)

Add files to Git staging

Add files to the Git staging area

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
    public class AddFilesExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new GitApi(httpClient, config, httpClientHandler);
            var request = new GitAddRequest(); // GitAddRequest | Add files request

            try
            {
                // Add files to Git staging
                apiInstance.AddFiles(request);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling GitApi.AddFiles: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the AddFilesWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Add files to Git staging
    apiInstance.AddFilesWithHttpInfo(request);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling GitApi.AddFilesWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **request** | [**GitAddRequest**](GitAddRequest.md) | Add files request |  |

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="checkoutbranch"></a>
# **CheckoutBranch**
> void CheckoutBranch (GitCheckoutRequest request)

Checkout branch or commit

Switch to a different branch or commit in the Git repository

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
    public class CheckoutBranchExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new GitApi(httpClient, config, httpClientHandler);
            var request = new GitCheckoutRequest(); // GitCheckoutRequest | Checkout request

            try
            {
                // Checkout branch or commit
                apiInstance.CheckoutBranch(request);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling GitApi.CheckoutBranch: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the CheckoutBranchWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Checkout branch or commit
    apiInstance.CheckoutBranchWithHttpInfo(request);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling GitApi.CheckoutBranchWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **request** | [**GitCheckoutRequest**](GitCheckoutRequest.md) | Checkout request |  |

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="clonerepository"></a>
# **CloneRepository**
> void CloneRepository (GitCloneRequest request)

Clone a Git repository

Clone a Git repository to the specified path

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
    public class CloneRepositoryExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new GitApi(httpClient, config, httpClientHandler);
            var request = new GitCloneRequest(); // GitCloneRequest | Clone repository request

            try
            {
                // Clone a Git repository
                apiInstance.CloneRepository(request);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling GitApi.CloneRepository: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the CloneRepositoryWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Clone a Git repository
    apiInstance.CloneRepositoryWithHttpInfo(request);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling GitApi.CloneRepositoryWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **request** | [**GitCloneRequest**](GitCloneRequest.md) | Clone repository request |  |

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="commitchanges"></a>
# **CommitChanges**
> GitCommitResponse CommitChanges (GitCommitRequest request)

Commit changes

Commit staged changes to the Git repository

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
    public class CommitChangesExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new GitApi(httpClient, config, httpClientHandler);
            var request = new GitCommitRequest(); // GitCommitRequest | Commit request

            try
            {
                // Commit changes
                GitCommitResponse result = apiInstance.CommitChanges(request);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling GitApi.CommitChanges: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the CommitChangesWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Commit changes
    ApiResponse<GitCommitResponse> response = apiInstance.CommitChangesWithHttpInfo(request);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling GitApi.CommitChangesWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **request** | [**GitCommitRequest**](GitCommitRequest.md) | Commit request |  |

### Return type

[**GitCommitResponse**](GitCommitResponse.md)

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

<a id="createbranch"></a>
# **CreateBranch**
> void CreateBranch (GitBranchRequest request)

Create a new branch

Create a new branch in the Git repository

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
    public class CreateBranchExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new GitApi(httpClient, config, httpClientHandler);
            var request = new GitBranchRequest(); // GitBranchRequest | Create branch request

            try
            {
                // Create a new branch
                apiInstance.CreateBranch(request);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling GitApi.CreateBranch: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the CreateBranchWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Create a new branch
    apiInstance.CreateBranchWithHttpInfo(request);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling GitApi.CreateBranchWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **request** | [**GitBranchRequest**](GitBranchRequest.md) | Create branch request |  |

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** | Created |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="deletebranch"></a>
# **DeleteBranch**
> void DeleteBranch (GitGitDeleteBranchRequest request)

Delete a branch

Delete a branch from the Git repository

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
    public class DeleteBranchExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new GitApi(httpClient, config, httpClientHandler);
            var request = new GitGitDeleteBranchRequest(); // GitGitDeleteBranchRequest | Delete branch request

            try
            {
                // Delete a branch
                apiInstance.DeleteBranch(request);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling GitApi.DeleteBranch: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DeleteBranchWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Delete a branch
    apiInstance.DeleteBranchWithHttpInfo(request);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling GitApi.DeleteBranchWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **request** | [**GitGitDeleteBranchRequest**](GitGitDeleteBranchRequest.md) | Delete branch request |  |

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | No Content |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getcommithistory"></a>
# **GetCommitHistory**
> List&lt;GitCommitInfo&gt; GetCommitHistory (string path)

Get commit history

Get the commit history of the Git repository

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
    public class GetCommitHistoryExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new GitApi(httpClient, config, httpClientHandler);
            var path = "path_example";  // string | Repository path

            try
            {
                // Get commit history
                List<GitCommitInfo> result = apiInstance.GetCommitHistory(path);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling GitApi.GetCommitHistory: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetCommitHistoryWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get commit history
    ApiResponse<List<GitCommitInfo>> response = apiInstance.GetCommitHistoryWithHttpInfo(path);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling GitApi.GetCommitHistoryWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **path** | **string** | Repository path |  |

### Return type

[**List&lt;GitCommitInfo&gt;**](GitCommitInfo.md)

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

<a id="getstatus"></a>
# **GetStatus**
> GitStatus GetStatus (string path)

Get Git status

Get the Git status of the repository at the specified path

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
    public class GetStatusExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new GitApi(httpClient, config, httpClientHandler);
            var path = "path_example";  // string | Repository path

            try
            {
                // Get Git status
                GitStatus result = apiInstance.GetStatus(path);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling GitApi.GetStatus: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetStatusWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get Git status
    ApiResponse<GitStatus> response = apiInstance.GetStatusWithHttpInfo(path);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling GitApi.GetStatusWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **path** | **string** | Repository path |  |

### Return type

[**GitStatus**](GitStatus.md)

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

<a id="listbranches"></a>
# **ListBranches**
> ListBranchResponse ListBranches (string path)

List branches

Get a list of all branches in the Git repository

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
    public class ListBranchesExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new GitApi(httpClient, config, httpClientHandler);
            var path = "path_example";  // string | Repository path

            try
            {
                // List branches
                ListBranchResponse result = apiInstance.ListBranches(path);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling GitApi.ListBranches: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ListBranchesWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // List branches
    ApiResponse<ListBranchResponse> response = apiInstance.ListBranchesWithHttpInfo(path);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling GitApi.ListBranchesWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **path** | **string** | Repository path |  |

### Return type

[**ListBranchResponse**](ListBranchResponse.md)

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

<a id="pullchanges"></a>
# **PullChanges**
> void PullChanges (GitRepoRequest request)

Pull changes from remote

Pull changes from the remote Git repository

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
    public class PullChangesExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new GitApi(httpClient, config, httpClientHandler);
            var request = new GitRepoRequest(); // GitRepoRequest | Pull request

            try
            {
                // Pull changes from remote
                apiInstance.PullChanges(request);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling GitApi.PullChanges: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the PullChangesWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Pull changes from remote
    apiInstance.PullChangesWithHttpInfo(request);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling GitApi.PullChangesWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **request** | [**GitRepoRequest**](GitRepoRequest.md) | Pull request |  |

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="pushchanges"></a>
# **PushChanges**
> void PushChanges (GitRepoRequest request)

Push changes to remote

Push local changes to the remote Git repository

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
    public class PushChangesExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new GitApi(httpClient, config, httpClientHandler);
            var request = new GitRepoRequest(); // GitRepoRequest | Push request

            try
            {
                // Push changes to remote
                apiInstance.PushChanges(request);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling GitApi.PushChanges: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the PushChangesWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Push changes to remote
    apiInstance.PushChangesWithHttpInfo(request);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling GitApi.PushChangesWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **request** | [**GitRepoRequest**](GitRepoRequest.md) | Push request |  |

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

