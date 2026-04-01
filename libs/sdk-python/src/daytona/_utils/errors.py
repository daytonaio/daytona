# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0
from __future__ import annotations

import functools
import inspect
import json
from collections.abc import Awaitable, Callable, Mapping
from typing import Any, NoReturn, TypeVar, cast

from daytona_api_client.exceptions import NotFoundException, OpenApiException
from daytona_api_client_async.exceptions import NotFoundException as NotFoundExceptionAsync
from daytona_api_client_async.exceptions import OpenApiException as OpenApiExceptionAsync
from daytona_toolbox_api_client.exceptions import NotFoundException as NotFoundExceptionToolbox
from daytona_toolbox_api_client.exceptions import OpenApiException as OpenApiExceptionToolbox
from daytona_toolbox_api_client_async.exceptions import NotFoundException as NotFoundExceptionToolboxAsync
from daytona_toolbox_api_client_async.exceptions import OpenApiException as OpenApiExceptionToolboxAsync

from ..common.errors import (
    DaytonaAuthenticationError,
    DaytonaBadRequestError,
    DaytonaConflictError,
    DaytonaConnectionError,
    DaytonaError,
    DaytonaForbiddenError,
    DaytonaNotFoundError,
    DaytonaRateLimitError,
    DaytonaServerError,
    DaytonaValidationError,
)
from .types import has_body

SESSION_IS_CLOSED_ERROR_MESSAGE = "Session is closed"

F = TypeVar("F", bound=Callable[..., object])


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
                raise e.__class__(msg, status_code=e.status_code, headers=e.headers) from None

            if isinstance(
                e, (OpenApiException, OpenApiExceptionAsync, OpenApiExceptionToolbox, OpenApiExceptionToolboxAsync)
            ):
                msg = _get_open_api_exception_message(e)
                status_code = getattr(e, "status", None)
                headers = cast(Mapping[str, Any] | None, getattr(e, "headers", None))

                if isinstance(
                    e,
                    (
                        NotFoundException,
                        NotFoundExceptionAsync,
                        NotFoundExceptionToolbox,
                        NotFoundExceptionToolboxAsync,
                    ),
                ):
                    raise DaytonaNotFoundError(
                        f"{message_prefix}{msg}", status_code=status_code, headers=headers
                    ) from None

                raise _map_status_code(status_code, f"{message_prefix}{msg}", headers=headers) from None

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


def _map_status_code(
    status_code: int | None,
    message: str,
    headers: Mapping[str, Any] | None = None,
) -> DaytonaError:
    """Map an HTTP status code to the appropriate DaytonaError subclass.

    Args:
        status_code: HTTP status code, or None / 0 for network-level errors.
        message: Error message.
        headers: Response headers if available.

    Returns:
        The most specific DaytonaError subclass for the given status code.
    """
    if status_code == 400:
        return DaytonaBadRequestError(message, status_code=status_code, headers=headers)
    if status_code == 401:
        return DaytonaAuthenticationError(message, status_code=status_code, headers=headers)
    if status_code == 403:
        return DaytonaForbiddenError(message, status_code=status_code, headers=headers)
    if status_code == 404:
        return DaytonaNotFoundError(message, status_code=status_code, headers=headers)
    if status_code == 409:
        return DaytonaConflictError(message, status_code=status_code, headers=headers)
    if status_code == 422:
        return DaytonaValidationError(message, status_code=status_code, headers=headers)
    if status_code == 429:
        return DaytonaRateLimitError(message, status_code=status_code, headers=headers)
    if status_code is not None and status_code >= 500:
        return DaytonaServerError(message, status_code=status_code, headers=headers)
    if not status_code:
        return DaytonaConnectionError(message, status_code=None, headers=headers)
    return DaytonaError(message, status_code=status_code, headers=headers)


def _get_open_api_exception_message(
    exception: OpenApiException | OpenApiExceptionAsync | OpenApiExceptionToolbox | OpenApiExceptionToolboxAsync,
) -> str:
    """Process API exceptions to extract the most meaningful error message.

    This method examines the exception's body attribute and attempts to extract
    the most informative error message using the following logic:
    1. If the body is missing or empty, returns the original exception
    2. If the body contains valid JSON with a 'message' field, uses that message
    3. If the body is not valid JSON or does not contain a 'message' field, uses the raw body string

    Args:
        exception (OpenApiException): The OpenApiException to process

    Returns:
        Processed message
    """
    if not has_body(exception):
        return str(exception)

    body_str: str = str(exception.body)
    message: str = body_str
    try:
        data = json.loads(body_str)
        if isinstance(data, dict):
            typed_data: dict[str, object] = cast(dict[str, object], data)
            msg: object | None = typed_data.get("message")
            if isinstance(msg, str):
                message = msg
    except json.JSONDecodeError:
        pass

    return message
