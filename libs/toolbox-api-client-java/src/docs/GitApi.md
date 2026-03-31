# GitApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**addFiles**](GitApi.md#addFiles) | **POST** /git/add | Add files to Git staging |
| [**checkoutBranch**](GitApi.md#checkoutBranch) | **POST** /git/checkout | Checkout branch or commit |
| [**cloneRepository**](GitApi.md#cloneRepository) | **POST** /git/clone | Clone a Git repository |
| [**commitChanges**](GitApi.md#commitChanges) | **POST** /git/commit | Commit changes |
| [**createBranch**](GitApi.md#createBranch) | **POST** /git/branches | Create a new branch |
| [**deleteBranch**](GitApi.md#deleteBranch) | **DELETE** /git/branches | Delete a branch |
| [**getCommitHistory**](GitApi.md#getCommitHistory) | **GET** /git/history | Get commit history |
| [**getStatus**](GitApi.md#getStatus) | **GET** /git/status | Get Git status |
| [**listBranches**](GitApi.md#listBranches) | **GET** /git/branches | List branches |
| [**pullChanges**](GitApi.md#pullChanges) | **POST** /git/pull | Pull changes from remote |
| [**pushChanges**](GitApi.md#pushChanges) | **POST** /git/push | Push changes to remote |


<a id="addFiles"></a>
# **addFiles**
> addFiles(request)

Add files to Git staging

Add files to the Git staging area

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.GitApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    GitApi apiInstance = new GitApi(defaultClient);
    GitAddRequest request = new GitAddRequest(); // GitAddRequest | Add files request
    try {
      apiInstance.addFiles(request);
    } catch (ApiException e) {
      System.err.println("Exception when calling GitApi#addFiles");
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
| **request** | [**GitAddRequest**](GitAddRequest.md)| Add files request | |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="checkoutBranch"></a>
# **checkoutBranch**
> checkoutBranch(request)

Checkout branch or commit

Switch to a different branch or commit in the Git repository

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.GitApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    GitApi apiInstance = new GitApi(defaultClient);
    GitCheckoutRequest request = new GitCheckoutRequest(); // GitCheckoutRequest | Checkout request
    try {
      apiInstance.checkoutBranch(request);
    } catch (ApiException e) {
      System.err.println("Exception when calling GitApi#checkoutBranch");
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
| **request** | [**GitCheckoutRequest**](GitCheckoutRequest.md)| Checkout request | |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="cloneRepository"></a>
# **cloneRepository**
> cloneRepository(request)

Clone a Git repository

Clone a Git repository to the specified path

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.GitApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    GitApi apiInstance = new GitApi(defaultClient);
    GitCloneRequest request = new GitCloneRequest(); // GitCloneRequest | Clone repository request
    try {
      apiInstance.cloneRepository(request);
    } catch (ApiException e) {
      System.err.println("Exception when calling GitApi#cloneRepository");
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
| **request** | [**GitCloneRequest**](GitCloneRequest.md)| Clone repository request | |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="commitChanges"></a>
# **commitChanges**
> GitCommitResponse commitChanges(request)

Commit changes

Commit staged changes to the Git repository

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.GitApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    GitApi apiInstance = new GitApi(defaultClient);
    GitCommitRequest request = new GitCommitRequest(); // GitCommitRequest | Commit request
    try {
      GitCommitResponse result = apiInstance.commitChanges(request);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling GitApi#commitChanges");
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
| **request** | [**GitCommitRequest**](GitCommitRequest.md)| Commit request | |

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

<a id="createBranch"></a>
# **createBranch**
> createBranch(request)

Create a new branch

Create a new branch in the Git repository

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.GitApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    GitApi apiInstance = new GitApi(defaultClient);
    GitBranchRequest request = new GitBranchRequest(); // GitBranchRequest | Create branch request
    try {
      apiInstance.createBranch(request);
    } catch (ApiException e) {
      System.err.println("Exception when calling GitApi#createBranch");
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
| **request** | [**GitBranchRequest**](GitBranchRequest.md)| Create branch request | |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** | Created |  -  |

<a id="deleteBranch"></a>
# **deleteBranch**
> deleteBranch(request)

Delete a branch

Delete a branch from the Git repository

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.GitApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    GitApi apiInstance = new GitApi(defaultClient);
    GitGitDeleteBranchRequest request = new GitGitDeleteBranchRequest(); // GitGitDeleteBranchRequest | Delete branch request
    try {
      apiInstance.deleteBranch(request);
    } catch (ApiException e) {
      System.err.println("Exception when calling GitApi#deleteBranch");
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
| **request** | [**GitGitDeleteBranchRequest**](GitGitDeleteBranchRequest.md)| Delete branch request | |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | No Content |  -  |

<a id="getCommitHistory"></a>
# **getCommitHistory**
> List&lt;GitCommitInfo&gt; getCommitHistory(path)

Get commit history

Get the commit history of the Git repository

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.GitApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    GitApi apiInstance = new GitApi(defaultClient);
    String path = "path_example"; // String | Repository path
    try {
      List<GitCommitInfo> result = apiInstance.getCommitHistory(path);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling GitApi#getCommitHistory");
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
| **path** | **String**| Repository path | |

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

<a id="getStatus"></a>
# **getStatus**
> GitStatus getStatus(path)

Get Git status

Get the Git status of the repository at the specified path

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.GitApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    GitApi apiInstance = new GitApi(defaultClient);
    String path = "path_example"; // String | Repository path
    try {
      GitStatus result = apiInstance.getStatus(path);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling GitApi#getStatus");
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
| **path** | **String**| Repository path | |

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

<a id="listBranches"></a>
# **listBranches**
> ListBranchResponse listBranches(path)

List branches

Get a list of all branches in the Git repository

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.GitApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    GitApi apiInstance = new GitApi(defaultClient);
    String path = "path_example"; // String | Repository path
    try {
      ListBranchResponse result = apiInstance.listBranches(path);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling GitApi#listBranches");
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
| **path** | **String**| Repository path | |

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

<a id="pullChanges"></a>
# **pullChanges**
> pullChanges(request)

Pull changes from remote

Pull changes from the remote Git repository

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.GitApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    GitApi apiInstance = new GitApi(defaultClient);
    GitRepoRequest request = new GitRepoRequest(); // GitRepoRequest | Pull request
    try {
      apiInstance.pullChanges(request);
    } catch (ApiException e) {
      System.err.println("Exception when calling GitApi#pullChanges");
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
| **request** | [**GitRepoRequest**](GitRepoRequest.md)| Pull request | |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="pushChanges"></a>
# **pushChanges**
> pushChanges(request)

Push changes to remote

Push local changes to the remote Git repository

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.GitApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    GitApi apiInstance = new GitApi(defaultClient);
    GitRepoRequest request = new GitRepoRequest(); // GitRepoRequest | Push request
    try {
      apiInstance.pushChanges(request);
    } catch (ApiException e) {
      System.err.println("Exception when calling GitApi#pushChanges");
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
| **request** | [**GitRepoRequest**](GitRepoRequest.md)| Push request | |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

