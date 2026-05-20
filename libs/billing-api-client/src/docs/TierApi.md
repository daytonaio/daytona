# TierApi

All URIs are relative to _http://localhost:6100_

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**listTiers**](#listtiers) | **GET** /tier | List tiers|

# **listTiers**
>
> Array<Tier> listTiers()

List tiers

### Example

```typescript
import {
    TierApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new TierApi(configuration);

const { status, data } = await apiInstance.listTiers();
```

### Parameters

This endpoint does not have any parameters.

### Return type

**Array<Tier>**

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
