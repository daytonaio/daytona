from __future__ import annotations

# flake8: noqa

# import apis into api package
import importlib
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from daytona_api_client.api.health_api import HealthApi
    from daytona_api_client.api.admin_api import AdminApi
    from daytona_api_client.api.api_keys_api import ApiKeysApi
    from daytona_api_client.api.audit_api import AuditApi
    from daytona_api_client.api.config_api import ConfigApi
    from daytona_api_client.api.docker_registry_api import DockerRegistryApi
    from daytona_api_client.api.jobs_api import JobsApi
    from daytona_api_client.api.object_storage_api import ObjectStorageApi
    from daytona_api_client.api.organizations_api import OrganizationsApi
    from daytona_api_client.api.preview_api import PreviewApi
    from daytona_api_client.api.regions_api import RegionsApi
    from daytona_api_client.api.runners_api import RunnersApi
    from daytona_api_client.api.sandbox_api import SandboxApi
    from daytona_api_client.api.snapshots_api import SnapshotsApi
    from daytona_api_client.api.toolbox_api import ToolboxApi
    from daytona_api_client.api.users_api import UsersApi
    from daytona_api_client.api.volumes_api import VolumesApi
    from daytona_api_client.api.webhooks_api import WebhooksApi
    from daytona_api_client.api.workspace_api import WorkspaceApi


_DYNAMIC_IMPORTS: dict[str, str] = {
    "HealthApi": "daytona_api_client.api.health_api",
    "AdminApi": "daytona_api_client.api.admin_api",
    "ApiKeysApi": "daytona_api_client.api.api_keys_api",
    "AuditApi": "daytona_api_client.api.audit_api",
    "ConfigApi": "daytona_api_client.api.config_api",
    "DockerRegistryApi": "daytona_api_client.api.docker_registry_api",
    "JobsApi": "daytona_api_client.api.jobs_api",
    "ObjectStorageApi": "daytona_api_client.api.object_storage_api",
    "OrganizationsApi": "daytona_api_client.api.organizations_api",
    "PreviewApi": "daytona_api_client.api.preview_api",
    "RegionsApi": "daytona_api_client.api.regions_api",
    "RunnersApi": "daytona_api_client.api.runners_api",
    "SandboxApi": "daytona_api_client.api.sandbox_api",
    "SnapshotsApi": "daytona_api_client.api.snapshots_api",
    "ToolboxApi": "daytona_api_client.api.toolbox_api",
    "UsersApi": "daytona_api_client.api.users_api",
    "VolumesApi": "daytona_api_client.api.volumes_api",
    "WebhooksApi": "daytona_api_client.api.webhooks_api",
    "WorkspaceApi": "daytona_api_client.api.workspace_api",

}


def __getattr__(attr_name: str) -> object:
    module_path = _DYNAMIC_IMPORTS.get(attr_name)
    if module_path is None:
        raise AttributeError(f"module {__name__!r} has no attribute {attr_name!r}")
    mod = importlib.import_module(module_path)
    value = getattr(mod, attr_name)
    globals()[attr_name] = value
    return value


def __dir__() -> list[str]:
    return list(_DYNAMIC_IMPORTS.keys())
