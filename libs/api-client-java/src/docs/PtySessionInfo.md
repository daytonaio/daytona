

# PtySessionInfo


## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
|**id** | **String** | The unique identifier for the PTY session |  |
|**cwd** | **String** | Starting directory for the PTY session, defaults to the sandbox&#39;s working directory |  |
|**envs** | **Object** | Environment variables for the PTY session |  |
|**cols** | **BigDecimal** | Number of terminal columns |  |
|**rows** | **BigDecimal** | Number of terminal rows |  |
|**createdAt** | **String** | When the PTY session was created |  |
|**active** | **Boolean** | Whether the PTY session is currently active |  |
|**lazyStart** | **Boolean** | Whether the PTY session uses lazy start (only start when first client connects) |  |



