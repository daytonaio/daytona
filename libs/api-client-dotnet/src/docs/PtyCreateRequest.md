# Daytona.ApiClient.Model.PtyCreateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The unique identifier for the PTY session | 
**Cwd** | **string** | Starting directory for the PTY session, defaults to the sandbox&#39;s working directory | [optional] 
**Envs** | **Object** | Environment variables for the PTY session | [optional] 
**Cols** | **decimal** | Number of terminal columns | [optional] 
**Rows** | **decimal** | Number of terminal rows | [optional] 
**LazyStart** | **bool** | Whether to start the PTY session lazily (only start when first client connects) | [optional] [default to false]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

