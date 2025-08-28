# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import asyncio
import concurrent.futures
import functools
import inspect
from typing import Any, Callable, Optional, ParamSpec, TypeVar

from .._utils.errors import DaytonaError

P = ParamSpec("P")
T = TypeVar("T")


def with_timeout(
    error_message: Optional[Callable[[Any, float], str]] = None,
) -> Callable[[Callable[P, T]], Callable[P, T]]:
    """Decorator to add a timeout mechanism with an optional custom error message.

    Args:
        error_message (Optional[Callable[[Any, float], str]]): A callable that accepts `self` and `timeout`,
                                                               and returns a string error message.
    """

    def decorator(func: Callable[P, T]) -> Callable[P, T]:
        # pull out `self` and `timeout` from args/kwargs
        def _extract(args: tuple, kwargs: dict) -> tuple[Any, Optional[float]]:
            names = func.__code__.co_varnames[: func.__code__.co_argcount]
            bound = dict(zip(names, args))
            self_inst = args[0] if args else None
            return self_inst, kwargs.get("timeout", bound.get("timeout", None))

        # produce the final TimeoutError message
        def _format_msg(self_inst: Any, timeout: float) -> str:
            return (
                error_message(self_inst, timeout)
                if error_message
                else f"Function '{func.__name__}' exceeded timeout of {timeout} seconds."
            )

        if inspect.iscoroutinefunction(func):

            @functools.wraps(func)
            async def async_wrapper(*args: P.args, **kwargs: P.kwargs) -> T:
                self_inst, timeout = _extract(args, kwargs)
                if timeout is None or timeout == 0:
                    return await func(*args, **kwargs)
                if timeout < 0:
                    raise DaytonaError("Timeout must be a non-negative number or None.")

                try:
                    return await asyncio.wait_for(func(*args, **kwargs), timeout)
                except asyncio.TimeoutError:
                    raise TimeoutError(_format_msg(self_inst, timeout))  # pylint: disable=raise-missing-from

            return async_wrapper

        @functools.wraps(func)
        def sync_wrapper(*args: P.args, **kwargs: P.kwargs) -> T:
            self_inst, timeout = _extract(args, kwargs)
            if timeout is None or timeout == 0:
                return func(*args, **kwargs)
            if timeout < 0:
                raise DaytonaError("Timeout must be a non-negative number or None.")

            with concurrent.futures.ThreadPoolExecutor() as executor:
                future = executor.submit(func, *args, **kwargs)
                try:
                    return future.result(timeout=timeout)
                except concurrent.futures.TimeoutError:
                    raise TimeoutError(_format_msg(self_inst, timeout))  # pylint: disable=raise-missing-from

        return sync_wrapper

    return decorator
