# \ProcessAPI

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ConnectPtySession**](ProcessAPI.md#ConnectPtySession) | **Get** /process/pty/{sessionId}/connect | Connect to PTY session via WebSocket
[**CreatePtySession**](ProcessAPI.md#CreatePtySession) | **Post** /process/pty | Create a new PTY session
[**CreateSession**](ProcessAPI.md#CreateSession) | **Post** /process/session | Create a new session
[**DeletePtySession**](ProcessAPI.md#DeletePtySession) | **Delete** /process/pty/{sessionId} | Delete a PTY session
[**DeleteSession**](ProcessAPI.md#DeleteSession) | **Delete** /process/session/{sessionId} | Delete a session
[**ExecuteCommand**](ProcessAPI.md#ExecuteCommand) | **Post** /process/execute | Execute a command
[**GetPtySession**](ProcessAPI.md#GetPtySession) | **Get** /process/pty/{sessionId} | Get PTY session information
[**GetSession**](ProcessAPI.md#GetSession) | **Get** /process/session/{sessionId} | Get session details
[**GetSessionCommand**](ProcessAPI.md#GetSessionCommand) | **Get** /process/session/{sessionId}/command/{commandId} | Get session command details
[**GetSessionCommandLogs**](ProcessAPI.md#GetSessionCommandLogs) | **Get** /process/session/{sessionId}/command/{commandId}/logs | Get session command logs
[**ListPtySessions**](ProcessAPI.md#ListPtySessions) | **Get** /process/pty | List all PTY sessions
[**ListSessions**](ProcessAPI.md#ListSessions) | **Get** /process/session | List all sessions
[**ResizePtySession**](ProcessAPI.md#ResizePtySession) | **Post** /process/pty/{sessionId}/resize | Resize a PTY session
[**SessionExecuteCommand**](ProcessAPI.md#SessionExecuteCommand) | **Post** /process/session/{sessionId}/exec | Execute command in session



## ConnectPtySession

> ConnectPtySession(ctx, sessionId).Execute()

Connect to PTY session via WebSocket



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	sessionId := "sessionId_example" // string | PTY session ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ProcessAPI.ConnectPtySession(context.Background(), sessionId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProcessAPI.ConnectPtySession``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**sessionId** | **string** | PTY session ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiConnectPtySessionRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreatePtySession

> PtyCreateResponse CreatePtySession(ctx).Request(request).Execute()

Create a new PTY session



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	request := *openapiclient.NewPtyCreateRequest() // PtyCreateRequest | PTY session creation request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProcessAPI.CreatePtySession(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProcessAPI.CreatePtySession``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreatePtySession`: PtyCreateResponse
	fmt.Fprintf(os.Stdout, "Response from `ProcessAPI.CreatePtySession`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreatePtySessionRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**PtyCreateRequest**](PtyCreateRequest.md) | PTY session creation request | 

### Return type

[**PtyCreateResponse**](PtyCreateResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateSession

> CreateSession(ctx).Request(request).Execute()

Create a new session



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	request := *openapiclient.NewCreateSessionRequest("SessionId_example") // CreateSessionRequest | Session creation request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ProcessAPI.CreateSession(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProcessAPI.CreateSession``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateSessionRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**CreateSessionRequest**](CreateSessionRequest.md) | Session creation request | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeletePtySession

> map[string]map[string]interface{} DeletePtySession(ctx, sessionId).Execute()

Delete a PTY session



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	sessionId := "sessionId_example" // string | PTY session ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProcessAPI.DeletePtySession(context.Background(), sessionId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProcessAPI.DeletePtySession``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeletePtySession`: map[string]map[string]interface{}
	fmt.Fprintf(os.Stdout, "Response from `ProcessAPI.DeletePtySession`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**sessionId** | **string** | PTY session ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeletePtySessionRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

**map[string]map[string]interface{}**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteSession

> DeleteSession(ctx, sessionId).Execute()

Delete a session



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	sessionId := "sessionId_example" // string | Session ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ProcessAPI.DeleteSession(context.Background(), sessionId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProcessAPI.DeleteSession``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**sessionId** | **string** | Session ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteSessionRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ExecuteCommand

> ExecuteResponse ExecuteCommand(ctx).Request(request).Execute()

Execute a command



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	request := *openapiclient.NewExecuteRequest("Command_example") // ExecuteRequest | Command execution request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProcessAPI.ExecuteCommand(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProcessAPI.ExecuteCommand``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ExecuteCommand`: ExecuteResponse
	fmt.Fprintf(os.Stdout, "Response from `ProcessAPI.ExecuteCommand`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiExecuteCommandRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**ExecuteRequest**](ExecuteRequest.md) | Command execution request | 

### Return type

[**ExecuteResponse**](ExecuteResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetPtySession

> PtySessionInfo GetPtySession(ctx, sessionId).Execute()

Get PTY session information



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	sessionId := "sessionId_example" // string | PTY session ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProcessAPI.GetPtySession(context.Background(), sessionId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProcessAPI.GetPtySession``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetPtySession`: PtySessionInfo
	fmt.Fprintf(os.Stdout, "Response from `ProcessAPI.GetPtySession`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**sessionId** | **string** | PTY session ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetPtySessionRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**PtySessionInfo**](PtySessionInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetSession

> Session GetSession(ctx, sessionId).Execute()

Get session details



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	sessionId := "sessionId_example" // string | Session ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProcessAPI.GetSession(context.Background(), sessionId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProcessAPI.GetSession``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetSession`: Session
	fmt.Fprintf(os.Stdout, "Response from `ProcessAPI.GetSession`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**sessionId** | **string** | Session ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetSessionRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Session**](Session.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetSessionCommand

> Command GetSessionCommand(ctx, sessionId, commandId).Execute()

Get session command details



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	sessionId := "sessionId_example" // string | Session ID
	commandId := "commandId_example" // string | Command ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProcessAPI.GetSessionCommand(context.Background(), sessionId, commandId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProcessAPI.GetSessionCommand``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetSessionCommand`: Command
	fmt.Fprintf(os.Stdout, "Response from `ProcessAPI.GetSessionCommand`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**sessionId** | **string** | Session ID | 
**commandId** | **string** | Command ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetSessionCommandRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**Command**](Command.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetSessionCommandLogs

> string GetSessionCommandLogs(ctx, sessionId, commandId).Follow(follow).Execute()

Get session command logs



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	sessionId := "sessionId_example" // string | Session ID
	commandId := "commandId_example" // string | Command ID
	follow := true // bool | Follow logs in real-time (WebSocket only) (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProcessAPI.GetSessionCommandLogs(context.Background(), sessionId, commandId).Follow(follow).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProcessAPI.GetSessionCommandLogs``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetSessionCommandLogs`: string
	fmt.Fprintf(os.Stdout, "Response from `ProcessAPI.GetSessionCommandLogs`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**sessionId** | **string** | Session ID | 
**commandId** | **string** | Command ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetSessionCommandLogsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **follow** | **bool** | Follow logs in real-time (WebSocket only) | 

### Return type

**string**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: text/plain

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListPtySessions

> PtyListResponse ListPtySessions(ctx).Execute()

List all PTY sessions



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProcessAPI.ListPtySessions(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProcessAPI.ListPtySessions``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListPtySessions`: PtyListResponse
	fmt.Fprintf(os.Stdout, "Response from `ProcessAPI.ListPtySessions`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiListPtySessionsRequest struct via the builder pattern


### Return type

[**PtyListResponse**](PtyListResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListSessions

> []Session ListSessions(ctx).Execute()

List all sessions



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProcessAPI.ListSessions(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProcessAPI.ListSessions``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListSessions`: []Session
	fmt.Fprintf(os.Stdout, "Response from `ProcessAPI.ListSessions`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiListSessionsRequest struct via the builder pattern


### Return type

[**[]Session**](Session.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ResizePtySession

> PtySessionInfo ResizePtySession(ctx, sessionId).Request(request).Execute()

Resize a PTY session



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	sessionId := "sessionId_example" // string | PTY session ID
	request := *openapiclient.NewPtyResizeRequest(int32(123), int32(123)) // PtyResizeRequest | Resize request with new dimensions

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProcessAPI.ResizePtySession(context.Background(), sessionId).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProcessAPI.ResizePtySession``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ResizePtySession`: PtySessionInfo
	fmt.Fprintf(os.Stdout, "Response from `ProcessAPI.ResizePtySession`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**sessionId** | **string** | PTY session ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiResizePtySessionRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **request** | [**PtyResizeRequest**](PtyResizeRequest.md) | Resize request with new dimensions | 

### Return type

[**PtySessionInfo**](PtySessionInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SessionExecuteCommand

> SessionExecuteResponse SessionExecuteCommand(ctx, sessionId).Request(request).Execute()

Execute command in session



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/daytonaio/daytona/libs/toolbox-api-go"
)

func main() {
	sessionId := "sessionId_example" // string | Session ID
	request := *openapiclient.NewSessionExecuteRequest("Command_example") // SessionExecuteRequest | Command execution request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProcessAPI.SessionExecuteCommand(context.Background(), sessionId).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProcessAPI.SessionExecuteCommand``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SessionExecuteCommand`: SessionExecuteResponse
	fmt.Fprintf(os.Stdout, "Response from `ProcessAPI.SessionExecuteCommand`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**sessionId** | **string** | Session ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiSessionExecuteCommandRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **request** | [**SessionExecuteRequest**](SessionExecuteRequest.md) | Command execution request | 

### Return type

[**SessionExecuteResponse**](SessionExecuteResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

