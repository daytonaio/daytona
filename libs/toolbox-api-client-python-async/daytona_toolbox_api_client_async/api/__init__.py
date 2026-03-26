from __future__ import annotations

# flake8: noqa

# import apis into api package
import importlib
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from daytona_toolbox_api_client_async.api.computer_use_api import ComputerUseApi
    from daytona_toolbox_api_client_async.api.file_system_api import FileSystemApi
    from daytona_toolbox_api_client_async.api.git_api import GitApi
    from daytona_toolbox_api_client_async.api.info_api import InfoApi
    from daytona_toolbox_api_client_async.api.interpreter_api import InterpreterApi
    from daytona_toolbox_api_client_async.api.lsp_api import LspApi
    from daytona_toolbox_api_client_async.api.port_api import PortApi
    from daytona_toolbox_api_client_async.api.process_api import ProcessApi
    from daytona_toolbox_api_client_async.api.server_api import ServerApi


_DYNAMIC_IMPORTS: dict[str, str] = {
    "ComputerUseApi": "daytona_toolbox_api_client_async.api.computer_use_api",
    "FileSystemApi": "daytona_toolbox_api_client_async.api.file_system_api",
    "GitApi": "daytona_toolbox_api_client_async.api.git_api",
    "InfoApi": "daytona_toolbox_api_client_async.api.info_api",
    "InterpreterApi": "daytona_toolbox_api_client_async.api.interpreter_api",
    "LspApi": "daytona_toolbox_api_client_async.api.lsp_api",
    "PortApi": "daytona_toolbox_api_client_async.api.port_api",
    "ProcessApi": "daytona_toolbox_api_client_async.api.process_api",
    "ServerApi": "daytona_toolbox_api_client_async.api.server_api",

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
