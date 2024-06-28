// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/daytonaio/daytona/pkg/builder/devcontainer"
	"github.com/daytonaio/daytona/pkg/ssh"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/google/uuid"
	"github.com/tailscale/hujson"

	log "github.com/sirupsen/logrus"
)

const dockerSockForwardContainer = "daytona-sock-forward"

type RemoteUser string

func (d *DockerClient) createProjectFromDevcontainer(opts *CreateProjectOptions, prebuild bool, sshClient *ssh.Client) (RemoteUser, error) {
	socketForwardId, err := d.ensureDockerSockForward(opts.LogWriter)
	if err != nil {
		return "", err
	}

	ctx := context.Background()

	overridesDir := filepath.Join(filepath.Dir(opts.ProjectDir), fmt.Sprintf("%s-data", filepath.Base(opts.ProjectDir)))
	overridesTarget := "/tmp/overrides"

	if sshClient != nil {
		err = sshClient.Exec(fmt.Sprintf("mkdir -p %s", overridesDir), opts.LogWriter)
		if err != nil {
			return "", err
		}
	} else {
		err = os.MkdirAll(overridesDir, 0755)
		if err != nil {
			return "", err
		}
	}

	projectTarget := path.Join("/project", filepath.Base(opts.ProjectDir))
	targetConfigFilePath := path.Join(projectTarget, opts.Project.Build.Devcontainer.DevContainerFilePath)

	config, err := d.readDevcontainerConfig(opts, socketForwardId, projectTarget, targetConfigFilePath)
	if err != nil {
		return "", err
	}

	workspaceFolder := config.Workspace.WorkspaceFolder
	if workspaceFolder == "" {
		return "", fmt.Errorf("unable to determine workspace folder from devcontainer configuration")
	}

	remoteUser := config.MergedConfiguration.RemoteUser
	if remoteUser == "" {
		return "", fmt.Errorf("unable to determine remote user from devcontainer configuration")
	}

	var devcontainerConfigContent []byte
	if sshClient != nil {
		configFilePath := path.Join(opts.ProjectDir, opts.Project.Build.Devcontainer.DevContainerFilePath)
		devcontainerConfigContent, err = sshClient.ReadFile(configFilePath)
	} else {
		configFilePath := filepath.Join(opts.ProjectDir, opts.Project.Build.Devcontainer.DevContainerFilePath)
		devcontainerConfigContent, err = os.ReadFile(configFilePath)
	}
	if err != nil {
		return "", err
	}

	var devcontainerConfig map[string]interface{}

	standardized, err := hujson.Standardize(devcontainerConfigContent)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(standardized, &devcontainerConfig)
	if err != nil {
		return "", err
	}

	envVars := map[string]string{}

	if _, ok := devcontainerConfig["containerEnv"]; ok {
		if containerEnv, ok := devcontainerConfig["containerEnv"].(map[string]interface{}); ok {
			for k, v := range containerEnv {
				envVars[k] = v.(string)
			}
		}
	}

	for k, v := range opts.Project.EnvVars {
		envVars[k] = v
	}

	// If the workspaceFolder is not set in the devcontainer.json, we set it to /workspaces/<project-name>
	if _, ok := devcontainerConfig["workspaceFolder"].(string); !ok {
		workspaceFolder = fmt.Sprintf("/workspaces/%s", opts.Project.Name)
		devcontainerConfig["workspaceFolder"] = workspaceFolder
	}
	devcontainerConfig["workspaceMount"] = fmt.Sprintf("source=%s,target=%s,type=bind", opts.ProjectDir, workspaceFolder)

	delete(devcontainerConfig, "initializeCommand")

	if _, ok := devcontainerConfig["dockerComposeFile"]; ok {
		composeFilePath := devcontainerConfig["dockerComposeFile"].(string)

		if opts.SshSessionConfig != nil {
			composeFilePath = path.Join(opts.ProjectDir, filepath.Dir(opts.Project.Build.Devcontainer.DevContainerFilePath), composeFilePath)

			composeFileContent, err := d.getRemoteComposeContent(opts, socketForwardId, opts.ProjectDir, composeFilePath)
			if err != nil {
				return "", err
			}

			composeFilePath = filepath.Join(os.TempDir(), fmt.Sprintf("daytona-compose-%s-%s.yml", opts.Project.WorkspaceId, opts.Project.Name))
			err = os.WriteFile(composeFilePath, []byte(composeFileContent), 0644)
			if err != nil {
				return "", err
			}
		} else {
			composeFilePath = filepath.Join(opts.ProjectDir, filepath.Dir(opts.Project.Build.Devcontainer.DevContainerFilePath), composeFilePath)
		}

		options, err := cli.NewProjectOptions([]string{composeFilePath}, cli.WithOsEnv, cli.WithDotEnv)
		if err != nil {
			return "", err
		}

		project, err := cli.ProjectFromOptions(ctx, options)
		if err != nil {
			return "", err
		}

		project.Name = fmt.Sprintf("%s-%s", opts.Project.WorkspaceId, opts.Project.Name)

		for _, service := range project.Services {
			if service.Build != nil {
				if strings.HasPrefix(service.Build.Context, opts.ProjectDir) {
					service.Build.Context = strings.Replace(service.Build.Context, opts.ProjectDir, projectTarget, 1)
				}
			}
		}

		overrideComposeContent, err := project.MarshalYAML()
		if err != nil {
			return "", err
		}

		if sshClient != nil {
			err = os.RemoveAll(composeFilePath)
			if err != nil {
				opts.LogWriter.Write([]byte(fmt.Sprintf("Error removing override compose file: %v\n", err)))
				return "", err
			}
			res, err := sshClient.WriteFile(string(overrideComposeContent), filepath.Join(overridesDir, "daytona-compose-override.yml"))
			if err != nil {
				opts.LogWriter.Write([]byte(fmt.Sprintf("Error writing override compose file: %s\n", string(res))))
				return "", err
			}
		} else {
			err = os.WriteFile(filepath.Join(overridesDir, "daytona-compose-override.yml"), overrideComposeContent, 0644)
			if err != nil {
				return "", err
			}
		}

		devcontainerConfig["dockerComposeFile"] = path.Join(overridesTarget, "daytona-compose-override.yml")
	}

	envVars["DAYTONA_PROJECT_DIR"] = workspaceFolder

	devcontainerConfig["containerEnv"] = envVars

	configString, err := json.MarshalIndent(devcontainerConfig, "", "  ")
	if err != nil {
		return "", err
	}

	if sshClient != nil {
		res, err := sshClient.WriteFile(string(configString), path.Join(overridesDir, "devcontainer.json"))
		if err != nil {
			opts.LogWriter.Write([]byte(fmt.Sprintf("Error writing override compose file: %s\n", string(res))))
			return "", err
		}
	} else {
		err = os.WriteFile(path.Join(overridesDir, "devcontainer.json"), configString, 0644)
		if err != nil {
			return "", err
		}
	}

	devcontainerCmd := []string{
		"devcontainer",
		"up",
		"--workspace-folder=" + projectTarget,
		"--config=" + targetConfigFilePath,
		"--override-config=" + path.Join(overridesTarget, "devcontainer.json"),
		"--id-label=daytona.workspace.id=" + opts.Project.WorkspaceId,
		"--id-label=daytona.project.name=" + opts.Project.Name,
	}

	if prebuild {
		devcontainerCmd = append(devcontainerCmd, "--prebuild")
	}

	cmd := []string{"-c", strings.Join(devcontainerCmd, " ")}

	if prebuild {
		err = d.runInitializeCommand(config.MergedConfiguration.InitializeCommand, opts.LogWriter, sshClient)
		if err != nil {
			return "", err
		}
	}

	c, err := d.apiClient.ContainerCreate(ctx, &container.Config{
		Image:        "daytonaio/workspace-project",
		Entrypoint:   []string{"sh"},
		Env:          []string{"DOCKER_HOST=tcp://localhost:2375"},
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	}, &container.HostConfig{
		Privileged:  true,
		NetworkMode: container.NetworkMode(fmt.Sprintf("container:%s", socketForwardId)),
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: opts.ProjectDir,
				Target: projectTarget,
			},
			{
				Type:   mount.TypeBind,
				Source: overridesDir,
				Target: overridesTarget,
			},
		},
	}, nil, nil, uuid.NewString())
	if err != nil {
		return "", err
	}

	defer d.removeContainer(c.ID) // nolint:errcheck

	waitResponse, errChan := d.apiClient.ContainerWait(ctx, c.ID, container.WaitConditionNextExit)

	err = d.apiClient.ContainerStart(ctx, c.ID, container.StartOptions{})
	if err != nil {
		return "", err
	}

	go func() {
		for {
			err = d.GetContainerLogs(c.ID, opts.LogWriter)
			if err == nil {
				break
			}
			log.Error(err)
			time.Sleep(100 * time.Millisecond)
		}
	}()

	select {
	case err := <-errChan:
		if err != nil {
			return "", err
		}
	case resp := <-waitResponse:
		if resp.StatusCode != 0 {
			return "", fmt.Errorf("container exited with status %d", resp.StatusCode)
		}
		if resp.Error != nil {
			return "", fmt.Errorf("container exited with error: %s", resp.Error.Message)
		}
	}

	return RemoteUser(remoteUser), nil
}

func (d *DockerClient) ensureDockerSockForward(logWriter io.Writer) (string, error) {
	ctx := context.Background()

	containers, err := d.apiClient.ContainerList(ctx, container.ListOptions{
		Filters: filters.NewArgs(filters.Arg("name", dockerSockForwardContainer)),
		All:     true,
	})
	if err != nil {
		return "", err
	}

	if len(containers) > 1 {
		return "", fmt.Errorf("multiple containers with name %s found", dockerSockForwardContainer)
	}

	if len(containers) == 1 {
		if containers[0].State == "running" {
			return containers[0].ID, nil
		}
		err := d.removeContainer(containers[0].ID)
		if err != nil {
			return "", err
		}
	}

	// TODO: This image should be configurable because it might be hosted on an alternative registry
	err = d.PullImage("alpine/socat", nil, logWriter)
	if err != nil {
		return "", err
	}

	c, err := d.apiClient.ContainerCreate(ctx, &container.Config{
		Image: "alpine/socat",
		User:  "root",
		Cmd:   []string{"tcp-listen:2375,fork,reuseaddr", "unix-connect:/var/run/docker.sock"},
	}, &container.HostConfig{
		Privileged: true,
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: "/var/run/docker.sock",
				Target: "/var/run/docker.sock",
			},
		},
	}, nil, nil, dockerSockForwardContainer)
	if err != nil {
		return "", err
	}

	return c.ID, d.apiClient.ContainerStart(ctx, dockerSockForwardContainer, container.StartOptions{})
}

func (d *DockerClient) readDevcontainerConfig(opts *CreateProjectOptions, socketForwardId, projectTarget, configFilePath string) (*devcontainer.Root, error) {
	opts.LogWriter.Write([]byte("Reading devcontainer configuration...\n"))

	devcontainerCmd := []string{
		"devcontainer",
		"read-configuration",
		"--workspace-folder=" + projectTarget,
		"--config=" + configFilePath,
		"--include-merged-configuration",
	}

	cmd := strings.Join(devcontainerCmd, " ")

	output, err := d.execInContainer(cmd, opts, projectTarget, socketForwardId, projectTarget)
	if err != nil {
		return nil, err
	}

	configStartIndex := strings.Index(output, "{")
	if configStartIndex == -1 {
		return nil, fmt.Errorf("unable to find start of JSON in devcontainer configuration")
	}

	var rootConfig devcontainer.Root
	err = json.Unmarshal([]byte(output[configStartIndex:]), &rootConfig)
	if err != nil {
		return nil, err
	}

	return &rootConfig, nil
}

func (d *DockerClient) runInitializeCommand(initializeCommand interface{}, logWriter io.Writer, sshClient *ssh.Client) error {
	if initializeCommand == nil {
		return nil
	}

	logWriter.Write([]byte("Running initialize command...\n"))

	switch initializeCommand := initializeCommand.(type) {
	case string:
		cmd := []string{"sh", "-c", initializeCommand}
		return execDevcontainerCommand(cmd, logWriter, sshClient)
	case []interface{}:
		var commandArray []string
		for _, arg := range initializeCommand {
			argString, ok := arg.(string)
			if !ok {
				return fmt.Errorf("invalid command type: %v", arg)
			}
			commandArray = append(commandArray, argString)
		}
		return execDevcontainerCommand(commandArray, logWriter, sshClient)
	case map[string]interface{}:
		commands := map[string][]string{}
		for name, command := range initializeCommand {
			switch command := command.(type) {
			case string:
				commands[name] = []string{"sh", "-c", command}
			case []interface{}:
				var cmd []string
				for _, arg := range command {
					argString, ok := arg.(string)
					if !ok {
						return fmt.Errorf("invalid command type: %v", command)
					}
					cmd = append(cmd, argString)
				}
				commands[name] = cmd
			}
		}
		errChan := make(chan error)
		for name, command := range commands {
			go func() {
				logWriter.Write([]byte(fmt.Sprintf("Running %s\n", name)))
				err := execDevcontainerCommand(command, logWriter, sshClient)
				if err != nil {
					logWriter.Write([]byte(fmt.Sprintf("Error running %s: %v\n", name, err)))
					errChan <- err
				} else {
					errChan <- nil
				}
			}()
		}

		errs := []error{}
		for range commands {
			err := <-errChan
			if err != nil {
				errs = append(errs, err)
			}
		}

		if len(errs) > 0 {
			return fmt.Errorf("errors running initialize commands: %v", errs)
		}

		return nil
	}

	return fmt.Errorf("invalid command type: %v", initializeCommand)
}

func (d *DockerClient) execInContainer(cmd string, opts *CreateProjectOptions, workDir, socketForwardId, projectTarget string) (string, error) {
	ctx := context.Background()

	c, err := d.apiClient.ContainerCreate(ctx, &container.Config{
		Image:      "daytonaio/workspace-project",
		Entrypoint: []string{"sh"},
		Env:        []string{"DOCKER_HOST=tcp://localhost:2375"},
		Cmd:        append([]string{"-c"}, cmd),
		Tty:        true,
		WorkingDir: workDir,
	}, &container.HostConfig{
		Privileged:  true,
		NetworkMode: container.NetworkMode(fmt.Sprintf("container:%s", socketForwardId)),
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: opts.ProjectDir,
				Target: projectTarget,
			},
		},
	}, nil, nil, uuid.NewString())
	if err != nil {
		return "", err
	}

	defer d.removeContainer(c.ID) // nolint:errcheck

	waitResponse, errChan := d.apiClient.ContainerWait(ctx, c.ID, container.WaitConditionNextExit)

	err = d.apiClient.ContainerStart(ctx, c.ID, container.StartOptions{})
	if err != nil {
		return "", err
	}

	output := ""

	r, w := io.Pipe()

	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			output += scanner.Text() + "\n"
		}
	}()

	go func() {
		err = d.GetContainerLogs(c.ID, w)
		if err != nil {
			opts.LogWriter.Write([]byte(fmt.Sprintf("Error reading devcontainer configuration: %v\n", err)))
		}
	}()

	select {
	case err := <-errChan:
		if err != nil {
			return "", err
		}
	case resp := <-waitResponse:
		if resp.StatusCode != 0 {
			return "", fmt.Errorf("container exited with status %d", resp.StatusCode)
		}
		if resp.Error != nil {
			return "", fmt.Errorf("container exited with error: %s", resp.Error.Message)
		}
	}

	return output, nil
}

func (d *DockerClient) getRemoteComposeContent(opts *CreateProjectOptions, socketForwardId, projectTarget, composePath string) (string, error) {
	if opts.SshSessionConfig == nil {
		return "", nil
	}

	output, err := d.execInContainer("docker compose config", opts, filepath.Dir(composePath), socketForwardId, projectTarget)
	if err != nil {
		return "", err
	}

	nameIndex := strings.Index(output, "name: ")
	if nameIndex == -1 {
		return "", fmt.Errorf("unable to find service name in compose config")
	}

	return output[nameIndex:], nil
}

func execDevcontainerCommand(command []string, logWriter io.Writer, sshClient *ssh.Client) error {
	if sshClient != nil {
		if command[0] == "sh" {
			cmd := fmt.Sprintf(`sh -c "%s"`, strings.Join(command[2:], " "))
			return sshClient.Exec(cmd, logWriter)
		}
		return sshClient.Exec(strings.Join(command, " "), logWriter)
	}

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	cmd.Env = os.Environ()
	return cmd.Run()
}
