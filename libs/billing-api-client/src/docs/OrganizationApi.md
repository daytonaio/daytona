# OrganizationApi

All URIs are relative to _http://localhost:6100_

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**addOrganizationEmail**](#addorganizationemail) | **POST** /organization/{organizationId}/email | Add organization email|
|[**deleteOrganizationEmail**](#deleteorganizationemail) | **DELETE** /organization/{organizationId}/email | Delete organization email|
|[**downgradeTier**](#downgradetier) | **POST** /organization/{organizationId}/tier/downgrade | Downgrade organization tier|
|[**getTier**](#gettier) | **GET** /organization/{organizationId}/tier | Get organization tier|
|[**listOrganizationEmails**](#listorganizationemails) | **GET** /organization/{organizationId}/email | List organization emails|
|[**redeemCoupon**](#redeemcoupon) | **POST** /organization/{organizationId}/redeem-coupon/{couponCode} | Redeem coupon|
|[**redeemV2Coupon**](#redeemv2coupon) | **POST** /v2/organization/{organizationId}/redeem-coupon/{couponCode} | Redeem coupon|
|[**resendVerificationEmail**](#resendverificationemail) | **POST** /organization/{organizationId}/email/resend | Resend verification email|
|[**upgradeTier**](#upgradetier) | **POST** /organization/{organizationId}/tier/upgrade | Upgrade organization tier|
|[**verifyEmail**](#verifyemail) | **POST** /organization/{organizationId}/email/verify | Verify email|
|[**verifyInternetAccess**](#verifyinternetaccess) | **POST** /organization/{organizationId}/verify-internet-access | Verify internet access|

# **addOrganizationEmail**
>
> OrganizationEmail addOrganizationEmail(data)

Add organization email

### Example

```typescript
import {
    OrganizationApi,
    Configuration,
    UpdateOrganizationEmail
} from './api';

const configuration = new Configuration();
const apiInstance = new OrganizationApi(configuration);

let organizationId: string; //Organization ID (default to undefined)
let data: UpdateOrganizationEmail; //Email

const { status, data } = await apiInstance.addOrganizationEmail(
    organizationId,
    data
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **data** | **UpdateOrganizationEmail**| Email | |
| **organizationId** | [**string**] | Organization ID | defaults to undefined|

### Return type

**OrganizationEmail**

### Authorization

[JwtAuth](../README.md#JwtAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: _/_

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | Created |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteOrganizationEmail**
>
> deleteOrganizationEmail(data)

Delete organization email

### Example

```typescript
import {
    OrganizationApi,
    Configuration,
    UpdateOrganizationEmail
} from './api';

const configuration = new Configuration();
const apiInstance = new OrganizationApi(configuration);

let organizationId: string; //Organization ID (default to undefined)
let data: UpdateOrganizationEmail; //Email

const { status, data } = await apiInstance.deleteOrganizationEmail(
    organizationId,
    data
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **data** | **UpdateOrganizationEmail**| Email | |
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

# **downgradeTier**
>
> downgradeTier(body)

Downgrade organization tier

### Example

```typescript
import {
    OrganizationApi,
    Configuration,
    OrganizationTierUpdate
} from './api';

const configuration = new Configuration();
const apiInstance = new OrganizationApi(configuration);

let organizationId: string; //Organization ID (default to undefined)
let body: OrganizationTierUpdate; //Organization Tier Update

const { status, data } = await apiInstance.downgradeTier(
    organizationId,
    body
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **body** | **OrganizationTierUpdate**| Organization Tier Update | |
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
|**202** | Accepted |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getTier**
>
> OrganizationTier getTier()

Get organization tier

### Example

```typescript
import {
    OrganizationApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new OrganizationApi(configuration);

let organizationId: string; //Organization ID (default to undefined)

const { status, data } = await apiInstance.getTier(
    organizationId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **organizationId** | [**string**] | Organization ID | defaults to undefined|

### Return type

**OrganizationTier**

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

# **listOrganizationEmails**
>
> Array<OrganizationEmail> listOrganizationEmails()

List organization emails

### Example

```typescript
import {
    OrganizationApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new OrganizationApi(configuration);

let organizationId: string; //Organization ID (default to undefined)

const { status, data } = await apiInstance.listOrganizationEmails(
    organizationId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **organizationId** | [**string**] | Organization ID | defaults to undefined|

### Return type

**Array<OrganizationEmail>**

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

# **redeemCoupon**
>
> Organization redeemCoupon()

Redeem coupon

### Example

```typescript
import {
    OrganizationApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new OrganizationApi(configuration);

let couponCode: string; //Coupon Code (default to undefined)
let organizationId: string; //Organization ID (default to undefined)

const { status, data } = await apiInstance.redeemCoupon(
    couponCode,
    organizationId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **couponCode** | [**string**] | Coupon Code | defaults to undefined|
| **organizationId** | [**string**] | Organization ID | defaults to undefined|

### Return type

**Organization**

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

# **redeemV2Coupon**
>
> { [key: string]: string; } redeemV2Coupon()

Redeem coupon using v2 billing

### Example

```typescript
import {
    OrganizationApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new OrganizationApi(configuration);

let couponCode: string; //Coupon Code (default to undefined)
let organizationId: string; //Organization ID (default to undefined)

const { status, data } = await apiInstance.redeemV2Coupon(
    couponCode,
    organizationId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **couponCode** | [**string**] | Coupon Code | defaults to undefined|
| **organizationId** | [**string**] | Organization ID | defaults to undefined|

### Return type

**{ [key: string]: string; }**

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

# **resendVerificationEmail**
>
> resendVerificationEmail(data)

Resend verification email

### Example

```typescript
import {
    OrganizationApi,
    Configuration,
    UpdateOrganizationEmail
} from './api';

const configuration = new Configuration();
const apiInstance = new OrganizationApi(configuration);

let organizationId: string; //Organization ID (default to undefined)
let data: UpdateOrganizationEmail; //Email

const { status, data } = await apiInstance.resendVerificationEmail(
    organizationId,
    data
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **data** | **UpdateOrganizationEmail**| Email | |
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
|**200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **upgradeTier**
>
> upgradeTier(body)

Upgrade organization tier

### Example

```typescript
import {
    OrganizationApi,
    Configuration,
    OrganizationTierUpdate
} from './api';

const configuration = new Configuration();
const apiInstance = new OrganizationApi(configuration);

let organizationId: string; //Organization ID (default to undefined)
let body: OrganizationTierUpdate; //Organization Tier Update

const { status, data } = await apiInstance.upgradeTier(
    organizationId,
    body
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **body** | **OrganizationTierUpdate**| Organization Tier Update | |
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
|**202** | Accepted |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **verifyEmail**
>
> verifyEmail(data)

Verify email

### Example

```typescript
import {
    OrganizationApi,
    Configuration,
    VerifyEmail
} from './api';

const configuration = new Configuration();
const apiInstance = new OrganizationApi(configuration);

let organizationId: string; //Organization ID (default to undefined)
let data: VerifyEmail; //Email

const { status, data } = await apiInstance.verifyEmail(
    organizationId,
    data
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **data** | **VerifyEmail**| Email | |
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
|**200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **verifyInternetAccess**
>
> verifyInternetAccess(data)

Verify internet access via Stripe Radar to unlock unrestricted network egress

### Example

```typescript
import {
    OrganizationApi,
    Configuration,
    VerifyInternetAccess
} from './api';

const configuration = new Configuration();
const apiInstance = new OrganizationApi(configuration);

let organizationId: string; //Organization ID (default to undefined)
let data: VerifyInternetAccess; //Radar session token

const { status, data } = await apiInstance.verifyInternetAccess(
    organizationId,
    data
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **data** | **VerifyInternetAccess**| Radar session token | |
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
|**200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)
