// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/pkg/jsonmessage"
	controlapi "github.com/moby/buildkit/api/services/control"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider"
	"google.golang.org/protobuf/proto"
)

func (d *DockerClient) runDockerImageBuildWithBuildKitSession(
	ctx context.Context,
	dockerBuildContext io.Reader,
	buildOpts build.ImageBuildOptions,
	writer io.Writer,
) error {
	sess, err := session.NewSession(ctx, "daytona-runner-image-build")
	if err != nil {
		return fmt.Errorf("buildkit session: %w", err)
	}

	sess.Allow(authprovider.NewDockerAuthProvider(authprovider.DockerAuthProviderConfig{
		AuthConfigProvider: authprovider.LoadAuthConfig(registryAuthConfigsToConfigFile(buildOpts.AuthConfigs)),
	}))

	dialer := func(c context.Context, protocol string, meta map[string][]string) (net.Conn, error) {
		return d.apiClient.DialHijack(c, "/session", protocol, meta)
	}

	runDone := make(chan error, 1)
	go func() {
		runDone <- sess.Run(ctx, dialer)
	}()

	buildOpts.SessionID = sess.ID()
	buildOpts.Version = build.BuilderBuildKit

	resp, err := d.apiClient.ImageBuild(ctx, dockerBuildContext, buildOpts)
	if err != nil {
		_ = sess.Close()
		<-runDone // join session goroutine after Close
		return fmt.Errorf("failed to build image: %w", err)
	}
	defer resp.Body.Close()

	startedVertices := make(map[string]bool)

	streamErr := jsonmessage.DisplayJSONMessagesStream(resp.Body, writer, 0, false, func(jm jsonmessage.JSONMessage) {
		if jm.Aux == nil {
			return
		}
		var rawData string
		if err := json.Unmarshal(*jm.Aux, &rawData); err != nil {
			return
		}
		protoBytes, err := base64.StdEncoding.DecodeString(rawData)
		if err != nil {
			return
		}
		var sr controlapi.StatusResponse
		if err := proto.Unmarshal(protoBytes, &sr); err != nil {
			return
		}
		for _, v := range sr.Vertexes {
			if v.Started != nil && !startedVertices[v.Digest] {
				startedVertices[v.Digest] = true
				if v.Cached {
					_, _ = fmt.Fprintf(writer, "#%d CACHED %s\n", len(startedVertices), v.Name)
				} else {
					_, _ = fmt.Fprintf(writer, "#%d %s\n", len(startedVertices), v.Name)
				}
			}
			if v.Error != "" {
				_, _ = fmt.Fprintf(writer, "ERROR: %s\n", v.Error)
			}
		}
		for _, l := range sr.Logs {
			if len(l.Msg) > 0 {
				_, _ = writer.Write(l.Msg)
			}
		}
	})
	closeErr := sess.Close()
	runErr := <-runDone

	if streamErr != nil {
		return streamErr
	}
	if closeErr != nil {
		return closeErr
	}
	if runErr != nil {
		return runErr
	}
	return nil
}
