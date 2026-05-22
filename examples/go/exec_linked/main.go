// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/options"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
)

func main() {
	client, err := daytona.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	owner, err := client.Create(ctx, types.SnapshotParams{})
	if err != nil {
		log.Fatalf("Failed to create owner sandbox: %v", err)
	}
	defer func() {
		log.Printf("Deleting owner %s", owner.ID)
		if err := owner.Delete(ctx); err != nil {
			log.Printf("Failed to delete owner: %v", err)
		}
	}()
	log.Printf("Owner sandbox ready: id=%s name=%s", owner.ID, owner.Name)

	// Linked sandboxes must be ephemeral — `Ephemeral: true` sets
	// `AutoDeleteInterval=0` automatically.
	follower, err := client.Create(ctx, types.SnapshotParams{
		SandboxBaseParams: types.SandboxBaseParams{
			LinkedSandbox: owner.ID,
			Ephemeral:     true,
		},
	})
	if err != nil {
		log.Fatalf("Failed to create follower sandbox: %v", err)
	}
	defer func() {
		log.Printf("Deleting follower %s", follower.ID)
		if err := follower.Delete(ctx); err != nil {
			log.Printf("Failed to delete follower: %v", err)
		}
	}()
	log.Printf("Follower sandbox ready: id=%s name=%s", follower.ID, follower.Name)

	// Background the http server with nohup, then poll locally until it
	// binds — so the follower's curl below doesn't race startup.
	log.Printf("Starting `python3 -m http.server 3000` in owner %q", owner.Name)
	startScript := `set -e
mkdir -p /tmp/lnk
echo 'hello from owner' > /tmp/lnk/index.html
cd /tmp/lnk
nohup python3 -m http.server 3000 > /tmp/lnk/srv.log 2>&1 &
for _ in $(seq 1 20); do
  if curl -sS --max-time 1 http://127.0.0.1:3000/ >/dev/null 2>&1; then
    echo READY
    exit 0
  fi
  sleep 0.5
done
echo "server failed to start"
cat /tmp/lnk/srv.log
exit 1
`
	startRes, err := owner.Process.ExecuteCommand(ctx, startScript, options.WithExecuteTimeout(30*time.Second))
	if err != nil {
		log.Fatalf("Failed to start server in owner: %v", err)
	}
	if startRes.ExitCode != 0 {
		log.Fatalf("Failed to start server in owner: %s", startRes.Result)
	}
	fmt.Println(startRes.Result)

	// The link network registers the owner under its sandbox name as a DNS
	// alias, so the follower can reach it by name.
	log.Printf("Reaching %q from the follower over the link network", owner.Name)
	curlRes, err := follower.Process.ExecuteCommand(ctx,
		fmt.Sprintf("curl -sS --max-time 5 http://%s:3000/", owner.Name),
		options.WithExecuteTimeout(10*time.Second),
	)
	if err != nil {
		log.Fatalf("Follower could not reach owner: %v", err)
	}
	if curlRes.ExitCode != 0 {
		log.Fatalf("Follower could not reach owner: exit=%d output=%s", curlRes.ExitCode, curlRes.Result)
	}
	log.Printf("Response from owner: %s", curlRes.Result)
}
