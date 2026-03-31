

# DockerRegistry


## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
|**id** | **String** | Registry ID |  |
|**name** | **String** | Registry name |  |
|**url** | **String** | Registry URL |  |
|**username** | **String** | Registry username |  |
|**project** | **String** | Registry project |  |
|**registryType** | [**RegistryTypeEnum**](#RegistryTypeEnum) | Registry type |  |
|**createdAt** | **OffsetDateTime** | Creation timestamp |  |
|**updatedAt** | **OffsetDateTime** | Last update timestamp |  |



## Enum: RegistryTypeEnum

| Name | Value |
|---- | -----|
| INTERNAL | &quot;internal&quot; |
| ORGANIZATION | &quot;organization&quot; |
| TRANSIENT | &quot;transient&quot; |
| BACKUP | &quot;backup&quot; |



