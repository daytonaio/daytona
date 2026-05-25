// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/daytonaio/common-go/pkg/timer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// androidAdbPort is the standard ADB port exposed inside android-device containers.
// We treat a successful TCP dial to it as "the emulator is ready to accept connections".
const androidAdbPort = 5555

// waitForAdbRunning polls the ADB port inside an android-device container until it accepts
// TCP connections or the androidBootTimeoutSec is exceeded. It does not speak the ADB
// protocol — a successful TCP handshake is enough to know the emulator has booted far
// enough to be reachable over the runner's container network. The Android emulator's
// cold boot regularly takes two or more minutes, so this probe uses a dedicated longer
// timeout instead of the generic sandbox start timeout.
func (d *DockerClient) waitForAdbRunning(ctx context.Context, containerIP string) error {
	defer timer.Timer()()

	tracer := otel.Tracer("runner")
	ctx, span := tracer.Start(ctx, "wait_for_adb_running",
		trace.WithAttributes(attribute.String("container.ip", containerIP)),
	)
	defer span.End()

	addr := net.JoinHostPort(containerIP, strconv.Itoa(androidAdbPort))

	timeout := time.Duration(d.androidBootTimeoutSec) * time.Second
	span.SetAttributes(attribute.Int64("timeout_sec", int64(d.androidBootTimeoutSec)))
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	dialer := &net.Dialer{Timeout: 500 * time.Millisecond}

	retries := 0
	for {
		select {
		case <-timeoutCtx.Done():
			span.SetAttributes(attribute.Int("retries", retries))
			var err error
			if ctx.Err() != nil {
				err = fmt.Errorf("waiting for ADB port on %s cancelled: %w", addr, ctx.Err())
				span.RecordError(err)
				span.SetStatus(codes.Error, "wait for adb cancelled")
			} else {
				err = fmt.Errorf("timeout waiting for ADB port on %s", addr)
				span.RecordError(err)
				span.SetStatus(codes.Error, "timeout waiting for adb")
			}
			return err
		case <-ticker.C:
			conn, err := dialer.DialContext(timeoutCtx, "tcp", addr)
			if err != nil {
				retries++
				continue
			}
			_ = conn.Close()
			span.SetAttributes(attribute.Int("retries", retries))
			span.SetStatus(codes.Ok, "adb ready")
			return nil
		}
	}
}
