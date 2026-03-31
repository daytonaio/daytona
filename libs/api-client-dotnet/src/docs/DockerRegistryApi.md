# Daytona.ApiClient.Api.DockerRegistryApi

All URIs are relative to *http://localhost:3000*

| Method | HTTP request | Description |
|--------|--------------|-------------|
| [**CreateRegistry**](DockerRegistryApi.md#createregistry) | **POST** /docker-registry | Create registry |
| [**DeleteRegistry**](DockerRegistryApi.md#deleteregistry) | **DELETE** /docker-registry/{id} | Delete registry |
| [**GetRegistry**](DockerRegistryApi.md#getregistry) | **GET** /docker-registry/{id} | Get registry |
| [**GetTransientPushAccess**](DockerRegistryApi.md#gettransientpushaccess) | **GET** /docker-registry/registry-push-access | Get temporary registry access for pushing snapshots |
| [**ListRegistries**](DockerRegistryApi.md#listregistries) | **GET** /docker-registry | List registries |
| [**SetDefaultRegistry**](DockerRegistryApi.md#setdefaultregistry) | **POST** /docker-registry/{id}/set-default | Set default registry |
| [**UpdateRegistry**](DockerRegistryApi.md#updateregistry) | **PATCH** /docker-registry/{id} | Update registry |

<a id="createregistry"></a>
# **CreateRegistry**
> DockerRegistry CreateRegistry (CreateDockerRegistry createDockerRegistry, string? xDaytonaOrganizationID = null)

Create registry

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
    public class CreateRegistryExample
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
            var apiInstance = new DockerRegistryApi(httpClient, config, httpClientHandler);
            var createDockerRegistry = new CreateDockerRegistry(); // CreateDockerRegistry | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Create registry
                DockerRegistry result = apiInstance.CreateRegistry(createDockerRegistry, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DockerRegistryApi.CreateRegistry: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the CreateRegistryWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Create registry
    ApiResponse<DockerRegistry> response = apiInstance.CreateRegistryWithHttpInfo(createDockerRegistry, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DockerRegistryApi.CreateRegistryWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **createDockerRegistry** | [**CreateDockerRegistry**](CreateDockerRegistry.md) |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**DockerRegistry**](DockerRegistry.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** | The docker registry has been successfully created. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="deleteregistry"></a>
# **DeleteRegistry**
> void DeleteRegistry (string id, string? xDaytonaOrganizationID = null)

Delete registry

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
    public class DeleteRegistryExample
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
            var apiInstance = new DockerRegistryApi(httpClient, config, httpClientHandler);
            var id = "id_example";  // string | ID of the docker registry
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Delete registry
                apiInstance.DeleteRegistry(id, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DockerRegistryApi.DeleteRegistry: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DeleteRegistryWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Delete registry
    apiInstance.DeleteRegistryWithHttpInfo(id, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DockerRegistryApi.DeleteRegistryWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **id** | **string** | ID of the docker registry |  |
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
| **204** | The docker registry has been successfully deleted. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getregistry"></a>
# **GetRegistry**
> DockerRegistry GetRegistry (string id, string? xDaytonaOrganizationID = null)

Get registry

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
    public class GetRegistryExample
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
            var apiInstance = new DockerRegistryApi(httpClient, config, httpClientHandler);
            var id = "id_example";  // string | ID of the docker registry
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Get registry
                DockerRegistry result = apiInstance.GetRegistry(id, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DockerRegistryApi.GetRegistry: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetRegistryWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get registry
    ApiResponse<DockerRegistry> response = apiInstance.GetRegistryWithHttpInfo(id, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DockerRegistryApi.GetRegistryWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **id** | **string** | ID of the docker registry |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**DockerRegistry**](DockerRegistry.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The docker registry |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="gettransientpushaccess"></a>
# **GetTransientPushAccess**
> RegistryPushAccessDto GetTransientPushAccess (string? xDaytonaOrganizationID = null, string? regionId = null)

Get temporary registry access for pushing snapshots

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
    public class GetTransientPushAccessExample
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
            var apiInstance = new DockerRegistryApi(httpClient, config, httpClientHandler);
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 
            var regionId = "regionId_example";  // string? | ID of the region where the snapshot will be available (defaults to organization default region) (optional) 

            try
            {
                // Get temporary registry access for pushing snapshots
                RegistryPushAccessDto result = apiInstance.GetTransientPushAccess(xDaytonaOrganizationID, regionId);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DockerRegistryApi.GetTransientPushAccess: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetTransientPushAccessWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get temporary registry access for pushing snapshots
    ApiResponse<RegistryPushAccessDto> response = apiInstance.GetTransientPushAccessWithHttpInfo(xDaytonaOrganizationID, regionId);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DockerRegistryApi.GetTransientPushAccessWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |
| **regionId** | **string?** | ID of the region where the snapshot will be available (defaults to organization default region) | [optional]  |

### Return type

[**RegistryPushAccessDto**](RegistryPushAccessDto.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Temporary registry access has been generated |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="listregistries"></a>
# **ListRegistries**
> List&lt;DockerRegistry&gt; ListRegistries (string? xDaytonaOrganizationID = null)

List registries

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
    public class ListRegistriesExample
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
            var apiInstance = new DockerRegistryApi(httpClient, config, httpClientHandler);
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // List registries
                List<DockerRegistry> result = apiInstance.ListRegistries(xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DockerRegistryApi.ListRegistries: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ListRegistriesWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // List registries
    ApiResponse<List<DockerRegistry>> response = apiInstance.ListRegistriesWithHttpInfo(xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DockerRegistryApi.ListRegistriesWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**List&lt;DockerRegistry&gt;**](DockerRegistry.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | List of all docker registries |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="setdefaultregistry"></a>
# **SetDefaultRegistry**
> DockerRegistry SetDefaultRegistry (string id, string? xDaytonaOrganizationID = null)

Set default registry

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
    public class SetDefaultRegistryExample
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
            var apiInstance = new DockerRegistryApi(httpClient, config, httpClientHandler);
            var id = "id_example";  // string | ID of the docker registry
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Set default registry
                DockerRegistry result = apiInstance.SetDefaultRegistry(id, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DockerRegistryApi.SetDefaultRegistry: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the SetDefaultRegistryWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Set default registry
    ApiResponse<DockerRegistry> response = apiInstance.SetDefaultRegistryWithHttpInfo(id, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DockerRegistryApi.SetDefaultRegistryWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **id** | **string** | ID of the docker registry |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**DockerRegistry**](DockerRegistry.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The docker registry has been set as default. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="updateregistry"></a>
# **UpdateRegistry**
> DockerRegistry UpdateRegistry (string id, UpdateDockerRegistry updateDockerRegistry, string? xDaytonaOrganizationID = null)

Update registry

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
    public class UpdateRegistryExample
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
            var apiInstance = new DockerRegistryApi(httpClient, config, httpClientHandler);
            var id = "id_example";  // string | ID of the docker registry
            var updateDockerRegistry = new UpdateDockerRegistry(); // UpdateDockerRegistry | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Update registry
                DockerRegistry result = apiInstance.UpdateRegistry(id, updateDockerRegistry, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DockerRegistryApi.UpdateRegistry: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UpdateRegistryWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Update registry
    ApiResponse<DockerRegistry> response = apiInstance.UpdateRegistryWithHttpInfo(id, updateDockerRegistry, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DockerRegistryApi.UpdateRegistryWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **id** | **string** | ID of the docker registry |  |
| **updateDockerRegistry** | [**UpdateDockerRegistry**](UpdateDockerRegistry.md) |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**DockerRegistry**](DockerRegistry.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The docker registry has been successfully updated. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

