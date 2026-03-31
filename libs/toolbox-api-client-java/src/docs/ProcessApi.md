# ProcessApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**connectPtySession**](ProcessApi.md#connectPtySession) | **GET** /process/pty/{sessionId}/connect | Connect to PTY session via WebSocket |
| [**createPtySession**](ProcessApi.md#createPtySession) | **POST** /process/pty | Create a new PTY session |
| [**createSession**](ProcessApi.md#createSession) | **POST** /process/session | Create a new session |
| [**deletePtySession**](ProcessApi.md#deletePtySession) | **DELETE** /process/pty/{sessionId} | Delete a PTY session |
| [**deleteSession**](ProcessApi.md#deleteSession) | **DELETE** /process/session/{sessionId} | Delete a session |
| [**executeCommand**](ProcessApi.md#executeCommand) | **POST** /process/execute | Execute a command |
| [**getEntrypointLogs**](ProcessApi.md#getEntrypointLogs) | **GET** /process/session/entrypoint/logs | Get entrypoint logs |
| [**getEntrypointSession**](ProcessApi.md#getEntrypointSession) | **GET** /process/session/entrypoint | Get entrypoint session details |
| [**getPtySession**](ProcessApi.md#getPtySession) | **GET** /process/pty/{sessionId} | Get PTY session information |
| [**getSession**](ProcessApi.md#getSession) | **GET** /process/session/{sessionId} | Get session details |
| [**getSessionCommand**](ProcessApi.md#getSessionCommand) | **GET** /process/session/{sessionId}/command/{commandId} | Get session command details |
| [**getSessionCommandLogs**](ProcessApi.md#getSessionCommandLogs) | **GET** /process/session/{sessionId}/command/{commandId}/logs | Get session command logs |
| [**listPtySessions**](ProcessApi.md#listPtySessions) | **GET** /process/pty | List all PTY sessions |
| [**listSessions**](ProcessApi.md#listSessions) | **GET** /process/session | List all sessions |
| [**resizePtySession**](ProcessApi.md#resizePtySession) | **POST** /process/pty/{sessionId}/resize | Resize a PTY session |
| [**sendInput**](ProcessApi.md#sendInput) | **POST** /process/session/{sessionId}/command/{commandId}/input | Send input to command |
| [**sessionExecuteCommand**](ProcessApi.md#sessionExecuteCommand) | **POST** /process/session/{sessionId}/exec | Execute command in session |


<a id="connectPtySession"></a>
# **connectPtySession**
> connectPtySession(sessionId)

Connect to PTY session via WebSocket

Establish a WebSocket connection to interact with a pseudo-terminal session

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ProcessApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ProcessApi apiInstance = new ProcessApi(defaultClient);
    String sessionId = "sessionId_example"; // String | PTY session ID
    try {
      apiInstance.connectPtySession(sessionId);
    } catch (ApiException e) {
      System.err.println("Exception when calling ProcessApi#connectPtySession");
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
| **sessionId** | **String**| PTY session ID | |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **101** | Switching Protocols - WebSocket connection established |  -  |

<a id="createPtySession"></a>
# **createPtySession**
> PtyCreateResponse createPtySession(request)

Create a new PTY session

Create a new pseudo-terminal session with specified configuration

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ProcessApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ProcessApi apiInstance = new ProcessApi(defaultClient);
    PtyCreateRequest request = new PtyCreateRequest(); // PtyCreateRequest | PTY session creation request
    try {
      PtyCreateResponse result = apiInstance.createPtySession(request);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ProcessApi#createPtySession");
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
| **request** | [**PtyCreateRequest**](PtyCreateRequest.md)| PTY session creation request | |

### Return type

[**PtyCreateResponse**](PtyCreateResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** | Created |  -  |

<a id="createSession"></a>
# **createSession**
> createSession(request)

Create a new session

Create a new shell session for command execution

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ProcessApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ProcessApi apiInstance = new ProcessApi(defaultClient);
    CreateSessionRequest request = new CreateSessionRequest(); // CreateSessionRequest | Session creation request
    try {
      apiInstance.createSession(request);
    } catch (ApiException e) {
      System.err.println("Exception when calling ProcessApi#createSession");
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
| **request** | [**CreateSessionRequest**](CreateSessionRequest.md)| Session creation request | |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** | Created |  -  |

<a id="deletePtySession"></a>
# **deletePtySession**
> Map&lt;String, Object&gt; deletePtySession(sessionId)

Delete a PTY session

Delete a pseudo-terminal session and terminate its process

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ProcessApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ProcessApi apiInstance = new ProcessApi(defaultClient);
    String sessionId = "sessionId_example"; // String | PTY session ID
    try {
      Map<String, Object> result = apiInstance.deletePtySession(sessionId);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ProcessApi#deletePtySession");
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
| **sessionId** | **String**| PTY session ID | |

### Return type

**Map&lt;String, Object&gt;**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="deleteSession"></a>
# **deleteSession**
> deleteSession(sessionId)

Delete a session

Delete an existing shell session

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ProcessApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ProcessApi apiInstance = new ProcessApi(defaultClient);
    String sessionId = "sessionId_example"; // String | Session ID
    try {
      apiInstance.deleteSession(sessionId);
    } catch (ApiException e) {
      System.err.println("Exception when calling ProcessApi#deleteSession");
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
| **sessionId** | **String**| Session ID | |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | No Content |  -  |

<a id="executeCommand"></a>
# **executeCommand**
> ExecuteResponse executeCommand(request)

Execute a command

Execute a shell command and return the output and exit code

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ProcessApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ProcessApi apiInstance = new ProcessApi(defaultClient);
    ExecuteRequest request = new ExecuteRequest(); // ExecuteRequest | Command execution request
    try {
      ExecuteResponse result = apiInstance.executeCommand(request);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ProcessApi#executeCommand");
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
| **request** | [**ExecuteRequest**](ExecuteRequest.md)| Command execution request | |

### Return type

[**ExecuteResponse**](ExecuteResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="getEntrypointLogs"></a>
# **getEntrypointLogs**
> String getEntrypointLogs(follow)

Get entrypoint logs

Get logs for a sandbox entrypoint session. Supports both HTTP and WebSocket streaming.

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ProcessApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ProcessApi apiInstance = new ProcessApi(defaultClient);
    Boolean follow = true; // Boolean | Follow logs in real-time (WebSocket only)
    try {
      String result = apiInstance.getEntrypointLogs(follow);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ProcessApi#getEntrypointLogs");
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
| **follow** | **Boolean**| Follow logs in real-time (WebSocket only) | [optional] |

### Return type

**String**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: text/plain

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Entrypoint log content |  -  |

<a id="getEntrypointSession"></a>
# **getEntrypointSession**
> Session getEntrypointSession()

Get entrypoint session details

Get details of an entrypoint session including its commands

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ProcessApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ProcessApi apiInstance = new ProcessApi(defaultClient);
    try {
      Session result = apiInstance.getEntrypointSession();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ProcessApi#getEntrypointSession");
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

[**Session**](Session.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="getPtySession"></a>
# **getPtySession**
> PtySessionInfo getPtySession(sessionId)

Get PTY session information

Get detailed information about a specific pseudo-terminal session

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ProcessApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ProcessApi apiInstance = new ProcessApi(defaultClient);
    String sessionId = "sessionId_example"; // String | PTY session ID
    try {
      PtySessionInfo result = apiInstance.getPtySession(sessionId);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ProcessApi#getPtySession");
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
| **sessionId** | **String**| PTY session ID | |

### Return type

[**PtySessionInfo**](PtySessionInfo.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="getSession"></a>
# **getSession**
> Session getSession(sessionId)

Get session details

Get details of a specific session including its commands

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ProcessApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ProcessApi apiInstance = new ProcessApi(defaultClient);
    String sessionId = "sessionId_example"; // String | Session ID
    try {
      Session result = apiInstance.getSession(sessionId);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ProcessApi#getSession");
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
| **sessionId** | **String**| Session ID | |

### Return type

[**Session**](Session.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="getSessionCommand"></a>
# **getSessionCommand**
> Command getSessionCommand(sessionId, commandId)

Get session command details

Get details of a specific command within a session

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ProcessApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ProcessApi apiInstance = new ProcessApi(defaultClient);
    String sessionId = "sessionId_example"; // String | Session ID
    String commandId = "commandId_example"; // String | Command ID
    try {
      Command result = apiInstance.getSessionCommand(sessionId, commandId);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ProcessApi#getSessionCommand");
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
| **sessionId** | **String**| Session ID | |
| **commandId** | **String**| Command ID | |

### Return type

[**Command**](Command.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="getSessionCommandLogs"></a>
# **getSessionCommandLogs**
> String getSessionCommandLogs(sessionId, commandId, follow)

Get session command logs

Get logs for a specific command within a session. Supports both HTTP and WebSocket streaming.

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ProcessApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ProcessApi apiInstance = new ProcessApi(defaultClient);
    String sessionId = "sessionId_example"; // String | Session ID
    String commandId = "commandId_example"; // String | Command ID
    Boolean follow = true; // Boolean | Follow logs in real-time (WebSocket only)
    try {
      String result = apiInstance.getSessionCommandLogs(sessionId, commandId, follow);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ProcessApi#getSessionCommandLogs");
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
| **sessionId** | **String**| Session ID | |
| **commandId** | **String**| Command ID | |
| **follow** | **Boolean**| Follow logs in real-time (WebSocket only) | [optional] |

### Return type

**String**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: text/plain

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Log content |  -  |

<a id="listPtySessions"></a>
# **listPtySessions**
> PtyListResponse listPtySessions()

List all PTY sessions

Get a list of all active pseudo-terminal sessions

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ProcessApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ProcessApi apiInstance = new ProcessApi(defaultClient);
    try {
      PtyListResponse result = apiInstance.listPtySessions();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ProcessApi#listPtySessions");
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

[**PtyListResponse**](PtyListResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="listSessions"></a>
# **listSessions**
> List&lt;Session&gt; listSessions()

List all sessions

Get a list of all active shell sessions

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ProcessApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ProcessApi apiInstance = new ProcessApi(defaultClient);
    try {
      List<Session> result = apiInstance.listSessions();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ProcessApi#listSessions");
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

[**List&lt;Session&gt;**](Session.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="resizePtySession"></a>
# **resizePtySession**
> PtySessionInfo resizePtySession(sessionId, request)

Resize a PTY session

Resize the terminal dimensions of a pseudo-terminal session

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ProcessApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ProcessApi apiInstance = new ProcessApi(defaultClient);
    String sessionId = "sessionId_example"; // String | PTY session ID
    PtyResizeRequest request = new PtyResizeRequest(); // PtyResizeRequest | Resize request with new dimensions
    try {
      PtySessionInfo result = apiInstance.resizePtySession(sessionId, request);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ProcessApi#resizePtySession");
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
| **sessionId** | **String**| PTY session ID | |
| **request** | [**PtyResizeRequest**](PtyResizeRequest.md)| Resize request with new dimensions | |

### Return type

[**PtySessionInfo**](PtySessionInfo.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |

<a id="sendInput"></a>
# **sendInput**
> sendInput(sessionId, commandId, request)

Send input to command

Send input data to a running command in a session for interactive execution

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ProcessApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ProcessApi apiInstance = new ProcessApi(defaultClient);
    String sessionId = "sessionId_example"; // String | Session ID
    String commandId = "commandId_example"; // String | Command ID
    SessionSendInputRequest request = new SessionSendInputRequest(); // SessionSendInputRequest | Input send request
    try {
      apiInstance.sendInput(sessionId, commandId, request);
    } catch (ApiException e) {
      System.err.println("Exception when calling ProcessApi#sendInput");
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
| **sessionId** | **String**| Session ID | |
| **commandId** | **String**| Command ID | |
| **request** | [**SessionSendInputRequest**](SessionSendInputRequest.md)| Input send request | |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | No Content |  -  |

<a id="sessionExecuteCommand"></a>
# **sessionExecuteCommand**
> SessionExecuteResponse sessionExecuteCommand(sessionId, request)

Execute command in session

Execute a command within an existing shell session

### Example
```java
// Import classes:
import io.daytona.toolbox.client.ApiClient;
import io.daytona.toolbox.client.ApiException;
import io.daytona.toolbox.client.Configuration;
import io.daytona.toolbox.client.models.*;
import io.daytona.toolbox.client.api.ProcessApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost");

    ProcessApi apiInstance = new ProcessApi(defaultClient);
    String sessionId = "sessionId_example"; // String | Session ID
    SessionExecuteRequest request = new SessionExecuteRequest(); // SessionExecuteRequest | Command execution request
    try {
      SessionExecuteResponse result = apiInstance.sessionExecuteCommand(sessionId, request);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ProcessApi#sessionExecuteCommand");
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
| **sessionId** | **String**| Session ID | |
| **request** | [**SessionExecuteRequest**](SessionExecuteRequest.md)| Command execution request | |

### Return type

[**SessionExecuteResponse**](SessionExecuteResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **202** | Accepted |  -  |

