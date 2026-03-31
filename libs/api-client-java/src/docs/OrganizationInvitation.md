

# OrganizationInvitation


## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
|**id** | **String** | Invitation ID |  |
|**email** | **String** | Email address of the invitee |  |
|**invitedBy** | **String** | Email address of the inviter |  |
|**organizationId** | **String** | Organization ID |  |
|**organizationName** | **String** | Organization name |  |
|**expiresAt** | **OffsetDateTime** | Expiration date of the invitation |  |
|**status** | [**StatusEnum**](#StatusEnum) | Invitation status |  |
|**role** | [**RoleEnum**](#RoleEnum) | Member role |  |
|**assignedRoles** | [**List&lt;OrganizationRole&gt;**](OrganizationRole.md) | Assigned roles |  |
|**createdAt** | **OffsetDateTime** | Creation timestamp |  |
|**updatedAt** | **OffsetDateTime** | Last update timestamp |  |



## Enum: StatusEnum

| Name | Value |
|---- | -----|
| PENDING | &quot;pending&quot; |
| ACCEPTED | &quot;accepted&quot; |
| DECLINED | &quot;declined&quot; |
| CANCELLED | &quot;cancelled&quot; |



## Enum: RoleEnum

| Name | Value |
|---- | -----|
| OWNER | &quot;owner&quot; |
| MEMBER | &quot;member&quot; |



