// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"context"
	"time"
)

// FollowBuildLogs streams build logs to stdout until pollState reports that
// the resource reached a terminal state. pollState returns done=true once the
// resource is terminal; an error returned alongside done=true is the terminal
// failure surfaced to the caller after the stream is drained, while an error
// with done=false aborts immediately (the state poll itself failed).
//
// When the resource is already terminal before streaming begins, a follow
// stream would race against the cancellation grace period and could truncate
// long output. In that case the complete log history is fetched with a single
// plain (non-follow) request and the terminal-state error semantics still
// apply.
func FollowBuildLogs(ctx context.Context, params ReadLogParams, pollState func(context.Context) (done bool, failed error)) error {
	done, failErr := pollState(ctx)
	if !done && failErr != nil {
		return failErr
	}
	if done {
		follow := false
		params.Follow = &follow
		err := ReadBuildLogs(ctx, params)
		if failErr != nil {
			return failErr
		}
		return err
	}

	follow := true
	params.Follow = &follow

	logsCtx, stopLogs := context.WithCancel(ctx)
	defer stopLogs()

	streamDone := make(chan error, 1)
	go func() {
		streamDone <- ReadBuildLogs(logsCtx, params)
	}()

	for {
		select {
		case err := <-streamDone:
			// A nil channel blocks forever, so after the stream ends the loop
			// keeps polling the resource state until it turns terminal.
			streamDone = nil
			if err != nil {
				return err
			}
		case <-time.After(time.Second):
		}

		done, failErr := pollState(ctx)
		if !done && failErr != nil {
			return failErr
		}
		if done {
			// Grace period so trailing log output is flushed before the
			// stream is canceled.
			time.Sleep(250 * time.Millisecond)
			stopLogs()
			if streamDone != nil {
				<-streamDone
			}
			return failErr
		}
	}
}
