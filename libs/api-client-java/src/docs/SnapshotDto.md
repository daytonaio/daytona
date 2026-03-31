

# SnapshotDto


## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
|**id** | **String** |  |  |
|**organizationId** | **String** |  |  [optional] |
|**general** | **Boolean** |  |  |
|**name** | **String** |  |  |
|**imageName** | **String** |  |  [optional] |
|**state** | **SnapshotState** |  |  |
|**size** | **BigDecimal** |  |  |
|**entrypoint** | **List&lt;String&gt;** |  |  |
|**cpu** | **BigDecimal** |  |  |
|**gpu** | **BigDecimal** |  |  |
|**mem** | **BigDecimal** |  |  |
|**disk** | **BigDecimal** |  |  |
|**errorReason** | **String** |  |  |
|**createdAt** | **OffsetDateTime** |  |  |
|**updatedAt** | **OffsetDateTime** |  |  |
|**lastUsedAt** | **OffsetDateTime** |  |  |
|**buildInfo** | [**BuildInfo**](BuildInfo.md) | Build information for the snapshot |  [optional] |
|**regionIds** | **List&lt;String&gt;** | IDs of regions where the snapshot is available |  [optional] |
|**initialRunnerId** | **String** | The initial runner ID of the snapshot |  [optional] |
|**ref** | **String** | The snapshot reference |  [optional] |



