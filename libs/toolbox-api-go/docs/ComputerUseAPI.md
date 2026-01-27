# \ComputerUseAPI

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**Click**](ComputerUseAPI.md#Click) | **Post** /computeruse/mouse/click | Click mouse button
[**Drag**](ComputerUseAPI.md#Drag) | **Post** /computeruse/mouse/drag | Drag mouse
[**GetComputerUseStatus**](ComputerUseAPI.md#GetComputerUseStatus) | **Get** /computeruse/process-status | Get computer use process status
[**GetComputerUseSystemStatus**](ComputerUseAPI.md#GetComputerUseSystemStatus) | **Get** /computeruse/status | Get computer use status
[**GetDisplayInfo**](ComputerUseAPI.md#GetDisplayInfo) | **Get** /computeruse/display/info | Get display information
[**GetMousePosition**](ComputerUseAPI.md#GetMousePosition) | **Get** /computeruse/mouse/position | Get mouse position
[**GetProcessErrors**](ComputerUseAPI.md#GetProcessErrors) | **Get** /computeruse/process/{processName}/errors | Get process errors
[**GetProcessLogs**](ComputerUseAPI.md#GetProcessLogs) | **Get** /computeruse/process/{processName}/logs | Get process logs
[**GetProcessStatus**](ComputerUseAPI.md#GetProcessStatus) | **Get** /computeruse/process/{processName}/status | Get specific process status
[**GetWindows**](ComputerUseAPI.md#GetWindows) | **Get** /computeruse/display/windows | Get windows information
[**MoveMouse**](ComputerUseAPI.md#MoveMouse) | **Post** /computeruse/mouse/move | Move mouse cursor
[**PressHotkey**](ComputerUseAPI.md#PressHotkey) | **Post** /computeruse/keyboard/hotkey | Press hotkey
[**PressKey**](ComputerUseAPI.md#PressKey) | **Post** /computeruse/keyboard/key | Press key
[**RestartProcess**](ComputerUseAPI.md#RestartProcess) | **Post** /computeruse/process/{processName}/restart | Restart specific process
[**Scroll**](ComputerUseAPI.md#Scroll) | **Post** /computeruse/mouse/scroll | Scroll mouse wheel
[**StartComputerUse**](ComputerUseAPI.md#StartComputerUse) | **Post** /computeruse/start | Start computer use processes
[**StopComputerUse**](ComputerUseAPI.md#StopComputerUse) | **Post** /computeruse/stop | Stop computer use processes
[**TakeCompressedRegionScreenshot**](ComputerUseAPI.md#TakeCompressedRegionScreenshot) | **Get** /computeruse/screenshot/region/compressed | Take a compressed region screenshot
[**TakeCompressedScreenshot**](ComputerUseAPI.md#TakeCompressedScreenshot) | **Get** /computeruse/screenshot/compressed | Take a compressed screenshot
[**TakeRegionScreenshot**](ComputerUseAPI.md#TakeRegionScreenshot) | **Get** /computeruse/screenshot/region | Take a region screenshot
[**TakeScreenshot**](ComputerUseAPI.md#TakeScreenshot) | **Get** /computeruse/screenshot | Take a screenshot
[**TypeText**](ComputerUseAPI.md#TypeText) | **Post** /computeruse/keyboard/type | Type text



## Click

> MouseClickResponse Click(ctx).Request(request).Execute()

Click mouse button



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
	request := *openapiclient.NewMouseClickRequest() // MouseClickRequest | Mouse click request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ComputerUseAPI.Click(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComputerUseAPI.Click``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `Click`: MouseClickResponse
	fmt.Fprintf(os.Stdout, "Response from `ComputerUseAPI.Click`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiClickRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**MouseClickRequest**](MouseClickRequest.md) | Mouse click request | 

### Return type

[**MouseClickResponse**](MouseClickResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## Drag

> MouseDragResponse Drag(ctx).Request(request).Execute()

Drag mouse



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
	request := *openapiclient.NewMouseDragRequest() // MouseDragRequest | Mouse drag request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ComputerUseAPI.Drag(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComputerUseAPI.Drag``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `Drag`: MouseDragResponse
	fmt.Fprintf(os.Stdout, "Response from `ComputerUseAPI.Drag`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDragRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**MouseDragRequest**](MouseDragRequest.md) | Mouse drag request | 

### Return type

[**MouseDragResponse**](MouseDragResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetComputerUseStatus

> ComputerUseStatusResponse GetComputerUseStatus(ctx).Execute()

Get computer use process status



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
	resp, r, err := apiClient.ComputerUseAPI.GetComputerUseStatus(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComputerUseAPI.GetComputerUseStatus``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetComputerUseStatus`: ComputerUseStatusResponse
	fmt.Fprintf(os.Stdout, "Response from `ComputerUseAPI.GetComputerUseStatus`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetComputerUseStatusRequest struct via the builder pattern


### Return type

[**ComputerUseStatusResponse**](ComputerUseStatusResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetComputerUseSystemStatus

> ComputerUseStatusResponse GetComputerUseSystemStatus(ctx).Execute()

Get computer use status



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
	resp, r, err := apiClient.ComputerUseAPI.GetComputerUseSystemStatus(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComputerUseAPI.GetComputerUseSystemStatus``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetComputerUseSystemStatus`: ComputerUseStatusResponse
	fmt.Fprintf(os.Stdout, "Response from `ComputerUseAPI.GetComputerUseSystemStatus`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetComputerUseSystemStatusRequest struct via the builder pattern


### Return type

[**ComputerUseStatusResponse**](ComputerUseStatusResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDisplayInfo

> DisplayInfoResponse GetDisplayInfo(ctx).Execute()

Get display information



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
	resp, r, err := apiClient.ComputerUseAPI.GetDisplayInfo(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComputerUseAPI.GetDisplayInfo``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetDisplayInfo`: DisplayInfoResponse
	fmt.Fprintf(os.Stdout, "Response from `ComputerUseAPI.GetDisplayInfo`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetDisplayInfoRequest struct via the builder pattern


### Return type

[**DisplayInfoResponse**](DisplayInfoResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetMousePosition

> MousePositionResponse GetMousePosition(ctx).Execute()

Get mouse position



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
	resp, r, err := apiClient.ComputerUseAPI.GetMousePosition(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComputerUseAPI.GetMousePosition``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetMousePosition`: MousePositionResponse
	fmt.Fprintf(os.Stdout, "Response from `ComputerUseAPI.GetMousePosition`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetMousePositionRequest struct via the builder pattern


### Return type

[**MousePositionResponse**](MousePositionResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetProcessErrors

> ProcessErrorsResponse GetProcessErrors(ctx, processName).Execute()

Get process errors



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
	processName := "processName_example" // string | Process name to get errors for

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ComputerUseAPI.GetProcessErrors(context.Background(), processName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComputerUseAPI.GetProcessErrors``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetProcessErrors`: ProcessErrorsResponse
	fmt.Fprintf(os.Stdout, "Response from `ComputerUseAPI.GetProcessErrors`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**processName** | **string** | Process name to get errors for | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetProcessErrorsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ProcessErrorsResponse**](ProcessErrorsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetProcessLogs

> ProcessLogsResponse GetProcessLogs(ctx, processName).Execute()

Get process logs



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
	processName := "processName_example" // string | Process name to get logs for

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ComputerUseAPI.GetProcessLogs(context.Background(), processName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComputerUseAPI.GetProcessLogs``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetProcessLogs`: ProcessLogsResponse
	fmt.Fprintf(os.Stdout, "Response from `ComputerUseAPI.GetProcessLogs`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**processName** | **string** | Process name to get logs for | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetProcessLogsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ProcessLogsResponse**](ProcessLogsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetProcessStatus

> ProcessStatusResponse GetProcessStatus(ctx, processName).Execute()

Get specific process status



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
	processName := "processName_example" // string | Process name to check

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ComputerUseAPI.GetProcessStatus(context.Background(), processName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComputerUseAPI.GetProcessStatus``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetProcessStatus`: ProcessStatusResponse
	fmt.Fprintf(os.Stdout, "Response from `ComputerUseAPI.GetProcessStatus`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**processName** | **string** | Process name to check | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetProcessStatusRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ProcessStatusResponse**](ProcessStatusResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetWindows

> WindowsResponse GetWindows(ctx).Execute()

Get windows information



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
	resp, r, err := apiClient.ComputerUseAPI.GetWindows(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComputerUseAPI.GetWindows``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWindows`: WindowsResponse
	fmt.Fprintf(os.Stdout, "Response from `ComputerUseAPI.GetWindows`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetWindowsRequest struct via the builder pattern


### Return type

[**WindowsResponse**](WindowsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## MoveMouse

> MousePositionResponse MoveMouse(ctx).Request(request).Execute()

Move mouse cursor



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
	request := *openapiclient.NewMouseMoveRequest() // MouseMoveRequest | Mouse move request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ComputerUseAPI.MoveMouse(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComputerUseAPI.MoveMouse``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `MoveMouse`: MousePositionResponse
	fmt.Fprintf(os.Stdout, "Response from `ComputerUseAPI.MoveMouse`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiMoveMouseRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**MouseMoveRequest**](MouseMoveRequest.md) | Mouse move request | 

### Return type

[**MousePositionResponse**](MousePositionResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## PressHotkey

> map[string]interface{} PressHotkey(ctx).Request(request).Execute()

Press hotkey



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
	request := *openapiclient.NewKeyboardHotkeyRequest() // KeyboardHotkeyRequest | Hotkey press request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ComputerUseAPI.PressHotkey(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComputerUseAPI.PressHotkey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `PressHotkey`: map[string]interface{}
	fmt.Fprintf(os.Stdout, "Response from `ComputerUseAPI.PressHotkey`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiPressHotkeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**KeyboardHotkeyRequest**](KeyboardHotkeyRequest.md) | Hotkey press request | 

### Return type

**map[string]interface{}**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## PressKey

> map[string]interface{} PressKey(ctx).Request(request).Execute()

Press key



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
	request := *openapiclient.NewKeyboardPressRequest() // KeyboardPressRequest | Key press request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ComputerUseAPI.PressKey(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComputerUseAPI.PressKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `PressKey`: map[string]interface{}
	fmt.Fprintf(os.Stdout, "Response from `ComputerUseAPI.PressKey`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiPressKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**KeyboardPressRequest**](KeyboardPressRequest.md) | Key press request | 

### Return type

**map[string]interface{}**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RestartProcess

> ProcessRestartResponse RestartProcess(ctx, processName).Execute()

Restart specific process



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
	processName := "processName_example" // string | Process name to restart

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ComputerUseAPI.RestartProcess(context.Background(), processName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComputerUseAPI.RestartProcess``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `RestartProcess`: ProcessRestartResponse
	fmt.Fprintf(os.Stdout, "Response from `ComputerUseAPI.RestartProcess`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**processName** | **string** | Process name to restart | 

### Other Parameters

Other parameters are passed through a pointer to a apiRestartProcessRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ProcessRestartResponse**](ProcessRestartResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## Scroll

> ScrollResponse Scroll(ctx).Request(request).Execute()

Scroll mouse wheel



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
	request := *openapiclient.NewMouseScrollRequest() // MouseScrollRequest | Mouse scroll request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ComputerUseAPI.Scroll(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComputerUseAPI.Scroll``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `Scroll`: ScrollResponse
	fmt.Fprintf(os.Stdout, "Response from `ComputerUseAPI.Scroll`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiScrollRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**MouseScrollRequest**](MouseScrollRequest.md) | Mouse scroll request | 

### Return type

[**ScrollResponse**](ScrollResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## StartComputerUse

> ComputerUseStartResponse StartComputerUse(ctx).Execute()

Start computer use processes



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
	resp, r, err := apiClient.ComputerUseAPI.StartComputerUse(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComputerUseAPI.StartComputerUse``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `StartComputerUse`: ComputerUseStartResponse
	fmt.Fprintf(os.Stdout, "Response from `ComputerUseAPI.StartComputerUse`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiStartComputerUseRequest struct via the builder pattern


### Return type

[**ComputerUseStartResponse**](ComputerUseStartResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## StopComputerUse

> ComputerUseStopResponse StopComputerUse(ctx).Execute()

Stop computer use processes



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
	resp, r, err := apiClient.ComputerUseAPI.StopComputerUse(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComputerUseAPI.StopComputerUse``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `StopComputerUse`: ComputerUseStopResponse
	fmt.Fprintf(os.Stdout, "Response from `ComputerUseAPI.StopComputerUse`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiStopComputerUseRequest struct via the builder pattern


### Return type

[**ComputerUseStopResponse**](ComputerUseStopResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## TakeCompressedRegionScreenshot

> ScreenshotResponse TakeCompressedRegionScreenshot(ctx).X(x).Y(y).Width(width).Height(height).ShowCursor(showCursor).Format(format).Quality(quality).Scale(scale).Execute()

Take a compressed region screenshot



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
	x := int32(56) // int32 | X coordinate of the region
	y := int32(56) // int32 | Y coordinate of the region
	width := int32(56) // int32 | Width of the region
	height := int32(56) // int32 | Height of the region
	showCursor := true // bool | Whether to show cursor in screenshot (optional)
	format := "format_example" // string | Image format (png or jpeg) (optional)
	quality := int32(56) // int32 | JPEG quality (1-100) (optional)
	scale := float32(8.14) // float32 | Scale factor (0.1-1.0) (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ComputerUseAPI.TakeCompressedRegionScreenshot(context.Background()).X(x).Y(y).Width(width).Height(height).ShowCursor(showCursor).Format(format).Quality(quality).Scale(scale).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComputerUseAPI.TakeCompressedRegionScreenshot``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `TakeCompressedRegionScreenshot`: ScreenshotResponse
	fmt.Fprintf(os.Stdout, "Response from `ComputerUseAPI.TakeCompressedRegionScreenshot`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiTakeCompressedRegionScreenshotRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **x** | **int32** | X coordinate of the region | 
 **y** | **int32** | Y coordinate of the region | 
 **width** | **int32** | Width of the region | 
 **height** | **int32** | Height of the region | 
 **showCursor** | **bool** | Whether to show cursor in screenshot | 
 **format** | **string** | Image format (png or jpeg) | 
 **quality** | **int32** | JPEG quality (1-100) | 
 **scale** | **float32** | Scale factor (0.1-1.0) | 

### Return type

[**ScreenshotResponse**](ScreenshotResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## TakeCompressedScreenshot

> ScreenshotResponse TakeCompressedScreenshot(ctx).ShowCursor(showCursor).Format(format).Quality(quality).Scale(scale).Execute()

Take a compressed screenshot



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
	showCursor := true // bool | Whether to show cursor in screenshot (optional)
	format := "format_example" // string | Image format (png or jpeg) (optional)
	quality := int32(56) // int32 | JPEG quality (1-100) (optional)
	scale := float32(8.14) // float32 | Scale factor (0.1-1.0) (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ComputerUseAPI.TakeCompressedScreenshot(context.Background()).ShowCursor(showCursor).Format(format).Quality(quality).Scale(scale).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComputerUseAPI.TakeCompressedScreenshot``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `TakeCompressedScreenshot`: ScreenshotResponse
	fmt.Fprintf(os.Stdout, "Response from `ComputerUseAPI.TakeCompressedScreenshot`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiTakeCompressedScreenshotRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **showCursor** | **bool** | Whether to show cursor in screenshot | 
 **format** | **string** | Image format (png or jpeg) | 
 **quality** | **int32** | JPEG quality (1-100) | 
 **scale** | **float32** | Scale factor (0.1-1.0) | 

### Return type

[**ScreenshotResponse**](ScreenshotResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## TakeRegionScreenshot

> ScreenshotResponse TakeRegionScreenshot(ctx).X(x).Y(y).Width(width).Height(height).ShowCursor(showCursor).Execute()

Take a region screenshot



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
	x := int32(56) // int32 | X coordinate of the region
	y := int32(56) // int32 | Y coordinate of the region
	width := int32(56) // int32 | Width of the region
	height := int32(56) // int32 | Height of the region
	showCursor := true // bool | Whether to show cursor in screenshot (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ComputerUseAPI.TakeRegionScreenshot(context.Background()).X(x).Y(y).Width(width).Height(height).ShowCursor(showCursor).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComputerUseAPI.TakeRegionScreenshot``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `TakeRegionScreenshot`: ScreenshotResponse
	fmt.Fprintf(os.Stdout, "Response from `ComputerUseAPI.TakeRegionScreenshot`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiTakeRegionScreenshotRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **x** | **int32** | X coordinate of the region | 
 **y** | **int32** | Y coordinate of the region | 
 **width** | **int32** | Width of the region | 
 **height** | **int32** | Height of the region | 
 **showCursor** | **bool** | Whether to show cursor in screenshot | 

### Return type

[**ScreenshotResponse**](ScreenshotResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## TakeScreenshot

> ScreenshotResponse TakeScreenshot(ctx).ShowCursor(showCursor).Execute()

Take a screenshot



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
	showCursor := true // bool | Whether to show cursor in screenshot (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ComputerUseAPI.TakeScreenshot(context.Background()).ShowCursor(showCursor).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComputerUseAPI.TakeScreenshot``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `TakeScreenshot`: ScreenshotResponse
	fmt.Fprintf(os.Stdout, "Response from `ComputerUseAPI.TakeScreenshot`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiTakeScreenshotRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **showCursor** | **bool** | Whether to show cursor in screenshot | 

### Return type

[**ScreenshotResponse**](ScreenshotResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## TypeText

> map[string]interface{} TypeText(ctx).Request(request).Execute()

Type text



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
	request := *openapiclient.NewKeyboardTypeRequest() // KeyboardTypeRequest | Text typing request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ComputerUseAPI.TypeText(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComputerUseAPI.TypeText``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `TypeText`: map[string]interface{}
	fmt.Fprintf(os.Stdout, "Response from `ComputerUseAPI.TypeText`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiTypeTextRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**KeyboardTypeRequest**](KeyboardTypeRequest.md) | Text typing request | 

### Return type

**map[string]interface{}**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

