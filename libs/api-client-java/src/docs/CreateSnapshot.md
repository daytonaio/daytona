

# CreateSnapshot


## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
|**name** | **String** | The name of the snapshot |  |
|**imageName** | **String** | The image name of the snapshot |  [optional] |
|**entrypoint** | **List&lt;String&gt;** | The entrypoint command for the snapshot |  [optional] |
|**general** | **Boolean** | Whether the snapshot is general |  [optional] |
|**cpu** | **Integer** | CPU cores allocated to the resulting sandbox |  [optional] |
|**gpu** | **Integer** | GPU units allocated to the resulting sandbox |  [optional] |
|**memory** | **Integer** | Memory allocated to the resulting sandbox in GB |  [optional] |
|**disk** | **Integer** | Disk space allocated to the sandbox in GB |  [optional] |
|**buildInfo** | [**CreateBuildInfo**](CreateBuildInfo.md) | Build information for the snapshot |  [optional] |
|**regionId** | **String** | ID of the region where the snapshot will be available. Defaults to organization default region if not specified. |  [optional] |



