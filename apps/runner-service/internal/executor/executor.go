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

	"github.com/docker/docker/client"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	apiclient "github.com/daytonaio/apiclient"
	"github.com/daytonaio/runner-service/internal/metrics"
)

// JobType constants
const (
	JobTypeCreateSandbox  = "CREATE_SANDBOX"
	JobTypeStartSandbox   = "START_SANDBOX"
	JobTypeStopSandbox    = "STOP_SANDBOX"
	JobTypeDestroySandbox = "DESTROY_SANDBOX"
	JobTypeCreateBackup   = "CREATE_BACKUP"
	JobTypeBuildSnapshot  = "BUILD_SNAPSHOT"
	JobTypePullSnapshot   = "PULL_SNAPSHOT"
	JobTypeRemoveSnapshot = "REMOVE_SNAPSHOT"
)

// JobStatus constants
const (
	JobStatusPending    = "PENDING"
	JobStatusInProgress = "IN_PROGRESS"
	JobStatusCompleted  = "COMPLETED"
	JobStatusFailed     = "FAILED"
)

// Executor handles job execution
type Executor struct {
	log          *slog.Logger
	client       *apiclient.APIClient
	dockerClient *client.Client
	collector    *metrics.Collector
	daemonPath   string
}

// NewExecutor creates a new job executor
func NewExecutor(apiClient *apiclient.APIClient, dockerClient *client.Client, collector *metrics.Collector, daemonPath string, logger *slog.Logger) *Executor {
	return &Executor{
		log:          logger.With(slog.String("component", "executor")),
		client:       apiClient,
		dockerClient: dockerClient,
		collector:    collector,
		daemonPath:   daemonPath,
	}
}

// Execute processes a job and updates its status
func (e *Executor) Execute(ctx context.Context, job *apiclient.Job) {
	// Extract trace context from job to continue distributed trace
	ctx = e.extractTraceContext(ctx, job)

	// Build log fields
	jobLog := e.log.With(
		slog.String("job_id", job.GetId()),
		slog.String("job_type", job.GetType()),
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
	err := e.executeJob(ctx, job)

	// Update job status
	status := JobStatusCompleted
	var errorMessage *string
	if err != nil {
		status = JobStatusFailed
		errMsg := err.Error()
		errorMessage = &errMsg
		jobLog.Error("Job failed", slog.Any("error", err))
	} else {
		jobLog.Info("Job completed successfully")
	}

	// Report status to API
	if err := e.updateJobStatus(ctx, job.GetId(), status, errorMessage); err != nil {
		jobLog.Error("Failed to update job status", slog.Any("error", err))
	}
}

// executeJob dispatches to the appropriate handler based on job type
func (e *Executor) executeJob(ctx context.Context, job *apiclient.Job) error {
	// Create a span for the job execution
	tracer := otel.Tracer("runner-service")
	ctx, span := tracer.Start(ctx, fmt.Sprintf("execute_%s", job.GetType()),
		trace.WithAttributes(
			attribute.String("job.id", job.GetId()),
			attribute.String("job.type", job.GetType()),
			attribute.String("job.status", job.GetStatus()),
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
	var err error
	switch job.GetType() {
	case JobTypeCreateSandbox:
		err = e.createSandbox(ctx, job)
	case JobTypeStartSandbox:
		err = e.startSandbox(ctx, job)
	case JobTypeStopSandbox:
		err = e.stopSandbox(ctx, job)
	case JobTypeDestroySandbox:
		err = e.destroySandbox(ctx, job)
	case JobTypeCreateBackup:
		err = e.createBackup(ctx, job)
	case JobTypeBuildSnapshot:
		err = e.buildSnapshot(ctx, job)
	case JobTypePullSnapshot:
		err = e.pullSnapshot(ctx, job)
	case JobTypeRemoveSnapshot:
		err = e.removeSnapshot(ctx, job)
	default:
		err = fmt.Errorf("unknown job type: %s", job.GetType())
	}

	// Record error in span if present
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Bool("error", true))
	}

	return err
}

// updateJobStatus reports job completion status to the API
func (e *Executor) updateJobStatus(ctx context.Context, jobID, status string, errorMessage *string) error {
	// Create a span for the API call - otelhttp will create a child span for the HTTP request
	tracer := otel.Tracer("runner-service")
	ctx, span := tracer.Start(ctx, "update_job_status",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("job.id", jobID),
			attribute.String("job.status", status),
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

	req := e.client.JobsAPI.UpdateJobStatus(ctx, jobID).UpdateJobStatus(*updateStatus)
	_, _, err := req.Execute()

	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Bool("error", true))
	}

	return err
}

// ParsePayload is a helper to parse job payload into a specific type
func ParsePayload(payload map[string]interface{}, target interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(jsonData, target); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return nil
}

// envMapToSlice converts env map to KEY=VALUE slice
func envMapToSlice(envMap map[string]string) []string {
	env := make([]string, 0, len(envMap))
	for k, v := range envMap {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return env
}

// extractTraceContext extracts OpenTelemetry trace context from the job
// and returns a new context with the trace information to continue distributed tracing
func (e *Executor) extractTraceContext(ctx context.Context, job *apiclient.Job) context.Context {
	traceContext := job.GetTraceContext()
	if traceContext == nil || len(traceContext) == 0 {
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
