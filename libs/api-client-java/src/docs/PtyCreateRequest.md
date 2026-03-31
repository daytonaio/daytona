

# PtyCreateRequest


## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
|**id** | **String** | The unique identifier for the PTY session |  |
|**cwd** | **String** | Starting directory for the PTY session, defaults to the sandbox&#39;s working directory |  [optional] |
|**envs** | **Object** | Environment variables for the PTY session |  [optional] |
|**cols** | **BigDecimal** | Number of terminal columns |  [optional] |
|**rows** | **BigDecimal** | Number of terminal rows |  [optional] |
|**lazyStart** | **Boolean** | Whether to start the PTY session lazily (only start when first client connects) |  [optional] |



