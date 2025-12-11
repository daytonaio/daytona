# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

"""OpenTelemetry instrumentation decorators for tracing and metrics."""

import functools
import time
from typing import Any, Callable, Dict, Optional, TypeVar, cast

from opentelemetry import metrics, trace
from opentelemetry.trace import Status, StatusCode

# Type variable for generic function decoration
F = TypeVar("F", bound=Callable[..., Any])

# Lazy initialization to ensure SDK is started before getting tracer/meter
_tracer = None
_meter = None
_execution_histograms: Dict[str, Any] = {}


def get_tracer():
    """Get or create the tracer instance."""
    global _tracer  # pylint: disable=global-statement
    if _tracer is None:
        _tracer = trace.get_tracer("")
    return _tracer


def get_meter():
    """Get or create the meter instance."""
    global _meter  # pylint: disable=global-statement
    if _meter is None:
        _meter = metrics.get_meter("")
    return _meter


def to_snake_case(string: str) -> str:
    """Converts a string to snake_case for Prometheus-friendly metric names."""
    result = ""
    for i, char in enumerate(string):
        if char.isupper() and i > 0:
            result += "_"
        result += char.lower()
    return result.replace(".", "_")


def with_span(
    name: Optional[str] = None,
    attributes: Optional[Dict[str, str]] = None,
) -> Callable[[F], F]:
    """Decorator for instrumenting methods with OpenTelemetry spans (traces only).

    Args:
        name: Custom name for the span. If not provided, uses `ClassName.methodName` format
        attributes: Additional attributes to attach to the span

    Returns:
        Decorated function with span instrumentation

    Example:
        ```python
        @with_span(name="custom_operation", attributes={"custom": "value"})
        async def my_method(self):
            pass
        ```
    """
    attrs = attributes or {}

    def decorator(func: F) -> F:
        @functools.wraps(func)
        async def wrapper(*args, **kwargs):
            # Get class name if this is a method
            class_name = args[0].__class__.__name__ if args and hasattr(args[0], "__class__") else ""
            method_name = func.__name__

            span_name = name or f"{class_name}.{method_name}" if class_name else method_name

            all_attributes = {
                "component": class_name,
                "method": method_name,
                **attrs,
            }

            tracer = get_tracer()
            with tracer.start_as_current_span(span_name, attributes=all_attributes) as span:
                try:
                    result = await func(*args, **kwargs)
                    span.set_status(Status(StatusCode.OK))
                    return result
                except Exception as error:
                    span.set_status(Status(StatusCode.ERROR, str(error)))
                    span.record_exception(error)
                    raise

        return cast(F, wrapper)

    return decorator


def with_metric(
    name: Optional[str] = None,
    description: Optional[str] = None,
    labels: Optional[Dict[str, str]] = None,
) -> Callable[[F], F]:
    """Decorator for instrumenting methods with OpenTelemetry metrics (metrics only).

    Collects histogram metric:
    - Histogram: `{name}_duration` - tracks execution duration in milliseconds

    Args:
        name: Custom name for the metric. If not provided, uses `ClassName.methodName` format
        description: Description for the metrics being collected
        labels: Additional labels to attach to the metrics

    Returns:
        Decorated function with metric instrumentation

    Example:
        ```python
        @with_metric(name="custom_operation", description="Custom operation duration")
        async def my_method(self):
            pass
        ```
    """
    metric_labels = labels or {}

    def decorator(func: F) -> F:
        @functools.wraps(func)
        async def wrapper(*args, **kwargs):
            # Get class name if this is a method
            class_name = args[0].__class__.__name__ if args and hasattr(args[0], "__class__") else ""
            method_name = func.__name__

            metric_name = to_snake_case(name or f"{class_name}.{method_name}" if class_name else method_name)

            all_labels = {
                "component": class_name,
                "method": method_name,
                **metric_labels,
            }

            # Get or create histogram for this method
            if metric_name not in _execution_histograms:
                meter = get_meter()
                _execution_histograms[metric_name] = meter.create_histogram(
                    f"{metric_name}_duration",
                    description=description or f"Duration of executions for {metric_name}",
                    unit="ms",
                )

            histogram = _execution_histograms[metric_name]
            start_time = time.time()
            status = "success"

            try:
                result = await func(*args, **kwargs)
                return result
            except Exception:
                status = "error"
                raise
            finally:
                duration = (time.time() - start_time) * 1000  # Convert to milliseconds
                histogram.record(duration, {**all_labels, "status": status})

        return cast(F, wrapper)

    return decorator


def with_instrumentation(
    name: Optional[str] = None,
    description: Optional[str] = None,
    labels: Optional[Dict[str, str]] = None,
    enable_traces: bool = True,
    enable_metrics: bool = True,
) -> Callable[[F], F]:
    """Decorator for instrumenting methods with both OpenTelemetry traces and metrics.

    This decorator composes @with_span and @with_metric to provide both trace and metric collection.
    You can selectively enable/disable traces or metrics using the config options.

    Args:
        name: Custom name for the instrumentation
        description: Description for the metrics being collected
        labels: Additional labels/attributes to attach to spans and metrics
        enable_traces: Enable trace collection (default: True)
        enable_metrics: Enable metrics collection (default: True)

    Returns:
        Decorated function with both span and metric instrumentation

    Example:
        ```python
        @with_instrumentation(name="create_sandbox", enable_metrics=True)
        async def create(self, params):
            pass
        ```
    """

    def decorator(func: F) -> F:
        decorated = func

        if enable_metrics:
            decorated = with_metric(name=name, description=description, labels=labels)(decorated)

        if enable_traces:
            decorated = with_span(name=name, attributes=labels)(decorated)

        return decorated

    return decorator
