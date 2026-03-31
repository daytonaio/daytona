# Daytona.ApiClient.Api.ObjectStorageApi

All URIs are relative to *http://localhost:3000*

| Method | HTTP request | Description |
|--------|--------------|-------------|
| [**GetPushAccess**](ObjectStorageApi.md#getpushaccess) | **GET** /object-storage/push-access | Get temporary storage access for pushing objects |

<a id="getpushaccess"></a>
# **GetPushAccess**
> StorageAccessDto GetPushAccess (string? xDaytonaOrganizationID = null)

Get temporary storage access for pushing objects

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
    public class GetPushAccessExample
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
            var apiInstance = new ObjectStorageApi(httpClient, config, httpClientHandler);
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Get temporary storage access for pushing objects
                StorageAccessDto result = apiInstance.GetPushAccess(xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling ObjectStorageApi.GetPushAccess: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetPushAccessWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get temporary storage access for pushing objects
    ApiResponse<StorageAccessDto> response = apiInstance.GetPushAccessWithHttpInfo(xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling ObjectStorageApi.GetPushAccessWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**StorageAccessDto**](StorageAccessDto.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Temporary storage access has been generated |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

