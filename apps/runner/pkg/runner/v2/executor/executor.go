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
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/daytonaio/common-go/pkg/utils"
	"github.com/daytonaio/runner/internal/metrics"
	"github.com/daytonaio/runner/pkg/docker"
	"github.com/daytonaio/runner/pkg/runner/v2/client"
	specsgen "github.com/daytonaio/runner/pkg/runner/v2/specs/gen"
)

type ExecutorConfig struct {
	Docker    *docker.DockerClient
	Collector *metrics.Collector
	Logger    *slog.Logger
}

// Executor handles job execution
type Executor struct {
	log       *slog.Logger
	client    *client.APIClient
	docker    *docker.DockerClient
	collector *metrics.Collector
}

// NewExecutor creates a new job executor
func NewExecutor(cfg *ExecutorConfig) (*Executor, error) {
	apiClient, err := client.NewAPIClient()
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
func (e *Executor) Execute(ctx context.Context, job *specsgen.Job) {
	ctx = e.extractTraceContext(ctx, job)

	jobLog := e.log.With(
		slog.String("job_id", job.GetId()),
		slog.String("job_type", job.GetType().String()),
	)

	if resourceType := job.GetResourceType(); resourceType != specsgen.ResourceType_RESOURCE_TYPE_UNSPECIFIED {
		jobLog = jobLog.With(slog.String("resource_type", resourceType.String()))
	}
	if resourceId := job.GetResourceId(); resourceId != "" {
		jobLog = jobLog.With(slog.String("resource_id", resourceId))
	}

	if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
		jobLog = jobLog.With(
			slog.String("trace_id", spanCtx.TraceID().String()),
			slog.String("span_id", spanCtx.SpanID().String()),
		)
	}

	jobLog.InfoContext(ctx, "Executing job")

	resultMetadata, err := e.executeJob(ctx, job)

	status := specsgen.JobStatus_COMPLETED
	var errorMessage *string
	if err != nil {
		status = specsgen.JobStatus_FAILED
		errMsg := err.Error()
		errorMessage = &errMsg
		jobLog.ErrorContext(ctx, "Job failed", "error", err)
	} else {
		jobLog.InfoContext(ctx, "Job completed successfully")
	}

	if err := e.updateJobStatus(ctx, job.GetId(), status, resultMetadata, errorMessage); err != nil {
		jobLog.ErrorContext(ctx, "Failed to update job status", "error", err)
	}
}

// executeJob dispatches to the appropriate handler based on job type
func (e *Executor) executeJob(ctx context.Context, job *specsgen.Job) (any, error) {
	tracer := otel.Tracer("runner")
	ctx, span := tracer.Start(ctx, fmt.Sprintf("execute_%s", job.GetType().String()),
		trace.WithAttributes(
			attribute.String("job.id", job.GetId()),
			attribute.String("job.type", job.GetType().String()),
			attribute.String("job.status", job.GetStatus().String()),
		),
	)
	defer span.End()

	if resourceType := job.GetResourceType(); resourceType != specsgen.ResourceType_RESOURCE_TYPE_UNSPECIFIED {
		span.SetAttributes(attribute.String("resource.type", resourceType.String()))
	}
	if resourceId := job.GetResourceId(); resourceId != "" {
		span.SetAttributes(attribute.String("resource.id", resourceId))
	}

	var resultMetadata any
	var err error
	switch job.GetType() {
	case specsgen.JobType_CREATE_SANDBOX:
		resultMetadata, err = e.createSandbox(ctx, job)
	case specsgen.JobType_START_SANDBOX:
		resultMetadata, err = e.startSandbox(ctx, job)
	case specsgen.JobType_STOP_SANDBOX:
		resultMetadata, err = e.stopSandbox(ctx, job)
	case specsgen.JobType_DESTROY_SANDBOX:
		resultMetadata, err = e.destroySandbox(ctx, job)
	case specsgen.JobType_RESIZE_SANDBOX:
		resultMetadata, err = e.resizeSandbox(ctx, job)
	case specsgen.JobType_CREATE_BACKUP:
		resultMetadata, err = e.createBackup(ctx, job)
	case specsgen.JobType_BUILD_SNAPSHOT:
		resultMetadata, err = e.buildSnapshot(ctx, job)
	case specsgen.JobType_PULL_SNAPSHOT:
		resultMetadata, err = e.pullSnapshot(ctx, job)
	case specsgen.JobType_REMOVE_SNAPSHOT:
		resultMetadata, err = e.removeSnapshot(ctx, job)
	case specsgen.JobType_UPDATE_SANDBOX_NETWORK_SETTINGS:
		resultMetadata, err = e.updateNetworkSettings(ctx, job)
	case specsgen.JobType_INSPECT_SNAPSHOT_IN_REGISTRY:
		resultMetadata, err = e.inspectSnapshotInRegistry(ctx, job)
	case specsgen.JobType_RECOVER_SANDBOX:
		resultMetadata, err = e.recoverSandbox(ctx, job)
	default:
		err = fmt.Errorf("unknown job type: %s", job.GetType().String())
	}

	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Bool("error", true))
		span.SetStatus(codes.Error, "job execution failed")
	}

	return resultMetadata, err
}

// updateJobStatus reports job completion status to the API
func (e *Executor) updateJobStatus(ctx context.Context, jobID string, status specsgen.JobStatus, resultMetadata any, errorMessage *string) error {
	tracer := otel.Tracer("runner")
	ctx, span := tracer.Start(ctx, "update_job_status",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("job.id", jobID),
			attribute.String("job.status", status.String()),
		),
	)
	defer span.End()

	if errorMessage != nil {
		span.SetAttributes(attribute.String("job.error", *errorMessage))
	}

	update := &specsgen.UpdateJobStatus{Status: status}
	if errorMessage != nil {
		update.ErrorMessage = errorMessage
	}

	if resultMetadata != nil {
		metaJSON, err := json.Marshal(resultMetadata)
		if err != nil {
			return fmt.Errorf("failed to marshal result metadata: %w", err)
		}
		metaStr := string(metaJSON)
		update.ResultMetadata = &metaStr
	}

	path := fmt.Sprintf("/jobs/%s/status", jobID)

	err := utils.RetryWithExponentialBackoff(
		ctx,
		fmt.Sprintf("update job %s status to %s", jobID, status.String()),
		utils.DEFAULT_MAX_RETRIES,
		utils.DEFAULT_BASE_DELAY,
		utils.DEFAULT_MAX_DELAY,
		func() error {
			httpResp, err := e.client.Do(ctx, "POST", path, update, nil)
			if err != nil && httpResp != nil && httpResp.StatusCode >= http.StatusBadRequest && httpResp.StatusCode < http.StatusInternalServerError {
				return &utils.NonRetryableError{Err: fmt.Errorf("HTTP %d: %w", httpResp.StatusCode, err)}
			}
			return err
		},
	)

	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Bool("error", true))
		span.SetStatus(codes.Error, "update job status failed")
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
func (e *Executor) extractTraceContext(ctx context.Context, job *specsgen.Job) context.Context {
	traceContext := job.GetTraceContext()
	if len(traceContext) == 0 {
		e.log.DebugContext(ctx, "no trace context in job", "job_id", job.GetId())
		return ctx
	}

	carrier := make(propagation.MapCarrier)
	for k, v := range traceContext {
		carrier[k] = v
	}

	propagator := propagation.TraceContext{}
	ctx = propagator.Extract(ctx, carrier)

	if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
		e.log.DebugContext(ctx, "extracted trace context from job",
			slog.String("job_id", job.GetId()),
			slog.String("trace_id", spanCtx.TraceID().String()),
			slog.String("span_id", spanCtx.SpanID().String()),
			slog.Bool("sampled", spanCtx.IsSampled()),
		)
	} else {
		e.log.WarnContext(ctx, "trace context present but invalid",
			slog.String("job_id", job.GetId()),
			slog.Any("trace_context", traceContext),
		)
	}

	return ctx
}
