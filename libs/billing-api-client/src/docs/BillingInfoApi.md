# BillingInfoApi

All URIs are relative to _http://localhost:6100_

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**getV2BillingInfo**](#getv2billinginfo) | **GET** /v2/organization/{organizationId}/billing-info | Get billing info|
|[**listV2Charges**](#listv2charges) | **GET** /v2/organization/{organizationId}/charges | List Stripe charges|
|[**listV2PaymentMethods**](#listv2paymentmethods) | **GET** /v2/organization/{organizationId}/payment-methods | List payment methods|

# **getV2BillingInfo**
>
> BillingInfo getV2BillingInfo()

Get organization billing contact and address from v2 billing

### Example

```typescript
import {
    BillingInfoApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new BillingInfoApi(configuration);

let organizationId: string; //Organization ID (default to undefined)

const { status, data } = await apiInstance.getV2BillingInfo(
    organizationId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **organizationId** | [**string**] | Organization ID | defaults to undefined|

### Return type

**BillingInfo**

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

# **listV2Charges**
>
> ChargeList listV2Charges()

List successful and failed Stripe charges for the organization

### Example

```typescript
import {
    BillingInfoApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new BillingInfoApi(configuration);

let organizationId: string; //Organization ID (default to undefined)
let limit: number; //Page size (1-100, default 25) (optional) (default to undefined)
let startingAfter: string; //Stripe cursor (charge ID) to paginate after (optional) (default to undefined)

const { status, data } = await apiInstance.listV2Charges(
    organizationId,
    limit,
    startingAfter
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **organizationId** | [**string**] | Organization ID | defaults to undefined|
| **limit** | [**number**] | Page size (1-100, default 25) | (optional) defaults to undefined|
| **startingAfter** | [**string**] | Stripe cursor (charge ID) to paginate after | (optional) defaults to undefined|

### Return type

**ChargeList**

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

# **listV2PaymentMethods**
>
> Array<PaymentMethod> listV2PaymentMethods()

List card payment methods attached to the Stripe customer

### Example

```typescript
import {
    BillingInfoApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new BillingInfoApi(configuration);

let organizationId: string; //Organization ID (default to undefined)

const { status, data } = await apiInstance.listV2PaymentMethods(
    organizationId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **organizationId** | [**string**] | Organization ID | defaults to undefined|

### Return type

**Array<PaymentMethod>**

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
