# WalletApi

All URIs are relative to _http://localhost:6100_

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**getV2Wallet**](#getv2wallet) | **GET** /v2/organization/{organizationId}/wallet | Get organization wallet|
|[**getWallet**](#getwallet) | **GET** /organization/{organizationId}/wallet | Get organization wallet|
|[**setAutomaticTopUp**](#setautomatictopup) | **PUT** /organization/{organizationId}/wallet/automatic-top-up | Set automatic top up|
|[**setV2AutomaticTopUp**](#setv2automatictopup) | **PUT** /v2/organization/{organizationId}/wallet/automatic-top-up | Set automatic top up|
|[**topUpV2Wallet**](#topupv2wallet) | **POST** /v2/organization/{organizationId}/wallet/top-up | Top up wallet|
|[**topUpWallet**](#topupwallet) | **POST** /organization/{organizationId}/wallet/top-up | Top up wallet|

# **getV2Wallet**
>
> OrganizationWallet getV2Wallet()

Get organization wallet from v2 billing

### Example

```typescript
import {
    WalletApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new WalletApi(configuration);

let organizationId: string; //Organization ID (default to undefined)

const { status, data } = await apiInstance.getV2Wallet(
    organizationId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **organizationId** | [**string**] | Organization ID | defaults to undefined|

### Return type

**OrganizationWallet**

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

# **getWallet**
>
> OrganizationWallet getWallet()

Get organization wallet

### Example

```typescript
import {
    WalletApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new WalletApi(configuration);

let organizationId: string; //Organization ID (default to undefined)

const { status, data } = await apiInstance.getWallet(
    organizationId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **organizationId** | [**string**] | Organization ID | defaults to undefined|

### Return type

**OrganizationWallet**

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

# **setAutomaticTopUp**
>
> setAutomaticTopUp()

Set automatic top up

### Example

```typescript
import {
    WalletApi,
    Configuration,
    AutomaticTopUp
} from './api';

const configuration = new Configuration();
const apiInstance = new WalletApi(configuration);

let organizationId: string; //Organization ID (default to undefined)
let automaticTopUp: AutomaticTopUp; //Automatic top up (optional)

const { status, data } = await apiInstance.setAutomaticTopUp(
    organizationId,
    automaticTopUp
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **automaticTopUp** | **AutomaticTopUp**| Automatic top up | |
| **organizationId** | [**string**] | Organization ID | defaults to undefined|

### Return type

void (empty response body)

### Authorization

[JwtAuth](../README.md#JwtAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**204** | No Content |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **setV2AutomaticTopUp**
>
> setV2AutomaticTopUp()

Set automatic top up using v2 billing

### Example

```typescript
import {
    WalletApi,
    Configuration,
    AutomaticTopUp
} from './api';

const configuration = new Configuration();
const apiInstance = new WalletApi(configuration);

let organizationId: string; //Organization ID (default to undefined)
let automaticTopUp: AutomaticTopUp; //Automatic top up (optional)

const { status, data } = await apiInstance.setV2AutomaticTopUp(
    organizationId,
    automaticTopUp
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **automaticTopUp** | **AutomaticTopUp**| Automatic top up | |
| **organizationId** | [**string**] | Organization ID | defaults to undefined|

### Return type

void (empty response body)

### Authorization

[JwtAuth](../README.md#JwtAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**204** | No Content |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **topUpV2Wallet**
>
> PaymentUrl topUpV2Wallet(topUpRequest)

Top up wallet with specified amount using v2 billing

### Example

```typescript
import {
    WalletApi,
    Configuration,
    WalletTopUpRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new WalletApi(configuration);

let organizationId: string; //Organization ID (default to undefined)
let topUpRequest: WalletTopUpRequest; //Top up request

const { status, data } = await apiInstance.topUpV2Wallet(
    organizationId,
    topUpRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **topUpRequest** | **WalletTopUpRequest**| Top up request | |
| **organizationId** | [**string**] | Organization ID | defaults to undefined|

### Return type

**PaymentUrl**

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

# **topUpWallet**
>
> PaymentUrl topUpWallet(topUpRequest)

Top up wallet with specified amount

### Example

```typescript
import {
    WalletApi,
    Configuration,
    WalletTopUpRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new WalletApi(configuration);

let organizationId: string; //Organization ID (default to undefined)
let topUpRequest: WalletTopUpRequest; //Top up request

const { status, data } = await apiInstance.topUpWallet(
    organizationId,
    topUpRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **topUpRequest** | **WalletTopUpRequest**| Top up request | |
| **organizationId** | [**string**] | Organization ID | defaults to undefined|

### Return type

**PaymentUrl**

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
