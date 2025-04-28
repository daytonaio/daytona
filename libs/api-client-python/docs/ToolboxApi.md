# daytona_api_client.ToolboxApi

All URIs are relative to _http://localhost_

| Method                                                                 | HTTP request                                                                                | Description                      |
| ---------------------------------------------------------------------- | ------------------------------------------------------------------------------------------- | -------------------------------- |
| [**create_folder**](ToolboxApi.md#create_folder)                       | **POST** /toolbox/{workspaceId}/toolbox/files/folder                                        | Create folder                    |
| [**create_session**](ToolboxApi.md#create_session)                     | **POST** /toolbox/{workspaceId}/toolbox/process/session                                     | Create session                   |
| [**delete_file**](ToolboxApi.md#delete_file)                           | **DELETE** /toolbox/{workspaceId}/toolbox/files                                             | Delete file                      |
| [**delete_session**](ToolboxApi.md#delete_session)                     | **DELETE** /toolbox/{workspaceId}/toolbox/process/session/{sessionId}                       | Delete session                   |
| [**download_file**](ToolboxApi.md#download_file)                       | **GET** /toolbox/{workspaceId}/toolbox/files/download                                       | Download file                    |
| [**execute_command**](ToolboxApi.md#execute_command)                   | **POST** /toolbox/{workspaceId}/toolbox/process/execute                                     | Execute command                  |
| [**execute_session_command**](ToolboxApi.md#execute_session_command)   | **POST** /toolbox/{workspaceId}/toolbox/process/session/{sessionId}/exec                    | Execute command in session       |
| [**find_in_files**](ToolboxApi.md#find_in_files)                       | **GET** /toolbox/{workspaceId}/toolbox/files/find                                           | Search for text/pattern in files |
| [**get_file_info**](ToolboxApi.md#get_file_info)                       | **GET** /toolbox/{workspaceId}/toolbox/files/info                                           | Get file info                    |
| [**get_project_dir**](ToolboxApi.md#get_project_dir)                   | **GET** /toolbox/{workspaceId}/toolbox/project-dir                                          | Get workspace project dir        |
| [**get_session**](ToolboxApi.md#get_session)                           | **GET** /toolbox/{workspaceId}/toolbox/process/session/{sessionId}                          | Get session                      |
| [**get_session_command**](ToolboxApi.md#get_session_command)           | **GET** /toolbox/{workspaceId}/toolbox/process/session/{sessionId}/command/{commandId}      | Get session command              |
| [**get_session_command_logs**](ToolboxApi.md#get_session_command_logs) | **GET** /toolbox/{workspaceId}/toolbox/process/session/{sessionId}/command/{commandId}/logs | Get command logs                 |
| [**git_add_files**](ToolboxApi.md#git_add_files)                       | **POST** /toolbox/{workspaceId}/toolbox/git/add                                             | Add files                        |
| [**git_clone_repository**](ToolboxApi.md#git_clone_repository)         | **POST** /toolbox/{workspaceId}/toolbox/git/clone                                           | Clone repository                 |
| [**git_commit_changes**](ToolboxApi.md#git_commit_changes)             | **POST** /toolbox/{workspaceId}/toolbox/git/commit                                          | Commit changes                   |
| [**git_create_branch**](ToolboxApi.md#git_create_branch)               | **POST** /toolbox/{workspaceId}/toolbox/git/branches                                        | Create branch                    |
| [**git_get_history**](ToolboxApi.md#git_get_history)                   | **GET** /toolbox/{workspaceId}/toolbox/git/history                                          | Get commit history               |
| [**git_get_status**](ToolboxApi.md#git_get_status)                     | **GET** /toolbox/{workspaceId}/toolbox/git/status                                           | Get git status                   |
| [**git_list_branches**](ToolboxApi.md#git_list_branches)               | **GET** /toolbox/{workspaceId}/toolbox/git/branches                                         | Get branch list                  |
| [**git_pull_changes**](ToolboxApi.md#git_pull_changes)                 | **POST** /toolbox/{workspaceId}/toolbox/git/pull                                            | Pull changes                     |
| [**git_push_changes**](ToolboxApi.md#git_push_changes)                 | **POST** /toolbox/{workspaceId}/toolbox/git/push                                            | Push changes                     |
| [**list_files**](ToolboxApi.md#list_files)                             | **GET** /toolbox/{workspaceId}/toolbox/files                                                | List files                       |
| [**list_sessions**](ToolboxApi.md#list_sessions)                       | **GET** /toolbox/{workspaceId}/toolbox/process/session                                      | List sessions                    |
| [**lsp_completions**](ToolboxApi.md#lsp_completions)                   | **POST** /toolbox/{workspaceId}/toolbox/lsp/completions                                     | Get Lsp Completions              |
| [**lsp_did_close**](ToolboxApi.md#lsp_did_close)                       | **POST** /toolbox/{workspaceId}/toolbox/lsp/did-close                                       | Call Lsp DidClose                |
| [**lsp_did_open**](ToolboxApi.md#lsp_did_open)                         | **POST** /toolbox/{workspaceId}/toolbox/lsp/did-open                                        | Call Lsp DidOpen                 |
| [**lsp_document_symbols**](ToolboxApi.md#lsp_document_symbols)         | **GET** /toolbox/{workspaceId}/toolbox/lsp/document-symbols                                 | Call Lsp DocumentSymbols         |
| [**lsp_start**](ToolboxApi.md#lsp_start)                               | **POST** /toolbox/{workspaceId}/toolbox/lsp/start                                           | Start Lsp server                 |
| [**lsp_stop**](ToolboxApi.md#lsp_stop)                                 | **POST** /toolbox/{workspaceId}/toolbox/lsp/stop                                            | Stop Lsp server                  |
| [**lsp_workspace_symbols**](ToolboxApi.md#lsp_workspace_symbols)       | **GET** /toolbox/{workspaceId}/toolbox/lsp/workspace-symbols                                | Call Lsp WorkspaceSymbols        |
| [**move_file**](ToolboxApi.md#move_file)                               | **POST** /toolbox/{workspaceId}/toolbox/files/move                                          | Move file                        |
| [**replace_in_files**](ToolboxApi.md#replace_in_files)                 | **POST** /toolbox/{workspaceId}/toolbox/files/replace                                       | Replace in files                 |
| [**search_files**](ToolboxApi.md#search_files)                         | **GET** /toolbox/{workspaceId}/toolbox/files/search                                         | Search files                     |
| [**set_file_permissions**](ToolboxApi.md#set_file_permissions)         | **POST** /toolbox/{workspaceId}/toolbox/files/permissions                                   | Set file permissions             |
| [**upload_file**](ToolboxApi.md#upload_file)                           | **POST** /toolbox/{workspaceId}/toolbox/files/upload                                        | Upload file                      |

# **create_folder**

> create_folder(workspace_id, path, mode, x_daytona_organization_id=x_daytona_organization_id)

Create folder

Create folder inside workspace

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    path = 'path_example' # str |
    mode = 'mode_example' # str |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Create folder
        api_instance.create_folder(workspace_id, path, mode, x_daytona_organization_id=x_daytona_organization_id)
    except Exception as e:
        print("Exception when calling ToolboxApi->create_folder: %s\n" % e)
```

### Parameters

| Name                          | Type    | Description                                 | Notes      |
| ----------------------------- | ------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str** |                                             |
| **path**                      | **str** |                                             |
| **mode**                      | **str** |                                             |
| **x_daytona_organization_id** | **str** | Use with JWT to specify the organization ID | [optional] |

### Return type

void (empty response body)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

### HTTP response details

| Status code | Description                 | Response headers |
| ----------- | --------------------------- | ---------------- |
| **200**     | Folder created successfully | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **create_session**

> create_session(workspace_id, create_session_request, x_daytona_organization_id=x_daytona_organization_id)

Create session

Create a new session in the workspace

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.models.create_session_request import CreateSessionRequest
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    create_session_request = daytona_api_client.CreateSessionRequest() # CreateSessionRequest |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Create session
        api_instance.create_session(workspace_id, create_session_request, x_daytona_organization_id=x_daytona_organization_id)
    except Exception as e:
        print("Exception when calling ToolboxApi->create_session: %s\n" % e)
```

### Parameters

| Name                          | Type                                                | Description                                 | Notes      |
| ----------------------------- | --------------------------------------------------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str**                                             |                                             |
| **create_session_request**    | [**CreateSessionRequest**](CreateSessionRequest.md) |                                             |
| **x_daytona_organization_id** | **str**                                             | Use with JWT to specify the organization ID | [optional] |

### Return type

void (empty response body)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

### HTTP response details

| Status code | Description | Response headers |
| ----------- | ----------- | ---------------- |
| **200**     |             | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **delete_file**

> delete_file(workspace_id, path, x_daytona_organization_id=x_daytona_organization_id)

Delete file

Delete file inside workspace

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    path = 'path_example' # str |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Delete file
        api_instance.delete_file(workspace_id, path, x_daytona_organization_id=x_daytona_organization_id)
    except Exception as e:
        print("Exception when calling ToolboxApi->delete_file: %s\n" % e)
```

### Parameters

| Name                          | Type    | Description                                 | Notes      |
| ----------------------------- | ------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str** |                                             |
| **path**                      | **str** |                                             |
| **x_daytona_organization_id** | **str** | Use with JWT to specify the organization ID | [optional] |

### Return type

void (empty response body)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

### HTTP response details

| Status code | Description               | Response headers |
| ----------- | ------------------------- | ---------------- |
| **200**     | File deleted successfully | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **delete_session**

> delete_session(workspace_id, session_id, x_daytona_organization_id=x_daytona_organization_id)

Delete session

Delete a specific session

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    session_id = 'session_id_example' # str |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Delete session
        api_instance.delete_session(workspace_id, session_id, x_daytona_organization_id=x_daytona_organization_id)
    except Exception as e:
        print("Exception when calling ToolboxApi->delete_session: %s\n" % e)
```

### Parameters

| Name                          | Type    | Description                                 | Notes      |
| ----------------------------- | ------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str** |                                             |
| **session_id**                | **str** |                                             |
| **x_daytona_organization_id** | **str** | Use with JWT to specify the organization ID | [optional] |

### Return type

void (empty response body)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

### HTTP response details

| Status code | Description                  | Response headers |
| ----------- | ---------------------------- | ---------------- |
| **200**     | Session deleted successfully | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **download_file**

> bytearray download_file(workspace_id, path, x_daytona_organization_id=x_daytona_organization_id)

Download file

Download file from workspace

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    path = 'path_example' # str |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Download file
        api_response = api_instance.download_file(workspace_id, path, x_daytona_organization_id=x_daytona_organization_id)
        print("The response of ToolboxApi->download_file:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ToolboxApi->download_file: %s\n" % e)
```

### Parameters

| Name                          | Type    | Description                                 | Notes      |
| ----------------------------- | ------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str** |                                             |
| **path**                      | **str** |                                             |
| **x_daytona_organization_id** | **str** | Use with JWT to specify the organization ID | [optional] |

### Return type

**bytearray**

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

### HTTP response details

| Status code | Description                  | Response headers |
| ----------- | ---------------------------- | ---------------- |
| **200**     | File downloaded successfully | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **execute_command**

> ExecuteResponse execute_command(workspace_id, execute_request, x_daytona_organization_id=x_daytona_organization_id)

Execute command

Execute command synchronously inside workspace

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.models.execute_request import ExecuteRequest
from daytona_api_client.models.execute_response import ExecuteResponse
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    execute_request = daytona_api_client.ExecuteRequest() # ExecuteRequest |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Execute command
        api_response = api_instance.execute_command(workspace_id, execute_request, x_daytona_organization_id=x_daytona_organization_id)
        print("The response of ToolboxApi->execute_command:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ToolboxApi->execute_command: %s\n" % e)
```

### Parameters

| Name                          | Type                                    | Description                                 | Notes      |
| ----------------------------- | --------------------------------------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str**                                 |                                             |
| **execute_request**           | [**ExecuteRequest**](ExecuteRequest.md) |                                             |
| **x_daytona_organization_id** | **str**                                 | Use with JWT to specify the organization ID | [optional] |

### Return type

[**ExecuteResponse**](ExecuteResponse.md)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

### HTTP response details

| Status code | Description                   | Response headers |
| ----------- | ----------------------------- | ---------------- |
| **200**     | Command executed successfully | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **execute_session_command**

> SessionExecuteResponse execute_session_command(workspace_id, session_id, session_execute_request, x_daytona_organization_id=x_daytona_organization_id)

Execute command in session

Execute a command in a specific session

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.models.session_execute_request import SessionExecuteRequest
from daytona_api_client.models.session_execute_response import SessionExecuteResponse
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    session_id = 'session_id_example' # str |
    session_execute_request = daytona_api_client.SessionExecuteRequest() # SessionExecuteRequest |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Execute command in session
        api_response = api_instance.execute_session_command(workspace_id, session_id, session_execute_request, x_daytona_organization_id=x_daytona_organization_id)
        print("The response of ToolboxApi->execute_session_command:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ToolboxApi->execute_session_command: %s\n" % e)
```

### Parameters

| Name                          | Type                                                  | Description                                 | Notes      |
| ----------------------------- | ----------------------------------------------------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str**                                               |                                             |
| **session_id**                | **str**                                               |                                             |
| **session_execute_request**   | [**SessionExecuteRequest**](SessionExecuteRequest.md) |                                             |
| **x_daytona_organization_id** | **str**                                               | Use with JWT to specify the organization ID | [optional] |

### Return type

[**SessionExecuteResponse**](SessionExecuteResponse.md)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

### HTTP response details

| Status code | Description                             | Response headers |
| ----------- | --------------------------------------- | ---------------- |
| **200**     | Command executed successfully           | -                |
| **202**     | Command accepted and is being processed | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **find_in_files**

> List[Match] find_in_files(workspace_id, path, pattern, x_daytona_organization_id=x_daytona_organization_id)

Search for text/pattern in files

Search for text/pattern inside workspace files

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.models.match import Match
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    path = 'path_example' # str |
    pattern = 'pattern_example' # str |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Search for text/pattern in files
        api_response = api_instance.find_in_files(workspace_id, path, pattern, x_daytona_organization_id=x_daytona_organization_id)
        print("The response of ToolboxApi->find_in_files:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ToolboxApi->find_in_files: %s\n" % e)
```

### Parameters

| Name                          | Type    | Description                                 | Notes      |
| ----------------------------- | ------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str** |                                             |
| **path**                      | **str** |                                             |
| **pattern**                   | **str** |                                             |
| **x_daytona_organization_id** | **str** | Use with JWT to specify the organization ID | [optional] |

### Return type

[**List[Match]**](Match.md)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

### HTTP response details

| Status code | Description                   | Response headers |
| ----------- | ----------------------------- | ---------------- |
| **200**     | Search completed successfully | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_file_info**

> FileInfo get_file_info(workspace_id, path, x_daytona_organization_id=x_daytona_organization_id)

Get file info

Get file info inside workspace

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.models.file_info import FileInfo
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    path = 'path_example' # str |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Get file info
        api_response = api_instance.get_file_info(workspace_id, path, x_daytona_organization_id=x_daytona_organization_id)
        print("The response of ToolboxApi->get_file_info:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ToolboxApi->get_file_info: %s\n" % e)
```

### Parameters

| Name                          | Type    | Description                                 | Notes      |
| ----------------------------- | ------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str** |                                             |
| **path**                      | **str** |                                             |
| **x_daytona_organization_id** | **str** | Use with JWT to specify the organization ID | [optional] |

### Return type

[**FileInfo**](FileInfo.md)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

### HTTP response details

| Status code | Description                      | Response headers |
| ----------- | -------------------------------- | ---------------- |
| **200**     | File info retrieved successfully | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_project_dir**

> ProjectDirResponse get_project_dir(workspace_id, x_daytona_organization_id=x_daytona_organization_id)

Get workspace project dir

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.models.project_dir_response import ProjectDirResponse
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Get workspace project dir
        api_response = api_instance.get_project_dir(workspace_id, x_daytona_organization_id=x_daytona_organization_id)
        print("The response of ToolboxApi->get_project_dir:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ToolboxApi->get_project_dir: %s\n" % e)
```

### Parameters

| Name                          | Type    | Description                                 | Notes      |
| ----------------------------- | ------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str** |                                             |
| **x_daytona_organization_id** | **str** | Use with JWT to specify the organization ID | [optional] |

### Return type

[**ProjectDirResponse**](ProjectDirResponse.md)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

### HTTP response details

| Status code | Description                              | Response headers |
| ----------- | ---------------------------------------- | ---------------- |
| **200**     | Project directory retrieved successfully | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_session**

> Session get_session(workspace_id, session_id, x_daytona_organization_id=x_daytona_organization_id)

Get session

Get session by ID

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.models.session import Session
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    session_id = 'session_id_example' # str |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Get session
        api_response = api_instance.get_session(workspace_id, session_id, x_daytona_organization_id=x_daytona_organization_id)
        print("The response of ToolboxApi->get_session:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ToolboxApi->get_session: %s\n" % e)
```

### Parameters

| Name                          | Type    | Description                                 | Notes      |
| ----------------------------- | ------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str** |                                             |
| **session_id**                | **str** |                                             |
| **x_daytona_organization_id** | **str** | Use with JWT to specify the organization ID | [optional] |

### Return type

[**Session**](Session.md)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

### HTTP response details

| Status code | Description                    | Response headers |
| ----------- | ------------------------------ | ---------------- |
| **200**     | Session retrieved successfully | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_session_command**

> Command get_session_command(workspace_id, session_id, command_id, x_daytona_organization_id=x_daytona_organization_id)

Get session command

Get session command by ID

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.models.command import Command
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    session_id = 'session_id_example' # str |
    command_id = 'command_id_example' # str |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Get session command
        api_response = api_instance.get_session_command(workspace_id, session_id, command_id, x_daytona_organization_id=x_daytona_organization_id)
        print("The response of ToolboxApi->get_session_command:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ToolboxApi->get_session_command: %s\n" % e)
```

### Parameters

| Name                          | Type    | Description                                 | Notes      |
| ----------------------------- | ------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str** |                                             |
| **session_id**                | **str** |                                             |
| **command_id**                | **str** |                                             |
| **x_daytona_organization_id** | **str** | Use with JWT to specify the organization ID | [optional] |

### Return type

[**Command**](Command.md)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

### HTTP response details

| Status code | Description                            | Response headers |
| ----------- | -------------------------------------- | ---------------- |
| **200**     | Session command retrieved successfully | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_session_command_logs**

> str get_session_command_logs(workspace_id, session_id, command_id, x_daytona_organization_id=x_daytona_organization_id, follow=follow)

Get command logs

Get logs for a specific command in a session

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    session_id = 'session_id_example' # str |
    command_id = 'command_id_example' # str |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)
    follow = True # bool |  (optional)

    try:
        # Get command logs
        api_response = api_instance.get_session_command_logs(workspace_id, session_id, command_id, x_daytona_organization_id=x_daytona_organization_id, follow=follow)
        print("The response of ToolboxApi->get_session_command_logs:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ToolboxApi->get_session_command_logs: %s\n" % e)
```

### Parameters

| Name                          | Type     | Description                                 | Notes      |
| ----------------------------- | -------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str**  |                                             |
| **session_id**                | **str**  |                                             |
| **command_id**                | **str**  |                                             |
| **x_daytona_organization_id** | **str**  | Use with JWT to specify the organization ID | [optional] |
| **follow**                    | **bool** |                                             | [optional] |

### Return type

**str**

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: text/plain

### HTTP response details

| Status code | Description        | Response headers |
| ----------- | ------------------ | ---------------- |
| **200**     | Command log stream | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **git_add_files**

> git_add_files(workspace_id, git_add_request, x_daytona_organization_id=x_daytona_organization_id)

Add files

Add files to git commit

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.models.git_add_request import GitAddRequest
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    git_add_request = daytona_api_client.GitAddRequest() # GitAddRequest |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Add files
        api_instance.git_add_files(workspace_id, git_add_request, x_daytona_organization_id=x_daytona_organization_id)
    except Exception as e:
        print("Exception when calling ToolboxApi->git_add_files: %s\n" % e)
```

### Parameters

| Name                          | Type                                  | Description                                 | Notes      |
| ----------------------------- | ------------------------------------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str**                               |                                             |
| **git_add_request**           | [**GitAddRequest**](GitAddRequest.md) |                                             |
| **x_daytona_organization_id** | **str**                               | Use with JWT to specify the organization ID | [optional] |

### Return type

void (empty response body)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

### HTTP response details

| Status code | Description                     | Response headers |
| ----------- | ------------------------------- | ---------------- |
| **200**     | Files added to git successfully | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **git_clone_repository**

> git_clone_repository(workspace_id, git_clone_request, x_daytona_organization_id=x_daytona_organization_id)

Clone repository

Clone git repository

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.models.git_clone_request import GitCloneRequest
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    git_clone_request = daytona_api_client.GitCloneRequest() # GitCloneRequest |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Clone repository
        api_instance.git_clone_repository(workspace_id, git_clone_request, x_daytona_organization_id=x_daytona_organization_id)
    except Exception as e:
        print("Exception when calling ToolboxApi->git_clone_repository: %s\n" % e)
```

### Parameters

| Name                          | Type                                      | Description                                 | Notes      |
| ----------------------------- | ----------------------------------------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str**                                   |                                             |
| **git_clone_request**         | [**GitCloneRequest**](GitCloneRequest.md) |                                             |
| **x_daytona_organization_id** | **str**                                   | Use with JWT to specify the organization ID | [optional] |

### Return type

void (empty response body)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

### HTTP response details

| Status code | Description                    | Response headers |
| ----------- | ------------------------------ | ---------------- |
| **200**     | Repository cloned successfully | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **git_commit_changes**

> GitCommitResponse git_commit_changes(workspace_id, git_commit_request, x_daytona_organization_id=x_daytona_organization_id)

Commit changes

Commit changes to git repository

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.models.git_commit_request import GitCommitRequest
from daytona_api_client.models.git_commit_response import GitCommitResponse
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    git_commit_request = daytona_api_client.GitCommitRequest() # GitCommitRequest |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Commit changes
        api_response = api_instance.git_commit_changes(workspace_id, git_commit_request, x_daytona_organization_id=x_daytona_organization_id)
        print("The response of ToolboxApi->git_commit_changes:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ToolboxApi->git_commit_changes: %s\n" % e)
```

### Parameters

| Name                          | Type                                        | Description                                 | Notes      |
| ----------------------------- | ------------------------------------------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str**                                     |                                             |
| **git_commit_request**        | [**GitCommitRequest**](GitCommitRequest.md) |                                             |
| **x_daytona_organization_id** | **str**                                     | Use with JWT to specify the organization ID | [optional] |

### Return type

[**GitCommitResponse**](GitCommitResponse.md)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

### HTTP response details

| Status code | Description                    | Response headers |
| ----------- | ------------------------------ | ---------------- |
| **200**     | Changes committed successfully | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **git_create_branch**

> git_create_branch(workspace_id, git_branch_request, x_daytona_organization_id=x_daytona_organization_id)

Create branch

Create branch on git repository

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.models.git_branch_request import GitBranchRequest
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    git_branch_request = daytona_api_client.GitBranchRequest() # GitBranchRequest |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Create branch
        api_instance.git_create_branch(workspace_id, git_branch_request, x_daytona_organization_id=x_daytona_organization_id)
    except Exception as e:
        print("Exception when calling ToolboxApi->git_create_branch: %s\n" % e)
```

### Parameters

| Name                          | Type                                        | Description                                 | Notes      |
| ----------------------------- | ------------------------------------------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str**                                     |                                             |
| **git_branch_request**        | [**GitBranchRequest**](GitBranchRequest.md) |                                             |
| **x_daytona_organization_id** | **str**                                     | Use with JWT to specify the organization ID | [optional] |

### Return type

void (empty response body)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

### HTTP response details

| Status code | Description                 | Response headers |
| ----------- | --------------------------- | ---------------- |
| **200**     | Branch created successfully | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **git_get_history**

> List[GitCommitInfo] git_get_history(workspace_id, path, x_daytona_organization_id=x_daytona_organization_id)

Get commit history

Get commit history from git repository

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.models.git_commit_info import GitCommitInfo
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    path = 'path_example' # str |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Get commit history
        api_response = api_instance.git_get_history(workspace_id, path, x_daytona_organization_id=x_daytona_organization_id)
        print("The response of ToolboxApi->git_get_history:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ToolboxApi->git_get_history: %s\n" % e)
```

### Parameters

| Name                          | Type    | Description                                 | Notes      |
| ----------------------------- | ------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str** |                                             |
| **path**                      | **str** |                                             |
| **x_daytona_organization_id** | **str** | Use with JWT to specify the organization ID | [optional] |

### Return type

[**List[GitCommitInfo]**](GitCommitInfo.md)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

### HTTP response details

| Status code | Description                           | Response headers |
| ----------- | ------------------------------------- | ---------------- |
| **200**     | Commit history retrieved successfully | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **git_get_status**

> GitStatus git_get_status(workspace_id, path, x_daytona_organization_id=x_daytona_organization_id)

Get git status

Get status from git repository

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.models.git_status import GitStatus
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    path = 'path_example' # str |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Get git status
        api_response = api_instance.git_get_status(workspace_id, path, x_daytona_organization_id=x_daytona_organization_id)
        print("The response of ToolboxApi->git_get_status:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ToolboxApi->git_get_status: %s\n" % e)
```

### Parameters

| Name                          | Type    | Description                                 | Notes      |
| ----------------------------- | ------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str** |                                             |
| **path**                      | **str** |                                             |
| **x_daytona_organization_id** | **str** | Use with JWT to specify the organization ID | [optional] |

### Return type

[**GitStatus**](GitStatus.md)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

### HTTP response details

| Status code | Description                       | Response headers |
| ----------- | --------------------------------- | ---------------- |
| **200**     | Git status retrieved successfully | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **git_list_branches**

> ListBranchResponse git_list_branches(workspace_id, path, x_daytona_organization_id=x_daytona_organization_id)

Get branch list

Get branch list from git repository

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.models.list_branch_response import ListBranchResponse
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    path = 'path_example' # str |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Get branch list
        api_response = api_instance.git_list_branches(workspace_id, path, x_daytona_organization_id=x_daytona_organization_id)
        print("The response of ToolboxApi->git_list_branches:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ToolboxApi->git_list_branches: %s\n" % e)
```

### Parameters

| Name                          | Type    | Description                                 | Notes      |
| ----------------------------- | ------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str** |                                             |
| **path**                      | **str** |                                             |
| **x_daytona_organization_id** | **str** | Use with JWT to specify the organization ID | [optional] |

### Return type

[**ListBranchResponse**](ListBranchResponse.md)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

### HTTP response details

| Status code | Description                        | Response headers |
| ----------- | ---------------------------------- | ---------------- |
| **200**     | Branch list retrieved successfully | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **git_pull_changes**

> git_pull_changes(workspace_id, git_repo_request, x_daytona_organization_id=x_daytona_organization_id)

Pull changes

Pull changes from remote

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.models.git_repo_request import GitRepoRequest
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    git_repo_request = daytona_api_client.GitRepoRequest() # GitRepoRequest |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Pull changes
        api_instance.git_pull_changes(workspace_id, git_repo_request, x_daytona_organization_id=x_daytona_organization_id)
    except Exception as e:
        print("Exception when calling ToolboxApi->git_pull_changes: %s\n" % e)
```

### Parameters

| Name                          | Type                                    | Description                                 | Notes      |
| ----------------------------- | --------------------------------------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str**                                 |                                             |
| **git_repo_request**          | [**GitRepoRequest**](GitRepoRequest.md) |                                             |
| **x_daytona_organization_id** | **str**                                 | Use with JWT to specify the organization ID | [optional] |

### Return type

void (empty response body)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

### HTTP response details

| Status code | Description                 | Response headers |
| ----------- | --------------------------- | ---------------- |
| **200**     | Changes pulled successfully | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **git_push_changes**

> git_push_changes(workspace_id, git_repo_request, x_daytona_organization_id=x_daytona_organization_id)

Push changes

Push changes to remote

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.models.git_repo_request import GitRepoRequest
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    git_repo_request = daytona_api_client.GitRepoRequest() # GitRepoRequest |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Push changes
        api_instance.git_push_changes(workspace_id, git_repo_request, x_daytona_organization_id=x_daytona_organization_id)
    except Exception as e:
        print("Exception when calling ToolboxApi->git_push_changes: %s\n" % e)
```

### Parameters

| Name                          | Type                                    | Description                                 | Notes      |
| ----------------------------- | --------------------------------------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str**                                 |                                             |
| **git_repo_request**          | [**GitRepoRequest**](GitRepoRequest.md) |                                             |
| **x_daytona_organization_id** | **str**                                 | Use with JWT to specify the organization ID | [optional] |

### Return type

void (empty response body)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

### HTTP response details

| Status code | Description                 | Response headers |
| ----------- | --------------------------- | ---------------- |
| **200**     | Changes pushed successfully | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **list_files**

> List[FileInfo] list_files(workspace_id, x_daytona_organization_id=x_daytona_organization_id, path=path)

List files

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.models.file_info import FileInfo
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)
    path = 'path_example' # str |  (optional)

    try:
        # List files
        api_response = api_instance.list_files(workspace_id, x_daytona_organization_id=x_daytona_organization_id, path=path)
        print("The response of ToolboxApi->list_files:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ToolboxApi->list_files: %s\n" % e)
```

### Parameters

| Name                          | Type    | Description                                 | Notes      |
| ----------------------------- | ------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str** |                                             |
| **x_daytona_organization_id** | **str** | Use with JWT to specify the organization ID | [optional] |
| **path**                      | **str** |                                             | [optional] |

### Return type

[**List[FileInfo]**](FileInfo.md)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

### HTTP response details

| Status code | Description               | Response headers |
| ----------- | ------------------------- | ---------------- |
| **200**     | Files listed successfully | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **list_sessions**

> List[Session] list_sessions(workspace_id, x_daytona_organization_id=x_daytona_organization_id)

List sessions

List all active sessions in the workspace

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.models.session import Session
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # List sessions
        api_response = api_instance.list_sessions(workspace_id, x_daytona_organization_id=x_daytona_organization_id)
        print("The response of ToolboxApi->list_sessions:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ToolboxApi->list_sessions: %s\n" % e)
```

### Parameters

| Name                          | Type    | Description                                 | Notes      |
| ----------------------------- | ------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str** |                                             |
| **x_daytona_organization_id** | **str** | Use with JWT to specify the organization ID | [optional] |

### Return type

[**List[Session]**](Session.md)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

### HTTP response details

| Status code | Description                     | Response headers |
| ----------- | ------------------------------- | ---------------- |
| **200**     | Sessions retrieved successfully | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **lsp_completions**

> CompletionList lsp_completions(workspace_id, lsp_completion_params, x_daytona_organization_id=x_daytona_organization_id)

Get Lsp Completions

The Completion request is sent from the client to the server to compute completion items at a given cursor position.

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.models.completion_list import CompletionList
from daytona_api_client.models.lsp_completion_params import LspCompletionParams
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    lsp_completion_params = daytona_api_client.LspCompletionParams() # LspCompletionParams |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Get Lsp Completions
        api_response = api_instance.lsp_completions(workspace_id, lsp_completion_params, x_daytona_organization_id=x_daytona_organization_id)
        print("The response of ToolboxApi->lsp_completions:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ToolboxApi->lsp_completions: %s\n" % e)
```

### Parameters

| Name                          | Type                                              | Description                                 | Notes      |
| ----------------------------- | ------------------------------------------------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str**                                           |                                             |
| **lsp_completion_params**     | [**LspCompletionParams**](LspCompletionParams.md) |                                             |
| **x_daytona_organization_id** | **str**                                           | Use with JWT to specify the organization ID | [optional] |

### Return type

[**CompletionList**](CompletionList.md)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
| ----------- | ----------- | ---------------- |
| **200**     | OK          | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **lsp_did_close**

> lsp_did_close(workspace_id, lsp_document_request, x_daytona_organization_id=x_daytona_organization_id)

Call Lsp DidClose

The document close notification is sent from the client to the server when the document got closed in the client.

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.models.lsp_document_request import LspDocumentRequest
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    lsp_document_request = daytona_api_client.LspDocumentRequest() # LspDocumentRequest |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Call Lsp DidClose
        api_instance.lsp_did_close(workspace_id, lsp_document_request, x_daytona_organization_id=x_daytona_organization_id)
    except Exception as e:
        print("Exception when calling ToolboxApi->lsp_did_close: %s\n" % e)
```

### Parameters

| Name                          | Type                                            | Description                                 | Notes      |
| ----------------------------- | ----------------------------------------------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str**                                         |                                             |
| **lsp_document_request**      | [**LspDocumentRequest**](LspDocumentRequest.md) |                                             |
| **x_daytona_organization_id** | **str**                                         | Use with JWT to specify the organization ID | [optional] |

### Return type

void (empty response body)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

### HTTP response details

| Status code | Description | Response headers |
| ----------- | ----------- | ---------------- |
| **200**     | OK          | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **lsp_did_open**

> lsp_did_open(workspace_id, lsp_document_request, x_daytona_organization_id=x_daytona_organization_id)

Call Lsp DidOpen

The document open notification is sent from the client to the server to signal newly opened text documents.

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.models.lsp_document_request import LspDocumentRequest
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    lsp_document_request = daytona_api_client.LspDocumentRequest() # LspDocumentRequest |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Call Lsp DidOpen
        api_instance.lsp_did_open(workspace_id, lsp_document_request, x_daytona_organization_id=x_daytona_organization_id)
    except Exception as e:
        print("Exception when calling ToolboxApi->lsp_did_open: %s\n" % e)
```

### Parameters

| Name                          | Type                                            | Description                                 | Notes      |
| ----------------------------- | ----------------------------------------------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str**                                         |                                             |
| **lsp_document_request**      | [**LspDocumentRequest**](LspDocumentRequest.md) |                                             |
| **x_daytona_organization_id** | **str**                                         | Use with JWT to specify the organization ID | [optional] |

### Return type

void (empty response body)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

### HTTP response details

| Status code | Description | Response headers |
| ----------- | ----------- | ---------------- |
| **200**     | OK          | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **lsp_document_symbols**

> List[LspSymbol] lsp_document_symbols(workspace_id, language_id, path_to_project, uri, x_daytona_organization_id=x_daytona_organization_id)

Call Lsp DocumentSymbols

The document symbol request is sent from the client to the server.

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.models.lsp_symbol import LspSymbol
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    language_id = 'language_id_example' # str |
    path_to_project = 'path_to_project_example' # str |
    uri = 'uri_example' # str |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Call Lsp DocumentSymbols
        api_response = api_instance.lsp_document_symbols(workspace_id, language_id, path_to_project, uri, x_daytona_organization_id=x_daytona_organization_id)
        print("The response of ToolboxApi->lsp_document_symbols:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ToolboxApi->lsp_document_symbols: %s\n" % e)
```

### Parameters

| Name                          | Type    | Description                                 | Notes      |
| ----------------------------- | ------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str** |                                             |
| **language_id**               | **str** |                                             |
| **path_to_project**           | **str** |                                             |
| **uri**                       | **str** |                                             |
| **x_daytona_organization_id** | **str** | Use with JWT to specify the organization ID | [optional] |

### Return type

[**List[LspSymbol]**](LspSymbol.md)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
| ----------- | ----------- | ---------------- |
| **200**     | OK          | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **lsp_start**

> lsp_start(workspace_id, lsp_server_request, x_daytona_organization_id=x_daytona_organization_id)

Start Lsp server

Start Lsp server process inside workspace project

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.models.lsp_server_request import LspServerRequest
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    lsp_server_request = daytona_api_client.LspServerRequest() # LspServerRequest |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Start Lsp server
        api_instance.lsp_start(workspace_id, lsp_server_request, x_daytona_organization_id=x_daytona_organization_id)
    except Exception as e:
        print("Exception when calling ToolboxApi->lsp_start: %s\n" % e)
```

### Parameters

| Name                          | Type                                        | Description                                 | Notes      |
| ----------------------------- | ------------------------------------------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str**                                     |                                             |
| **lsp_server_request**        | [**LspServerRequest**](LspServerRequest.md) |                                             |
| **x_daytona_organization_id** | **str**                                     | Use with JWT to specify the organization ID | [optional] |

### Return type

void (empty response body)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

### HTTP response details

| Status code | Description | Response headers |
| ----------- | ----------- | ---------------- |
| **200**     | OK          | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **lsp_stop**

> lsp_stop(workspace_id, lsp_server_request, x_daytona_organization_id=x_daytona_organization_id)

Stop Lsp server

Stop Lsp server process inside workspace project

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.models.lsp_server_request import LspServerRequest
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    lsp_server_request = daytona_api_client.LspServerRequest() # LspServerRequest |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Stop Lsp server
        api_instance.lsp_stop(workspace_id, lsp_server_request, x_daytona_organization_id=x_daytona_organization_id)
    except Exception as e:
        print("Exception when calling ToolboxApi->lsp_stop: %s\n" % e)
```

### Parameters

| Name                          | Type                                        | Description                                 | Notes      |
| ----------------------------- | ------------------------------------------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str**                                     |                                             |
| **lsp_server_request**        | [**LspServerRequest**](LspServerRequest.md) |                                             |
| **x_daytona_organization_id** | **str**                                     | Use with JWT to specify the organization ID | [optional] |

### Return type

void (empty response body)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

### HTTP response details

| Status code | Description | Response headers |
| ----------- | ----------- | ---------------- |
| **200**     | OK          | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **lsp_workspace_symbols**

> List[LspSymbol] lsp_workspace_symbols(workspace_id, language_id, path_to_project, query, x_daytona_organization_id=x_daytona_organization_id)

Call Lsp WorkspaceSymbols

The workspace symbol request is sent from the client to the server to list project-wide symbols matching the query string.

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.models.lsp_symbol import LspSymbol
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    language_id = 'language_id_example' # str |
    path_to_project = 'path_to_project_example' # str |
    query = 'query_example' # str |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Call Lsp WorkspaceSymbols
        api_response = api_instance.lsp_workspace_symbols(workspace_id, language_id, path_to_project, query, x_daytona_organization_id=x_daytona_organization_id)
        print("The response of ToolboxApi->lsp_workspace_symbols:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ToolboxApi->lsp_workspace_symbols: %s\n" % e)
```

### Parameters

| Name                          | Type    | Description                                 | Notes      |
| ----------------------------- | ------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str** |                                             |
| **language_id**               | **str** |                                             |
| **path_to_project**           | **str** |                                             |
| **query**                     | **str** |                                             |
| **x_daytona_organization_id** | **str** | Use with JWT to specify the organization ID | [optional] |

### Return type

[**List[LspSymbol]**](LspSymbol.md)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
| ----------- | ----------- | ---------------- |
| **200**     | OK          | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **move_file**

> move_file(workspace_id, source, destination, x_daytona_organization_id=x_daytona_organization_id)

Move file

Move file inside workspace

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    source = 'source_example' # str |
    destination = 'destination_example' # str |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Move file
        api_instance.move_file(workspace_id, source, destination, x_daytona_organization_id=x_daytona_organization_id)
    except Exception as e:
        print("Exception when calling ToolboxApi->move_file: %s\n" % e)
```

### Parameters

| Name                          | Type    | Description                                 | Notes      |
| ----------------------------- | ------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str** |                                             |
| **source**                    | **str** |                                             |
| **destination**               | **str** |                                             |
| **x_daytona_organization_id** | **str** | Use with JWT to specify the organization ID | [optional] |

### Return type

void (empty response body)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

### HTTP response details

| Status code | Description             | Response headers |
| ----------- | ----------------------- | ---------------- |
| **200**     | File moved successfully | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **replace_in_files**

> List[ReplaceResult] replace_in_files(workspace_id, replace_request, x_daytona_organization_id=x_daytona_organization_id)

Replace in files

Replace text/pattern in multiple files inside workspace

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.models.replace_request import ReplaceRequest
from daytona_api_client.models.replace_result import ReplaceResult
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    replace_request = daytona_api_client.ReplaceRequest() # ReplaceRequest |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Replace in files
        api_response = api_instance.replace_in_files(workspace_id, replace_request, x_daytona_organization_id=x_daytona_organization_id)
        print("The response of ToolboxApi->replace_in_files:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ToolboxApi->replace_in_files: %s\n" % e)
```

### Parameters

| Name                          | Type                                    | Description                                 | Notes      |
| ----------------------------- | --------------------------------------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str**                                 |                                             |
| **replace_request**           | [**ReplaceRequest**](ReplaceRequest.md) |                                             |
| **x_daytona_organization_id** | **str**                                 | Use with JWT to specify the organization ID | [optional] |

### Return type

[**List[ReplaceResult]**](ReplaceResult.md)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

### HTTP response details

| Status code | Description                | Response headers |
| ----------- | -------------------------- | ---------------- |
| **200**     | Text replaced successfully | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **search_files**

> SearchFilesResponse search_files(workspace_id, path, pattern, x_daytona_organization_id=x_daytona_organization_id)

Search files

Search for files inside workspace

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.models.search_files_response import SearchFilesResponse
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    path = 'path_example' # str |
    pattern = 'pattern_example' # str |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)

    try:
        # Search files
        api_response = api_instance.search_files(workspace_id, path, pattern, x_daytona_organization_id=x_daytona_organization_id)
        print("The response of ToolboxApi->search_files:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ToolboxApi->search_files: %s\n" % e)
```

### Parameters

| Name                          | Type    | Description                                 | Notes      |
| ----------------------------- | ------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str** |                                             |
| **path**                      | **str** |                                             |
| **pattern**                   | **str** |                                             |
| **x_daytona_organization_id** | **str** | Use with JWT to specify the organization ID | [optional] |

### Return type

[**SearchFilesResponse**](SearchFilesResponse.md)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

### HTTP response details

| Status code | Description                   | Response headers |
| ----------- | ----------------------------- | ---------------- |
| **200**     | Search completed successfully | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **set_file_permissions**

> set_file_permissions(workspace_id, path, x_daytona_organization_id=x_daytona_organization_id, owner=owner, group=group, mode=mode)

Set file permissions

Set file owner/group/permissions inside workspace

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    path = 'path_example' # str |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)
    owner = 'owner_example' # str |  (optional)
    group = 'group_example' # str |  (optional)
    mode = 'mode_example' # str |  (optional)

    try:
        # Set file permissions
        api_instance.set_file_permissions(workspace_id, path, x_daytona_organization_id=x_daytona_organization_id, owner=owner, group=group, mode=mode)
    except Exception as e:
        print("Exception when calling ToolboxApi->set_file_permissions: %s\n" % e)
```

### Parameters

| Name                          | Type    | Description                                 | Notes      |
| ----------------------------- | ------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str** |                                             |
| **path**                      | **str** |                                             |
| **x_daytona_organization_id** | **str** | Use with JWT to specify the organization ID | [optional] |
| **owner**                     | **str** |                                             | [optional] |
| **group**                     | **str** |                                             | [optional] |
| **mode**                      | **str** |                                             | [optional] |

### Return type

void (empty response body)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

### HTTP response details

| Status code | Description                           | Response headers |
| ----------- | ------------------------------------- | ---------------- |
| **200**     | File permissions updated successfully | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **upload_file**

> upload_file(workspace_id, path, x_daytona_organization_id=x_daytona_organization_id, file=file)

Upload file

Upload file inside workspace

### Example

- OAuth Authentication (oauth2):

```python
import daytona_api_client
from daytona_api_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = daytona_api_client.Configuration(
    host = "http://localhost"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

configuration.access_token = os.environ["ACCESS_TOKEN"]

# Enter a context with an instance of the API client
with daytona_api_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = daytona_api_client.ToolboxApi(api_client)
    workspace_id = 'workspace_id_example' # str |
    path = 'path_example' # str |
    x_daytona_organization_id = 'x_daytona_organization_id_example' # str | Use with JWT to specify the organization ID (optional)
    file = None # bytearray |  (optional)

    try:
        # Upload file
        api_instance.upload_file(workspace_id, path, x_daytona_organization_id=x_daytona_organization_id, file=file)
    except Exception as e:
        print("Exception when calling ToolboxApi->upload_file: %s\n" % e)
```

### Parameters

| Name                          | Type          | Description                                 | Notes      |
| ----------------------------- | ------------- | ------------------------------------------- | ---------- |
| **workspace_id**              | **str**       |                                             |
| **path**                      | **str**       |                                             |
| **x_daytona_organization_id** | **str**       | Use with JWT to specify the organization ID | [optional] |
| **file**                      | **bytearray** |                                             | [optional] |

### Return type

void (empty response body)

### Authorization

[oauth2](../README.md#oauth2)

### HTTP request headers

- **Content-Type**: multipart/form-data
- **Accept**: Not defined

### HTTP response details

| Status code | Description                | Response headers |
| ----------- | -------------------------- | ---------------- |
| **200**     | File uploaded successfully | -                |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)
