

# CreateUser


## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
|**id** | **String** |  |  |
|**name** | **String** |  |  |
|**email** | **String** |  |  [optional] |
|**personalOrganizationQuota** | [**CreateOrganizationQuota**](CreateOrganizationQuota.md) |  |  [optional] |
|**personalOrganizationDefaultRegionId** | **String** |  |  [optional] |
|**role** | [**RoleEnum**](#RoleEnum) |  |  [optional] |
|**emailVerified** | **Boolean** |  |  [optional] |



## Enum: RoleEnum

| Name | Value |
|---- | -----|
| ADMIN | &quot;admin&quot; |
| USER | &quot;user&quot; |



