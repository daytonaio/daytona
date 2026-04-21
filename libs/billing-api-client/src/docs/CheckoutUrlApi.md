# CheckoutUrlApi

All URIs are relative to _http://localhost:6100_

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**getCheckoutUrl**](#getcheckouturl) | **GET** /organization/{organizationId}/checkout-url | Get organization checkout url|
|[**getV2CheckoutURL**](#getv2checkouturl) | **GET** /v2/organization/{organizationId}/checkout-url | Get organization checkout url|

# **getCheckoutUrl**
>
> string getCheckoutUrl()

Get organization checkout url

### Example

```typescript
import {
    CheckoutUrlApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CheckoutUrlApi(configuration);

let organizationId: string; //Organization ID (default to undefined)

const { status, data } = await apiInstance.getCheckoutUrl(
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

# **getV2CheckoutURL**
>
> string getV2CheckoutURL()

Get organization checkout url from v2 billing

### Example

```typescript
import {
    CheckoutUrlApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new CheckoutUrlApi(configuration);

let organizationId: string; //Organization ID (default to undefined)

const { status, data } = await apiInstance.getV2CheckoutURL(
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
