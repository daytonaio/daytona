# Daytona.ApiClient.Api.HealthApi

All URIs are relative to *http://localhost:3000*

| Method | HTTP request | Description |
|--------|--------------|-------------|
| [**HealthControllerCheck**](HealthApi.md#healthcontrollercheck) | **GET** /health/ready |  |
| [**HealthControllerLive**](HealthApi.md#healthcontrollerlive) | **GET** /health |  |

<a id="healthcontrollercheck"></a>
# **HealthControllerCheck**
> HealthControllerCheck200Response HealthControllerCheck ()



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
    public class HealthControllerCheckExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:3000";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new HealthApi(httpClient, config, httpClientHandler);

            try
            {
                HealthControllerCheck200Response result = apiInstance.HealthControllerCheck();
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling HealthApi.HealthControllerCheck: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the HealthControllerCheckWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    ApiResponse<HealthControllerCheck200Response> response = apiInstance.HealthControllerCheckWithHttpInfo();
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling HealthApi.HealthControllerCheckWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters
This endpoint does not need any parameter.
### Return type

[**HealthControllerCheck200Response**](HealthControllerCheck200Response.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The Health Check is successful |  -  |
| **503** | The Health Check is not successful |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="healthcontrollerlive"></a>
# **HealthControllerLive**
> void HealthControllerLive ()



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
    public class HealthControllerLiveExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:3000";
            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new HealthApi(httpClient, config, httpClientHandler);

            try
            {
                apiInstance.HealthControllerLive();
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling HealthApi.HealthControllerLive: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the HealthControllerLiveWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    apiInstance.HealthControllerLiveWithHttpInfo();
}
catch (ApiException e)
{
    Debug.Print("Exception when calling HealthApi.HealthControllerLiveWithHttpInfo: " + e.Message);
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
| **200** |  |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

