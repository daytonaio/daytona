package headscale

import (
	"context"
	"os"

	v1 "github.com/juanfont/headscale/gen/go/headscale/v1"
	"github.com/juanfont/headscale/hscontrol/types"
	"github.com/juanfont/headscale/hscontrol/util"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func GetClient() (context.Context, v1.HeadscaleServiceClient, *grpc.ClientConn, context.CancelFunc) {
	cfg, err := types.GetHeadscaleConfig()
	if err != nil {
		log.Fatal().
			Err(err).
			Caller().
			Msgf("Failed to load configuration")
		os.Exit(-1)
	}

	log.Debug().
		Dur("timeout", cfg.CLI.Timeout).
		Msgf("Setting timeout")

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
		log.Fatal().Caller().Err(err).Msgf("Could not connect: %v", err)
		os.Exit(-1)
	}

	client := v1.NewHeadscaleServiceClient(conn)

	return ctx, client, conn, cancel
}
