# Daytona.ApiClient.Api.OrganizationsApi

All URIs are relative to *http://localhost:3000*

| Method | HTTP request | Description |
|--------|--------------|-------------|
| [**AcceptOrganizationInvitation**](OrganizationsApi.md#acceptorganizationinvitation) | **POST** /organizations/invitations/{invitationId}/accept | Accept organization invitation |
| [**CancelOrganizationInvitation**](OrganizationsApi.md#cancelorganizationinvitation) | **POST** /organizations/{organizationId}/invitations/{invitationId}/cancel | Cancel organization invitation |
| [**CreateOrganization**](OrganizationsApi.md#createorganization) | **POST** /organizations | Create organization |
| [**CreateOrganizationInvitation**](OrganizationsApi.md#createorganizationinvitation) | **POST** /organizations/{organizationId}/invitations | Create organization invitation |
| [**CreateOrganizationRole**](OrganizationsApi.md#createorganizationrole) | **POST** /organizations/{organizationId}/roles | Create organization role |
| [**CreateRegion**](OrganizationsApi.md#createregion) | **POST** /regions | Create a new region |
| [**DeclineOrganizationInvitation**](OrganizationsApi.md#declineorganizationinvitation) | **POST** /organizations/invitations/{invitationId}/decline | Decline organization invitation |
| [**DeleteOrganization**](OrganizationsApi.md#deleteorganization) | **DELETE** /organizations/{organizationId} | Delete organization |
| [**DeleteOrganizationMember**](OrganizationsApi.md#deleteorganizationmember) | **DELETE** /organizations/{organizationId}/users/{userId} | Delete organization member |
| [**DeleteOrganizationRole**](OrganizationsApi.md#deleteorganizationrole) | **DELETE** /organizations/{organizationId}/roles/{roleId} | Delete organization role |
| [**DeleteRegion**](OrganizationsApi.md#deleteregion) | **DELETE** /regions/{id} | Delete a region |
| [**GetOrganization**](OrganizationsApi.md#getorganization) | **GET** /organizations/{organizationId} | Get organization by ID |
| [**GetOrganizationBySandboxId**](OrganizationsApi.md#getorganizationbysandboxid) | **GET** /organizations/by-sandbox-id/{sandboxId} | Get organization by sandbox ID |
| [**GetOrganizationInvitationsCountForAuthenticatedUser**](OrganizationsApi.md#getorganizationinvitationscountforauthenticateduser) | **GET** /organizations/invitations/count | Get count of organization invitations for authenticated user |
| [**GetOrganizationOtelConfigBySandboxAuthToken**](OrganizationsApi.md#getorganizationotelconfigbysandboxauthtoken) | **GET** /organizations/otel-config/by-sandbox-auth-token/{authToken} | Get organization OTEL config by sandbox auth token |
| [**GetOrganizationUsageOverview**](OrganizationsApi.md#getorganizationusageoverview) | **GET** /organizations/{organizationId}/usage | Get organization current usage overview |
| [**GetRegionById**](OrganizationsApi.md#getregionbyid) | **GET** /regions/{id} | Get region by ID |
| [**GetRegionQuotaBySandboxId**](OrganizationsApi.md#getregionquotabysandboxid) | **GET** /organizations/region-quota/by-sandbox-id/{sandboxId} | Get region quota by sandbox ID |
| [**LeaveOrganization**](OrganizationsApi.md#leaveorganization) | **POST** /organizations/{organizationId}/leave | Leave organization |
| [**ListAvailableRegions**](OrganizationsApi.md#listavailableregions) | **GET** /regions | List all available regions for the organization |
| [**ListOrganizationInvitations**](OrganizationsApi.md#listorganizationinvitations) | **GET** /organizations/{organizationId}/invitations | List pending organization invitations |
| [**ListOrganizationInvitationsForAuthenticatedUser**](OrganizationsApi.md#listorganizationinvitationsforauthenticateduser) | **GET** /organizations/invitations | List organization invitations for authenticated user |
| [**ListOrganizationMembers**](OrganizationsApi.md#listorganizationmembers) | **GET** /organizations/{organizationId}/users | List organization members |
| [**ListOrganizationRoles**](OrganizationsApi.md#listorganizationroles) | **GET** /organizations/{organizationId}/roles | List organization roles |
| [**ListOrganizations**](OrganizationsApi.md#listorganizations) | **GET** /organizations | List organizations |
| [**RegenerateProxyApiKey**](OrganizationsApi.md#regenerateproxyapikey) | **POST** /regions/{id}/regenerate-proxy-api-key | Regenerate proxy API key for a region |
| [**RegenerateSnapshotManagerCredentials**](OrganizationsApi.md#regeneratesnapshotmanagercredentials) | **POST** /regions/{id}/regenerate-snapshot-manager-credentials | Regenerate snapshot manager credentials for a region |
| [**RegenerateSshGatewayApiKey**](OrganizationsApi.md#regeneratesshgatewayapikey) | **POST** /regions/{id}/regenerate-ssh-gateway-api-key | Regenerate SSH gateway API key for a region |
| [**SetOrganizationDefaultRegion**](OrganizationsApi.md#setorganizationdefaultregion) | **PATCH** /organizations/{organizationId}/default-region | Set default region for organization |
| [**SuspendOrganization**](OrganizationsApi.md#suspendorganization) | **POST** /organizations/{organizationId}/suspend | Suspend organization |
| [**UnsuspendOrganization**](OrganizationsApi.md#unsuspendorganization) | **POST** /organizations/{organizationId}/unsuspend | Unsuspend organization |
| [**UpdateAccessForOrganizationMember**](OrganizationsApi.md#updateaccessfororganizationmember) | **POST** /organizations/{organizationId}/users/{userId}/access | Update access for organization member |
| [**UpdateExperimentalConfig**](OrganizationsApi.md#updateexperimentalconfig) | **PUT** /organizations/{organizationId}/experimental-config | Update experimental configuration |
| [**UpdateOrganizationInvitation**](OrganizationsApi.md#updateorganizationinvitation) | **PUT** /organizations/{organizationId}/invitations/{invitationId} | Update organization invitation |
| [**UpdateOrganizationQuota**](OrganizationsApi.md#updateorganizationquota) | **PATCH** /organizations/{organizationId}/quota | Update organization quota |
| [**UpdateOrganizationRegionQuota**](OrganizationsApi.md#updateorganizationregionquota) | **PATCH** /organizations/{organizationId}/quota/{regionId} | Update organization region quota |
| [**UpdateOrganizationRole**](OrganizationsApi.md#updateorganizationrole) | **PUT** /organizations/{organizationId}/roles/{roleId} | Update organization role |
| [**UpdateRegion**](OrganizationsApi.md#updateregion) | **PATCH** /regions/{id} | Update region configuration |
| [**UpdateSandboxDefaultLimitedNetworkEgress**](OrganizationsApi.md#updatesandboxdefaultlimitednetworkegress) | **POST** /organizations/{organizationId}/sandbox-default-limited-network-egress | Update sandbox default limited network egress |

<a id="acceptorganizationinvitation"></a>
# **AcceptOrganizationInvitation**
> OrganizationInvitation AcceptOrganizationInvitation (string invitationId)

Accept organization invitation

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
    public class AcceptOrganizationInvitationExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var invitationId = "invitationId_example";  // string | Invitation ID

            try
            {
                // Accept organization invitation
                OrganizationInvitation result = apiInstance.AcceptOrganizationInvitation(invitationId);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.AcceptOrganizationInvitation: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the AcceptOrganizationInvitationWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Accept organization invitation
    ApiResponse<OrganizationInvitation> response = apiInstance.AcceptOrganizationInvitationWithHttpInfo(invitationId);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.AcceptOrganizationInvitationWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **invitationId** | **string** | Invitation ID |  |

### Return type

[**OrganizationInvitation**](OrganizationInvitation.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Organization invitation accepted successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="cancelorganizationinvitation"></a>
# **CancelOrganizationInvitation**
> void CancelOrganizationInvitation (string organizationId, string invitationId)

Cancel organization invitation

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
    public class CancelOrganizationInvitationExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var organizationId = "organizationId_example";  // string | Organization ID
            var invitationId = "invitationId_example";  // string | Invitation ID

            try
            {
                // Cancel organization invitation
                apiInstance.CancelOrganizationInvitation(organizationId, invitationId);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.CancelOrganizationInvitation: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the CancelOrganizationInvitationWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Cancel organization invitation
    apiInstance.CancelOrganizationInvitationWithHttpInfo(organizationId, invitationId);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.CancelOrganizationInvitationWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **organizationId** | **string** | Organization ID |  |
| **invitationId** | **string** | Invitation ID |  |

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
| **204** | Organization invitation cancelled successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="createorganization"></a>
# **CreateOrganization**
> Organization CreateOrganization (CreateOrganization createOrganization)

Create organization

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
    public class CreateOrganizationExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var createOrganization = new CreateOrganization(); // CreateOrganization | 

            try
            {
                // Create organization
                Organization result = apiInstance.CreateOrganization(createOrganization);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.CreateOrganization: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the CreateOrganizationWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Create organization
    ApiResponse<Organization> response = apiInstance.CreateOrganizationWithHttpInfo(createOrganization);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.CreateOrganizationWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **createOrganization** | [**CreateOrganization**](CreateOrganization.md) |  |  |

### Return type

[**Organization**](Organization.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** | Organization created successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="createorganizationinvitation"></a>
# **CreateOrganizationInvitation**
> OrganizationInvitation CreateOrganizationInvitation (string organizationId, CreateOrganizationInvitation createOrganizationInvitation)

Create organization invitation

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
    public class CreateOrganizationInvitationExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var organizationId = "organizationId_example";  // string | Organization ID
            var createOrganizationInvitation = new CreateOrganizationInvitation(); // CreateOrganizationInvitation | 

            try
            {
                // Create organization invitation
                OrganizationInvitation result = apiInstance.CreateOrganizationInvitation(organizationId, createOrganizationInvitation);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.CreateOrganizationInvitation: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the CreateOrganizationInvitationWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Create organization invitation
    ApiResponse<OrganizationInvitation> response = apiInstance.CreateOrganizationInvitationWithHttpInfo(organizationId, createOrganizationInvitation);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.CreateOrganizationInvitationWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **organizationId** | **string** | Organization ID |  |
| **createOrganizationInvitation** | [**CreateOrganizationInvitation**](CreateOrganizationInvitation.md) |  |  |

### Return type

[**OrganizationInvitation**](OrganizationInvitation.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** | Organization invitation created successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="createorganizationrole"></a>
# **CreateOrganizationRole**
> OrganizationRole CreateOrganizationRole (string organizationId, CreateOrganizationRole createOrganizationRole)

Create organization role

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
    public class CreateOrganizationRoleExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var organizationId = "organizationId_example";  // string | Organization ID
            var createOrganizationRole = new CreateOrganizationRole(); // CreateOrganizationRole | 

            try
            {
                // Create organization role
                OrganizationRole result = apiInstance.CreateOrganizationRole(organizationId, createOrganizationRole);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.CreateOrganizationRole: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the CreateOrganizationRoleWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Create organization role
    ApiResponse<OrganizationRole> response = apiInstance.CreateOrganizationRoleWithHttpInfo(organizationId, createOrganizationRole);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.CreateOrganizationRoleWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **organizationId** | **string** | Organization ID |  |
| **createOrganizationRole** | [**CreateOrganizationRole**](CreateOrganizationRole.md) |  |  |

### Return type

[**OrganizationRole**](OrganizationRole.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** | Organization role created successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="createregion"></a>
# **CreateRegion**
> CreateRegionResponse CreateRegion (CreateRegion createRegion, string? xDaytonaOrganizationID = null)

Create a new region

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
    public class CreateRegionExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var createRegion = new CreateRegion(); // CreateRegion | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Create a new region
                CreateRegionResponse result = apiInstance.CreateRegion(createRegion, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.CreateRegion: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the CreateRegionWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Create a new region
    ApiResponse<CreateRegionResponse> response = apiInstance.CreateRegionWithHttpInfo(createRegion, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.CreateRegionWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **createRegion** | [**CreateRegion**](CreateRegion.md) |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**CreateRegionResponse**](CreateRegionResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** | The region has been successfully created. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="declineorganizationinvitation"></a>
# **DeclineOrganizationInvitation**
> void DeclineOrganizationInvitation (string invitationId)

Decline organization invitation

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
    public class DeclineOrganizationInvitationExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var invitationId = "invitationId_example";  // string | Invitation ID

            try
            {
                // Decline organization invitation
                apiInstance.DeclineOrganizationInvitation(invitationId);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.DeclineOrganizationInvitation: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DeclineOrganizationInvitationWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Decline organization invitation
    apiInstance.DeclineOrganizationInvitationWithHttpInfo(invitationId);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.DeclineOrganizationInvitationWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **invitationId** | **string** | Invitation ID |  |

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
| **200** | Organization invitation declined successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="deleteorganization"></a>
# **DeleteOrganization**
> void DeleteOrganization (string organizationId)

Delete organization

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
    public class DeleteOrganizationExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var organizationId = "organizationId_example";  // string | Organization ID

            try
            {
                // Delete organization
                apiInstance.DeleteOrganization(organizationId);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.DeleteOrganization: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DeleteOrganizationWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Delete organization
    apiInstance.DeleteOrganizationWithHttpInfo(organizationId);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.DeleteOrganizationWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **organizationId** | **string** | Organization ID |  |

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
| **204** | Organization deleted successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="deleteorganizationmember"></a>
# **DeleteOrganizationMember**
> void DeleteOrganizationMember (string organizationId, string userId)

Delete organization member

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
    public class DeleteOrganizationMemberExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var organizationId = "organizationId_example";  // string | Organization ID
            var userId = "userId_example";  // string | User ID

            try
            {
                // Delete organization member
                apiInstance.DeleteOrganizationMember(organizationId, userId);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.DeleteOrganizationMember: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DeleteOrganizationMemberWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Delete organization member
    apiInstance.DeleteOrganizationMemberWithHttpInfo(organizationId, userId);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.DeleteOrganizationMemberWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **organizationId** | **string** | Organization ID |  |
| **userId** | **string** | User ID |  |

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
| **204** | User removed from organization successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="deleteorganizationrole"></a>
# **DeleteOrganizationRole**
> void DeleteOrganizationRole (string organizationId, string roleId)

Delete organization role

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
    public class DeleteOrganizationRoleExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var organizationId = "organizationId_example";  // string | Organization ID
            var roleId = "roleId_example";  // string | Role ID

            try
            {
                // Delete organization role
                apiInstance.DeleteOrganizationRole(organizationId, roleId);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.DeleteOrganizationRole: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DeleteOrganizationRoleWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Delete organization role
    apiInstance.DeleteOrganizationRoleWithHttpInfo(organizationId, roleId);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.DeleteOrganizationRoleWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **organizationId** | **string** | Organization ID |  |
| **roleId** | **string** | Role ID |  |

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
| **204** | Organization role deleted successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="deleteregion"></a>
# **DeleteRegion**
> void DeleteRegion (string id, string? xDaytonaOrganizationID = null)

Delete a region

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
    public class DeleteRegionExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var id = "id_example";  // string | Region ID
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Delete a region
                apiInstance.DeleteRegion(id, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.DeleteRegion: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DeleteRegionWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Delete a region
    apiInstance.DeleteRegionWithHttpInfo(id, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.DeleteRegionWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **id** | **string** | Region ID |  |
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
| **204** | The region has been successfully deleted. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getorganization"></a>
# **GetOrganization**
> Organization GetOrganization (string organizationId)

Get organization by ID

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
    public class GetOrganizationExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var organizationId = "organizationId_example";  // string | Organization ID

            try
            {
                // Get organization by ID
                Organization result = apiInstance.GetOrganization(organizationId);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.GetOrganization: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetOrganizationWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get organization by ID
    ApiResponse<Organization> response = apiInstance.GetOrganizationWithHttpInfo(organizationId);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.GetOrganizationWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **organizationId** | **string** | Organization ID |  |

### Return type

[**Organization**](Organization.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Organization details |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getorganizationbysandboxid"></a>
# **GetOrganizationBySandboxId**
> Organization GetOrganizationBySandboxId (string sandboxId)

Get organization by sandbox ID

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
    public class GetOrganizationBySandboxIdExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | Sandbox ID

            try
            {
                // Get organization by sandbox ID
                Organization result = apiInstance.GetOrganizationBySandboxId(sandboxId);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.GetOrganizationBySandboxId: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetOrganizationBySandboxIdWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get organization by sandbox ID
    ApiResponse<Organization> response = apiInstance.GetOrganizationBySandboxIdWithHttpInfo(sandboxId);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.GetOrganizationBySandboxIdWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** | Sandbox ID |  |

### Return type

[**Organization**](Organization.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Organization |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getorganizationinvitationscountforauthenticateduser"></a>
# **GetOrganizationInvitationsCountForAuthenticatedUser**
> decimal GetOrganizationInvitationsCountForAuthenticatedUser ()

Get count of organization invitations for authenticated user

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
    public class GetOrganizationInvitationsCountForAuthenticatedUserExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);

            try
            {
                // Get count of organization invitations for authenticated user
                decimal result = apiInstance.GetOrganizationInvitationsCountForAuthenticatedUser();
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.GetOrganizationInvitationsCountForAuthenticatedUser: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetOrganizationInvitationsCountForAuthenticatedUserWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get count of organization invitations for authenticated user
    ApiResponse<decimal> response = apiInstance.GetOrganizationInvitationsCountForAuthenticatedUserWithHttpInfo();
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.GetOrganizationInvitationsCountForAuthenticatedUserWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters
This endpoint does not need any parameter.
### Return type

**decimal**

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Count of organization invitations |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getorganizationotelconfigbysandboxauthtoken"></a>
# **GetOrganizationOtelConfigBySandboxAuthToken**
> OtelConfig GetOrganizationOtelConfigBySandboxAuthToken (string authToken)

Get organization OTEL config by sandbox auth token

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
    public class GetOrganizationOtelConfigBySandboxAuthTokenExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var authToken = "authToken_example";  // string | Sandbox Auth Token

            try
            {
                // Get organization OTEL config by sandbox auth token
                OtelConfig result = apiInstance.GetOrganizationOtelConfigBySandboxAuthToken(authToken);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.GetOrganizationOtelConfigBySandboxAuthToken: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetOrganizationOtelConfigBySandboxAuthTokenWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get organization OTEL config by sandbox auth token
    ApiResponse<OtelConfig> response = apiInstance.GetOrganizationOtelConfigBySandboxAuthTokenWithHttpInfo(authToken);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.GetOrganizationOtelConfigBySandboxAuthTokenWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **authToken** | **string** | Sandbox Auth Token |  |

### Return type

[**OtelConfig**](OtelConfig.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OTEL Config |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getorganizationusageoverview"></a>
# **GetOrganizationUsageOverview**
> OrganizationUsageOverview GetOrganizationUsageOverview (string organizationId)

Get organization current usage overview

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
    public class GetOrganizationUsageOverviewExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var organizationId = "organizationId_example";  // string | Organization ID

            try
            {
                // Get organization current usage overview
                OrganizationUsageOverview result = apiInstance.GetOrganizationUsageOverview(organizationId);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.GetOrganizationUsageOverview: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetOrganizationUsageOverviewWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get organization current usage overview
    ApiResponse<OrganizationUsageOverview> response = apiInstance.GetOrganizationUsageOverviewWithHttpInfo(organizationId);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.GetOrganizationUsageOverviewWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **organizationId** | **string** | Organization ID |  |

### Return type

[**OrganizationUsageOverview**](OrganizationUsageOverview.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Current usage overview |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getregionbyid"></a>
# **GetRegionById**
> Region GetRegionById (string id, string? xDaytonaOrganizationID = null)

Get region by ID

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
    public class GetRegionByIdExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var id = "id_example";  // string | Region ID
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Get region by ID
                Region result = apiInstance.GetRegionById(id, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.GetRegionById: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetRegionByIdWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get region by ID
    ApiResponse<Region> response = apiInstance.GetRegionByIdWithHttpInfo(id, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.GetRegionByIdWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **id** | **string** | Region ID |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**Region**](Region.md)

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

<a id="getregionquotabysandboxid"></a>
# **GetRegionQuotaBySandboxId**
> RegionQuota GetRegionQuotaBySandboxId (string sandboxId)

Get region quota by sandbox ID

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
    public class GetRegionQuotaBySandboxIdExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var sandboxId = "sandboxId_example";  // string | Sandbox ID

            try
            {
                // Get region quota by sandbox ID
                RegionQuota result = apiInstance.GetRegionQuotaBySandboxId(sandboxId);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.GetRegionQuotaBySandboxId: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetRegionQuotaBySandboxIdWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get region quota by sandbox ID
    ApiResponse<RegionQuota> response = apiInstance.GetRegionQuotaBySandboxIdWithHttpInfo(sandboxId);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.GetRegionQuotaBySandboxIdWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **sandboxId** | **string** | Sandbox ID |  |

### Return type

[**RegionQuota**](RegionQuota.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Region quota |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="leaveorganization"></a>
# **LeaveOrganization**
> void LeaveOrganization (string organizationId)

Leave organization

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
    public class LeaveOrganizationExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var organizationId = "organizationId_example";  // string | Organization ID

            try
            {
                // Leave organization
                apiInstance.LeaveOrganization(organizationId);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.LeaveOrganization: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the LeaveOrganizationWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Leave organization
    apiInstance.LeaveOrganizationWithHttpInfo(organizationId);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.LeaveOrganizationWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **organizationId** | **string** | Organization ID |  |

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
| **204** | Organization left successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="listavailableregions"></a>
# **ListAvailableRegions**
> List&lt;Region&gt; ListAvailableRegions (string? xDaytonaOrganizationID = null)

List all available regions for the organization

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
    public class ListAvailableRegionsExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // List all available regions for the organization
                List<Region> result = apiInstance.ListAvailableRegions(xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.ListAvailableRegions: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ListAvailableRegionsWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // List all available regions for the organization
    ApiResponse<List<Region>> response = apiInstance.ListAvailableRegionsWithHttpInfo(xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.ListAvailableRegionsWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**List&lt;Region&gt;**](Region.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | List of all available regions |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="listorganizationinvitations"></a>
# **ListOrganizationInvitations**
> List&lt;OrganizationInvitation&gt; ListOrganizationInvitations (string organizationId)

List pending organization invitations

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
    public class ListOrganizationInvitationsExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var organizationId = "organizationId_example";  // string | Organization ID

            try
            {
                // List pending organization invitations
                List<OrganizationInvitation> result = apiInstance.ListOrganizationInvitations(organizationId);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.ListOrganizationInvitations: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ListOrganizationInvitationsWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // List pending organization invitations
    ApiResponse<List<OrganizationInvitation>> response = apiInstance.ListOrganizationInvitationsWithHttpInfo(organizationId);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.ListOrganizationInvitationsWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **organizationId** | **string** | Organization ID |  |

### Return type

[**List&lt;OrganizationInvitation&gt;**](OrganizationInvitation.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | List of pending organization invitations |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="listorganizationinvitationsforauthenticateduser"></a>
# **ListOrganizationInvitationsForAuthenticatedUser**
> List&lt;OrganizationInvitation&gt; ListOrganizationInvitationsForAuthenticatedUser ()

List organization invitations for authenticated user

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
    public class ListOrganizationInvitationsForAuthenticatedUserExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);

            try
            {
                // List organization invitations for authenticated user
                List<OrganizationInvitation> result = apiInstance.ListOrganizationInvitationsForAuthenticatedUser();
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.ListOrganizationInvitationsForAuthenticatedUser: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ListOrganizationInvitationsForAuthenticatedUserWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // List organization invitations for authenticated user
    ApiResponse<List<OrganizationInvitation>> response = apiInstance.ListOrganizationInvitationsForAuthenticatedUserWithHttpInfo();
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.ListOrganizationInvitationsForAuthenticatedUserWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters
This endpoint does not need any parameter.
### Return type

[**List&lt;OrganizationInvitation&gt;**](OrganizationInvitation.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | List of organization invitations |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="listorganizationmembers"></a>
# **ListOrganizationMembers**
> List&lt;OrganizationUser&gt; ListOrganizationMembers (string organizationId)

List organization members

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
    public class ListOrganizationMembersExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var organizationId = "organizationId_example";  // string | Organization ID

            try
            {
                // List organization members
                List<OrganizationUser> result = apiInstance.ListOrganizationMembers(organizationId);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.ListOrganizationMembers: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ListOrganizationMembersWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // List organization members
    ApiResponse<List<OrganizationUser>> response = apiInstance.ListOrganizationMembersWithHttpInfo(organizationId);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.ListOrganizationMembersWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **organizationId** | **string** | Organization ID |  |

### Return type

[**List&lt;OrganizationUser&gt;**](OrganizationUser.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | List of organization members |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="listorganizationroles"></a>
# **ListOrganizationRoles**
> List&lt;OrganizationRole&gt; ListOrganizationRoles (string organizationId)

List organization roles

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
    public class ListOrganizationRolesExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var organizationId = "organizationId_example";  // string | Organization ID

            try
            {
                // List organization roles
                List<OrganizationRole> result = apiInstance.ListOrganizationRoles(organizationId);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.ListOrganizationRoles: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ListOrganizationRolesWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // List organization roles
    ApiResponse<List<OrganizationRole>> response = apiInstance.ListOrganizationRolesWithHttpInfo(organizationId);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.ListOrganizationRolesWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **organizationId** | **string** | Organization ID |  |

### Return type

[**List&lt;OrganizationRole&gt;**](OrganizationRole.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | List of organization roles |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="listorganizations"></a>
# **ListOrganizations**
> List&lt;Organization&gt; ListOrganizations ()

List organizations

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
    public class ListOrganizationsExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);

            try
            {
                // List organizations
                List<Organization> result = apiInstance.ListOrganizations();
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.ListOrganizations: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ListOrganizationsWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // List organizations
    ApiResponse<List<Organization>> response = apiInstance.ListOrganizationsWithHttpInfo();
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.ListOrganizationsWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters
This endpoint does not need any parameter.
### Return type

[**List&lt;Organization&gt;**](Organization.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | List of organizations |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="regenerateproxyapikey"></a>
# **RegenerateProxyApiKey**
> RegenerateApiKeyResponse RegenerateProxyApiKey (string id, string? xDaytonaOrganizationID = null)

Regenerate proxy API key for a region

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
    public class RegenerateProxyApiKeyExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var id = "id_example";  // string | Region ID
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Regenerate proxy API key for a region
                RegenerateApiKeyResponse result = apiInstance.RegenerateProxyApiKey(id, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.RegenerateProxyApiKey: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the RegenerateProxyApiKeyWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Regenerate proxy API key for a region
    ApiResponse<RegenerateApiKeyResponse> response = apiInstance.RegenerateProxyApiKeyWithHttpInfo(id, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.RegenerateProxyApiKeyWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **id** | **string** | Region ID |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**RegenerateApiKeyResponse**](RegenerateApiKeyResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The proxy API key has been successfully regenerated. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="regeneratesnapshotmanagercredentials"></a>
# **RegenerateSnapshotManagerCredentials**
> SnapshotManagerCredentials RegenerateSnapshotManagerCredentials (string id, string? xDaytonaOrganizationID = null)

Regenerate snapshot manager credentials for a region

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
    public class RegenerateSnapshotManagerCredentialsExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var id = "id_example";  // string | Region ID
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Regenerate snapshot manager credentials for a region
                SnapshotManagerCredentials result = apiInstance.RegenerateSnapshotManagerCredentials(id, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.RegenerateSnapshotManagerCredentials: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the RegenerateSnapshotManagerCredentialsWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Regenerate snapshot manager credentials for a region
    ApiResponse<SnapshotManagerCredentials> response = apiInstance.RegenerateSnapshotManagerCredentialsWithHttpInfo(id, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.RegenerateSnapshotManagerCredentialsWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **id** | **string** | Region ID |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**SnapshotManagerCredentials**](SnapshotManagerCredentials.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The snapshot manager credentials have been successfully regenerated. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="regeneratesshgatewayapikey"></a>
# **RegenerateSshGatewayApiKey**
> RegenerateApiKeyResponse RegenerateSshGatewayApiKey (string id, string? xDaytonaOrganizationID = null)

Regenerate SSH gateway API key for a region

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
    public class RegenerateSshGatewayApiKeyExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var id = "id_example";  // string | Region ID
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Regenerate SSH gateway API key for a region
                RegenerateApiKeyResponse result = apiInstance.RegenerateSshGatewayApiKey(id, xDaytonaOrganizationID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.RegenerateSshGatewayApiKey: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the RegenerateSshGatewayApiKeyWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Regenerate SSH gateway API key for a region
    ApiResponse<RegenerateApiKeyResponse> response = apiInstance.RegenerateSshGatewayApiKeyWithHttpInfo(id, xDaytonaOrganizationID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.RegenerateSshGatewayApiKeyWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **id** | **string** | Region ID |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

[**RegenerateApiKeyResponse**](RegenerateApiKeyResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The SSH gateway API key has been successfully regenerated. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="setorganizationdefaultregion"></a>
# **SetOrganizationDefaultRegion**
> void SetOrganizationDefaultRegion (string organizationId, UpdateOrganizationDefaultRegion updateOrganizationDefaultRegion)

Set default region for organization

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
    public class SetOrganizationDefaultRegionExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var organizationId = "organizationId_example";  // string | Organization ID
            var updateOrganizationDefaultRegion = new UpdateOrganizationDefaultRegion(); // UpdateOrganizationDefaultRegion | 

            try
            {
                // Set default region for organization
                apiInstance.SetOrganizationDefaultRegion(organizationId, updateOrganizationDefaultRegion);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.SetOrganizationDefaultRegion: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the SetOrganizationDefaultRegionWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Set default region for organization
    apiInstance.SetOrganizationDefaultRegionWithHttpInfo(organizationId, updateOrganizationDefaultRegion);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.SetOrganizationDefaultRegionWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **organizationId** | **string** | Organization ID |  |
| **updateOrganizationDefaultRegion** | [**UpdateOrganizationDefaultRegion**](UpdateOrganizationDefaultRegion.md) |  |  |

### Return type

void (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | Default region set successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="suspendorganization"></a>
# **SuspendOrganization**
> void SuspendOrganization (string organizationId, OrganizationSuspension? organizationSuspension = null)

Suspend organization

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
    public class SuspendOrganizationExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var organizationId = "organizationId_example";  // string | Organization ID
            var organizationSuspension = new OrganizationSuspension?(); // OrganizationSuspension? |  (optional) 

            try
            {
                // Suspend organization
                apiInstance.SuspendOrganization(organizationId, organizationSuspension);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.SuspendOrganization: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the SuspendOrganizationWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Suspend organization
    apiInstance.SuspendOrganizationWithHttpInfo(organizationId, organizationSuspension);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.SuspendOrganizationWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **organizationId** | **string** | Organization ID |  |
| **organizationSuspension** | [**OrganizationSuspension?**](OrganizationSuspension?.md) |  | [optional]  |

### Return type

void (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | Organization suspended successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="unsuspendorganization"></a>
# **UnsuspendOrganization**
> void UnsuspendOrganization (string organizationId)

Unsuspend organization

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
    public class UnsuspendOrganizationExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var organizationId = "organizationId_example";  // string | Organization ID

            try
            {
                // Unsuspend organization
                apiInstance.UnsuspendOrganization(organizationId);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.UnsuspendOrganization: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UnsuspendOrganizationWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Unsuspend organization
    apiInstance.UnsuspendOrganizationWithHttpInfo(organizationId);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.UnsuspendOrganizationWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **organizationId** | **string** | Organization ID |  |

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
| **204** | Organization unsuspended successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="updateaccessfororganizationmember"></a>
# **UpdateAccessForOrganizationMember**
> OrganizationUser UpdateAccessForOrganizationMember (string organizationId, string userId, UpdateOrganizationMemberAccess updateOrganizationMemberAccess)

Update access for organization member

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
    public class UpdateAccessForOrganizationMemberExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var organizationId = "organizationId_example";  // string | Organization ID
            var userId = "userId_example";  // string | User ID
            var updateOrganizationMemberAccess = new UpdateOrganizationMemberAccess(); // UpdateOrganizationMemberAccess | 

            try
            {
                // Update access for organization member
                OrganizationUser result = apiInstance.UpdateAccessForOrganizationMember(organizationId, userId, updateOrganizationMemberAccess);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.UpdateAccessForOrganizationMember: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UpdateAccessForOrganizationMemberWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Update access for organization member
    ApiResponse<OrganizationUser> response = apiInstance.UpdateAccessForOrganizationMemberWithHttpInfo(organizationId, userId, updateOrganizationMemberAccess);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.UpdateAccessForOrganizationMemberWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **organizationId** | **string** | Organization ID |  |
| **userId** | **string** | User ID |  |
| **updateOrganizationMemberAccess** | [**UpdateOrganizationMemberAccess**](UpdateOrganizationMemberAccess.md) |  |  |

### Return type

[**OrganizationUser**](OrganizationUser.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Access updated successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="updateexperimentalconfig"></a>
# **UpdateExperimentalConfig**
> void UpdateExperimentalConfig (string organizationId, Dictionary<string, Object>? requestBody = null)

Update experimental configuration

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
    public class UpdateExperimentalConfigExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var organizationId = "organizationId_example";  // string | Organization ID
            var requestBody = new Dictionary<string, Object>?(); // Dictionary<string, Object>? | Experimental configuration as a JSON object. Set to null to clear the configuration. (optional) 

            try
            {
                // Update experimental configuration
                apiInstance.UpdateExperimentalConfig(organizationId, requestBody);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.UpdateExperimentalConfig: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UpdateExperimentalConfigWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Update experimental configuration
    apiInstance.UpdateExperimentalConfigWithHttpInfo(organizationId, requestBody);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.UpdateExperimentalConfigWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **organizationId** | **string** | Organization ID |  |
| **requestBody** | [**Dictionary&lt;string, Object&gt;?**](Object.md) | Experimental configuration as a JSON object. Set to null to clear the configuration. | [optional]  |

### Return type

void (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** |  |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="updateorganizationinvitation"></a>
# **UpdateOrganizationInvitation**
> OrganizationInvitation UpdateOrganizationInvitation (string organizationId, string invitationId, UpdateOrganizationInvitation updateOrganizationInvitation)

Update organization invitation

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
    public class UpdateOrganizationInvitationExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var organizationId = "organizationId_example";  // string | Organization ID
            var invitationId = "invitationId_example";  // string | Invitation ID
            var updateOrganizationInvitation = new UpdateOrganizationInvitation(); // UpdateOrganizationInvitation | 

            try
            {
                // Update organization invitation
                OrganizationInvitation result = apiInstance.UpdateOrganizationInvitation(organizationId, invitationId, updateOrganizationInvitation);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.UpdateOrganizationInvitation: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UpdateOrganizationInvitationWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Update organization invitation
    ApiResponse<OrganizationInvitation> response = apiInstance.UpdateOrganizationInvitationWithHttpInfo(organizationId, invitationId, updateOrganizationInvitation);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.UpdateOrganizationInvitationWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **organizationId** | **string** | Organization ID |  |
| **invitationId** | **string** | Invitation ID |  |
| **updateOrganizationInvitation** | [**UpdateOrganizationInvitation**](UpdateOrganizationInvitation.md) |  |  |

### Return type

[**OrganizationInvitation**](OrganizationInvitation.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Organization invitation updated successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="updateorganizationquota"></a>
# **UpdateOrganizationQuota**
> void UpdateOrganizationQuota (string organizationId, UpdateOrganizationQuota updateOrganizationQuota)

Update organization quota

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
    public class UpdateOrganizationQuotaExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var organizationId = "organizationId_example";  // string | Organization ID
            var updateOrganizationQuota = new UpdateOrganizationQuota(); // UpdateOrganizationQuota | 

            try
            {
                // Update organization quota
                apiInstance.UpdateOrganizationQuota(organizationId, updateOrganizationQuota);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.UpdateOrganizationQuota: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UpdateOrganizationQuotaWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Update organization quota
    apiInstance.UpdateOrganizationQuotaWithHttpInfo(organizationId, updateOrganizationQuota);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.UpdateOrganizationQuotaWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **organizationId** | **string** | Organization ID |  |
| **updateOrganizationQuota** | [**UpdateOrganizationQuota**](UpdateOrganizationQuota.md) |  |  |

### Return type

void (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | Organization quota updated successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="updateorganizationregionquota"></a>
# **UpdateOrganizationRegionQuota**
> void UpdateOrganizationRegionQuota (string organizationId, string regionId, UpdateOrganizationRegionQuota updateOrganizationRegionQuota)

Update organization region quota

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
    public class UpdateOrganizationRegionQuotaExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var organizationId = "organizationId_example";  // string | Organization ID
            var regionId = "regionId_example";  // string | ID of the region where the updated quota will be applied
            var updateOrganizationRegionQuota = new UpdateOrganizationRegionQuota(); // UpdateOrganizationRegionQuota | 

            try
            {
                // Update organization region quota
                apiInstance.UpdateOrganizationRegionQuota(organizationId, regionId, updateOrganizationRegionQuota);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.UpdateOrganizationRegionQuota: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UpdateOrganizationRegionQuotaWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Update organization region quota
    apiInstance.UpdateOrganizationRegionQuotaWithHttpInfo(organizationId, regionId, updateOrganizationRegionQuota);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.UpdateOrganizationRegionQuotaWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **organizationId** | **string** | Organization ID |  |
| **regionId** | **string** | ID of the region where the updated quota will be applied |  |
| **updateOrganizationRegionQuota** | [**UpdateOrganizationRegionQuota**](UpdateOrganizationRegionQuota.md) |  |  |

### Return type

void (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | Region quota updated successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="updateorganizationrole"></a>
# **UpdateOrganizationRole**
> OrganizationRole UpdateOrganizationRole (string organizationId, string roleId, UpdateOrganizationRole updateOrganizationRole)

Update organization role

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
    public class UpdateOrganizationRoleExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var organizationId = "organizationId_example";  // string | Organization ID
            var roleId = "roleId_example";  // string | Role ID
            var updateOrganizationRole = new UpdateOrganizationRole(); // UpdateOrganizationRole | 

            try
            {
                // Update organization role
                OrganizationRole result = apiInstance.UpdateOrganizationRole(organizationId, roleId, updateOrganizationRole);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.UpdateOrganizationRole: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UpdateOrganizationRoleWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Update organization role
    ApiResponse<OrganizationRole> response = apiInstance.UpdateOrganizationRoleWithHttpInfo(organizationId, roleId, updateOrganizationRole);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.UpdateOrganizationRoleWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **organizationId** | **string** | Organization ID |  |
| **roleId** | **string** | Role ID |  |
| **updateOrganizationRole** | [**UpdateOrganizationRole**](UpdateOrganizationRole.md) |  |  |

### Return type

[**OrganizationRole**](OrganizationRole.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Role updated successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="updateregion"></a>
# **UpdateRegion**
> void UpdateRegion (string id, UpdateRegion updateRegion, string? xDaytonaOrganizationID = null)

Update region configuration

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
    public class UpdateRegionExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var id = "id_example";  // string | Region ID
            var updateRegion = new UpdateRegion(); // UpdateRegion | 
            var xDaytonaOrganizationID = "xDaytonaOrganizationID_example";  // string? | Use with JWT to specify the organization ID (optional) 

            try
            {
                // Update region configuration
                apiInstance.UpdateRegion(id, updateRegion, xDaytonaOrganizationID);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.UpdateRegion: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UpdateRegionWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Update region configuration
    apiInstance.UpdateRegionWithHttpInfo(id, updateRegion, xDaytonaOrganizationID);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.UpdateRegionWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **id** | **string** | Region ID |  |
| **updateRegion** | [**UpdateRegion**](UpdateRegion.md) |  |  |
| **xDaytonaOrganizationID** | **string?** | Use with JWT to specify the organization ID | [optional]  |

### Return type

void (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** |  |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="updatesandboxdefaultlimitednetworkegress"></a>
# **UpdateSandboxDefaultLimitedNetworkEgress**
> void UpdateSandboxDefaultLimitedNetworkEgress (string organizationId, OrganizationSandboxDefaultLimitedNetworkEgress organizationSandboxDefaultLimitedNetworkEgress)

Update sandbox default limited network egress

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
    public class UpdateSandboxDefaultLimitedNetworkEgressExample
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
            var apiInstance = new OrganizationsApi(httpClient, config, httpClientHandler);
            var organizationId = "organizationId_example";  // string | Organization ID
            var organizationSandboxDefaultLimitedNetworkEgress = new OrganizationSandboxDefaultLimitedNetworkEgress(); // OrganizationSandboxDefaultLimitedNetworkEgress | 

            try
            {
                // Update sandbox default limited network egress
                apiInstance.UpdateSandboxDefaultLimitedNetworkEgress(organizationId, organizationSandboxDefaultLimitedNetworkEgress);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling OrganizationsApi.UpdateSandboxDefaultLimitedNetworkEgress: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UpdateSandboxDefaultLimitedNetworkEgressWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Update sandbox default limited network egress
    apiInstance.UpdateSandboxDefaultLimitedNetworkEgressWithHttpInfo(organizationId, organizationSandboxDefaultLimitedNetworkEgress);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling OrganizationsApi.UpdateSandboxDefaultLimitedNetworkEgressWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **organizationId** | **string** | Organization ID |  |
| **organizationSandboxDefaultLimitedNetworkEgress** | [**OrganizationSandboxDefaultLimitedNetworkEgress**](OrganizationSandboxDefaultLimitedNetworkEgress.md) |  |  |

### Return type

void (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | Sandbox default limited network egress updated successfully |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

