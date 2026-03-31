# Daytona.ApiClient.Model.PtySessionInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The unique identifier for the PTY session | 
**Cwd** | **string** | Starting directory for the PTY session, defaults to the sandbox&#39;s working directory | 
**Envs** | **Object** | Environment variables for the PTY session | 
**Cols** | **decimal** | Number of terminal columns | 
**Rows** | **decimal** | Number of terminal rows | 
**CreatedAt** | **string** | When the PTY session was created | 
**Active** | **bool** | Whether the PTY session is currently active | 
**LazyStart** | **bool** | Whether the PTY session uses lazy start (only start when first client connects) | [default to false]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

