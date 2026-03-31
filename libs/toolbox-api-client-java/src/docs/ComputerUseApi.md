# ComputerUseApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**click**](ComputerUseApi.md#click) | **POST** /computeruse/mouse/click | Click mouse button |
| [**deleteRecording**](ComputerUseApi.md#deleteRecording) | **DELETE** /computeruse/recordings/{id} | Delete a recording |
| [**downloadRecording**](ComputerUseApi.md#downloadRecording) | **GET** /computeruse/recordings/{id}/download | Download a recording |
| [**drag**](ComputerUseApi.md#drag) | **POST** /computeruse/mouse/drag | Drag mouse |
| [**getComputerUseStatus**](ComputerUseApi.md#getComputerUseStatus) | **GET** /computeruse/process-status | Get computer use process status |
| [**getComputerUseSystemStatus**](ComputerUseApi.md#getComputerUseSystemStatus) | **GET** /computeruse/status | Get computer use status |
| [**getDisplayInfo**](ComputerUseApi.md#getDisplayInfo) | **GET** /computeruse/display/info | Get display information |
| [**getMousePosition**](ComputerUseApi.md#getMousePosition) | **GET** /computeruse/mouse/position | Get mouse position |
| [**getProcessErrors**](ComputerUseApi.md#getProcessErrors) | **GET** /computeruse/process/{processName}/errors | Get process errors |
| [**getProcessLogs**](ComputerUseApi.md#getProcessLogs) | **GET** /computeruse/process/{processName}/logs | Get process logs |
| [**getProcessStatus**](ComputerUseApi.md#getProcessStatus) | **GET** /computeruse/process/{processName}/status | Get specific process status |
| [**getRecording**](ComputerUseApi.md#getRecording) | **GET** /computeruse/recordings/{id} | Get recording details |
| [**getWindows**](ComputerUseApi.md#getWindows) | **GET** /computeruse/display/windows | Get windows information |
| [**listRecordings**](ComputerUseApi.md#listRecordings) | **GET** /computeruse/recordings | List all recordings |
| [**moveMouse**](ComputerUseApi.md#moveMouse) | **POST** /computeruse/mouse/move | Move mouse cursor |
| [**pressHotkey**](ComputerUseApi.md#pressHotkey) | **POST** /computeruse/keyboard/hotkey | Press hotkey |
| [**pressKey**](ComputerUseApi.md#pressKey) | **POST** /computeruse/keyboard/key | Press key |
| [**restartProcess**](ComputerUseApi.md#restartProcess) | **POST** /computeruse/process/{processName}/restart | Restart specific process |
| [**scroll**](ComputerUseApi.md#scroll) | **POST** /computeruse/mouse/scroll | Scroll mouse wheel |
| [**startComputerUse**](ComputerUseApi.md#startComputerUse) | **POST** /computeruse/start | Start computer use processes |
| [**startRecording**](ComputerUseApi.md#startRecording) | **POST** /computeruse/recordings/start | Start a new recording |
| [**stopComputerUse**](ComputerUseApi.md#stopComputerUse) | **POST** /computeruse/stop | Stop computer use processes |
| [**stopRecording**](ComputerUseApi.md#stopRecording) | **POST** /computeruse/recordings/stop | Stop a recording |
| [**takeCompressedRegionScreenshot**](ComputerUseApi.md#takeCompressedRegionScreenshot) | **GET** /computeruse/screenshot/region/compressed | Take a compressed region screenshot |
| [**takeCompressedScreenshot**](ComputerUseApi.md#takeCompressedScreenshot) | **GET** /computeruse/screenshot/compressed | Take a compressed screenshot |
| [**takeRegionScreenshot**](ComputerUseApi.md#takeRegionScreenshot) | **GET** /computeruse/screenshot/region | Take a region screenshot |
| [**takeScreenshot**](ComputerUseApi.md#takeScreenshot) | **GET** /computeruse/screenshot | Take a screenshot |
| [**typeText**](ComputerUseApi.md#typeText) | **POST** /computeruse/keyboard/type | Type text |


<a id="click"></a>
# **click**
> MouseClickResponse click(request)

Click mouse button

Click the mouse button at the specified coordinates

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ComputerUseApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ComputerUseApi apiInstance = new ComputerUseApi(defaultClient);
    MouseClickRequest request = new MouseClickRequest(); // MouseClickRequest | Mouse click request
    try {
      MouseClickResponse result = apiInstance.click(request);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ComputerUseApi#click");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **request** | [**MouseClickRequest**](MouseClickRequest.md)| Mouse click request | |

### Return type

[**MouseClickResponse**](MouseClickResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="deleteRecording"></a>
# **deleteRecording**
> deleteRecording(id)

Delete a recording

Delete a recording file by ID

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ComputerUseApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ComputerUseApi apiInstance = new ComputerUseApi(defaultClient);
    String id = "id_example"; // String | Recording ID
    try {
      apiInstance.deleteRecording(id);
    } catch (ApiException e) {
      System.err.println("Exception when calling ComputerUseApi#deleteRecording");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | **String**| Recording ID | |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: */*

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | No Content |  -  |
| **400** | Bad Request |  -  |
| **404** | Not Found |  -  |
| **500** | Internal Server Error |  -  |

<a id="downloadRecording"></a>
# **downloadRecording**
> File downloadRecording(id)

Download a recording

Download a recording by providing its ID

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ComputerUseApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ComputerUseApi apiInstance = new ComputerUseApi(defaultClient);
    String id = "id_example"; // String | Recording ID
    try {
      File result = apiInstance.downloadRecording(id);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ComputerUseApi#downloadRecording");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | **String**| Recording ID | |

### Return type

[**File**](File.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/octet-stream

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **404** | Not Found |  -  |
| **500** | Internal Server Error |  -  |

<a id="drag"></a>
# **drag**
> MouseDragResponse drag(request)

Drag mouse

Drag the mouse from start to end coordinates

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ComputerUseApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ComputerUseApi apiInstance = new ComputerUseApi(defaultClient);
    MouseDragRequest request = new MouseDragRequest(); // MouseDragRequest | Mouse drag request
    try {
      MouseDragResponse result = apiInstance.drag(request);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ComputerUseApi#drag");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **request** | [**MouseDragRequest**](MouseDragRequest.md)| Mouse drag request | |

### Return type

[**MouseDragResponse**](MouseDragResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="getComputerUseStatus"></a>
# **getComputerUseStatus**
> ComputerUseStatusResponse getComputerUseStatus()

Get computer use process status

Get the status of all computer use processes

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ComputerUseApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ComputerUseApi apiInstance = new ComputerUseApi(defaultClient);
    try {
      ComputerUseStatusResponse result = apiInstance.getComputerUseStatus();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ComputerUseApi#getComputerUseStatus");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**ComputerUseStatusResponse**](ComputerUseStatusResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="getComputerUseSystemStatus"></a>
# **getComputerUseSystemStatus**
> ComputerUseStatusResponse getComputerUseSystemStatus()

Get computer use status

Get the current status of the computer use system

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ComputerUseApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ComputerUseApi apiInstance = new ComputerUseApi(defaultClient);
    try {
      ComputerUseStatusResponse result = apiInstance.getComputerUseSystemStatus();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ComputerUseApi#getComputerUseSystemStatus");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**ComputerUseStatusResponse**](ComputerUseStatusResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="getDisplayInfo"></a>
# **getDisplayInfo**
> DisplayInfoResponse getDisplayInfo()

Get display information

Get information about all available displays

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ComputerUseApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ComputerUseApi apiInstance = new ComputerUseApi(defaultClient);
    try {
      DisplayInfoResponse result = apiInstance.getDisplayInfo();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ComputerUseApi#getDisplayInfo");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**DisplayInfoResponse**](DisplayInfoResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="getMousePosition"></a>
# **getMousePosition**
> MousePositionResponse getMousePosition()

Get mouse position

Get the current mouse cursor position

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ComputerUseApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ComputerUseApi apiInstance = new ComputerUseApi(defaultClient);
    try {
      MousePositionResponse result = apiInstance.getMousePosition();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ComputerUseApi#getMousePosition");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**MousePositionResponse**](MousePositionResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="getProcessErrors"></a>
# **getProcessErrors**
> ProcessErrorsResponse getProcessErrors(processName)

Get process errors

Get errors for a specific computer use process

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ComputerUseApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ComputerUseApi apiInstance = new ComputerUseApi(defaultClient);
    String processName = "processName_example"; // String | Process name to get errors for
    try {
      ProcessErrorsResponse result = apiInstance.getProcessErrors(processName);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ComputerUseApi#getProcessErrors");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **processName** | **String**| Process name to get errors for | |

### Return type

[**ProcessErrorsResponse**](ProcessErrorsResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="getProcessLogs"></a>
# **getProcessLogs**
> ProcessLogsResponse getProcessLogs(processName)

Get process logs

Get logs for a specific computer use process

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ComputerUseApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ComputerUseApi apiInstance = new ComputerUseApi(defaultClient);
    String processName = "processName_example"; // String | Process name to get logs for
    try {
      ProcessLogsResponse result = apiInstance.getProcessLogs(processName);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ComputerUseApi#getProcessLogs");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **processName** | **String**| Process name to get logs for | |

### Return type

[**ProcessLogsResponse**](ProcessLogsResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="getProcessStatus"></a>
# **getProcessStatus**
> ProcessStatusResponse getProcessStatus(processName)

Get specific process status

Check if a specific computer use process is running

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ComputerUseApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ComputerUseApi apiInstance = new ComputerUseApi(defaultClient);
    String processName = "processName_example"; // String | Process name to check
    try {
      ProcessStatusResponse result = apiInstance.getProcessStatus(processName);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ComputerUseApi#getProcessStatus");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **processName** | **String**| Process name to check | |

### Return type

[**ProcessStatusResponse**](ProcessStatusResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="getRecording"></a>
# **getRecording**
> Recording getRecording(id)

Get recording details

Get details of a specific recording by ID

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ComputerUseApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ComputerUseApi apiInstance = new ComputerUseApi(defaultClient);
    String id = "id_example"; // String | Recording ID
    try {
      Recording result = apiInstance.getRecording(id);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ComputerUseApi#getRecording");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | **String**| Recording ID | |

### Return type

[**Recording**](Recording.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **404** | Not Found |  -  |
| **500** | Internal Server Error |  -  |

<a id="getWindows"></a>
# **getWindows**
> WindowsResponse getWindows()

Get windows information

Get information about all open windows

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ComputerUseApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ComputerUseApi apiInstance = new ComputerUseApi(defaultClient);
    try {
      WindowsResponse result = apiInstance.getWindows();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ComputerUseApi#getWindows");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**WindowsResponse**](WindowsResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="listRecordings"></a>
# **listRecordings**
> ListRecordingsResponse listRecordings()

List all recordings

Get a list of all recordings (active and completed)

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ComputerUseApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ComputerUseApi apiInstance = new ComputerUseApi(defaultClient);
    try {
      ListRecordingsResponse result = apiInstance.listRecordings();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ComputerUseApi#listRecordings");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**ListRecordingsResponse**](ListRecordingsResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **500** | Internal Server Error |  -  |

<a id="moveMouse"></a>
# **moveMouse**
> MousePositionResponse moveMouse(request)

Move mouse cursor

Move the mouse cursor to the specified coordinates

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ComputerUseApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ComputerUseApi apiInstance = new ComputerUseApi(defaultClient);
    MouseMoveRequest request = new MouseMoveRequest(); // MouseMoveRequest | Mouse move request
    try {
      MousePositionResponse result = apiInstance.moveMouse(request);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ComputerUseApi#moveMouse");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **request** | [**MouseMoveRequest**](MouseMoveRequest.md)| Mouse move request | |

### Return type

[**MousePositionResponse**](MousePositionResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="pressHotkey"></a>
# **pressHotkey**
> Object pressHotkey(request)

Press hotkey

Press a hotkey combination (e.g., ctrl+c, cmd+v)

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ComputerUseApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ComputerUseApi apiInstance = new ComputerUseApi(defaultClient);
    KeyboardHotkeyRequest request = new KeyboardHotkeyRequest(); // KeyboardHotkeyRequest | Hotkey press request
    try {
      Object result = apiInstance.pressHotkey(request);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ComputerUseApi#pressHotkey");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **request** | [**KeyboardHotkeyRequest**](KeyboardHotkeyRequest.md)| Hotkey press request | |

### Return type

**Object**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="pressKey"></a>
# **pressKey**
> Object pressKey(request)

Press key

Press a key with optional modifiers

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ComputerUseApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ComputerUseApi apiInstance = new ComputerUseApi(defaultClient);
    KeyboardPressRequest request = new KeyboardPressRequest(); // KeyboardPressRequest | Key press request
    try {
      Object result = apiInstance.pressKey(request);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ComputerUseApi#pressKey");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **request** | [**KeyboardPressRequest**](KeyboardPressRequest.md)| Key press request | |

### Return type

**Object**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="restartProcess"></a>
# **restartProcess**
> ProcessRestartResponse restartProcess(processName)

Restart specific process

Restart a specific computer use process

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ComputerUseApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ComputerUseApi apiInstance = new ComputerUseApi(defaultClient);
    String processName = "processName_example"; // String | Process name to restart
    try {
      ProcessRestartResponse result = apiInstance.restartProcess(processName);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ComputerUseApi#restartProcess");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **processName** | **String**| Process name to restart | |

### Return type

[**ProcessRestartResponse**](ProcessRestartResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="scroll"></a>
# **scroll**
> ScrollResponse scroll(request)

Scroll mouse wheel

Scroll the mouse wheel at the specified coordinates

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ComputerUseApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ComputerUseApi apiInstance = new ComputerUseApi(defaultClient);
    MouseScrollRequest request = new MouseScrollRequest(); // MouseScrollRequest | Mouse scroll request
    try {
      ScrollResponse result = apiInstance.scroll(request);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ComputerUseApi#scroll");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **request** | [**MouseScrollRequest**](MouseScrollRequest.md)| Mouse scroll request | |

### Return type

[**ScrollResponse**](ScrollResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="startComputerUse"></a>
# **startComputerUse**
> ComputerUseStartResponse startComputerUse()

Start computer use processes

Start all computer use processes and return their status

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ComputerUseApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ComputerUseApi apiInstance = new ComputerUseApi(defaultClient);
    try {
      ComputerUseStartResponse result = apiInstance.startComputerUse();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ComputerUseApi#startComputerUse");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**ComputerUseStartResponse**](ComputerUseStartResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="startRecording"></a>
# **startRecording**
> Recording startRecording(request)

Start a new recording

Start a new screen recording session

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ComputerUseApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ComputerUseApi apiInstance = new ComputerUseApi(defaultClient);
    StartRecordingRequest request = new StartRecordingRequest(); // StartRecordingRequest | Recording options
    try {
      Recording result = apiInstance.startRecording(request);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ComputerUseApi#startRecording");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **request** | [**StartRecordingRequest**](StartRecordingRequest.md)| Recording options | [optional] |

### Return type

[**Recording**](Recording.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** | Created |  -  |
| **400** | Bad Request |  -  |
| **500** | Internal Server Error |  -  |

<a id="stopComputerUse"></a>
# **stopComputerUse**
> ComputerUseStopResponse stopComputerUse()

Stop computer use processes

Stop all computer use processes and return their status

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ComputerUseApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ComputerUseApi apiInstance = new ComputerUseApi(defaultClient);
    try {
      ComputerUseStopResponse result = apiInstance.stopComputerUse();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ComputerUseApi#stopComputerUse");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**ComputerUseStopResponse**](ComputerUseStopResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="stopRecording"></a>
# **stopRecording**
> Recording stopRecording(request)

Stop a recording

Stop an active screen recording session

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ComputerUseApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ComputerUseApi apiInstance = new ComputerUseApi(defaultClient);
    StopRecordingRequest request = new StopRecordingRequest(); // StopRecordingRequest | Recording ID to stop
    try {
      Recording result = apiInstance.stopRecording(request);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ComputerUseApi#stopRecording");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **request** | [**StopRecordingRequest**](StopRecordingRequest.md)| Recording ID to stop | |

### Return type

[**Recording**](Recording.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **400** | Bad Request |  -  |
| **404** | Not Found |  -  |

<a id="takeCompressedRegionScreenshot"></a>
# **takeCompressedRegionScreenshot**
> ScreenshotResponse takeCompressedRegionScreenshot(x, y, width, height, showCursor, format, quality, scale)

Take a compressed region screenshot

Take a compressed screenshot of a specific region of the screen

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ComputerUseApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ComputerUseApi apiInstance = new ComputerUseApi(defaultClient);
    Integer x = 56; // Integer | X coordinate of the region
    Integer y = 56; // Integer | Y coordinate of the region
    Integer width = 56; // Integer | Width of the region
    Integer height = 56; // Integer | Height of the region
    Boolean showCursor = true; // Boolean | Whether to show cursor in screenshot
    String format = "format_example"; // String | Image format (png or jpeg)
    Integer quality = 56; // Integer | JPEG quality (1-100)
    BigDecimal scale = new BigDecimal(78); // BigDecimal | Scale factor (0.1-1.0)
    try {
      ScreenshotResponse result = apiInstance.takeCompressedRegionScreenshot(x, y, width, height, showCursor, format, quality, scale);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ComputerUseApi#takeCompressedRegionScreenshot");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **x** | **Integer**| X coordinate of the region | |
| **y** | **Integer**| Y coordinate of the region | |
| **width** | **Integer**| Width of the region | |
| **height** | **Integer**| Height of the region | |
| **showCursor** | **Boolean**| Whether to show cursor in screenshot | [optional] |
| **format** | **String**| Image format (png or jpeg) | [optional] |
| **quality** | **Integer**| JPEG quality (1-100) | [optional] |
| **scale** | **BigDecimal**| Scale factor (0.1-1.0) | [optional] |

### Return type

[**ScreenshotResponse**](ScreenshotResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="takeCompressedScreenshot"></a>
# **takeCompressedScreenshot**
> ScreenshotResponse takeCompressedScreenshot(showCursor, format, quality, scale)

Take a compressed screenshot

Take a compressed screenshot of the entire screen

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ComputerUseApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ComputerUseApi apiInstance = new ComputerUseApi(defaultClient);
    Boolean showCursor = true; // Boolean | Whether to show cursor in screenshot
    String format = "format_example"; // String | Image format (png or jpeg)
    Integer quality = 56; // Integer | JPEG quality (1-100)
    BigDecimal scale = new BigDecimal(78); // BigDecimal | Scale factor (0.1-1.0)
    try {
      ScreenshotResponse result = apiInstance.takeCompressedScreenshot(showCursor, format, quality, scale);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ComputerUseApi#takeCompressedScreenshot");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **showCursor** | **Boolean**| Whether to show cursor in screenshot | [optional] |
| **format** | **String**| Image format (png or jpeg) | [optional] |
| **quality** | **Integer**| JPEG quality (1-100) | [optional] |
| **scale** | **BigDecimal**| Scale factor (0.1-1.0) | [optional] |

### Return type

[**ScreenshotResponse**](ScreenshotResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="takeRegionScreenshot"></a>
# **takeRegionScreenshot**
> ScreenshotResponse takeRegionScreenshot(x, y, width, height, showCursor)

Take a region screenshot

Take a screenshot of a specific region of the screen

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ComputerUseApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ComputerUseApi apiInstance = new ComputerUseApi(defaultClient);
    Integer x = 56; // Integer | X coordinate of the region
    Integer y = 56; // Integer | Y coordinate of the region
    Integer width = 56; // Integer | Width of the region
    Integer height = 56; // Integer | Height of the region
    Boolean showCursor = true; // Boolean | Whether to show cursor in screenshot
    try {
      ScreenshotResponse result = apiInstance.takeRegionScreenshot(x, y, width, height, showCursor);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ComputerUseApi#takeRegionScreenshot");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **x** | **Integer**| X coordinate of the region | |
| **y** | **Integer**| Y coordinate of the region | |
| **width** | **Integer**| Width of the region | |
| **height** | **Integer**| Height of the region | |
| **showCursor** | **Boolean**| Whether to show cursor in screenshot | [optional] |

### Return type

[**ScreenshotResponse**](ScreenshotResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="takeScreenshot"></a>
# **takeScreenshot**
> ScreenshotResponse takeScreenshot(showCursor)

Take a screenshot

Take a screenshot of the entire screen

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ComputerUseApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ComputerUseApi apiInstance = new ComputerUseApi(defaultClient);
    Boolean showCursor = true; // Boolean | Whether to show cursor in screenshot
    try {
      ScreenshotResponse result = apiInstance.takeScreenshot(showCursor);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ComputerUseApi#takeScreenshot");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **showCursor** | **Boolean**| Whether to show cursor in screenshot | [optional] |

### Return type

[**ScreenshotResponse**](ScreenshotResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="typeText"></a>
# **typeText**
> Object typeText(request)

Type text

Type text with optional delay between keystrokes

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ComputerUseApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ComputerUseApi apiInstance = new ComputerUseApi(defaultClient);
    KeyboardTypeRequest request = new KeyboardTypeRequest(); // KeyboardTypeRequest | Text typing request
    try {
      Object result = apiInstance.typeText(request);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ComputerUseApi#typeText");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **request** | [**KeyboardTypeRequest**](KeyboardTypeRequest.md)| Text typing request | |

### Return type

**Object**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

