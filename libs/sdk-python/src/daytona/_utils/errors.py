# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0
from __future__ import annotations

import functools
import inspect
import json
from collections.abc import AsyncIterator, Awaitable, Callable, Iterator, Mapping
from typing import Any, NoReturn, TypeVar, Union, cast

import aiohttp
import httpx
import urllib3.exceptions

from daytona_api_client.exceptions import OpenApiException
from daytona_api_client_async.exceptions import OpenApiException as OpenApiExceptionAsync
from daytona_toolbox_api_client.exceptions import OpenApiException as OpenApiExceptionToolbox
from daytona_toolbox_api_client_async.exceptions import OpenApiException as OpenApiExceptionToolboxAsync

from ..common.errors import DaytonaConnectionError, DaytonaConnectionTimeoutError, DaytonaError, _resolve_error_class
from .types import has_body

SESSION_IS_CLOSED_ERROR_MESSAGE = "Session is closed"

F = TypeVar("F", bound=Callable[..., object])

OpenApiDaytonaException = Union[
    OpenApiException,
    OpenApiExceptionAsync,
    OpenApiExceptionToolbox,
    OpenApiExceptionToolboxAsync,
]

OPENAPI_EXCEPTIONS = (OpenApiException, OpenApiExceptionAsync, OpenApiExceptionToolbox, OpenApiExceptionToolboxAsync)

# Order matters: subclasses before their bases. ConnectionError does NOT catch
# the broader OSError family.
TRANSPORT_ERROR_TO_DAYTONA_ERROR: tuple[tuple[type[BaseException], type[DaytonaError]], ...] = (
    (aiohttp.ServerTimeoutError, DaytonaConnectionTimeoutError),
    (urllib3.exceptions.ReadTimeoutError, DaytonaConnectionTimeoutError),
    (urllib3.exceptions.ConnectTimeoutError, DaytonaConnectionTimeoutError),
    (httpx.TimeoutException, DaytonaConnectionTimeoutError),
    (TimeoutError, DaytonaConnectionTimeoutError),
    (aiohttp.ClientConnectorError, DaytonaConnectionError),
    (urllib3.exceptions.NewConnectionError, DaytonaConnectionError),
    (aiohttp.ServerDisconnectedError, DaytonaConnectionError),
    (aiohttp.ClientPayloadError, DaytonaConnectionError),
    (aiohttp.ClientOSError, DaytonaConnectionError),
    (urllib3.exceptions.ProtocolError, DaytonaConnectionError),
    (aiohttp.ClientConnectionError, DaytonaConnectionError),
    (httpx.NetworkError, DaytonaConnectionError),
    (ConnectionError, DaytonaConnectionError),
)


def _prefix_message(message_prefix: str, message: str) -> str:
    if not message_prefix:
        return message

    return f"{message_prefix}{message}"


def _parse_openapi_exception(
    exception: OpenApiDaytonaException,
) -> tuple[str, str | None, str | None]:
    """Extract (message, code, source) from a Daytona wire envelope in the exception body."""
    if not has_body(exception):
        return str(exception), None, None

    body_str: str = str(exception.body)
    message: str = body_str
    code: str | None = None
    source: str | None = None
    try:
        data = json.loads(body_str)
        if isinstance(data, dict):
            typed_data: dict[str, object] = cast(dict[str, object], data)
            msg: object | None = typed_data.get("message")
            if isinstance(msg, str):
                message = msg
            code_value: object | None = typed_data.get("code")
            if isinstance(code_value, str):
                code = code_value
            source_value: object | None = typed_data.get("source")
            if isinstance(source_value, str):
                source = source_value
    except json.JSONDecodeError:
        pass

    return message, code, source


def intercept_errors(
    message_prefix: str = "",
) -> Callable[[F], F]:
    """Decorator translating generated-client and transport errors into DaytonaError.

    All re-raises preserve ``__cause__`` via ``raise ... from e``.
    """

    def decorator(func: F) -> F:
        def process_n_raise_exception(e: Exception) -> NoReturn:
            if isinstance(e, DaytonaError):
                raise e.__class__(
                    _prefix_message(message_prefix, str(e)),
                    status_code=e.status_code,
                    headers=e.headers,
                    code=e.code,
                    source=e.source,
                ) from e

            if isinstance(e, OPENAPI_EXCEPTIONS):
                msg, code, source = _parse_openapi_exception(e)
                status_code = getattr(e, "status", None)
                headers = cast(Mapping[str, Any] | None, getattr(e, "headers", None))
                error_cls = _resolve_error_class(status_code, code, source)
                raise error_cls(
                    _prefix_message(message_prefix, msg),
                    status_code=status_code,
                    headers=headers,
                    code=code,
                    source=source,
                ) from e

            for source_error, daytona_error_cls in TRANSPORT_ERROR_TO_DAYTONA_ERROR:
                if isinstance(e, source_error):
                    raise daytona_error_cls(_prefix_message(message_prefix, str(e))) from e

            if isinstance(e, RuntimeError) and SESSION_IS_CLOSED_ERROR_MESSAGE in str(e):
                raise DaytonaError(
                    (
                        f"{_prefix_message(message_prefix, str(e))}: Daytona client is closed"
                        " — sandbox is used outside its parent's context. "
                        "Ensure sandboxes are only used within the scope of their parent Daytona object."
                    )
                ) from e

            raise DaytonaError(_prefix_message(message_prefix, str(e))) from e

        if inspect.isasyncgenfunction(func):
            async_gen_func = cast(Callable[..., AsyncIterator[Any]], func)

            @functools.wraps(func)
            async def async_gen_wrapper(*args: object, **kwargs: object) -> AsyncIterator[Any]:
                try:
                    async for item in async_gen_func(*args, **kwargs):
                        yield item
                except Exception as e:
                    process_n_raise_exception(e)

            return cast(F, async_gen_wrapper)

        if inspect.isgeneratorfunction(func):
            sync_gen_func = cast(Callable[..., Iterator[Any]], func)

            @functools.wraps(func)
            def sync_gen_wrapper(*args: object, **kwargs: object) -> Iterator[Any]:
                try:
                    yield from sync_gen_func(*args, **kwargs)
                except Exception as e:
                    process_n_raise_exception(e)

            return cast(F, sync_gen_wrapper)

        if inspect.iscoroutinefunction(func):
            async_func = cast(Callable[..., Awaitable[object]], func)

            @functools.wraps(func)
            async def async_wrapper(*args: object, **kwargs: object) -> object:
                try:
                    return await async_func(*args, **kwargs)
                except Exception as e:
                    process_n_raise_exception(e)

            return cast(F, async_wrapper)

        sync_func = cast(Callable[..., object], func)

        @functools.wraps(func)
        def sync_wrapper(*args: object, **kwargs: object) -> object:
            try:
                return sync_func(*args, **kwargs)
            except Exception as e:
                process_n_raise_exception(e)

        return cast(F, sync_wrapper)

    return decorator
