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

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/daytonaio/daytona/pkg/build/devcontainer"
	"github.com/daytonaio/daytona/pkg/ssh"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/google/uuid"
)

const dockerSockForwardContainer = "daytona-sock-forward"

type RemoteUser string

type DevcontainerPaths struct {
	OverridesDir         string
	OverridesTarget      string
	ProjectTarget        string
	TargetConfigFilePath string
}

func (d *DockerClient) createProjectFromDevcontainer(opts *CreateProjectOptions, prebuild bool) (RemoteUser, error) {
	socketForwardId, err := d.ensureDockerSockForward(opts.LogWriter)
	if err != nil {
		return "", err
	}

	ctx := context.Background()

	paths := d.getDevcontainerPaths(opts)

	if opts.SshClient != nil {
		err = opts.SshClient.Exec(fmt.Sprintf("mkdir -p %s", paths.OverridesDir), opts.LogWriter)
		if err != nil {
			return "", err
		}
	} else {
		err = os.MkdirAll(paths.OverridesDir, 0755)
		if err != nil {
			return "", err
		}
	}

	rawConfig, config, err := d.readDevcontainerConfig(opts, paths, socketForwardId)
	if err != nil {
		return "", err
	}

	workspaceFolder := config.Workspace.WorkspaceFolder
	if workspaceFolder == "" {
		return "", fmt.Errorf("unable to determine workspace folder from devcontainer configuration")
	}

	remoteUser := config.MergedConfiguration.RemoteUser

	var mergedConfig map[string]interface{}

	err = json.Unmarshal([]byte(rawConfig), &mergedConfig)
	if err != nil {
		return "", err
	}

	devcontainerConfig, ok := mergedConfig["configuration"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unable to find devcontainer configuration in merged configuration")
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

		if opts.SshClient != nil {
			composeFilePath = path.Join(opts.ProjectDir, filepath.Dir(opts.Project.BuildConfig.Devcontainer.FilePath), composeFilePath)

			composeFileContent, err := d.getRemoteComposeContent(opts, paths, socketForwardId, composeFilePath)
			if err != nil {
				return "", err
			}

			composeFilePath = filepath.Join(os.TempDir(), fmt.Sprintf("daytona-compose-%s-%s.yml", opts.Project.WorkspaceId, opts.Project.Name))
			err = os.WriteFile(composeFilePath, []byte(composeFileContent), 0644)
			if err != nil {
				return "", err
			}
		} else {
			composeFilePath = filepath.Join(opts.ProjectDir, filepath.Dir(opts.Project.BuildConfig.Devcontainer.FilePath), composeFilePath)
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
					service.Build.Context = strings.Replace(service.Build.Context, opts.ProjectDir, paths.ProjectTarget, 1)
				}
			}
		}

		overrideComposeContent, err := project.MarshalYAML()
		if err != nil {
			return "", err
		}

		if opts.SshClient != nil {
			err = os.RemoveAll(composeFilePath)
			if err != nil {
				opts.LogWriter.Write([]byte(fmt.Sprintf("Error removing override compose file: %v\n", err)))
				return "", err
			}
			res, err := opts.SshClient.WriteFile(string(overrideComposeContent), filepath.Join(paths.OverridesDir, "daytona-compose-override.yml"))
			if err != nil {
				opts.LogWriter.Write([]byte(fmt.Sprintf("Error writing override compose file: %s\n", string(res))))
				return "", err
			}
		} else {
			err = os.WriteFile(filepath.Join(paths.OverridesDir, "daytona-compose-override.yml"), overrideComposeContent, 0644)
			if err != nil {
				return "", err
			}
		}

		devcontainerConfig["dockerComposeFile"] = path.Join(paths.OverridesTarget, "daytona-compose-override.yml")
	}

	envVars["DAYTONA_PROJECT_DIR"] = workspaceFolder

	devcontainerConfig["containerEnv"] = envVars

	configString, err := json.MarshalIndent(devcontainerConfig, "", "  ")
	if err != nil {
		return "", err
	}

	if opts.SshClient != nil {
		res, err := opts.SshClient.WriteFile(string(configString), path.Join(paths.OverridesDir, "devcontainer.json"))
		if err != nil {
			opts.LogWriter.Write([]byte(fmt.Sprintf("Error writing override compose file: %s\n", string(res))))
			return "", err
		}
	} else {
		err = os.WriteFile(path.Join(paths.OverridesDir, "devcontainer.json"), configString, 0644)
		if err != nil {
			return "", err
		}
	}

	devcontainerCmd := []string{
		"devcontainer",
		"up",
		"--workspace-folder=" + paths.ProjectTarget,
		"--config=" + paths.TargetConfigFilePath,
		"--override-config=" + path.Join(paths.OverridesTarget, "devcontainer.json"),
		"--id-label=daytona.workspace.id=" + opts.Project.WorkspaceId,
		"--id-label=daytona.project.name=" + opts.Project.Name,
		"--skip-non-blocking-commands",
	}

	if prebuild {
		devcontainerCmd = append(devcontainerCmd, "--prebuild")
	}

	err = d.runInitializeCommand(config.MergedConfiguration.InitializeCommand, opts.LogWriter, opts.SshClient)
	if err != nil {
		return "", err
	}

	output, err := d.execInContainer(strings.Join(devcontainerCmd, " "), opts, paths, paths.ProjectTarget, socketForwardId, true, []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: paths.OverridesDir,
			Target: paths.OverridesTarget,
		},
	})
	if err != nil {
		return "", err
	}

	if remoteUser != "" {
		return RemoteUser(remoteUser), nil
	}

	resultIndex := strings.LastIndex(output, "{")
	if resultIndex == -1 {
		return "", fmt.Errorf("unable to find result in devcontainer output")
	}

	resultRaw := output[resultIndex:]

	var result devcontainer.DevcontainerUpResult
	err = json.Unmarshal([]byte(resultRaw), &result)
	if err != nil {
		return "", err
	}

	return RemoteUser(result.RemoteUser), nil
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

func (d *DockerClient) readDevcontainerConfig(opts *CreateProjectOptions, paths DevcontainerPaths, socketForwardId string) (string, *devcontainer.Root, error) {
	opts.LogWriter.Write([]byte("Reading devcontainer configuration...\n"))

	// Sleep is there to make sure the logs get read
	cmd := []string{"cat", paths.TargetConfigFilePath, "&&", "sleep", "1"}

	// We need to override localEnvs to the host env variables
	// FIXME: This will not work for features that require localEnv
	configEnvOverride, err := d.execInContainer(strings.Join(cmd, " "), opts, paths, paths.ProjectTarget, socketForwardId, false, nil)
	if err != nil {
		return "", nil, err
	}

	envVars, err := d.getHostEnvVars(opts.SshClient)
	if err != nil {
		return "", nil, err
	}

	for k, v := range envVars {
		configEnvOverride = strings.ReplaceAll(configEnvOverride, fmt.Sprintf("${localEnv:%s}", k), v)
	}

	writeOverrideCmd := []string{"echo", fmt.Sprintf(`'%s'`, configEnvOverride), ">", "/tmp/devcontainer.json", "&&"}

	devcontainerCmd := append(writeOverrideCmd, []string{
		"devcontainer",
		"read-configuration",
		"--workspace-folder=" + paths.ProjectTarget,
		"--config=" + paths.TargetConfigFilePath,
		"--override-config=/tmp/devcontainer.json",
		"--include-merged-configuration",
	}...)

	output, err := d.execInContainer(strings.Join(devcontainerCmd, " "), opts, paths, paths.ProjectTarget, socketForwardId, false, nil)
	if err != nil {
		return "", nil, err
	}

	configStartIndex := strings.Index(output, "{")
	if configStartIndex == -1 {
		return "", nil, fmt.Errorf("unable to find start of JSON in devcontainer configuration")
	}

	rawConfig := output[configStartIndex:]

	var rootConfig devcontainer.Root
	err = json.Unmarshal([]byte(rawConfig), &rootConfig)
	if err != nil {
		return "", nil, err
	}

	return rawConfig, &rootConfig, nil
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

func (d *DockerClient) execInContainer(cmd string, opts *CreateProjectOptions, paths DevcontainerPaths, workdir, socketForwardId string, writeOutput bool, extraMounts []mount.Mount) (string, error) {
	ctx := context.Background()

	mounts := []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: opts.ProjectDir,
			Target: paths.ProjectTarget,
		},
	}

	if extraMounts != nil {
		mounts = append(mounts, extraMounts...)
	}

	c, err := d.apiClient.ContainerCreate(ctx, &container.Config{
		Image:      "daytonaio/workspace-project",
		Entrypoint: []string{"sh"},
		Env:        []string{"DOCKER_HOST=tcp://localhost:2375"},
		Cmd:        append([]string{"-c"}, cmd),
		Tty:        true,
		WorkingDir: workdir,
	}, &container.HostConfig{
		Privileged:  true,
		NetworkMode: container.NetworkMode(fmt.Sprintf("container:%s", socketForwardId)),
		Mounts:      mounts,
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

	writer := io.MultiWriter(w)
	if writeOutput {
		writer = io.MultiWriter(w, opts.LogWriter)
	}

	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			output += scanner.Text() + "\n"
		}
	}()

	go func() {
		err = d.GetContainerLogs(c.ID, writer)
		if err != nil {
			opts.LogWriter.Write([]byte(fmt.Sprintf("Error running command in container: %v\n", err)))
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

func (d *DockerClient) getRemoteComposeContent(opts *CreateProjectOptions, paths DevcontainerPaths, socketForwardId, composePath string) (string, error) {
	if opts.SshClient == nil {
		return "", nil
	}

	output, err := d.execInContainer("docker compose config", opts, paths, filepath.Dir(composePath), socketForwardId, false, nil)
	if err != nil {
		return "", err
	}

	nameIndex := strings.Index(output, "name: ")
	if nameIndex == -1 {
		return "", fmt.Errorf("unable to find service name in compose config")
	}

	return output[nameIndex:], nil
}

func (d *DockerClient) getDevcontainerPaths(opts *CreateProjectOptions) DevcontainerPaths {
	projectTarget := path.Join("/project", filepath.Base(opts.ProjectDir))
	targetConfigFilePath := path.Join(projectTarget, opts.Project.BuildConfig.Devcontainer.FilePath)

	overridesDir := filepath.Join(filepath.Dir(opts.ProjectDir), fmt.Sprintf("%s-data", filepath.Base(opts.ProjectDir)))
	overridesTarget := "/tmp/overrides"

	return DevcontainerPaths{
		OverridesDir:         overridesDir,
		OverridesTarget:      overridesTarget,
		ProjectTarget:        projectTarget,
		TargetConfigFilePath: targetConfigFilePath,
	}
}

func (d *DockerClient) getHostEnvVars(sshClient *ssh.Client) (map[string]string, error) {
	env := os.Environ()
	if sshClient != nil {
		var err error
		env, err = sshClient.GetEnv(nil)
		if err != nil {
			return nil, err
		}
	}

	envMap := map[string]string{}
	for _, el := range env {
		parts := strings.Split(el, "=")
		envMap[parts[0]] = parts[1]
	}

	return envMap, nil
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
