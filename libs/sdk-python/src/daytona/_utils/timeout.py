# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import asyncio
import contextvars
import ctypes
import functools
import inspect
import logging
import math
import signal
import threading
import time
from types import FrameType, TracebackType
from typing import Any, Callable, TypeVar, cast

from ..common.errors import DaytonaError, DaytonaTimeoutError

F = TypeVar("F", bound=Callable[..., Any])

logger = logging.getLogger(__name__)

# Context-local storage for nested async timeouts (similar to _signal_timeout_stack)
# Using contextvars ensures proper isolation across async contexts
_async_timeout_stack: contextvars.ContextVar[list[dict[str, Any]]] = contextvars.ContextVar(
    "_async_timeout_stack", default=[]
)


class _TimeoutMarker(BaseException):
    """Internal marker exception for threading timeout.

    This is raised via PyThreadState_SetAsyncExc and then caught and converted
    to DaytonaTimeoutError with a proper error message. We use BaseException
    (not Exception) to ensure it's not accidentally caught by generic exception handlers.
    """


def _async_raise(target_tid: int, exception: type) -> bool:
    """Raises an exception asynchronously in another thread.

    This uses the CPython API PyThreadState_SetAsyncExc to raise an exception
    in a different thread. This allows interrupting blocking operations.

    Args:
        target_tid: Target thread identifier
        exception: Exception class to be raised in that thread

    Returns:
        True if successful, False otherwise

    Note:
        Requires Python 3.7+ where thread IDs are unsigned long.
        This is an undocumented CPython API and may fail in certain scenarios.
    """
    try:
        ret = ctypes.pythonapi.PyThreadState_SetAsyncExc(ctypes.c_ulong(target_tid), ctypes.py_object(exception))
        if ret == 0:
            # Thread ID is invalid or thread has already exited
            logger.debug(f"Failed to raise exception in thread {target_tid}: thread not found")
            return False
        if ret > 1:
            # Multiple threads affected - this should never happen but we handle it
            ctypes.pythonapi.PyThreadState_SetAsyncExc(ctypes.c_ulong(target_tid), None)
            logger.error(f"PyThreadState_SetAsyncExc affected {ret} threads, expected 1")
            return False
        return True
    except Exception as e:
        logger.debug(f"Failed to raise exception in thread {target_tid}: {e}")
        return False


def _round_up_to_nearest_second(seconds: float) -> int:
    return int(seconds) + (1 if seconds % 1 > 0 else 0)


def http_timeout(timeout: float | None) -> float | None:
    """Calculate HTTP client timeout to prevent race condition with decorator timeout.

    For sub-second timeouts, uses ceil() to match signal.alarm rounding behavior.
    For timeouts >= 1 second, returns unchanged (natural execution overhead is sufficient).

    This prevents a race condition where:
    - User specifies timeout=0.2s
    - Decorator rounds to 1s (signal.alarm limitation)
    - HTTP client times out at 0.2s first → raises DaytonaError (wrong!)

    With this fix:
    - timeout=0.2 → HTTP timeout: 1s (matches decorator) → DaytonaTimeoutError
    - timeout=5.0 → HTTP timeout: 5s (unchanged) → DaytonaTimeoutError

    Args:
        timeout: User's desired timeout in seconds, or None for no timeout

    Returns:
        - None if timeout is None or 0
        - ceil(timeout) if timeout < 1 (matches signal rounding)
        - timeout unchanged if timeout >= 1 (natural overhead is sufficient)

    Example:
        >>> http_timeout(0.2)
        1
        >>> http_timeout(0.9)
        1
        >>> http_timeout(5.0)
        5.0
        >>> http_timeout(None)
        None
    """
    if not timeout:
        return None
    return math.ceil(timeout) if timeout < 1 else timeout


class _ThreadingTimeout:
    """Context manager for timeout using threading.Timer and async exception raising.

    This implements a timeout mechanism that raises DaytonaTimeoutError in the calling
    thread after a specified duration. It properly executes finally blocks because
    the exception is raised as if it came from within the protected code.

    Nested Timeout Handling:
        Each instance tracks whether IT caused the timeout via the `_timed_out` flag.
        When _TimeoutMarker propagates through nested contexts:
        - The instance that set `_timed_out=True` converts it to DaytonaTimeoutError
        - Other instances let the marker propagate unchanged
        This ensures correct error attribution even with nested timeouts.

    Limitations:
        - **Cannot interrupt blocking C extensions**: The PyThreadState_SetAsyncExc API
          only raises exceptions at Python bytecode boundaries. Code blocked in C extensions
          (e.g., time.sleep(), socket operations, database calls) will not be interrupted
          until control returns to Python. The timeout will fire, but the exception will
          only be raised when the blocking C call completes.
        - Race conditions possible if thread completes just as timeout fires
        - Uses undocumented CPython API (PyThreadState_SetAsyncExc)
    """

    def __init__(self, seconds: float, func_name: str):
        """Initialize the threading timeout.

        Args:
            seconds: Timeout duration in seconds
            func_name: Name of the function being timed (for error messages)
        """
        self.seconds: float = seconds
        self.func_name: str = func_name
        self.target_tid: int | None = threading.current_thread().ident
        self.timer: threading.Timer | None = None
        self._lock: threading.Lock = threading.Lock()
        self._timed_out: bool = False
        self._completed: bool = False

    def _timeout_handler(self) -> None:
        """Called by timer thread when timeout occurs."""
        with self._lock:
            if self._completed:
                # Function completed just before timeout, don't interrupt
                return
            self._timed_out = True

        # Raise exception in target thread
        if self.target_tid:
            success = _async_raise(self.target_tid, _TimeoutMarker)
            if not success:
                logger.warning(
                    "Timeout occurred for '%s' but could not interrupt thread. "
                    + "This may happen if the function completed or is blocked in a C extension.",
                    self.func_name,
                )

    def __enter__(self) -> "_ThreadingTimeout":
        """Start the timeout timer."""
        self.timer = threading.Timer(self.seconds, self._timeout_handler)
        self.timer.daemon = True  # Don't prevent program exit
        self.timer.start()
        return self

    def __exit__(
        self,
        exc_type: type[BaseException] | None,
        exc_val: BaseException | None,
        exc_tb: TracebackType | None,
    ) -> bool:
        """Stop the timeout timer and handle timeout exception."""
        # Mark as completed and cancel timer
        with self._lock:
            self._completed = True
            timed_out = self._timed_out

        if self.timer:
            self.timer.cancel()
            # Wait briefly for timer thread to finish if it was running
            self.timer.join(timeout=0.1)

        # If we timed out via our marker exception, convert to DaytonaTimeoutError
        if exc_type is _TimeoutMarker:
            if timed_out:
                raise DaytonaTimeoutError(
                    f"Function '{self.func_name}' exceeded timeout of {self.seconds} seconds."
                ) from None
            # If we got _TimeoutMarker but we didn't set it, something else raised it
            # Let it propagate as-is

        return False  # Don't suppress any exceptions


# Thread-local storage for nested signal timeouts
_signal_timeout_stack = threading.local()


class _SignalTimeout:
    """Context manager for signal-based timeout with support for nested timeouts.

    This properly handles nested SIGALRM timeouts by maintaining a stack of handlers
    and only setting the alarm for the shortest remaining timeout.

    Note:
        SIGALRM signals can only be received in the main thread. This class should
        only be used when `threading.current_thread() is threading.main_thread()`.

    Nested Timeout Handling:
        When entering a nested timeout, the alarm is set for the minimum of:
        - The new timeout's duration
        - The remaining time of any outer timeout

        When the alarm fires, the handler checks all timeouts on the stack to
        determine which one actually expired, ensuring correct error attribution.
    """

    def __init__(self, seconds: float, func_name: str):
        """Initialize signal timeout.

        Args:
            seconds: Timeout duration in seconds
            func_name: Name of the function being timed (for error messages)
        """
        self.seconds: float = seconds
        self.func_name: str = func_name
        self.old_handler: signal.Handlers | Callable[[int, FrameType | None], Any] | int | None = None
        self.start_time: float | None = None

    def _timeout_handler(self, _signum: int, _frame: FrameType | None) -> None:
        """Signal handler that raises timeout exception.

        When nested timeouts exist, checks which timeout actually expired
        (could be an outer timeout that should have fired first).
        """
        # Check all timeouts on stack to find which one actually expired
        if hasattr(_signal_timeout_stack, "stack"):
            current_time = time.time()
            for item in _signal_timeout_stack.stack:
                elapsed = current_time - item.start_time
                if elapsed >= item.seconds:
                    raise DaytonaTimeoutError(
                        f"Function '{item.func_name}' exceeded timeout of {item.seconds} seconds."
                    )
        # Fallback - raise our own timeout
        raise DaytonaTimeoutError(f"Function '{self.func_name}' exceeded timeout of {self.seconds} seconds.")

    def __enter__(self) -> "_SignalTimeout":
        """Set up the signal handler and alarm."""
        # Initialize stack if needed
        if not hasattr(_signal_timeout_stack, "stack"):
            _signal_timeout_stack.stack = []

        self.start_time = time.time()
        self.old_handler = signal.signal(signal.SIGALRM, self._timeout_handler)

        # Add ourselves to the stack
        _ = _signal_timeout_stack.stack.append(self)

        # Calculate the minimum remaining time across all timeouts on the stack
        # This ensures outer timeouts with shorter remaining time still fire first
        min_remaining = self.seconds
        stack = cast(list["_SignalTimeout"], _signal_timeout_stack.stack)
        for item in stack[:-1]:  # All outer timeouts
            assert self.start_time is not None and item.start_time is not None
            item_elapsed = self.start_time - item.start_time
            item_remaining = item.seconds - item_elapsed
            if item_remaining > 0:
                min_remaining = min(min_remaining, item_remaining)

        # Set alarm for the shortest timeout
        # Round up to avoid sub-second precision issues
        alarm_seconds = _round_up_to_nearest_second(min_remaining)
        _ = signal.alarm(alarm_seconds)

        return self

    def __exit__(
        self, exc_type: type[BaseException] | None, exc_val: BaseException | None, exc_tb: TracebackType | None
    ) -> bool:
        """Restore previous signal handler and alarm."""
        # Cancel current alarm
        _ = signal.alarm(0)

        # Remove ourselves from the stack
        if hasattr(_signal_timeout_stack, "stack") and _signal_timeout_stack.stack:
            _ = _signal_timeout_stack.stack.pop()

        # Restore the signal handler
        if self.old_handler is not None:
            _ = signal.signal(signal.SIGALRM, self.old_handler)

        # If there are outer timeouts, restore their alarm
        # Note: If remaining <= 0, the outer timeout would have already fired during
        # the inner function's execution (since we set alarm for min remaining time).
        # So we only need to restore alarm if there's still time remaining.
        if hasattr(_signal_timeout_stack, "stack") and _signal_timeout_stack.stack:
            outer = _signal_timeout_stack.stack[-1]
            elapsed = time.time() - outer.start_time
            remaining = outer.seconds - elapsed
            if remaining > 0:
                alarm_seconds = _round_up_to_nearest_second(remaining)
                _ = signal.alarm(alarm_seconds)

        return False  # Don't suppress exceptions


def with_timeout() -> Callable[[F], F]:
    """Decorator to add timeout mechanism that executes finally blocks properly.

    This decorator ensures that finally blocks and context managers execute properly
    when a timeout occurs, allowing for proper resource cleanup. The `DaytonaTimeoutError` is
    raised as if it originated from within the decorated function's workflow.

    Wrapped method must have parameter `timeout` which is the timeout duration in seconds.

    Platform Support:
        - **Async functions**: All platforms (uses asyncio task cancellation). The exception that can
            be caught within the function is `asyncio.CancelledError`, since the task is cancelled;
            this wrapper converts it to DaytonaTimeoutError only after the wrapped method returns.
        - **Sync functions (Unix/Linux, main thread)**: Uses SIGALRM signal. Throws `DaytonaTimeoutError`
            directly inside the wrapped method.
        - **Sync functions (Windows or threads)**: Uses threading.Timer with async exception raising.
            Throws `DaytonaTimeoutError` directly inside the wrapped method.

    Behavior:
        - **Finally blocks**: Execute properly on timeout for both sync and async
        - **Context managers**: __exit__ methods are called with the timeout exception
        - **Resource cleanup**: Guaranteed to execute cleanup code
        - **Nested timeouts**: Supported on all platforms with correct error attribution

    Nested Timeout Handling (All Platforms):
        When timeouts are nested, the decorator correctly identifies which function's timeout
        actually expired:
        - **Async**: Uses context-local stack to track active timeouts, checks which expired
        - **Signal**: Uses thread-local stack, handler checks all timeouts to find expired one
        - **Threading**: Each instance tracks its own timeout flag via _timed_out

        Example: If outer(timeout=5) calls inner(timeout=10) and 5 seconds elapse, the error
        will correctly report "outer" exceeded timeout, not "inner".

    Limitations:
        - **Async with blocking code**: Cannot interrupt blocking operations like time.sleep().
            Use proper async code (await asyncio.sleep()) instead.
        - **Threading timeout**: Cannot interrupt code blocked in C extensions. The exception
            will be raised at the next Python bytecode instruction.
        - **Signal precision**: Unix signal-based timeout rounds up to nearest second.

    Returns:
        Decorated function with timeout enforcement.

    Raises:
        DaytonaTimeoutError: When the function exceeds the specified timeout.
        DaytonaError: If timeout is negative.

    Example:
        ```python
        @with_timeout()
        async def create_resource(self, timeout=60):
            resource = None
            try:
                resource = await allocate_resource()
                await resource.initialize()
                return resource
            finally:
                # This cleanup ALWAYS executes, even on timeout
                if resource and not resource.initialized:
                    await resource.cleanup()
        ```

    Example with nested timeouts:
        ```python
        @with_timeout()
        async def outer_operation(self, timeout=30):
            await self.step1(timeout=40)  # Inner timeout
            await self.step2(timeout=50)  # Inner timeout
            # If outer timeout (30s) fires first, both inner timeouts are cancelled
        ```
    """

    def decorator(func: F) -> F:
        # Extract timeout from args/kwargs
        def _extract_timeout(args: tuple[Any, ...], kwargs: dict[str, Any]) -> float | None:
            names = func.__code__.co_varnames[: func.__code__.co_argcount]
            bound = dict(zip(names, args))
            return kwargs.get("timeout", bound.get("timeout", None))

        if inspect.iscoroutinefunction(func):
            # Async function: Use asyncio.wait_for (works on all platforms)
            @functools.wraps(func)
            async def async_wrapper(*args: Any, **kwargs: Any) -> Any:
                timeout = _extract_timeout(args, kwargs)
                if timeout is None or timeout == 0:
                    return await func(*args, **kwargs)
                if timeout < 0:
                    raise DaytonaError("Timeout must be a non-negative number or None.")

                # Add ourselves to the async timeout stack for nested timeout tracking
                stack = _async_timeout_stack.get().copy()
                timeout_info = {"func_name": func.__name__, "timeout": timeout, "start_time": time.time()}
                _ = stack.append(timeout_info)
                _ = _async_timeout_stack.set(stack)

                try:
                    # Use asyncio.wait_for with task cancellation
                    # This executes finally blocks via CancelledError propagation
                    task = asyncio.create_task(func(*args, **kwargs))
                    return await asyncio.wait_for(task, timeout=timeout)
                except asyncio.TimeoutError:
                    # Our own wait_for timed out - check if outer timeout also expired
                    # (could happen if outer and inner timeouts are very close)
                    current_time = time.time()
                    current_stack = _async_timeout_stack.get()
                    for item in current_stack:
                        elapsed = current_time - item["start_time"]
                        if elapsed >= item["timeout"]:
                            # pylint: disable=raise-missing-from
                            raise DaytonaTimeoutError(
                                f"Function '{item['func_name']}' exceeded timeout of {item['timeout']} seconds."
                            )
                    # Our timeout fired (shouldn't reach here but fallback)
                    # pylint: disable=raise-missing-from
                    raise DaytonaTimeoutError(f"Function '{func.__name__}' exceeded timeout of {timeout} seconds.")
                except asyncio.CancelledError:
                    # Task was cancelled - could be from outer timeout or external cancellation
                    # Check all timeouts on stack to determine which one expired
                    current_time = time.time()
                    current_stack = _async_timeout_stack.get()
                    for item in current_stack:
                        elapsed = current_time - item["start_time"]
                        if elapsed >= item["timeout"]:
                            # An outer timeout expired - report it correctly
                            # pylint: disable=raise-missing-from
                            raise DaytonaTimeoutError(
                                f"Function '{item['func_name']}' exceeded timeout of {item['timeout']} seconds."
                            )
                    # Not from a timeout, just a regular cancellation - re-raise as-is
                    raise
                finally:
                    # Remove ourselves from the stack
                    stack = _async_timeout_stack.get().copy()
                    if stack and stack[-1] == timeout_info:
                        _ = stack.pop()
                        _ = _async_timeout_stack.set(stack)

            return async_wrapper  # pyright: ignore[reportReturnType]

        # Sync function: Use best available method
        @functools.wraps(func)
        def sync_wrapper(*args: Any, **kwargs: Any) -> Any:
            timeout = _extract_timeout(args, kwargs)
            if timeout is None or timeout == 0:
                return func(*args, **kwargs)
            if timeout < 0:
                raise DaytonaError("Timeout must be a non-negative number or None.")

            # Strategy 1: Unix/Linux main thread - use signals with nested timeout support
            if hasattr(signal, "SIGALRM") and threading.current_thread() is threading.main_thread():
                with _SignalTimeout(timeout, func.__name__):
                    return func(*args, **kwargs)

            # Strategy 2: Windows or non-main thread - use threading timeout (cross-platform)
            else:
                with _ThreadingTimeout(timeout, func.__name__):
                    return func(*args, **kwargs)

        return sync_wrapper  # pyright: ignore[reportReturnType]

    return decorator
