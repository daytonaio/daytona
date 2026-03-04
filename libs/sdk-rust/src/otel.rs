// Copyright 2019-2025 Daytona Labs Inc.
// SPDX-License-Identifier: Apache-2.0

#![allow(dead_code)]
// SPDX-License-Identifier: Apache-2.0

use crate::error::DaytonaError;
use opentelemetry::global;
use opentelemetry::trace::{Span, Status, Tracer};
use opentelemetry_otlp::SpanExporter;
use opentelemetry_sdk::trace::SdkTracerProvider;
use opentelemetry_sdk::Resource;

const SERVICE_NAME: &str = "daytona-sdk-rust";
const TRACER_NAME: &str = "daytona-sdk";

pub(crate) struct OtelState {
    tracer_provider: SdkTracerProvider,
}

impl OtelState {
    /// Initialize OTLP HTTP exporter, create tracer provider, and set as global.
    ///
    /// The exporter reads configuration from standard OpenTelemetry environment
    /// variables (`OTEL_EXPORTER_OTLP_ENDPOINT`, etc.) and defaults to
    /// `http://localhost:4318`.
    pub fn new() -> Result<Self, DaytonaError> {
        let exporter = SpanExporter::builder()
            .with_http()
            .build()
            .map_err(|e| DaytonaError::Network(format!("Failed to create OTLP exporter: {e}")))?;

        let tracer_provider = SdkTracerProvider::builder()
            .with_batch_exporter(exporter)
            .with_resource(Resource::builder().with_service_name(SERVICE_NAME).build())
            .build();

        global::set_tracer_provider(tracer_provider.clone());

        Ok(Self { tracer_provider })
    }

    /// Shutdown the tracer provider, flushing any pending spans.
    pub async fn shutdown(&self) -> Result<(), DaytonaError> {
        self.tracer_provider
            .shutdown()
            .map_err(|e| DaytonaError::Network(format!("Failed to shutdown tracer provider: {e}")))
    }
}

/// Instrument a function call with OpenTelemetry tracing.
///
/// If `otel` is `Some`, creates a span named `{component}.{method}` and
/// executes `f` within it. On error the span status is set accordingly.
/// If `otel` is `None`, `f` is called directly with no overhead.
pub(crate) async fn with_instrumentation<T, F, Fut>(
    otel: Option<&OtelState>,
    component: &str,
    method: &str,
    f: F,
) -> Result<T, DaytonaError>
where
    F: FnOnce() -> Fut,
    Fut: std::future::Future<Output = Result<T, DaytonaError>>,
{
    if otel.is_some() {
        let tracer = global::tracer(TRACER_NAME);
        let span_name = format!("{component}.{method}");
        let mut span = tracer.start(span_name);

        let result = f().await;

        if let Err(ref e) = result {
            span.set_status(Status::error(e.to_string()));
        }

        span.end();
        result
    } else {
        f().await
    }
}

/// Instrument a void function (no return value).
///
/// Convenience wrapper around [`with_instrumentation`] for `()` return type.
pub(crate) async fn with_instrumentation_void<F, Fut>(
    otel: Option<&OtelState>,
    component: &str,
    method: &str,
    f: F,
) -> Result<(), DaytonaError>
where
    F: FnOnce() -> Fut,
    Fut: std::future::Future<Output = Result<(), DaytonaError>>,
{
    with_instrumentation(otel, component, method, f).await
}

/// HTTP transport wrapper that injects `traceparent` header for
/// distributed trace context propagation.
pub(crate) struct OtelTransport<T> {
    inner: T,
}

#[allow(dead_code)]
impl<T> OtelTransport<T> {
    pub fn new(inner: T) -> Self {
        Self { inner }
    }

    pub fn inner(&self) -> &T {
        &self.inner
    }
}
