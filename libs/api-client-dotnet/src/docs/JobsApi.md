# Daytona.ApiClient.Api.JobsApi

All URIs are relative to *http://localhost:3000*

| Method | HTTP request | Description |
|--------|--------------|-------------|
| [**GetJob**](JobsApi.md#getjob) | **GET** /jobs/{jobId} | Get job details |
| [**ListJobs**](JobsApi.md#listjobs) | **GET** /jobs | List jobs for the runner |
| [**PollJobs**](JobsApi.md#polljobs) | **GET** /jobs/poll | Long poll for jobs |
| [**UpdateJobStatus**](JobsApi.md#updatejobstatus) | **POST** /jobs/{jobId}/status | Update job status |

<a id="getjob"></a>
# **GetJob**
> Job GetJob (string jobId)

Get job details

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
    public class GetJobExample
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
            var apiInstance = new JobsApi(httpClient, config, httpClientHandler);
            var jobId = "jobId_example";  // string | ID of the job

            try
            {
                // Get job details
                Job result = apiInstance.GetJob(jobId);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling JobsApi.GetJob: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetJobWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get job details
    ApiResponse<Job> response = apiInstance.GetJobWithHttpInfo(jobId);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling JobsApi.GetJobWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **jobId** | **string** | ID of the job |  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="listjobs"></a>
# **ListJobs**
> PaginatedJobs ListJobs (decimal? page = null, decimal? limit = null, JobStatus? status = null, decimal? offset = null)

List jobs for the runner

Returns a paginated list of jobs for the runner, optionally filtered by status.

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
    public class ListJobsExample
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
            var apiInstance = new JobsApi(httpClient, config, httpClientHandler);
            var page = 1MD;  // decimal? | Page number of the results (optional)  (default to 1M)
            var limit = 100MD;  // decimal? | Maximum number of jobs to return (default: 100, max: 500) (optional)  (default to 100M)
            var status = new JobStatus?(); // JobStatus? | Filter jobs by status (optional) 
            var offset = 8.14D;  // decimal? | Number of jobs to skip for pagination (default: 0) (optional) 

            try
            {
                // List jobs for the runner
                PaginatedJobs result = apiInstance.ListJobs(page, limit, status, offset);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling JobsApi.ListJobs: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ListJobsWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // List jobs for the runner
    ApiResponse<PaginatedJobs> response = apiInstance.ListJobsWithHttpInfo(page, limit, status, offset);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling JobsApi.ListJobsWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **page** | **decimal?** | Page number of the results | [optional] [default to 1M] |
| **limit** | **decimal?** | Maximum number of jobs to return (default: 100, max: 500) | [optional] [default to 100M] |
| **status** | [**JobStatus?**](JobStatus?.md) | Filter jobs by status | [optional]  |
| **offset** | **decimal?** | Number of jobs to skip for pagination (default: 0) | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="polljobs"></a>
# **PollJobs**
> PollJobsResponse PollJobs (decimal? timeout = null, decimal? limit = null)

Long poll for jobs

Long poll endpoint for runners to fetch pending jobs. Returns immediately if jobs are available, otherwise waits up to timeout seconds.

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
    public class PollJobsExample
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
            var apiInstance = new JobsApi(httpClient, config, httpClientHandler);
            var timeout = 8.14D;  // decimal? | Timeout in seconds for long polling (default: 30, max: 60) (optional) 
            var limit = 8.14D;  // decimal? | Maximum number of jobs to return (default: 10, max: 100) (optional) 

            try
            {
                // Long poll for jobs
                PollJobsResponse result = apiInstance.PollJobs(timeout, limit);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling JobsApi.PollJobs: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the PollJobsWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Long poll for jobs
    ApiResponse<PollJobsResponse> response = apiInstance.PollJobsWithHttpInfo(timeout, limit);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling JobsApi.PollJobsWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **timeout** | **decimal?** | Timeout in seconds for long polling (default: 30, max: 60) | [optional]  |
| **limit** | **decimal?** | Maximum number of jobs to return (default: 10, max: 100) | [optional]  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="updatejobstatus"></a>
# **UpdateJobStatus**
> Job UpdateJobStatus (string jobId, UpdateJobStatus updateJobStatus)

Update job status

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
    public class UpdateJobStatusExample
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
            var apiInstance = new JobsApi(httpClient, config, httpClientHandler);
            var jobId = "jobId_example";  // string | ID of the job
            var updateJobStatus = new UpdateJobStatus(); // UpdateJobStatus | 

            try
            {
                // Update job status
                Job result = apiInstance.UpdateJobStatus(jobId, updateJobStatus);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling JobsApi.UpdateJobStatus: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UpdateJobStatusWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Update job status
    ApiResponse<Job> response = apiInstance.UpdateJobStatusWithHttpInfo(jobId, updateJobStatus);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling JobsApi.UpdateJobStatusWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **jobId** | **string** | ID of the job |  |
| **updateJobStatus** | [**UpdateJobStatus**](UpdateJobStatus.md) |  |  |

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

