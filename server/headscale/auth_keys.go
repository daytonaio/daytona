package headscale

import (
	"fmt"
	"time"

	v1 "github.com/juanfont/headscale/gen/go/headscale/v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	log "github.com/sirupsen/logrus"
)

func CreateAuthKey() (string, error) {
	log.Debug("Creating headscale auth key")

	request := &v1.CreatePreAuthKeyRequest{
		Reusable: true,
		User:     "daytona",
	}

	expiration := time.Now().UTC().Add(time.Duration(24) * time.Hour)
	request.Expiration = timestamppb.New(expiration)

	ctx, client, conn, cancel := GetClient()
	defer cancel()
	defer conn.Close()

	response, err := client.CreatePreAuthKey(ctx, request)
	if err != nil {
		return "", fmt.Errorf("failed to create ApiKey: %w", err)
	}

	log.Debug("Headscale auth key created")

	return response.PreAuthKey.Key, nil
}
