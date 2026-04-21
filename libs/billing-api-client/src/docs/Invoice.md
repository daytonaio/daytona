# Invoice

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**currency** | **string** |  | [optional] [default to undefined]
**errorDetails** | [**Array&lt;InvoiceErrorDetail&gt;**](InvoiceErrorDetail.md) |  | [optional] [default to undefined]
**fileUrl** | **string** |  | [optional] [default to undefined]
**id** | **string** |  | [optional] [default to undefined]
**issuingDate** | **string** |  | [optional] [default to undefined]
**number** | **string** |  | [optional] [default to undefined]
**paymentDueDate** | **string** |  | [optional] [default to undefined]
**paymentOverdue** | **boolean** |  | [optional] [default to undefined]
**paymentStatus** | [**InvoicePaymentStatus**](InvoicePaymentStatus.md) |  | [optional] [default to undefined]
**sequentialId** | **number** |  | [optional] [default to undefined]
**status** | [**InvoiceStatus**](InvoiceStatus.md) |  | [optional] [default to undefined]
**totalAmountCents** | **number** |  | [optional] [default to undefined]
**totalDueAmountCents** | **number** |  | [optional] [default to undefined]
**type** | [**InvoiceType**](InvoiceType.md) |  | [optional] [default to undefined]

## Example

```typescript
import { Invoice } from './api';

const instance: Invoice = {
    currency,
    errorDetails,
    fileUrl,
    id,
    issuingDate,
    number,
    paymentDueDate,
    paymentOverdue,
    paymentStatus,
    sequentialId,
    status,
    totalAmountCents,
    totalDueAmountCents,
    type,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
