# AuditApi

All URIs are relative to *http://localhost:3000*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**getAllAuditLogs**](AuditApi.md#getAllAuditLogs) | **GET** /audit | Get all audit logs |
| [**getOrganizationAuditLogs**](AuditApi.md#getOrganizationAuditLogs) | **GET** /audit/organizations/{organizationId} | Get audit logs for organization |


<a id="getAllAuditLogs"></a>
# **getAllAuditLogs**
> PaginatedAuditLogs getAllAuditLogs(page, limit, from, to, nextToken)

Get all audit logs

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.AuditApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    AuditApi apiInstance = new AuditApi(defaultClient);
    BigDecimal page = new BigDecimal("1"); // BigDecimal | Page number of the results
    BigDecimal limit = new BigDecimal("100"); // BigDecimal | Number of results per page
    OffsetDateTime from = OffsetDateTime.now(); // OffsetDateTime | From date (ISO 8601 format)
    OffsetDateTime to = OffsetDateTime.now(); // OffsetDateTime | To date (ISO 8601 format)
    String nextToken = "nextToken_example"; // String | Token for cursor-based pagination. When provided, takes precedence over page parameter.
    try {
      PaginatedAuditLogs result = apiInstance.getAllAuditLogs(page, limit, from, to, nextToken);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling AuditApi#getAllAuditLogs");
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
| **page** | **BigDecimal**| Page number of the results | [optional] [default to 1] |
| **limit** | **BigDecimal**| Number of results per page | [optional] [default to 100] |
| **from** | **OffsetDateTime**| From date (ISO 8601 format) | [optional] |
| **to** | **OffsetDateTime**| To date (ISO 8601 format) | [optional] |
| **nextToken** | **String**| Token for cursor-based pagination. When provided, takes precedence over page parameter. | [optional] |

### Return type

[**PaginatedAuditLogs**](PaginatedAuditLogs.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Paginated list of all audit logs |  -  |

<a id="getOrganizationAuditLogs"></a>
# **getOrganizationAuditLogs**
> PaginatedAuditLogs getOrganizationAuditLogs(organizationId, page, limit, from, to, nextToken)

Get audit logs for organization

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.AuditApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    AuditApi apiInstance = new AuditApi(defaultClient);
    String organizationId = "organizationId_example"; // String | Organization ID
    BigDecimal page = new BigDecimal("1"); // BigDecimal | Page number of the results
    BigDecimal limit = new BigDecimal("100"); // BigDecimal | Number of results per page
    OffsetDateTime from = OffsetDateTime.now(); // OffsetDateTime | From date (ISO 8601 format)
    OffsetDateTime to = OffsetDateTime.now(); // OffsetDateTime | To date (ISO 8601 format)
    String nextToken = "nextToken_example"; // String | Token for cursor-based pagination. When provided, takes precedence over page parameter.
    try {
      PaginatedAuditLogs result = apiInstance.getOrganizationAuditLogs(organizationId, page, limit, from, to, nextToken);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling AuditApi#getOrganizationAuditLogs");
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
| **page** | **BigDecimal**| Page number of the results | [optional] [default to 1] |
| **limit** | **BigDecimal**| Number of results per page | [optional] [default to 100] |
| **from** | **OffsetDateTime**| From date (ISO 8601 format) | [optional] |
| **to** | **OffsetDateTime**| To date (ISO 8601 format) | [optional] |
| **nextToken** | **String**| Token for cursor-based pagination. When provided, takes precedence over page parameter. | [optional] |

### Return type

[**PaginatedAuditLogs**](PaginatedAuditLogs.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Paginated list of organization audit logs |  -  |

