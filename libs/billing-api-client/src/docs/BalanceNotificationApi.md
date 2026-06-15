# BalanceNotificationApi

All URIs are relative to _http://localhost:6100_

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**getV2BalanceNotification**](#getv2balancenotification) | **GET** /v2/organization/{organizationId}/balance-notification | Get balance notification|
|[**setV2BalanceNotification**](#setv2balancenotification) | **PUT** /v2/organization/{organizationId}/balance-notification | Set balance notification|

# **getV2BalanceNotification**
>
> BalanceNotification getV2BalanceNotification()

Get the organization\'s wallet balance notification settings

### Example

```typescript
import {
    BalanceNotificationApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new BalanceNotificationApi(configuration);

let organizationId: string; //Organization ID (default to undefined)

const { status, data } = await apiInstance.getV2BalanceNotification(
    organizationId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **organizationId** | [**string**] | Organization ID | defaults to undefined|

### Return type

**BalanceNotification**

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

# **setV2BalanceNotification**
>
> BalanceNotification setV2BalanceNotification(balanceNotification)

Create, update, or disable the organization\'s wallet balance notification

### Example

```typescript
import {
    BalanceNotificationApi,
    Configuration,
    BalanceNotification
} from './api';

const configuration = new Configuration();
const apiInstance = new BalanceNotificationApi(configuration);

let organizationId: string; //Organization ID (default to undefined)
let balanceNotification: BalanceNotification; //Balance notification

const { status, data } = await apiInstance.setV2BalanceNotification(
    organizationId,
    balanceNotification
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **balanceNotification** | **BalanceNotification**| Balance notification | |
| **organizationId** | [**string**] | Organization ID | defaults to undefined|

### Return type

**BalanceNotification**

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
