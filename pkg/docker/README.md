## Package Purpose

The purpose of the `docker` package is to provide a single library that providers can use to easily create workspaces and projects.
Most providers will import the package and call commands from the `DockerClient` in order to create workspaces and projects on provided targets.

## Usage

To use the Daytona `DockerClient`, the consumer of the library must provide a Docker API client from `github.com/docker/docker/client`.

Example:
```golang
import (
  "github.com/docker/docker/client"
  "github.com/daytonaio/daytona/pkg/docker"
)

func GetDockerClient() (docker.IDockerClient, error) {
	client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return docker.NewDockerClient(docker.DockerClientConfig{
		ApiClient: client,
	}), nil
}
```

## Testing

To test changes made in this library, other than writing tests, we recommend using the library locally with the [Docker provider](https://github.com/daytonaio/daytona-provider-docker).

The procedure would be as follows:
1. Clone the [Docker provider](https://github.com/daytonaio/daytona-provider-docker)
1. Open the `go.mod` file.
1. Add `replace github.com/daytonaio/daytona => <PATH_TO_DAYTONA_REPO>`
1. Run `go mod tidy`
1. Build the Docker provider and store it to your local provider directory.
 - `go build -o <PATH_TO_PROVIDER_DIR>/docker-provider/docker-provider main.go`
 - If you're developing Daytona in a devcontainer, `<PATH_TO_PROVIDER_DIR> = ~/.config/daytona/providers`

**Note**: Any change to the `docker` library will require that providers update the version of `github.com/daytonaio/daytona` after the next release and post a release of their own.
