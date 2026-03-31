# OrganizationsApi

All URIs are relative to *http://localhost:3000*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**acceptOrganizationInvitation**](OrganizationsApi.md#acceptOrganizationInvitation) | **POST** /organizations/invitations/{invitationId}/accept | Accept organization invitation |
| [**cancelOrganizationInvitation**](OrganizationsApi.md#cancelOrganizationInvitation) | **POST** /organizations/{organizationId}/invitations/{invitationId}/cancel | Cancel organization invitation |
| [**createOrganization**](OrganizationsApi.md#createOrganization) | **POST** /organizations | Create organization |
| [**createOrganizationInvitation**](OrganizationsApi.md#createOrganizationInvitation) | **POST** /organizations/{organizationId}/invitations | Create organization invitation |
| [**createOrganizationRole**](OrganizationsApi.md#createOrganizationRole) | **POST** /organizations/{organizationId}/roles | Create organization role |
| [**createRegion**](OrganizationsApi.md#createRegion) | **POST** /regions | Create a new region |
| [**declineOrganizationInvitation**](OrganizationsApi.md#declineOrganizationInvitation) | **POST** /organizations/invitations/{invitationId}/decline | Decline organization invitation |
| [**deleteOrganization**](OrganizationsApi.md#deleteOrganization) | **DELETE** /organizations/{organizationId} | Delete organization |
| [**deleteOrganizationMember**](OrganizationsApi.md#deleteOrganizationMember) | **DELETE** /organizations/{organizationId}/users/{userId} | Delete organization member |
| [**deleteOrganizationRole**](OrganizationsApi.md#deleteOrganizationRole) | **DELETE** /organizations/{organizationId}/roles/{roleId} | Delete organization role |
| [**deleteRegion**](OrganizationsApi.md#deleteRegion) | **DELETE** /regions/{id} | Delete a region |
| [**getOrganization**](OrganizationsApi.md#getOrganization) | **GET** /organizations/{organizationId} | Get organization by ID |
| [**getOrganizationBySandboxId**](OrganizationsApi.md#getOrganizationBySandboxId) | **GET** /organizations/by-sandbox-id/{sandboxId} | Get organization by sandbox ID |
| [**getOrganizationInvitationsCountForAuthenticatedUser**](OrganizationsApi.md#getOrganizationInvitationsCountForAuthenticatedUser) | **GET** /organizations/invitations/count | Get count of organization invitations for authenticated user |
| [**getOrganizationOtelConfigBySandboxAuthToken**](OrganizationsApi.md#getOrganizationOtelConfigBySandboxAuthToken) | **GET** /organizations/otel-config/by-sandbox-auth-token/{authToken} | Get organization OTEL config by sandbox auth token |
| [**getOrganizationUsageOverview**](OrganizationsApi.md#getOrganizationUsageOverview) | **GET** /organizations/{organizationId}/usage | Get organization current usage overview |
| [**getRegionById**](OrganizationsApi.md#getRegionById) | **GET** /regions/{id} | Get region by ID |
| [**getRegionQuotaBySandboxId**](OrganizationsApi.md#getRegionQuotaBySandboxId) | **GET** /organizations/region-quota/by-sandbox-id/{sandboxId} | Get region quota by sandbox ID |
| [**leaveOrganization**](OrganizationsApi.md#leaveOrganization) | **POST** /organizations/{organizationId}/leave | Leave organization |
| [**listAvailableRegions**](OrganizationsApi.md#listAvailableRegions) | **GET** /regions | List all available regions for the organization |
| [**listOrganizationInvitations**](OrganizationsApi.md#listOrganizationInvitations) | **GET** /organizations/{organizationId}/invitations | List pending organization invitations |
| [**listOrganizationInvitationsForAuthenticatedUser**](OrganizationsApi.md#listOrganizationInvitationsForAuthenticatedUser) | **GET** /organizations/invitations | List organization invitations for authenticated user |
| [**listOrganizationMembers**](OrganizationsApi.md#listOrganizationMembers) | **GET** /organizations/{organizationId}/users | List organization members |
| [**listOrganizationRoles**](OrganizationsApi.md#listOrganizationRoles) | **GET** /organizations/{organizationId}/roles | List organization roles |
| [**listOrganizations**](OrganizationsApi.md#listOrganizations) | **GET** /organizations | List organizations |
| [**regenerateProxyApiKey**](OrganizationsApi.md#regenerateProxyApiKey) | **POST** /regions/{id}/regenerate-proxy-api-key | Regenerate proxy API key for a region |
| [**regenerateSnapshotManagerCredentials**](OrganizationsApi.md#regenerateSnapshotManagerCredentials) | **POST** /regions/{id}/regenerate-snapshot-manager-credentials | Regenerate snapshot manager credentials for a region |
| [**regenerateSshGatewayApiKey**](OrganizationsApi.md#regenerateSshGatewayApiKey) | **POST** /regions/{id}/regenerate-ssh-gateway-api-key | Regenerate SSH gateway API key for a region |
| [**setOrganizationDefaultRegion**](OrganizationsApi.md#setOrganizationDefaultRegion) | **PATCH** /organizations/{organizationId}/default-region | Set default region for organization |
| [**suspendOrganization**](OrganizationsApi.md#suspendOrganization) | **POST** /organizations/{organizationId}/suspend | Suspend organization |
| [**unsuspendOrganization**](OrganizationsApi.md#unsuspendOrganization) | **POST** /organizations/{organizationId}/unsuspend | Unsuspend organization |
| [**updateAccessForOrganizationMember**](OrganizationsApi.md#updateAccessForOrganizationMember) | **POST** /organizations/{organizationId}/users/{userId}/access | Update access for organization member |
| [**updateExperimentalConfig**](OrganizationsApi.md#updateExperimentalConfig) | **PUT** /organizations/{organizationId}/experimental-config | Update experimental configuration |
| [**updateOrganizationInvitation**](OrganizationsApi.md#updateOrganizationInvitation) | **PUT** /organizations/{organizationId}/invitations/{invitationId} | Update organization invitation |
| [**updateOrganizationQuota**](OrganizationsApi.md#updateOrganizationQuota) | **PATCH** /organizations/{organizationId}/quota | Update organization quota |
| [**updateOrganizationRegionQuota**](OrganizationsApi.md#updateOrganizationRegionQuota) | **PATCH** /organizations/{organizationId}/quota/{regionId} | Update organization region quota |
| [**updateOrganizationRole**](OrganizationsApi.md#updateOrganizationRole) | **PUT** /organizations/{organizationId}/roles/{roleId} | Update organization role |
| [**updateRegion**](OrganizationsApi.md#updateRegion) | **PATCH** /regions/{id} | Update region configuration |
| [**updateSandboxDefaultLimitedNetworkEgress**](OrganizationsApi.md#updateSandboxDefaultLimitedNetworkEgress) | **POST** /organizations/{organizationId}/sandbox-default-limited-network-egress | Update sandbox default limited network egress |


<a id="acceptOrganizationInvitation"></a>
# **acceptOrganizationInvitation**
> OrganizationInvitation acceptOrganizationInvitation(invitationId)

Accept organization invitation

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String invitationId = "invitationId_example"; // String | Invitation ID
    try {
      OrganizationInvitation result = apiInstance.acceptOrganizationInvitation(invitationId);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#acceptOrganizationInvitation");
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
| **invitationId** | **String**| Invitation ID | |

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

<a id="cancelOrganizationInvitation"></a>
# **cancelOrganizationInvitation**
> cancelOrganizationInvitation(organizationId, invitationId)

Cancel organization invitation

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String organizationId = "organizationId_example"; // String | Organization ID
    String invitationId = "invitationId_example"; // String | Invitation ID
    try {
      apiInstance.cancelOrganizationInvitation(organizationId, invitationId);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#cancelOrganizationInvitation");
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
| **organizationId** | **String**| Organization ID | |
| **invitationId** | **String**| Invitation ID | |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | Organization invitation cancelled successfully |  -  |

<a id="createOrganization"></a>
# **createOrganization**
> Organization createOrganization(createOrganization)

Create organization

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    CreateOrganization createOrganization = new CreateOrganization(); // CreateOrganization | 
    try {
      Organization result = apiInstance.createOrganization(createOrganization);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#createOrganization");
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
| **createOrganization** | [**CreateOrganization**](CreateOrganization.md)|  | |

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

<a id="createOrganizationInvitation"></a>
# **createOrganizationInvitation**
> OrganizationInvitation createOrganizationInvitation(organizationId, createOrganizationInvitation)

Create organization invitation

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String organizationId = "organizationId_example"; // String | Organization ID
    CreateOrganizationInvitation createOrganizationInvitation = new CreateOrganizationInvitation(); // CreateOrganizationInvitation | 
    try {
      OrganizationInvitation result = apiInstance.createOrganizationInvitation(organizationId, createOrganizationInvitation);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#createOrganizationInvitation");
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
| **organizationId** | **String**| Organization ID | |
| **createOrganizationInvitation** | [**CreateOrganizationInvitation**](CreateOrganizationInvitation.md)|  | |

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

<a id="createOrganizationRole"></a>
# **createOrganizationRole**
> OrganizationRole createOrganizationRole(organizationId, createOrganizationRole)

Create organization role

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String organizationId = "organizationId_example"; // String | Organization ID
    CreateOrganizationRole createOrganizationRole = new CreateOrganizationRole(); // CreateOrganizationRole | 
    try {
      OrganizationRole result = apiInstance.createOrganizationRole(organizationId, createOrganizationRole);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#createOrganizationRole");
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
| **organizationId** | **String**| Organization ID | |
| **createOrganizationRole** | [**CreateOrganizationRole**](CreateOrganizationRole.md)|  | |

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

<a id="createRegion"></a>
# **createRegion**
> CreateRegionResponse createRegion(createRegion, xDaytonaOrganizationID)

Create a new region

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    CreateRegion createRegion = new CreateRegion(); // CreateRegion | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      CreateRegionResponse result = apiInstance.createRegion(createRegion, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#createRegion");
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
| **createRegion** | [**CreateRegion**](CreateRegion.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="declineOrganizationInvitation"></a>
# **declineOrganizationInvitation**
> declineOrganizationInvitation(invitationId)

Decline organization invitation

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String invitationId = "invitationId_example"; // String | Invitation ID
    try {
      apiInstance.declineOrganizationInvitation(invitationId);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#declineOrganizationInvitation");
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
| **invitationId** | **String**| Invitation ID | |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Organization invitation declined successfully |  -  |

<a id="deleteOrganization"></a>
# **deleteOrganization**
> deleteOrganization(organizationId)

Delete organization

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String organizationId = "organizationId_example"; // String | Organization ID
    try {
      apiInstance.deleteOrganization(organizationId);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#deleteOrganization");
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
| **organizationId** | **String**| Organization ID | |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | Organization deleted successfully |  -  |

<a id="deleteOrganizationMember"></a>
# **deleteOrganizationMember**
> deleteOrganizationMember(organizationId, userId)

Delete organization member

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String organizationId = "organizationId_example"; // String | Organization ID
    String userId = "userId_example"; // String | User ID
    try {
      apiInstance.deleteOrganizationMember(organizationId, userId);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#deleteOrganizationMember");
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
| **organizationId** | **String**| Organization ID | |
| **userId** | **String**| User ID | |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | User removed from organization successfully |  -  |

<a id="deleteOrganizationRole"></a>
# **deleteOrganizationRole**
> deleteOrganizationRole(organizationId, roleId)

Delete organization role

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String organizationId = "organizationId_example"; // String | Organization ID
    String roleId = "roleId_example"; // String | Role ID
    try {
      apiInstance.deleteOrganizationRole(organizationId, roleId);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#deleteOrganizationRole");
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
| **organizationId** | **String**| Organization ID | |
| **roleId** | **String**| Role ID | |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | Organization role deleted successfully |  -  |

<a id="deleteRegion"></a>
# **deleteRegion**
> deleteRegion(id, xDaytonaOrganizationID)

Delete a region

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String id = "id_example"; // String | Region ID
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.deleteRegion(id, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#deleteRegion");
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
| **id** | **String**| Region ID | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | The region has been successfully deleted. |  -  |

<a id="getOrganization"></a>
# **getOrganization**
> Organization getOrganization(organizationId)

Get organization by ID

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String organizationId = "organizationId_example"; // String | Organization ID
    try {
      Organization result = apiInstance.getOrganization(organizationId);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#getOrganization");
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
| **organizationId** | **String**| Organization ID | |

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

<a id="getOrganizationBySandboxId"></a>
# **getOrganizationBySandboxId**
> Organization getOrganizationBySandboxId(sandboxId)

Get organization by sandbox ID

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | Sandbox ID
    try {
      Organization result = apiInstance.getOrganizationBySandboxId(sandboxId);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#getOrganizationBySandboxId");
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
| **sandboxId** | **String**| Sandbox ID | |

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

<a id="getOrganizationInvitationsCountForAuthenticatedUser"></a>
# **getOrganizationInvitationsCountForAuthenticatedUser**
> BigDecimal getOrganizationInvitationsCountForAuthenticatedUser()

Get count of organization invitations for authenticated user

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    try {
      BigDecimal result = apiInstance.getOrganizationInvitationsCountForAuthenticatedUser();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#getOrganizationInvitationsCountForAuthenticatedUser");
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

[**BigDecimal**](BigDecimal.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Count of organization invitations |  -  |

<a id="getOrganizationOtelConfigBySandboxAuthToken"></a>
# **getOrganizationOtelConfigBySandboxAuthToken**
> OtelConfig getOrganizationOtelConfigBySandboxAuthToken(authToken)

Get organization OTEL config by sandbox auth token

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String authToken = "authToken_example"; // String | Sandbox Auth Token
    try {
      OtelConfig result = apiInstance.getOrganizationOtelConfigBySandboxAuthToken(authToken);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#getOrganizationOtelConfigBySandboxAuthToken");
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
| **authToken** | **String**| Sandbox Auth Token | |

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

<a id="getOrganizationUsageOverview"></a>
# **getOrganizationUsageOverview**
> OrganizationUsageOverview getOrganizationUsageOverview(organizationId)

Get organization current usage overview

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String organizationId = "organizationId_example"; // String | Organization ID
    try {
      OrganizationUsageOverview result = apiInstance.getOrganizationUsageOverview(organizationId);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#getOrganizationUsageOverview");
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
| **organizationId** | **String**| Organization ID | |

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

<a id="getRegionById"></a>
# **getRegionById**
> Region getRegionById(id, xDaytonaOrganizationID)

Get region by ID

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String id = "id_example"; // String | Region ID
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      Region result = apiInstance.getRegionById(id, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#getRegionById");
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
| **id** | **String**| Region ID | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="getRegionQuotaBySandboxId"></a>
# **getRegionQuotaBySandboxId**
> RegionQuota getRegionQuotaBySandboxId(sandboxId)

Get region quota by sandbox ID

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String sandboxId = "sandboxId_example"; // String | Sandbox ID
    try {
      RegionQuota result = apiInstance.getRegionQuotaBySandboxId(sandboxId);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#getRegionQuotaBySandboxId");
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
| **sandboxId** | **String**| Sandbox ID | |

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

<a id="leaveOrganization"></a>
# **leaveOrganization**
> leaveOrganization(organizationId)

Leave organization

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String organizationId = "organizationId_example"; // String | Organization ID
    try {
      apiInstance.leaveOrganization(organizationId);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#leaveOrganization");
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
| **organizationId** | **String**| Organization ID | |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | Organization left successfully |  -  |

<a id="listAvailableRegions"></a>
# **listAvailableRegions**
> List&lt;Region&gt; listAvailableRegions(xDaytonaOrganizationID)

List all available regions for the organization

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      List<Region> result = apiInstance.listAvailableRegions(xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#listAvailableRegions");
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
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="listOrganizationInvitations"></a>
# **listOrganizationInvitations**
> List&lt;OrganizationInvitation&gt; listOrganizationInvitations(organizationId)

List pending organization invitations

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String organizationId = "organizationId_example"; // String | Organization ID
    try {
      List<OrganizationInvitation> result = apiInstance.listOrganizationInvitations(organizationId);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#listOrganizationInvitations");
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
| **organizationId** | **String**| Organization ID | |

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

<a id="listOrganizationInvitationsForAuthenticatedUser"></a>
# **listOrganizationInvitationsForAuthenticatedUser**
> List&lt;OrganizationInvitation&gt; listOrganizationInvitationsForAuthenticatedUser()

List organization invitations for authenticated user

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    try {
      List<OrganizationInvitation> result = apiInstance.listOrganizationInvitationsForAuthenticatedUser();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#listOrganizationInvitationsForAuthenticatedUser");
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

<a id="listOrganizationMembers"></a>
# **listOrganizationMembers**
> List&lt;OrganizationUser&gt; listOrganizationMembers(organizationId)

List organization members

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String organizationId = "organizationId_example"; // String | Organization ID
    try {
      List<OrganizationUser> result = apiInstance.listOrganizationMembers(organizationId);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#listOrganizationMembers");
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
| **organizationId** | **String**| Organization ID | |

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

<a id="listOrganizationRoles"></a>
# **listOrganizationRoles**
> List&lt;OrganizationRole&gt; listOrganizationRoles(organizationId)

List organization roles

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String organizationId = "organizationId_example"; // String | Organization ID
    try {
      List<OrganizationRole> result = apiInstance.listOrganizationRoles(organizationId);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#listOrganizationRoles");
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
| **organizationId** | **String**| Organization ID | |

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

<a id="listOrganizations"></a>
# **listOrganizations**
> List&lt;Organization&gt; listOrganizations()

List organizations

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    try {
      List<Organization> result = apiInstance.listOrganizations();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#listOrganizations");
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

<a id="regenerateProxyApiKey"></a>
# **regenerateProxyApiKey**
> RegenerateApiKeyResponse regenerateProxyApiKey(id, xDaytonaOrganizationID)

Regenerate proxy API key for a region

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String id = "id_example"; // String | Region ID
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      RegenerateApiKeyResponse result = apiInstance.regenerateProxyApiKey(id, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#regenerateProxyApiKey");
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
| **id** | **String**| Region ID | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="regenerateSnapshotManagerCredentials"></a>
# **regenerateSnapshotManagerCredentials**
> SnapshotManagerCredentials regenerateSnapshotManagerCredentials(id, xDaytonaOrganizationID)

Regenerate snapshot manager credentials for a region

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String id = "id_example"; // String | Region ID
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      SnapshotManagerCredentials result = apiInstance.regenerateSnapshotManagerCredentials(id, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#regenerateSnapshotManagerCredentials");
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
| **id** | **String**| Region ID | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="regenerateSshGatewayApiKey"></a>
# **regenerateSshGatewayApiKey**
> RegenerateApiKeyResponse regenerateSshGatewayApiKey(id, xDaytonaOrganizationID)

Regenerate SSH gateway API key for a region

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String id = "id_example"; // String | Region ID
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      RegenerateApiKeyResponse result = apiInstance.regenerateSshGatewayApiKey(id, xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#regenerateSshGatewayApiKey");
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
| **id** | **String**| Region ID | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

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

<a id="setOrganizationDefaultRegion"></a>
# **setOrganizationDefaultRegion**
> setOrganizationDefaultRegion(organizationId, updateOrganizationDefaultRegion)

Set default region for organization

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String organizationId = "organizationId_example"; // String | Organization ID
    UpdateOrganizationDefaultRegion updateOrganizationDefaultRegion = new UpdateOrganizationDefaultRegion(); // UpdateOrganizationDefaultRegion | 
    try {
      apiInstance.setOrganizationDefaultRegion(organizationId, updateOrganizationDefaultRegion);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#setOrganizationDefaultRegion");
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
| **organizationId** | **String**| Organization ID | |
| **updateOrganizationDefaultRegion** | [**UpdateOrganizationDefaultRegion**](UpdateOrganizationDefaultRegion.md)|  | |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | Default region set successfully |  -  |

<a id="suspendOrganization"></a>
# **suspendOrganization**
> suspendOrganization(organizationId, organizationSuspension)

Suspend organization

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String organizationId = "organizationId_example"; // String | Organization ID
    OrganizationSuspension organizationSuspension = new OrganizationSuspension(); // OrganizationSuspension | 
    try {
      apiInstance.suspendOrganization(organizationId, organizationSuspension);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#suspendOrganization");
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
| **organizationId** | **String**| Organization ID | |
| **organizationSuspension** | [**OrganizationSuspension**](OrganizationSuspension.md)|  | [optional] |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | Organization suspended successfully |  -  |

<a id="unsuspendOrganization"></a>
# **unsuspendOrganization**
> unsuspendOrganization(organizationId)

Unsuspend organization

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String organizationId = "organizationId_example"; // String | Organization ID
    try {
      apiInstance.unsuspendOrganization(organizationId);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#unsuspendOrganization");
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
| **organizationId** | **String**| Organization ID | |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | Organization unsuspended successfully |  -  |

<a id="updateAccessForOrganizationMember"></a>
# **updateAccessForOrganizationMember**
> OrganizationUser updateAccessForOrganizationMember(organizationId, userId, updateOrganizationMemberAccess)

Update access for organization member

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String organizationId = "organizationId_example"; // String | Organization ID
    String userId = "userId_example"; // String | User ID
    UpdateOrganizationMemberAccess updateOrganizationMemberAccess = new UpdateOrganizationMemberAccess(); // UpdateOrganizationMemberAccess | 
    try {
      OrganizationUser result = apiInstance.updateAccessForOrganizationMember(organizationId, userId, updateOrganizationMemberAccess);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#updateAccessForOrganizationMember");
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
| **organizationId** | **String**| Organization ID | |
| **userId** | **String**| User ID | |
| **updateOrganizationMemberAccess** | [**UpdateOrganizationMemberAccess**](UpdateOrganizationMemberAccess.md)|  | |

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

<a id="updateExperimentalConfig"></a>
# **updateExperimentalConfig**
> updateExperimentalConfig(organizationId, requestBody)

Update experimental configuration

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String organizationId = "organizationId_example"; // String | Organization ID
    Map<String, Object> requestBody = null; // Map<String, Object> | Experimental configuration as a JSON object. Set to null to clear the configuration.
    try {
      apiInstance.updateExperimentalConfig(organizationId, requestBody);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#updateExperimentalConfig");
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
| **organizationId** | **String**| Organization ID | |
| **requestBody** | [**Map&lt;String, Object&gt;**](Object.md)| Experimental configuration as a JSON object. Set to null to clear the configuration. | [optional] |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** |  |  -  |

<a id="updateOrganizationInvitation"></a>
# **updateOrganizationInvitation**
> OrganizationInvitation updateOrganizationInvitation(organizationId, invitationId, updateOrganizationInvitation)

Update organization invitation

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String organizationId = "organizationId_example"; // String | Organization ID
    String invitationId = "invitationId_example"; // String | Invitation ID
    UpdateOrganizationInvitation updateOrganizationInvitation = new UpdateOrganizationInvitation(); // UpdateOrganizationInvitation | 
    try {
      OrganizationInvitation result = apiInstance.updateOrganizationInvitation(organizationId, invitationId, updateOrganizationInvitation);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#updateOrganizationInvitation");
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
| **organizationId** | **String**| Organization ID | |
| **invitationId** | **String**| Invitation ID | |
| **updateOrganizationInvitation** | [**UpdateOrganizationInvitation**](UpdateOrganizationInvitation.md)|  | |

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

<a id="updateOrganizationQuota"></a>
# **updateOrganizationQuota**
> updateOrganizationQuota(organizationId, updateOrganizationQuota)

Update organization quota

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String organizationId = "organizationId_example"; // String | Organization ID
    UpdateOrganizationQuota updateOrganizationQuota = new UpdateOrganizationQuota(); // UpdateOrganizationQuota | 
    try {
      apiInstance.updateOrganizationQuota(organizationId, updateOrganizationQuota);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#updateOrganizationQuota");
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
| **organizationId** | **String**| Organization ID | |
| **updateOrganizationQuota** | [**UpdateOrganizationQuota**](UpdateOrganizationQuota.md)|  | |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | Organization quota updated successfully |  -  |

<a id="updateOrganizationRegionQuota"></a>
# **updateOrganizationRegionQuota**
> updateOrganizationRegionQuota(organizationId, regionId, updateOrganizationRegionQuota)

Update organization region quota

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String organizationId = "organizationId_example"; // String | Organization ID
    String regionId = "regionId_example"; // String | ID of the region where the updated quota will be applied
    UpdateOrganizationRegionQuota updateOrganizationRegionQuota = new UpdateOrganizationRegionQuota(); // UpdateOrganizationRegionQuota | 
    try {
      apiInstance.updateOrganizationRegionQuota(organizationId, regionId, updateOrganizationRegionQuota);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#updateOrganizationRegionQuota");
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
| **organizationId** | **String**| Organization ID | |
| **regionId** | **String**| ID of the region where the updated quota will be applied | |
| **updateOrganizationRegionQuota** | [**UpdateOrganizationRegionQuota**](UpdateOrganizationRegionQuota.md)|  | |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | Region quota updated successfully |  -  |

<a id="updateOrganizationRole"></a>
# **updateOrganizationRole**
> OrganizationRole updateOrganizationRole(organizationId, roleId, updateOrganizationRole)

Update organization role

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String organizationId = "organizationId_example"; // String | Organization ID
    String roleId = "roleId_example"; // String | Role ID
    UpdateOrganizationRole updateOrganizationRole = new UpdateOrganizationRole(); // UpdateOrganizationRole | 
    try {
      OrganizationRole result = apiInstance.updateOrganizationRole(organizationId, roleId, updateOrganizationRole);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#updateOrganizationRole");
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
| **organizationId** | **String**| Organization ID | |
| **roleId** | **String**| Role ID | |
| **updateOrganizationRole** | [**UpdateOrganizationRole**](UpdateOrganizationRole.md)|  | |

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

<a id="updateRegion"></a>
# **updateRegion**
> updateRegion(id, updateRegion, xDaytonaOrganizationID)

Update region configuration

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String id = "id_example"; // String | Region ID
    UpdateRegion updateRegion = new UpdateRegion(); // UpdateRegion | 
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      apiInstance.updateRegion(id, updateRegion, xDaytonaOrganizationID);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#updateRegion");
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
| **id** | **String**| Region ID | |
| **updateRegion** | [**UpdateRegion**](UpdateRegion.md)|  | |
| **xDaytonaOrganizationID** | **String**| Use with JWT to specify the organization ID | [optional] |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** |  |  -  |

<a id="updateSandboxDefaultLimitedNetworkEgress"></a>
# **updateSandboxDefaultLimitedNetworkEgress**
> updateSandboxDefaultLimitedNetworkEgress(organizationId, organizationSandboxDefaultLimitedNetworkEgress)

Update sandbox default limited network egress

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.OrganizationsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    OrganizationsApi apiInstance = new OrganizationsApi(defaultClient);
    String organizationId = "organizationId_example"; // String | Organization ID
    OrganizationSandboxDefaultLimitedNetworkEgress organizationSandboxDefaultLimitedNetworkEgress = new OrganizationSandboxDefaultLimitedNetworkEgress(); // OrganizationSandboxDefaultLimitedNetworkEgress | 
    try {
      apiInstance.updateSandboxDefaultLimitedNetworkEgress(organizationId, organizationSandboxDefaultLimitedNetworkEgress);
    } catch (ApiException e) {
      System.err.println("Exception when calling OrganizationsApi#updateSandboxDefaultLimitedNetworkEgress");
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
| **organizationId** | **String**| Organization ID | |
| **organizationSandboxDefaultLimitedNetworkEgress** | [**OrganizationSandboxDefaultLimitedNetworkEgress**](OrganizationSandboxDefaultLimitedNetworkEgress.md)|  | |

### Return type

null (empty response body)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | Sandbox default limited network egress updated successfully |  -  |

