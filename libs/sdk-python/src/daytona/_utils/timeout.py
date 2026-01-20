# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import asyncio
import concurrent.futures
import functools
import inspect
from collections.abc import Awaitable, Callable
from typing import TypeVar, cast

from typing_extensions import ParamSpec

from ..common.errors import DaytonaError

P = ParamSpec("P")
R = TypeVar("R")  # return type (may be coroutine result or regular)

C = TypeVar("C", bound=Callable[..., object])


def with_timeout(
    error_message: Callable[[object | None, float], str] | None = None,
) -> Callable[[C], C]:
    """Decorator to add a timeout mechanism with an optional custom error message.

    Args:
        error_message (Callable[[Any, float], str] | None): A callable that accepts `self` and `timeout`,
                                                               and returns a string error message.
    """

    def decorator(func: C) -> C:
        typed_func = cast(Callable[..., object], func)

        # pull out `self` and `timeout` from args/kwargs
        def _extract(args: tuple[object, ...], kwargs: dict[str, object]) -> tuple[object | None, float | None]:
            names = typed_func.__code__.co_varnames[: typed_func.__code__.co_argcount]
            bound: dict[str, object] = dict(zip(names, args))
            self_inst = args[0] if args else None
            raw_timeout = kwargs.get("timeout", bound.get("timeout"))
            timeout: float | None
            if raw_timeout is None:
                timeout = None
            elif isinstance(raw_timeout, (int, float)):
                timeout = float(raw_timeout)
            else:
                raise DaytonaError("Timeout must be a number or None.")
            return self_inst, timeout

        # produce the final TimeoutError message
        def _format_msg(self_inst: object | None, timeout: float) -> str:
            return (
                error_message(self_inst, timeout)
                if error_message
                else f"Function '{func.__name__}' exceeded timeout of {timeout} seconds."
            )

        if inspect.iscoroutinefunction(func):
            async_func = cast(Callable[..., Awaitable[object]], func)

            @functools.wraps(func)
            async def async_wrapper(*args: P.args, **kwargs: P.kwargs) -> object:
                self_inst, timeout = _extract(args, kwargs)
                if timeout is None or timeout == 0:
                    return await async_func(*args, **kwargs)
                if timeout < 0:
                    raise DaytonaError("Timeout must be a non-negative number or None.")

                try:
                    return await asyncio.wait_for(async_func(*args, **kwargs), timeout)
                except asyncio.TimeoutError:
                    raise TimeoutError(_format_msg(self_inst, timeout))  # pylint: disable=raise-missing-from

            return cast(C, async_wrapper)

        @functools.wraps(func)
        def sync_wrapper(*args: P.args, **kwargs: P.kwargs) -> object:
            self_inst, timeout = _extract(args, kwargs)
            if timeout is None or timeout == 0:
                return typed_func(*args, **kwargs)
            if timeout < 0:
                raise DaytonaError("Timeout must be a non-negative number or None.")

            with concurrent.futures.ThreadPoolExecutor() as executor:
                future: concurrent.futures.Future[object] = executor.submit(typed_func, *args, **kwargs)
                try:
                    return future.result(timeout=timeout)
                except concurrent.futures.TimeoutError:
                    raise TimeoutError(_format_msg(self_inst, timeout))  # pylint: disable=raise-missing-from

        return cast(C, sync_wrapper)

    return decorator
