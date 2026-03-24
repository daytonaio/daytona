# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0
from __future__ import annotations

import functools
import inspect
import json
from collections.abc import Awaitable, Callable, Mapping
from typing import Any, NoReturn, TypeVar, cast

import httpx
from daytona_api_client.exceptions import (
    BadRequestException,
    ConflictException,
    ForbiddenException,
    NotFoundException,
    OpenApiException,
    UnauthorizedException,
)
from daytona_api_client_async.exceptions import BadRequestException as BadRequestExceptionAsync
from daytona_api_client_async.exceptions import ConflictException as ConflictExceptionAsync
from daytona_api_client_async.exceptions import ForbiddenException as ForbiddenExceptionAsync
from daytona_api_client_async.exceptions import NotFoundException as NotFoundExceptionAsync
from daytona_api_client_async.exceptions import OpenApiException as OpenApiExceptionAsync
from daytona_api_client_async.exceptions import UnauthorizedException as UnauthorizedExceptionAsync
from daytona_toolbox_api_client.exceptions import BadRequestException as BadRequestExceptionToolbox
from daytona_toolbox_api_client.exceptions import ConflictException as ConflictExceptionToolbox
from daytona_toolbox_api_client.exceptions import ForbiddenException as ForbiddenExceptionToolbox
from daytona_toolbox_api_client.exceptions import NotFoundException as NotFoundExceptionToolbox
from daytona_toolbox_api_client.exceptions import OpenApiException as OpenApiExceptionToolbox
from daytona_toolbox_api_client.exceptions import UnauthorizedException as UnauthorizedExceptionToolbox
from daytona_toolbox_api_client_async.exceptions import BadRequestException as BadRequestExceptionToolboxAsync
from daytona_toolbox_api_client_async.exceptions import ConflictException as ConflictExceptionToolboxAsync
from daytona_toolbox_api_client_async.exceptions import ForbiddenException as ForbiddenExceptionToolboxAsync
from daytona_toolbox_api_client_async.exceptions import NotFoundException as NotFoundExceptionToolboxAsync
from daytona_toolbox_api_client_async.exceptions import OpenApiException as OpenApiExceptionToolboxAsync
from daytona_toolbox_api_client_async.exceptions import UnauthorizedException as UnauthorizedExceptionToolboxAsync

from ..common.errors import (
    DaytonaAuthenticationError,
    DaytonaAuthorizationError,
    DaytonaConflictError,
    DaytonaConnectionError,
    DaytonaError,
    DaytonaNotFoundError,
    DaytonaRateLimitError,
    DaytonaTimeoutError,
    DaytonaValidationError,
)
from .types import has_body

SESSION_IS_CLOSED_ERROR_MESSAGE = "Session is closed"

F = TypeVar("F", bound=Callable[..., object])

STATUS_CODE_TO_ERROR: dict[int, type[DaytonaError]] = {
    400: DaytonaValidationError,
    401: DaytonaAuthenticationError,
    403: DaytonaAuthorizationError,
    404: DaytonaNotFoundError,
    409: DaytonaConflictError,
    429: DaytonaRateLimitError,
}


def intercept_errors(
    message_prefix: str = "",
) -> Callable[[F], F]:
    """Decorator to intercept errors, process them, and optionally add a message prefix.
    If the error is an OpenApiException, it will be processed to extract the most meaningful error message.

    Args:
        message_prefix (str): Custom message prefix for the error.
    """

    def decorator(func: F) -> F:
        def process_n_raise_exception(e: Exception) -> NoReturn:
            if isinstance(e, DaytonaError):
                msg = f"{message_prefix}{str(e)}" if message_prefix else str(e)
                raise e.__class__(msg, status_code=e.status_code, headers=e.headers, error_code=e.error_code) from None

            if isinstance(
                e, (OpenApiException, OpenApiExceptionAsync, OpenApiExceptionToolbox, OpenApiExceptionToolboxAsync)
            ):
                msg, error_code = _get_open_api_exception_message(e)
                status_code = getattr(e, "status", None)
                headers = cast(Mapping[str, Any] | None, getattr(e, "headers", None))

                raise create_daytona_error(
                    f"{message_prefix}{msg}",
                    status_code=status_code,
                    headers=headers,
                    error_code=error_code,
                    exception=e,
                ) from None

            # Preserve typed transport failures from the manual httpx streaming paths.
            if isinstance(e, httpx.TimeoutException):
                msg = f"{message_prefix}{str(e)}" if message_prefix else str(e)
                raise DaytonaTimeoutError(msg) from None

            if isinstance(e, httpx.NetworkError):
                msg = f"{message_prefix}{str(e)}" if message_prefix else str(e)
                raise DaytonaConnectionError(msg) from None

            if isinstance(e, TimeoutError):
                msg = f"{message_prefix}{str(e)}" if message_prefix else str(e)
                raise DaytonaTimeoutError(msg) from None

            # Network/connection errors (ConnectionError covers ConnectionRefusedError,
            # ConnectionResetError, etc. — but not the broader OSError which includes
            # local filesystem errors like PermissionError and FileNotFoundError)
            if isinstance(e, ConnectionError):
                msg = f"{message_prefix}{str(e)}" if message_prefix else str(e)
                raise DaytonaConnectionError(msg) from None

            if isinstance(e, RuntimeError) and SESSION_IS_CLOSED_ERROR_MESSAGE in str(e):
                raise DaytonaError(
                    (
                        f"{message_prefix}{str(e)}: Daytona client is closed"
                        " — sandbox is used outside its parent's context. "
                        "Ensure sandboxes are only used within the scope of their parent Daytona object."
                    )
                ) from e

            msg = f"{message_prefix}{str(e)}" if message_prefix else str(e)
            raise DaytonaError(msg)  # pylint: disable=raise-missing-from

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


def _map_api_exception_to_error(
    e: OpenApiException | OpenApiExceptionAsync | OpenApiExceptionToolbox | OpenApiExceptionToolboxAsync,
    status_code: int | None,
) -> type[DaytonaError]:
    """Map an OpenAPI exception to the appropriate DaytonaError subclass."""
    # Map by exception type first (most reliable)
    if isinstance(
        e,
        (NotFoundException, NotFoundExceptionAsync, NotFoundExceptionToolbox, NotFoundExceptionToolboxAsync),
    ):
        return DaytonaNotFoundError

    if isinstance(
        e,
        (
            UnauthorizedException,
            UnauthorizedExceptionAsync,
            UnauthorizedExceptionToolbox,
            UnauthorizedExceptionToolboxAsync,
        ),
    ):
        return DaytonaAuthenticationError

    if isinstance(
        e,
        (ForbiddenException, ForbiddenExceptionAsync, ForbiddenExceptionToolbox, ForbiddenExceptionToolboxAsync),
    ):
        return DaytonaAuthorizationError

    if isinstance(
        e,
        (BadRequestException, BadRequestExceptionAsync, BadRequestExceptionToolbox, BadRequestExceptionToolboxAsync),
    ):
        return DaytonaValidationError

    if isinstance(
        e,
        (ConflictException, ConflictExceptionAsync, ConflictExceptionToolbox, ConflictExceptionToolboxAsync),
    ):
        return DaytonaConflictError

    return error_class_from_status_code(status_code)


def error_class_from_status_code(status_code: int | None) -> type[DaytonaError]:
    """Map an HTTP status code to the corresponding DaytonaError subclass."""

    if status_code is None:
        return DaytonaError

    return STATUS_CODE_TO_ERROR.get(status_code, DaytonaError)


def create_daytona_error(
    message: str,
    status_code: int | None = None,
    headers: Mapping[str, Any] | None = None,
    error_code: str | None = None,
    exception: OpenApiException
    | OpenApiExceptionAsync
    | OpenApiExceptionToolbox
    | OpenApiExceptionToolboxAsync
    | None = None,
) -> DaytonaError:
    """Create the appropriate DaytonaError subclass from structured error metadata."""

    error_cls = (
        _map_api_exception_to_error(exception, status_code) if exception else error_class_from_status_code(status_code)
    )
    return error_cls(message, status_code=status_code, headers=headers, error_code=error_code)


def _get_open_api_exception_message(
    exception: OpenApiException | OpenApiExceptionAsync | OpenApiExceptionToolbox | OpenApiExceptionToolboxAsync,
) -> tuple[str, str | None]:
    """Process API exceptions to extract the most meaningful error message and error code.

    This method examines the exception's body attribute and attempts to extract
    the most informative error message using the following logic:
    1. If the body is missing or empty, returns the original exception
    2. If the body contains valid JSON with a 'message' field, uses that message
    3. If the body is not valid JSON or does not contain a 'message' field, uses the raw body string

    Args:
        exception (OpenApiException): The OpenApiException to process

    Returns:
        Tuple of (message, error_code). error_code is None if not present in the response.
    """
    if not has_body(exception):
        return str(exception), None

    body_str: str = str(exception.body)
    message: str = body_str
    error_code: str | None = None
    try:
        data = json.loads(body_str)
        if isinstance(data, dict):
            typed_data: dict[str, object] = cast(dict[str, object], data)
            msg: object | None = typed_data.get("message")
            if isinstance(msg, str):
                message = msg
            code: object | None = typed_data.get("error") or typed_data.get("code") or typed_data.get("error_code")
            if isinstance(code, str):
                error_code = code
    except json.JSONDecodeError:
        pass

    return message, error_code
