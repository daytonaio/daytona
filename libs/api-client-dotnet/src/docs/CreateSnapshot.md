# Daytona.ApiClient.Model.CreateSnapshot

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the snapshot | 
**ImageName** | **string** | The image name of the snapshot | [optional] 
**Entrypoint** | **List&lt;string&gt;** | The entrypoint command for the snapshot | [optional] 
**General** | **bool** | Whether the snapshot is general | [optional] 
**Cpu** | **int** | CPU cores allocated to the resulting sandbox | [optional] 
**Gpu** | **int** | GPU units allocated to the resulting sandbox | [optional] 
**Memory** | **int** | Memory allocated to the resulting sandbox in GB | [optional] 
**Disk** | **int** | Disk space allocated to the sandbox in GB | [optional] 
**BuildInfo** | [**CreateBuildInfo**](CreateBuildInfo.md) | Build information for the snapshot | [optional] 
**RegionId** | **string** | ID of the region where the snapshot will be available. Defaults to organization default region if not specified. | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

