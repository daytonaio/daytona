# PortalUrlApi

All URIs are relative to _http://localhost:6100_

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**getPortalUrl**](#getportalurl) | **GET** /organization/{organizationId}/portal-url | Get organization billing portal url|
|[**getV2PortalURL**](#getv2portalurl) | **GET** /v2/organization/{organizationId}/portal-url | Get organization billing portal url|

# **getPortalUrl**
>
> string getPortalUrl()

Get organization billing portal url

### Example

```typescript
import {
    PortalUrlApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new PortalUrlApi(configuration);

let organizationId: string; //Organization ID (default to undefined)

const { status, data } = await apiInstance.getPortalUrl(
    organizationId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **organizationId** | [**string**] | Organization ID | defaults to undefined|

### Return type

**string**

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

# **getV2PortalURL**
>
> string getV2PortalURL()

Get organization billing portal url from v2 billing

### Example

```typescript
import {
    PortalUrlApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new PortalUrlApi(configuration);

let organizationId: string; //Organization ID (default to undefined)

const { status, data } = await apiInstance.getV2PortalURL(
    organizationId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **organizationId** | [**string**] | Organization ID | defaults to undefined|

### Return type

**string**

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
