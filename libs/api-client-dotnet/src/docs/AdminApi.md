# Daytona.ApiClient.Api.AdminApi

All URIs are relative to *http://localhost:3000*

| Method | HTTP request | Description |
|--------|--------------|-------------|
| [**AdminCreateRunner**](AdminApi.md#admincreaterunner) | **POST** /admin/runners | Create runner |
| [**AdminDeleteRunner**](AdminApi.md#admindeleterunner) | **DELETE** /admin/runners/{id} | Delete runner |
| [**AdminGetRunnerById**](AdminApi.md#admingetrunnerbyid) | **GET** /admin/runners/{id} | Get runner by ID |
| [**AdminListRunners**](AdminApi.md#adminlistrunners) | **GET** /admin/runners | List all runners |
| [**AdminRecoverSandbox**](AdminApi.md#adminrecoversandbox) | **POST** /admin/sandbox/{sandboxId}/recover | Recover sandbox from error state as an admin |
| [**AdminUpdateRunnerScheduling**](AdminApi.md#adminupdaterunnerscheduling) | **PATCH** /admin/runners/{id}/scheduling | Update runner scheduling status |

<a id="admincreaterunner"></a>
# **AdminCreateRunner**
> CreateRunnerResponse AdminCreateRunner (AdminCreateRunner adminCreateRunner)

Create runner

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
    public class AdminCreateRunnerExample
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
            var apiInstance = new AdminApi(httpClient, config, httpClientHandler);
            var adminCreateRunner = new AdminCreateRunner(); // AdminCreateRunner | 

            try
            {
                // Create runner
                CreateRunnerResponse result = apiInstance.AdminCreateRunner(adminCreateRunner);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling AdminApi.AdminCreateRunner: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the AdminCreateRunnerWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Create runner
    ApiResponse<CreateRunnerResponse> response = apiInstance.AdminCreateRunnerWithHttpInfo(adminCreateRunner);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling AdminApi.AdminCreateRunnerWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **adminCreateRunner** | [**AdminCreateRunner**](AdminCreateRunner.md) |  |  |

### Return type

[**CreateRunnerResponse**](CreateRunnerResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** |  |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="admindeleterunner"></a>
# **AdminDeleteRunner**
> void AdminDeleteRunner (string id)

Delete runner

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
    public class AdminDeleteRunnerExample
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
            var apiInstance = new AdminApi(httpClient, config, httpClientHandler);
            var id = "id_example";  // string | Runner ID

            try
            {
                // Delete runner
                apiInstance.AdminDeleteRunner(id);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling AdminApi.AdminDeleteRunner: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the AdminDeleteRunnerWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Delete runner
    apiInstance.AdminDeleteRunnerWithHttpInfo(id);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling AdminApi.AdminDeleteRunnerWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **id** | **string** | Runner ID |  |

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
| **204** |  |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="admingetrunnerbyid"></a>
# **AdminGetRunnerById**
> RunnerFull AdminGetRunnerById (string id)

Get runner by ID

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
    public class AdminGetRunnerByIdExample
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
            var apiInstance = new AdminApi(httpClient, config, httpClientHandler);
            var id = "id_example";  // string | Runner ID

            try
            {
                // Get runner by ID
                RunnerFull result = apiInstance.AdminGetRunnerById(id);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling AdminApi.AdminGetRunnerById: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the AdminGetRunnerByIdWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get runner by ID
    ApiResponse<RunnerFull> response = apiInstance.AdminGetRunnerByIdWithHttpInfo(id);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling AdminApi.AdminGetRunnerByIdWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **id** | **string** | Runner ID |  |

### Return type

[**RunnerFull**](RunnerFull.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** |  |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="adminlistrunners"></a>
# **AdminListRunners**
> List&lt;RunnerFull&gt; AdminListRunners (string? regionId = null)

List all runners

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
    public class AdminListRunnersExample
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
            var apiInstance = new AdminApi(httpClient, config, httpClientHandler);
            var regionId = "regionId_example";  // string? | Filter runners by region ID (optional) 

            try
            {
                // List all runners
                List<RunnerFull> result = apiInstance.AdminListRunners(regionId);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling AdminApi.AdminListRunners: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the AdminListRunnersWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // List all runners
    ApiResponse<List<RunnerFull>> response = apiInstance.AdminListRunnersWithHttpInfo(regionId);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling AdminApi.AdminListRunnersWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **regionId** | **string?** | Filter runners by region ID | [optional]  |

### Return type

[**List&lt;RunnerFull&gt;**](RunnerFull.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** |  |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="adminrecoversandbox"></a>
# **AdminRecoverSandbox**
> Sandbox AdminRecoverSandbox (string sandboxId)

Recover sandbox from error state as an admin

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
    public class AdminRecoverSandboxExample
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
            var apiInstance = new AdminApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | ID of the sandbox

            try
            {
                // Recover sandbox from error state as an admin
                Sandbox result = apiInstance.AdminRecoverSandbox(sandboxId);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling AdminApi.AdminRecoverSandbox: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the AdminRecoverSandboxWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Recover sandbox from error state as an admin
    ApiResponse<Sandbox> response = apiInstance.AdminRecoverSandboxWithHttpInfo(sandboxId);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling AdminApi.AdminRecoverSandboxWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** | ID of the sandbox |  |

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

<a id="adminupdaterunnerscheduling"></a>
# **AdminUpdateRunnerScheduling**
> void AdminUpdateRunnerScheduling (string id)

Update runner scheduling status

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
    public class AdminUpdateRunnerSchedulingExample
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
            var apiInstance = new AdminApi(httpClient, config, httpClientHandler);
            var id = "id_example";  // string | 

            try
            {
                // Update runner scheduling status
                apiInstance.AdminUpdateRunnerScheduling(id);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling AdminApi.AdminUpdateRunnerScheduling: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the AdminUpdateRunnerSchedulingWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Update runner scheduling status
    apiInstance.AdminUpdateRunnerSchedulingWithHttpInfo(id);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling AdminApi.AdminUpdateRunnerSchedulingWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **id** | **string** |  |  |

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
| **204** |  |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

