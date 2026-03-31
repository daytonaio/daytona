# Daytona.ApiClient.Api.PreviewApi

All URIs are relative to *http://localhost:3000*

| Method | HTTP request | Description |
|--------|--------------|-------------|
| [**GetSandboxIdFromSignedPreviewUrlToken**](PreviewApi.md#getsandboxidfromsignedpreviewurltoken) | **GET** /preview/{signedPreviewToken}/{port}/sandbox-id | Get sandbox ID from signed preview URL token |
| [**HasSandboxAccess**](PreviewApi.md#hassandboxaccess) | **GET** /preview/{sandboxId}/access | Check if user has access to the sandbox |
| [**IsSandboxPublic**](PreviewApi.md#issandboxpublic) | **GET** /preview/{sandboxId}/public | Check if sandbox is public |
| [**IsValidAuthToken**](PreviewApi.md#isvalidauthtoken) | **GET** /preview/{sandboxId}/validate/{authToken} | Check if sandbox auth token is valid |

<a id="getsandboxidfromsignedpreviewurltoken"></a>
# **GetSandboxIdFromSignedPreviewUrlToken**
> string GetSandboxIdFromSignedPreviewUrlToken (string signedPreviewToken, decimal port)

Get sandbox ID from signed preview URL token

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
    public class GetSandboxIdFromSignedPreviewUrlTokenExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:3000";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new PreviewApi(httpClient, config, httpClientHandler);
            var signedPreviewToken = "signedPreviewToken_example";  // string | Signed preview URL token
            var port = 8.14D;  // decimal | Port number to get sandbox ID from signed preview URL token

            try
            {
                // Get sandbox ID from signed preview URL token
                string result = apiInstance.GetSandboxIdFromSignedPreviewUrlToken(signedPreviewToken, port);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling PreviewApi.GetSandboxIdFromSignedPreviewUrlToken: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetSandboxIdFromSignedPreviewUrlTokenWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get sandbox ID from signed preview URL token
    ApiResponse<string> response = apiInstance.GetSandboxIdFromSignedPreviewUrlTokenWithHttpInfo(signedPreviewToken, port);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling PreviewApi.GetSandboxIdFromSignedPreviewUrlTokenWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **signedPreviewToken** | **string** | Signed preview URL token |  |
| **port** | **decimal** | Port number to get sandbox ID from signed preview URL token |  |

### Return type

**string**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Sandbox ID from signed preview URL token |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="hassandboxaccess"></a>
# **HasSandboxAccess**
> bool HasSandboxAccess (string sandboxId)

Check if user has access to the sandbox

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
    public class HasSandboxAccessExample
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
            var apiInstance = new PreviewApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | 

            try
            {
                // Check if user has access to the sandbox
                bool result = apiInstance.HasSandboxAccess(sandboxId);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling PreviewApi.HasSandboxAccess: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the HasSandboxAccessWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Check if user has access to the sandbox
    ApiResponse<bool> response = apiInstance.HasSandboxAccessWithHttpInfo(sandboxId);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling PreviewApi.HasSandboxAccessWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** |  |  |

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
| **200** | User access status to the sandbox |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="issandboxpublic"></a>
# **IsSandboxPublic**
> bool IsSandboxPublic (string sandboxId)

Check if sandbox is public

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
    public class IsSandboxPublicExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:3000";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new PreviewApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | ID of the sandbox

            try
            {
                // Check if sandbox is public
                bool result = apiInstance.IsSandboxPublic(sandboxId);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling PreviewApi.IsSandboxPublic: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the IsSandboxPublicWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Check if sandbox is public
    ApiResponse<bool> response = apiInstance.IsSandboxPublicWithHttpInfo(sandboxId);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling PreviewApi.IsSandboxPublicWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** | ID of the sandbox |  |

### Return type

**bool**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Public status of the sandbox |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="isvalidauthtoken"></a>
# **IsValidAuthToken**
> bool IsValidAuthToken (string sandboxId, string authToken)

Check if sandbox auth token is valid

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
    public class IsValidAuthTokenExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:3000";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new PreviewApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | ID of the sandbox
            var authToken = "authToken_example";  // string | Auth token of the sandbox

            try
            {
                // Check if sandbox auth token is valid
                bool result = apiInstance.IsValidAuthToken(sandboxId, authToken);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling PreviewApi.IsValidAuthToken: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the IsValidAuthTokenWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Check if sandbox auth token is valid
    ApiResponse<bool> response = apiInstance.IsValidAuthTokenWithHttpInfo(sandboxId, authToken);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling PreviewApi.IsValidAuthTokenWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** | ID of the sandbox |  |
| **authToken** | **string** | Auth token of the sandbox |  |

### Return type

**bool**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Sandbox auth token validation status |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

