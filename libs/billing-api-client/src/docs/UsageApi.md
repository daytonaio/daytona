# UsageApi

All URIs are relative to _http://localhost:6100_

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**getCurrentUsage**](#getcurrentusage) | **GET** /organization/{organizationId}/usage | Get organization usage|
|[**getPastUsage**](#getpastusage) | **GET** /organization/{organizationId}/usage/past | Get organization usage|
|[**getV2CurrentUsage**](#getv2currentusage) | **GET** /v2/organization/{organizationId}/usage | Get organization usage|
|[**getV2PastUsage**](#getv2pastusage) | **GET** /v2/organization/{organizationId}/usage/past | Get organization usage|

# **getCurrentUsage**
>
> OrganizationUsage getCurrentUsage()

Get organization usage

### Example

```typescript
import {
    UsageApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new UsageApi(configuration);

let organizationId: string; //Organization ID (default to undefined)

const { status, data } = await apiInstance.getCurrentUsage(
    organizationId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **organizationId** | [**string**] | Organization ID | defaults to undefined|

### Return type

**OrganizationUsage**

### Authorization

[JwtAuth](../README.md#JwtAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getPastUsage**
>
> OrganizationUsage getPastUsage()

Get organization usage

### Example

```typescript
import {
    UsageApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new UsageApi(configuration);

let organizationId: string; //Organization ID (default to undefined)
let periods: number; //Periods (optional) (default to undefined)

const { status, data } = await apiInstance.getPastUsage(
    organizationId,
    periods
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **organizationId** | [**string**] | Organization ID | defaults to undefined|
| **periods** | [**number**] | Periods | (optional) defaults to undefined|

### Return type

**OrganizationUsage**

### Authorization

[JwtAuth](../README.md#JwtAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getV2CurrentUsage**
>
> OrganizationUsage getV2CurrentUsage()

Get organization usage from v2 billing

### Example

```typescript
import {
    UsageApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new UsageApi(configuration);

let organizationId: string; //Organization ID (default to undefined)

const { status, data } = await apiInstance.getV2CurrentUsage(
    organizationId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **organizationId** | [**string**] | Organization ID | defaults to undefined|

### Return type

**OrganizationUsage**

### Authorization

[JwtAuth](../README.md#JwtAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getV2PastUsage**
>
> Array<OrganizationUsage> getV2PastUsage()

Get historical organization usage from v2 billing

### Example

```typescript
import {
    UsageApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new UsageApi(configuration);

let organizationId: string; //Organization ID (default to undefined)
let periods: number; //Periods (optional) (default to undefined)

const { status, data } = await apiInstance.getV2PastUsage(
    organizationId,
    periods
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **organizationId** | [**string**] | Organization ID | defaults to undefined|
| **periods** | [**number**] | Periods | (optional) defaults to undefined|

### Return type

**Array<OrganizationUsage>**

### Authorization

[JwtAuth](../README.md#JwtAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)
