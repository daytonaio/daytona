package headscale

import (
	"fmt"

	v1 "github.com/juanfont/headscale/gen/go/headscale/v1"
	log "github.com/sirupsen/logrus"
)

func DeleteNode(nodeName string) error {
	log.Debug("Deleting headscale node")

	ctx, client, conn, cancel, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}
	defer cancel()
	defer conn.Close()

	response, err := client.ListNodes(ctx, &v1.ListNodesRequest{
		User: "daytona",
	})
	if err != nil {
		return fmt.Errorf("failed to list nodes: %w", err)
	}

	for _, node := range response.Nodes {
		if node.Name == nodeName {
			_, err := client.DeleteNode(ctx, &v1.DeleteNodeRequest{
				NodeId: node.Id,
			})
			if err != nil {
				return fmt.Errorf("failed to delete node: %w", err)
			}
			log.Debug("Headscale node deleted")
			return nil
		}
	}

	return fmt.Errorf("node not found")
}
