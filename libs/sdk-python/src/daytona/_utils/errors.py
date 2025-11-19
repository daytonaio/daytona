# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import functools
import inspect
import json
import sys
from typing import Callable, NoReturn, TypeVar, Union

from daytona_api_client.exceptions import NotFoundException, OpenApiException
from daytona_api_client_async.exceptions import NotFoundException as NotFoundExceptionAsync
from daytona_api_client_async.exceptions import OpenApiException as OpenApiExceptionAsync
from daytona_toolbox_api_client.exceptions import NotFoundException as NotFoundExceptionToolbox
from daytona_toolbox_api_client.exceptions import OpenApiException as OpenApiExceptionToolbox
from daytona_toolbox_api_client_async.exceptions import NotFoundException as NotFoundExceptionToolboxAsync
from daytona_toolbox_api_client_async.exceptions import OpenApiException as OpenApiExceptionToolboxAsync

from ..common.errors import DaytonaError, DaytonaNotFoundError

if sys.version_info >= (3, 10):
    from typing import ParamSpec
else:
    from typing_extensions import ParamSpec

SESSION_IS_CLOSED_ERROR_MESSAGE = "Session is closed"

P = ParamSpec("P")
T = TypeVar("T")


def intercept_errors(
    message_prefix: str = "",
) -> Callable[[Callable[P, T]], Callable[P, T]]:
    """Decorator to intercept errors, process them, and optionally add a message prefix.
    If the error is an OpenApiException, it will be processed to extract the most meaningful error message.

    Args:
        message_prefix (str): Custom message prefix for the error.
    """

    def decorator(func: Callable[P, T]) -> Callable[P, T]:
        def process_n_raise_exception(e: Exception) -> NoReturn:
            if isinstance(e, DaytonaError):
                msg = f"{message_prefix}{str(e)}" if message_prefix else str(e)
                raise e.__class__(msg) from None

            if isinstance(
                e, (OpenApiException, OpenApiExceptionAsync, OpenApiExceptionToolbox, OpenApiExceptionToolboxAsync)
            ):
                msg = _get_open_api_exception_message(e)
                if isinstance(
                    e,
                    (
                        NotFoundException,
                        NotFoundExceptionAsync,
                        NotFoundExceptionToolbox,
                        NotFoundExceptionToolboxAsync,
                    ),
                ):
                    raise DaytonaNotFoundError(f"{message_prefix}{msg}") from None
                raise DaytonaError(f"{message_prefix}{msg}") from None

            if isinstance(e, RuntimeError) and SESSION_IS_CLOSED_ERROR_MESSAGE in str(e):
                raise DaytonaError(
                    f"{message_prefix}{str(e)}: Daytona client is closed"
                    " — sandbox is used outside its parent's context. "
                    "Ensure sandboxes are only used within the scope of their parent Daytona object."
                ) from e

            msg = f"{message_prefix}{str(e)}" if message_prefix else str(e)
            raise DaytonaError(msg)  # pylint: disable=raise-missing-from

        if inspect.iscoroutinefunction(func):

            @functools.wraps(func)
            async def async_wrapper(*args: P.args, **kwargs: P.kwargs) -> T:
                try:
                    return await func(*args, **kwargs)
                except Exception as e:
                    process_n_raise_exception(e)

            return async_wrapper

        @functools.wraps(func)
        def sync_wrapper(*args: P.args, **kwargs: P.kwargs) -> T:
            try:
                return func(*args, **kwargs)
            except Exception as e:
                process_n_raise_exception(e)

        return sync_wrapper

    return decorator


def _get_open_api_exception_message(exception: Union[OpenApiException, OpenApiExceptionAsync]) -> str:
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
    if not hasattr(exception, "body") or not exception.body:
        return str(exception)

    body_str = str(exception.body)
    try:
        data = json.loads(body_str)
        message = data.get("message", body_str) if isinstance(data, dict) else body_str
    except json.JSONDecodeError:
        message = body_str

    return message
