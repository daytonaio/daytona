---
title: "Daytona"
hideTitleOnPage: true
---

## Daytona

```python
class Daytona()
```

Main class for interacting with the Daytona API.

This class provides methods to create, manage, and interact with Daytona Sandboxes.
It can be initialized either with explicit configuration or using environment variables.

**Attributes**:

- `volume` _VolumeService_ - Service for managing volumes.
- `snapshot` _SnapshotService_ - Service for managing snapshots.
  

**Example**:

  Using environment variables:
```python
daytona = Daytona()  # Uses DAYTONA_API_KEY, DAYTONA_API_URL
sandbox = daytona.create()
```
  
  Using explicit configuration:
```python
config = DaytonaConfig(
    api_key="your-api-key",
    api_url="https://your-api.com",
    target="us"
)
daytona = Daytona(config)
sandbox = daytona.create()
```
  
  Using OpenTelemetry tracing:
```python
config = DaytonaConfig(
    api_key="your-api-key",
    experimental={"otelEnabled": True}  # Enable OpenTelemetry tracing through experimental config
)
async with Daytona(config) as daytona:
    sandbox = daytona.create()
    # All SDK operations will be traced
# OpenTelemetry traces are flushed on close
```

#### Daytona.\_\_init\_\_

```python
def __init__(config: DaytonaConfig | None = None)
```

Initializes Daytona instance with optional configuration.

If no config is provided, reads from environment variables:
- `DAYTONA_API_KEY`: Required API key for authentication
- `DAYTONA_API_URL`: Required api URL
- `DAYTONA_TARGET`: Optional target environment (if not provided, default region for the organization is used)

**Arguments**:

- `config` _DaytonaConfig | None_ - Object containing api_key, api_url, and target.
  

**Raises**:

- `DaytonaError` - If API key is not provided either through config or environment variables
  

**Example**:

```python
from daytona import Daytona, DaytonaConfig
# Using environment variables
daytona1 = Daytona()

# Using explicit configuration
config = DaytonaConfig(
    api_key="your-api-key",
    api_url="https://your-api.com",
    target="us"
)
daytona2 = Daytona(config)

```

#### Daytona.create

```python
@overload
def create(params: CreateSandboxFromSnapshotParams | None = None,
           *,
           timeout: float = 60) -> Sandbox
```

Creates Sandboxes from specified or default snapshot. You can specify various parameters,
including language, image, environment variables, and volumes.

**Arguments**:

- `params` _CreateSandboxFromSnapshotParams | None_ - Parameters for Sandbox creation. If not provided,
  defaults to default Daytona snapshot and Python language.
- `timeout` _float_ - Timeout (in seconds) for sandbox creation. 0 means no timeout.
  Default is 60 seconds.
  

**Returns**:

- `Sandbox` - The created Sandbox instance.
  

**Raises**:

- `DaytonaError` - If timeout, auto_stop_interval or auto_archive_interval is negative;
  If sandbox fails to start or times out
  

**Example**:

  Create a default Python Sandbox:
```python
sandbox = daytona.create()
```
  
  Create a custom Sandbox:
```python
params = CreateSandboxFromSnapshotParams(
    language="python",
    snapshot="my-snapshot-id",
    env_vars={"DEBUG": "true"},
    auto_stop_interval=0,
    auto_archive_interval=60,
    auto_delete_interval=120
)
sandbox = daytona.create(params, timeout=40)
```

#### Daytona.create

```python
@overload
def create(
        params: CreateSandboxFromImageParams | None = None,
        *,
        timeout: float = 60,
        on_snapshot_create_logs: Callable[[str], None] | None = None
) -> Sandbox
```

Creates Sandboxes from specified image available on some registry or declarative Daytona Image.
You can specify various parameters, including resources, language, image, environment variables,
and volumes. Daytona creates snapshot from provided image and uses it to create Sandbox.

**Arguments**:

- `params` _CreateSandboxFromImageParams | None_ - Parameters for Sandbox creation from image.
- `timeout` _float_ - Timeout (in seconds) for sandbox creation. 0 means no timeout.
  Default is 60 seconds.
- `on_snapshot_create_logs` _Callable[[str], None] | None_ - This callback function
  handles snapshot creation logs.
  

**Returns**:

- `Sandbox` - The created Sandbox instance.
  

**Raises**:

- `DaytonaError` - If timeout, auto_stop_interval or auto_archive_interval is negative;
  If sandbox fails to start or times out
  

**Example**:

  Create a default Python Sandbox from image:
```python
sandbox = daytona.create(CreateSandboxFromImageParams(image="debian:12.9"))
```
  
  Create a custom Sandbox from declarative Image definition:
```python
declarative_image = (
    Image.base("alpine:3.18")
    .pipInstall(["numpy", "pandas"])
    .env({"MY_ENV_VAR": "My Environment Variable"})
)
params = CreateSandboxFromImageParams(
    language="python",
    image=declarative_image,
    env_vars={"DEBUG": "true"},
    resources=Resources(cpu=2, memory=4),
    auto_stop_interval=0,
    auto_archive_interval=60,
    auto_delete_interval=120
)
sandbox = daytona.create(
    params,
    timeout=40,
    on_snapshot_create_logs=lambda chunk: print(chunk, end=""),
)
```

#### Daytona.delete

```python
@with_instrumentation()
def delete(sandbox: Sandbox, timeout: float = 60) -> None
```

Deletes a Sandbox.

**Arguments**:

- `sandbox` _Sandbox_ - The Sandbox instance to delete.
- `timeout` _float_ - Timeout (in seconds) for sandbox deletion. 0 means no timeout.
  Default is 60 seconds.
  

**Raises**:

- `DaytonaError` - If sandbox fails to delete or times out
  

**Example**:

```python
sandbox = daytona.create()
# ... use sandbox ...
daytona.delete(sandbox)  # Clean up when done
```

#### Daytona.get

```python
@intercept_errors(message_prefix="Failed to get sandbox: ")
@with_instrumentation()
def get(sandbox_id_or_name: str) -> Sandbox
```

Gets a Sandbox by its ID or name.

**Arguments**:

- `sandbox_id_or_name` _str_ - The ID or name of the Sandbox to retrieve.
  

**Returns**:

- `Sandbox` - The Sandbox instance.
  

**Raises**:

- `DaytonaError` - If sandbox_id_or_name is not provided.
  

**Example**:

```python
sandbox = daytona.get("my-sandbox-id-or-name")
print(sandbox.state)
```

#### Daytona.find\_one

```python
@intercept_errors(message_prefix="Failed to find sandbox: ")
@with_instrumentation()
def find_one(sandbox_id_or_name: str | None = None,
             labels: dict[str, str] | None = None) -> Sandbox
```

Finds a Sandbox by its ID or name or labels.

**Arguments**:

- `sandbox_id_or_name` _str | None_ - The ID or name of the Sandbox to retrieve.
- `labels` _dict[str, str] | None_ - Labels to filter Sandboxes.
  

**Returns**:

- `Sandbox` - First Sandbox that matches the ID or name or labels.
  

**Raises**:

- `DaytonaError` - If no Sandbox is found.
  

**Example**:

```python
sandbox = daytona.find_one(labels={"my-label": "my-value"})
print(f"Sandbox ID: {sandbox.id} State: {sandbox.state}")
```

#### Daytona.list

```python
@intercept_errors(message_prefix="Failed to list sandboxes: ")
@with_instrumentation()
def list(labels: dict[str, str] | None = None,
         page: int | None = None,
         limit: int | None = None) -> PaginatedSandboxes
```

Returns paginated list of Sandboxes filtered by labels.

**Arguments**:

- `labels` _dict[str, str] | None_ - Labels to filter Sandboxes.
- `page` _int | None_ - Page number for pagination (starting from 1).
- `limit` _int | None_ - Maximum number of items per page.
  

**Returns**:

- `PaginatedSandboxes` - Paginated list of Sandbox instances that match the labels.
  

**Example**:

```python
result = daytona.list(labels={"my-label": "my-value"}, page=2, limit=10)
for sandbox in result.items:
    print(f"{sandbox.id}: {sandbox.state}")
```

#### Daytona.start

```python
@with_instrumentation()
def start(sandbox: Sandbox, timeout: float = 60) -> None
```

Starts a Sandbox and waits for it to be ready.

**Arguments**:

- `sandbox` _Sandbox_ - The Sandbox to start.
- `timeout` _float_ - Optional timeout in seconds to wait for the Sandbox to start.
  0 means no timeout. Default is 60 seconds.
  

**Raises**:

- `DaytonaError` - If timeout is negative; If Sandbox fails to start or times out

#### Daytona.stop

```python
@with_instrumentation()
def stop(sandbox: Sandbox, timeout: float = 60) -> None
```

Stops a Sandbox and waits for it to be stopped.

**Arguments**:

- `sandbox` _Sandbox_ - The sandbox to stop
- `timeout` _float_ - Optional timeout (in seconds) for sandbox stop.
  0 means no timeout. Default is 60 seconds.
  

**Raises**:

- `DaytonaError` - If timeout is negative; If Sandbox fails to stop or times out


## CodeLanguage

```python
class CodeLanguage(str, Enum)
```

Programming languages supported by Daytona

**Enum Members**:
    - `PYTHON` ("python")
    - `TYPESCRIPT` ("typescript")
    - `JAVASCRIPT` ("javascript")

## DaytonaConfig

```python
class DaytonaConfig(BaseModel)
```

Configuration options for initializing the Daytona client.

**Attributes**:

- `api_key` _str | None_ - API key for authentication with the Daytona API. If not set, it must be provided
  via the environment variable `DAYTONA_API_KEY`, or a JWT token must be provided instead.
- `jwt_token` _str | None_ - JWT token for authentication with the Daytona API. If not set, it must be provided
  via the environment variable `DAYTONA_JWT_TOKEN`, or an API key must be provided instead.
- `organization_id` _str | None_ - Organization ID used for JWT-based authentication. Required if a JWT token
  is provided, and must be set either here or in the environment variable `DAYTONA_ORGANIZATION_ID`.
- `api_url` _str | None_ - URL of the Daytona API. Defaults to `'https://app.daytona.io/api'` if not set
  here or in the environment variable `DAYTONA_API_URL`.
- `server_url` _str | None_ - Deprecated. Use `api_url` instead. This property will be removed
  in a future version.
- `target` _str | None_ - Target runner location for the Sandbox. Default region for the organization is used
  if not set here or in the environment variable `DAYTONA_TARGET`.
- `_experimental` _dict[str, any] | None_ - Configuration for experimental features.
  

**Example**:

```python
config = DaytonaConfig(api_key="your-api-key")
```
```python
config = DaytonaConfig(jwt_token="your-jwt-token", organization_id="your-organization-id")
```

## CreateSandboxBaseParams

```python
class CreateSandboxBaseParams(BaseModel)
```

Base parameters for creating a new Sandbox.

**Attributes**:

- `name` _str | None_ - Name of the Sandbox.
- `language` _CodeLanguage | CodeLanguageLiteral | None_ - Programming language for the Sandbox.
  Defaults to "python".
- `os_user` _str | None_ - OS user for the Sandbox.
- `env_vars` _dict[str, str] | None_ - Environment variables to set in the Sandbox.
- `labels` _dict[str, str] | None_ - Custom labels for the Sandbox.
- `public` _bool | None_ - Whether the Sandbox should be public.
- `timeout` _float | None_ - Timeout in seconds for Sandbox to be created and started.
- `auto_stop_interval` _int | None_ - Interval in minutes after which Sandbox will
  automatically stop if no Sandbox event occurs during that time. Default is 15 minutes.
  0 means no auto-stop.
- `auto_archive_interval` _int | None_ - Interval in minutes after which a continuously stopped Sandbox will
  automatically archive. Default is 7 days.
  0 means the maximum interval will be used.
- `auto_delete_interval` _int | None_ - Interval in minutes after which a continuously stopped Sandbox will
  automatically be deleted. By default, auto-delete is disabled.
  Negative value means disabled, 0 means delete immediately upon stopping.
- `volumes` _list[VolumeMount] | None_ - List of volumes mounts to attach to the Sandbox.
- `network_block_all` _bool | None_ - Whether to block all network access for the Sandbox.
- `network_allow_list` _str | None_ - Comma-separated list of allowed CIDR network addresses for the Sandbox.
- `ephemeral` _bool | None_ - Whether the Sandbox should be ephemeral.
  If True, auto_delete_interval will be set to 0.

## CreateSandboxFromImageParams

```python
class CreateSandboxFromImageParams(CreateSandboxBaseParams)
```

Parameters for creating a new Sandbox from an image.

**Attributes**:

- `image` _str | Image_ - Custom Docker image to use for the Sandbox. If an Image object is provided,
  the image will be dynamically built.
- `resources` _Resources | None_ - Resource configuration for the Sandbox. If not provided, sandbox will
  have default resources.

## CreateSandboxFromSnapshotParams

```python
class CreateSandboxFromSnapshotParams(CreateSandboxBaseParams)
```

Parameters for creating a new Sandbox from a snapshot.

**Attributes**:

- `snapshot` _str | None_ - Name of the snapshot to use for the Sandbox.

