# AutomaticTopUp

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**disabled** | **boolean** | Disabled is true when the threshold rule still has amounts configured but is not currently active. The most common cause is Metronome auto-disabling the rule after a Stripe charge for a threshold top-up fails. The customer can re-enable by re-submitting the rule through SetAutomaticTopUp. | [optional] [default to undefined]
**targetAmount** | **number** |  | [optional] [default to undefined]
**thresholdAmount** | **number** |  | [optional] [default to undefined]

## Example

```typescript
import { AutomaticTopUp } from './api';

const instance: AutomaticTopUp = {
    disabled,
    targetAmount,
    thresholdAmount,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
