# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import functools
import json
from typing import Callable, ParamSpec, TypeVar

from daytona_api_client.exceptions import OpenApiException

from ..common.errors import DaytonaError

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
        @functools.wraps(func)
        def wrapper(*args: P.args, **kwargs: P.kwargs) -> T:
            try:
                return func(*args, **kwargs)
            except OpenApiException as e:
                message = _get_open_api_exception_message(e)

                raise DaytonaError(f"{message_prefix}{message}") from None
            except Exception as e:
                if message_prefix:
                    message = f"{message_prefix}{str(e)}"
                    raise DaytonaError(message)  # pylint: disable=raise-missing-from
                raise DaytonaError(str(e))  # pylint: disable=raise-missing-from

        return wrapper

    return decorator


def _get_open_api_exception_message(exception: OpenApiException) -> str:
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
