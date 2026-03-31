# JobsApi

All URIs are relative to *http://localhost:3000*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**getJob**](JobsApi.md#getJob) | **GET** /jobs/{jobId} | Get job details |
| [**listJobs**](JobsApi.md#listJobs) | **GET** /jobs | List jobs for the runner |
| [**pollJobs**](JobsApi.md#pollJobs) | **GET** /jobs/poll | Long poll for jobs |
| [**updateJobStatus**](JobsApi.md#updateJobStatus) | **POST** /jobs/{jobId}/status | Update job status |


<a id="getJob"></a>
# **getJob**
> Job getJob(jobId)

Get job details

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.JobsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    JobsApi apiInstance = new JobsApi(defaultClient);
    String jobId = "jobId_example"; // String | ID of the job
    try {
      Job result = apiInstance.getJob(jobId);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling JobsApi#getJob");
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
| **jobId** | **String**| ID of the job | |

### Return type

[**Job**](Job.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Job details |  -  |

<a id="listJobs"></a>
# **listJobs**
> PaginatedJobs listJobs(page, limit, status, offset)

List jobs for the runner

Returns a paginated list of jobs for the runner, optionally filtered by status.

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.JobsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    JobsApi apiInstance = new JobsApi(defaultClient);
    BigDecimal page = new BigDecimal("1"); // BigDecimal | Page number of the results
    BigDecimal limit = new BigDecimal("100"); // BigDecimal | Maximum number of jobs to return (default: 100, max: 500)
    JobStatus status = JobStatus.fromValue("PENDING"); // JobStatus | Filter jobs by status
    BigDecimal offset = new BigDecimal(78); // BigDecimal | Number of jobs to skip for pagination (default: 0)
    try {
      PaginatedJobs result = apiInstance.listJobs(page, limit, status, offset);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling JobsApi#listJobs");
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
| **limit** | **BigDecimal**| Maximum number of jobs to return (default: 100, max: 500) | [optional] [default to 100] |
| **status** | [**JobStatus**](.md)| Filter jobs by status | [optional] [enum: PENDING, IN_PROGRESS, COMPLETED, FAILED] |
| **offset** | **BigDecimal**| Number of jobs to skip for pagination (default: 0) | [optional] |

### Return type

[**PaginatedJobs**](PaginatedJobs.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | List of jobs for the runner |  -  |

<a id="pollJobs"></a>
# **pollJobs**
> PollJobsResponse pollJobs(timeout, limit)

Long poll for jobs

Long poll endpoint for runners to fetch pending jobs. Returns immediately if jobs are available, otherwise waits up to timeout seconds.

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.JobsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    JobsApi apiInstance = new JobsApi(defaultClient);
    BigDecimal timeout = new BigDecimal(78); // BigDecimal | Timeout in seconds for long polling (default: 30, max: 60)
    BigDecimal limit = new BigDecimal(78); // BigDecimal | Maximum number of jobs to return (default: 10, max: 100)
    try {
      PollJobsResponse result = apiInstance.pollJobs(timeout, limit);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling JobsApi#pollJobs");
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
| **timeout** | **BigDecimal**| Timeout in seconds for long polling (default: 30, max: 60) | [optional] |
| **limit** | **BigDecimal**| Maximum number of jobs to return (default: 10, max: 100) | [optional] |

### Return type

[**PollJobsResponse**](PollJobsResponse.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | List of jobs for the runner |  -  |

<a id="updateJobStatus"></a>
# **updateJobStatus**
> Job updateJobStatus(jobId, updateJobStatus)

Update job status

### Example
```java
// Import classes:
import io.daytona.api.client.ApiClient;
import io.daytona.api.client.ApiException;
import io.daytona.api.client.Configuration;
import io.daytona.api.client.auth.*;
import io.daytona.api.client.models.*;
import io.daytona.api.client.api.JobsApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:3000");
    
    // Configure HTTP bearer authorization: bearer
    HttpBearerAuth bearer = (HttpBearerAuth) defaultClient.getAuthentication("bearer");
    bearer.setBearerToken("BEARER TOKEN");


    JobsApi apiInstance = new JobsApi(defaultClient);
    String jobId = "jobId_example"; // String | ID of the job
    UpdateJobStatus updateJobStatus = new UpdateJobStatus(); // UpdateJobStatus | 
    try {
      Job result = apiInstance.updateJobStatus(jobId, updateJobStatus);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling JobsApi#updateJobStatus");
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
| **jobId** | **String**| ID of the job | |
| **updateJobStatus** | [**UpdateJobStatus**](UpdateJobStatus.md)|  | |

### Return type

[**Job**](Job.md)

### Authorization

[bearer](../README.md#bearer), [oauth2](../README.md#oauth2)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Job status updated successfully |  -  |

