# SetupCheckoutUrlApi

All URIs are relative to _http://localhost:6100_

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**getV2SetupCheckoutURL**](#getv2setupcheckouturl) | **GET** /v2/organization/{organizationId}/setup-checkout-url | Get setup checkout url for adding a payment method|

# **getV2SetupCheckoutURL**
>
> string getV2SetupCheckoutURL()

Returns a Stripe Checkout Session URL in setup mode with free-trial Radar metadata

### Example

```typescript
import {
    SetupCheckoutUrlApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new SetupCheckoutUrlApi(configuration);

let organizationId: string; //Organization ID (default to undefined)

const { status, data } = await apiInstance.getV2SetupCheckoutURL(
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
