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
	request.Expiration = timestamppb.New(time.Now().Add(100000 * time.Hour))

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

func RevokeAuthKey(key string) error {
	log.Debug("Revoking headscale auth key")

	request := &v1.ExpirePreAuthKeyRequest{
		Key:  key,
		User: "daytona",
	}

	ctx, client, conn, cancel := GetClient()
	defer cancel()
	defer conn.Close()

	_, err := client.ExpirePreAuthKey(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to revoke ApiKey: %w", err)
	}

	log.Debug("Headscale auth key revoked")
	return nil
}
