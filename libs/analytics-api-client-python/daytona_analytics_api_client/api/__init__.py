from __future__ import annotations

# flake8: noqa

# import apis into api package
import importlib
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from daytona_analytics_api_client.api.telemetry_api import TelemetryApi
    from daytona_analytics_api_client.api.usage_api import UsageApi


_DYNAMIC_IMPORTS: dict[str, str] = {
    "TelemetryApi": "daytona_analytics_api_client.api.telemetry_api",
    "UsageApi": "daytona_analytics_api_client.api.usage_api",

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
