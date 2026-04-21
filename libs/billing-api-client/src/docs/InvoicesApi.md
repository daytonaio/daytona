# InvoicesApi

All URIs are relative to _http://localhost:6100_

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**createPaymentUrl**](#createpaymenturl) | **POST** /organization/{organizationId}/invoices/{invoiceId}/payment-url | Create payment URL for invoice|
|[**createV2PaymentURL**](#createv2paymenturl) | **POST** /v2/organization/{organizationId}/invoices/{invoiceId}/payment-url | Create payment URL for invoice|
|[**listInvoices**](#listinvoices) | **GET** /organization/{organizationId}/invoices | Get organization invoices|
|[**listV2Invoices**](#listv2invoices) | **GET** /v2/organization/{organizationId}/invoices | Get organization invoices|
|[**voidInvoice**](#voidinvoice) | **POST** /organization/{organizationId}/invoices/{invoiceId}/void | Void an invoice|

# **createPaymentUrl**
>
> PaymentUrl createPaymentUrl()

Create payment URL for invoice

### Example

```typescript
import {
    InvoicesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new InvoicesApi(configuration);

let organizationId: string; //Organization ID (default to undefined)
let invoiceId: string; //Invoice ID (default to undefined)

const { status, data } = await apiInstance.createPaymentUrl(
    organizationId,
    invoiceId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **organizationId** | [**string**] | Organization ID | defaults to undefined|
| **invoiceId** | [**string**] | Invoice ID | defaults to undefined|

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

# **createV2PaymentURL**
>
> PaymentUrl createV2PaymentURL()

Create payment URL for invoice using v2 billing

### Example

```typescript
import {
    InvoicesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new InvoicesApi(configuration);

let organizationId: string; //Organization ID (default to undefined)
let invoiceId: string; //Invoice ID (default to undefined)

const { status, data } = await apiInstance.createV2PaymentURL(
    organizationId,
    invoiceId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **organizationId** | [**string**] | Organization ID | defaults to undefined|
| **invoiceId** | [**string**] | Invoice ID | defaults to undefined|

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

# **listInvoices**
>
> PaginatedTInvoice listInvoices()

Get organization invoices

### Example

```typescript
import {
    InvoicesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new InvoicesApi(configuration);

let organizationId: string; //Organization ID (default to undefined)
let page: number; //Page number (optional) (default to undefined)
let perPage: number; //Number of items per page (optional) (default to undefined)

const { status, data } = await apiInstance.listInvoices(
    organizationId,
    page,
    perPage
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **organizationId** | [**string**] | Organization ID | defaults to undefined|
| **page** | [**number**] | Page number | (optional) defaults to undefined|
| **perPage** | [**number**] | Number of items per page | (optional) defaults to undefined|

### Return type

**PaginatedTInvoice**

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

# **listV2Invoices**
>
> PaginatedTInvoice listV2Invoices()

Get organization invoices from v2 billing

### Example

```typescript
import {
    InvoicesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new InvoicesApi(configuration);

let organizationId: string; //Organization ID (default to undefined)
let page: number; //Page number (optional) (default to undefined)
let perPage: number; //Number of items per page (optional) (default to undefined)

const { status, data } = await apiInstance.listV2Invoices(
    organizationId,
    page,
    perPage
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **organizationId** | [**string**] | Organization ID | defaults to undefined|
| **page** | [**number**] | Page number | (optional) defaults to undefined|
| **perPage** | [**number**] | Number of items per page | (optional) defaults to undefined|

### Return type

**PaginatedTInvoice**

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

# **voidInvoice**
>
> voidInvoice()

Void an invoice

### Example

```typescript
import {
    InvoicesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new InvoicesApi(configuration);

let organizationId: string; //Organization ID (default to undefined)
let invoiceId: string; //Invoice ID (default to undefined)

const { status, data } = await apiInstance.voidInvoice(
    organizationId,
    invoiceId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **organizationId** | [**string**] | Organization ID | defaults to undefined|
| **invoiceId** | [**string**] | Invoice ID | defaults to undefined|

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
