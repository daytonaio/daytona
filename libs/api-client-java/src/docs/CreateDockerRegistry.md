

# CreateDockerRegistry


## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
|**name** | **String** | Registry name |  |
|**url** | **String** | Registry URL |  |
|**username** | **String** | Registry username |  |
|**password** | **String** | Registry password |  |
|**project** | **String** | Registry project |  [optional] |
|**registryType** | [**RegistryTypeEnum**](#RegistryTypeEnum) | Registry type |  |
|**isDefault** | **Boolean** | Set as default registry |  [optional] |



## Enum: RegistryTypeEnum

| Name | Value |
|---- | -----|
| INTERNAL | &quot;internal&quot; |
| ORGANIZATION | &quot;organization&quot; |
| TRANSIENT | &quot;transient&quot; |
| BACKUP | &quot;backup&quot; |



