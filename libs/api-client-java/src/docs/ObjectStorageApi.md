# ObjectStorageApi

All URIs are relative to *http://localhost:3000*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**getPushAccess**](ObjectStorageApi.md#getPushAccess) | **GET** /object-storage/push-access | Get temporary storage access for pushing objects |


<a id="getPushAccess"></a>
# **getPushAccess**
> StorageAccessDto getPushAccess(xDaytonaOrganizationID)

Get temporary storage access for pushing objects

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.ObjectStorageApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    ObjectStorageApi apiInstance = new ObjectStorageApi(defaultClient);
    String xDaytonaOrganizationID = "xDaytonaOrganizationID_example"; // String | Use with JWT to specify the organization ID
    try {
      StorageAccessDto result = apiInstance.getPushAccess(xDaytonaOrganizationID);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ObjectStorageApi#getPushAccess");
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

