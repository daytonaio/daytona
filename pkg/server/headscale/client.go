// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package headscale

import (
	"context"

	"github.com/daytonaio/daytona/pkg/server/config"
	v1 "github.com/juanfont/headscale/gen/go/headscale/v1"
	"github.com/juanfont/headscale/hscontrol/util"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func getClient() (context.Context, v1.HeadscaleServiceClient, *grpc.ClientConn, context.CancelFunc, error) {
	c, err := config.GetConfig()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	cfg, err := getConfig(c)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.CLI.Timeout)

	grpcOptions := []grpc.DialOption{
		grpc.WithBlock(),
	}

	address := cfg.UnixSocket

	grpcOptions = append(
		grpcOptions,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(util.GrpcSocketDialer),
	)

	log.Trace().Caller().Str("address", address).Msg("Connecting via gRPC")
	conn, err := grpc.DialContext(ctx, address, grpcOptions...)
	if err != nil {
		cancel()
		return nil, nil, nil, nil, err
	}

	client := v1.NewHeadscaleServiceClient(conn)

	return ctx, client, conn, cancel, nil
}
