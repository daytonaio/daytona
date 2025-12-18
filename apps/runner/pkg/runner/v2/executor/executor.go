/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	apiclient "github.com/daytonaio/apiclient"
	"github.com/daytonaio/runner/internal/metrics"
	runnerapiclient "github.com/daytonaio/runner/pkg/apiclient"
	"github.com/daytonaio/runner/pkg/docker"
)

type ExecutorConfig struct {
	Docker    *docker.DockerClient
	Collector *metrics.Collector
	Logger    *slog.Logger
}

// Executor handles job execution
type Executor struct {
	log       *slog.Logger
	client    *apiclient.APIClient
	docker    *docker.DockerClient
	collector *metrics.Collector
}

// NewExecutor creates a new job executor
func NewExecutor(cfg *ExecutorConfig) (*Executor, error) {
	apiClient, err := runnerapiclient.GetApiClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}

	return &Executor{
		log:       cfg.Logger.With(slog.String("component", "executor")),
		client:    apiClient,
		docker:    cfg.Docker,
		collector: cfg.Collector,
	}, nil
}

// Execute processes a job and updates its status
func (e *Executor) Execute(ctx context.Context, job *apiclient.Job) {
	// Extract trace context from job to continue distributed trace
	ctx = e.extractTraceContext(ctx, job)

	// Build log fields
	jobLog := e.log.With(
		slog.String("job_id", job.GetId()),
		slog.String("job_type", string(job.GetType())),
	)

	// Add resource info if present
	if resourceType := job.GetResourceType(); resourceType != "" {
		jobLog = jobLog.With(slog.String("resource_type", resourceType))
	}
	if resourceId := job.GetResourceId(); resourceId != "" {
		jobLog = jobLog.With(slog.String("resource_id", resourceId))
	}

	// Add trace info to logs if available
	if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
		jobLog = jobLog.With(
			slog.String("trace_id", spanCtx.TraceID().String()),
			slog.String("span_id", spanCtx.SpanID().String()),
		)
	}

	jobLog.Info("Executing job")

	// Execute the job based on type
	resultMetadata, err := e.executeJob(ctx, job)

	// Update job status
	status := apiclient.JOBSTATUS_COMPLETED
	var errorMessage *string
	if err != nil {
		status = apiclient.JOBSTATUS_FAILED
		errMsg := err.Error()
		errorMessage = &errMsg
		jobLog.Error("Job failed", slog.Any("error", err))
	} else {
		jobLog.Info("Job completed successfully")
	}

	// Report status to API
	if err := e.updateJobStatus(ctx, job.GetId(), status, resultMetadata, errorMessage); err != nil {
		jobLog.Error("Failed to update job status", slog.Any("error", err))
	}
}

// executeJob dispatches to the appropriate handler based on job type
func (e *Executor) executeJob(ctx context.Context, job *apiclient.Job) (any, error) {
	// Create a span for the job execution
	tracer := otel.Tracer("runner")
	ctx, span := tracer.Start(ctx, fmt.Sprintf("execute_%s", job.GetType()),
		trace.WithAttributes(
			attribute.String("job.id", job.GetId()),
			attribute.String("job.type", string(job.GetType())),
			attribute.String("job.status", string(job.GetStatus())),
		),
	)
	defer span.End()

	// Add resource attributes if present
	if resourceType := job.GetResourceType(); resourceType != "" {
		span.SetAttributes(attribute.String("resource.type", resourceType))
	}
	if resourceId := job.GetResourceId(); resourceId != "" {
		span.SetAttributes(attribute.String("resource.id", resourceId))
	}

	// Dispatch to handler
	var resultMetadata any
	var err error
	switch job.GetType() {
	case apiclient.JOBTYPE_CREATE_SANDBOX:
		resultMetadata, err = e.createSandbox(ctx, job)
	case apiclient.JOBTYPE_START_SANDBOX:
		resultMetadata, err = e.startSandbox(ctx, job)
	case apiclient.JOBTYPE_STOP_SANDBOX:
		resultMetadata, err = e.stopSandbox(ctx, job)
	case apiclient.JOBTYPE_DESTROY_SANDBOX:
		resultMetadata, err = e.destroySandbox(ctx, job)
	case apiclient.JOBTYPE_CREATE_BACKUP:
		resultMetadata, err = e.createBackup(ctx, job)
	case apiclient.JOBTYPE_BUILD_SNAPSHOT:
		resultMetadata, err = e.buildSnapshot(ctx, job)
	case apiclient.JOBTYPE_PULL_SNAPSHOT:
		resultMetadata, err = e.pullSnapshot(ctx, job)
	case apiclient.JOBTYPE_REMOVE_SNAPSHOT:
		resultMetadata, err = e.removeSnapshot(ctx, job)
	case apiclient.JOBTYPE_UPDATE_SANDBOX_NETWORK_SETTINGS:
		resultMetadata, err = e.updateNetworkSettings(ctx, job)
	default:
		err = fmt.Errorf("unknown job type: %s", job.GetType())
	}

	// Record error in span if present
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Bool("error", true))
	}

	return resultMetadata, err
}

// updateJobStatus reports job completion status to the API
func (e *Executor) updateJobStatus(ctx context.Context, jobID string, status apiclient.JobStatus, resultMetadata any, errorMessage *string) error {
	// Create a span for the API call - otelhttp will create a child span for the HTTP request
	tracer := otel.Tracer("runner")
	ctx, span := tracer.Start(ctx, "update_job_status",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("job.id", jobID),
			attribute.String("job.status", string(status)),
		),
	)
	defer span.End()

	if errorMessage != nil {
		span.SetAttributes(attribute.String("job.error", *errorMessage))
	}

	updateStatus := apiclient.NewUpdateJobStatus(status)
	if errorMessage != nil {
		updateStatus.SetErrorMessage(*errorMessage)
	}

	if resultMetadata != nil {
		resultMetadataJSON, err := json.Marshal(resultMetadata)
		if err != nil {
			return fmt.Errorf("failed to marshal result metadata: %w", err)
		}
		updateStatus.SetResultMetadata(string(resultMetadataJSON))
	}

	req := e.client.JobsAPI.UpdateJobStatus(ctx, jobID).UpdateJobStatus(*updateStatus)
	_, _, err := req.Execute()

	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Bool("error", true))
	}

	return err
}

// parsePayload is a helper to parse job payload into a specific type
func (e *Executor) parsePayload(payload *string, target interface{}) error {
	if payload == nil || *payload == "" {
		return fmt.Errorf("payload is required")
	}

	if err := json.Unmarshal([]byte(*payload), target); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return nil
}

// extractTraceContext extracts OpenTelemetry trace context from the job
// and returns a new context with the trace information to continue distributed tracing
func (e *Executor) extractTraceContext(ctx context.Context, job *apiclient.Job) context.Context {
	traceContext := job.GetTraceContext()
	if len(traceContext) == 0 {
		e.log.Debug("no trace context in job", slog.String("job_id", job.GetId()))
		return ctx
	}

	// Convert map[string]interface{} to map[string]string for propagation
	carrier := make(propagation.MapCarrier)
	for k, v := range traceContext {
		if strVal, ok := v.(string); ok {
			carrier[k] = strVal
		}
	}

	// Use W3C Trace Context propagator to extract trace info
	propagator := propagation.TraceContext{}
	ctx = propagator.Extract(ctx, carrier)

	// Log trace information if extracted successfully
	if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
		e.log.Debug("extracted trace context from job",
			slog.String("job_id", job.GetId()),
			slog.String("trace_id", spanCtx.TraceID().String()),
			slog.String("span_id", spanCtx.SpanID().String()),
			slog.Bool("sampled", spanCtx.IsSampled()),
		)
	} else {
		e.log.Warn("trace context present but invalid",
			slog.String("job_id", job.GetId()),
			slog.Any("trace_context", traceContext),
		)
	}

	return ctx
}
