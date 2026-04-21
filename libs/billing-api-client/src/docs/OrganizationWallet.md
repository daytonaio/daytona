# OrganizationWallet

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**automaticTopUp** | [**AutomaticTopUp**](AutomaticTopUp.md) |  | [optional] [default to undefined]
**balanceCents** | **number** |  | [optional] [default to undefined]
**billingType** | [**BillingType**](BillingType.md) |  | [optional] [default to undefined]
**creditCardConnected** | **boolean** |  | [optional] [default to undefined]
**hasFailedOrPendingInvoice** | **boolean** |  | [optional] [default to undefined]
**name** | **string** |  | [optional] [default to undefined]
**ongoingBalanceCents** | **number** |  | [optional] [default to undefined]

## Example

```typescript
import { OrganizationWallet } from './api';

const instance: OrganizationWallet = {
    automaticTopUp,
    balanceCents,
    billingType,
    creditCardConnected,
    hasFailedOrPendingInvoice,
    name,
    ongoingBalanceCents,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
