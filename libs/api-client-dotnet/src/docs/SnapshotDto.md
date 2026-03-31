# Daytona.ApiClient.Model.SnapshotDto

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** |  | 
**OrganizationId** | **string** |  | [optional] 
**General** | **bool** |  | 
**Name** | **string** |  | 
**ImageName** | **string** |  | [optional] 
**State** | **SnapshotState** |  | 
**Size** | **decimal?** |  | 
**Entrypoint** | **List&lt;string&gt;** |  | 
**Cpu** | **decimal** |  | 
**Gpu** | **decimal** |  | 
**Mem** | **decimal** |  | 
**Disk** | **decimal** |  | 
**ErrorReason** | **string** |  | 
**CreatedAt** | **DateTime** |  | 
**UpdatedAt** | **DateTime** |  | 
**LastUsedAt** | **DateTime?** |  | 
**BuildInfo** | [**BuildInfo**](BuildInfo.md) | Build information for the snapshot | [optional] 
**RegionIds** | **List&lt;string&gt;** | IDs of regions where the snapshot is available | [optional] 
**InitialRunnerId** | **string** | The initial runner ID of the snapshot | [optional] 
**Ref** | **string** | The snapshot reference | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

