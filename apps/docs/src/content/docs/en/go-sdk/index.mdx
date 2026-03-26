---
title: Go SDK Reference
description: Interact with Daytona Sandboxes using the Go SDK
next: /docs/go-sdk/daytona
---

The Daytona Go SDK provides a powerful interface for programmatically interacting with Daytona Sandboxes.

## Installation

Install the Daytona Go SDK using go get:

```bash
go get github.com/daytonaio/daytona/libs/sdk-go
```

## Getting Started

### Create a Sandbox

Create a Daytona Sandbox to run your code securely in an isolated environment. The following snippet is an example "Hello World" program that runs securely inside a Daytona Sandbox.

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
)

func main() {
	// Initialize the SDK (uses environment variables by default)
	client, err := daytona.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	// Create a new sandbox
	sandbox, err := client.Create(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	// Execute a command
	response, err := sandbox.Process.ExecuteCommand(context.Background(), "echo 'Hello, World!'")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(response.Result)
}
```

## Configuration

The Daytona SDK can be configured using environment variables or by passing options to the constructor:

```go
package main

import (
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
)

func main() {
	// Using environment variables (DAYTONA_API_KEY, DAYTONA_API_URL, DAYTONA_TARGET)
	client, _ := daytona.NewClient()

	// Using explicit configuration
	config := &types.DaytonaConfig{
		APIKey: "YOUR_API_KEY",
		APIUrl: "https://app.daytona.io/api",
		Target: "us",
	}
	client, _ = daytona.NewClientWithConfig(config)
}
```

For more information on configuring the Daytona SDK, see [configuration](/docs/en/configuration).
