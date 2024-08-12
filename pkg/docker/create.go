// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/daytonaio/daytona/pkg/build/detect"
	"github.com/daytonaio/daytona/pkg/git"
	"github.com/daytonaio/daytona/pkg/ssh"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	log "github.com/sirupsen/logrus"
)

func (d *DockerClient) CreateWorkspace(workspace *workspace.Workspace, workspaceDir string, logWriter io.Writer, sshClient *ssh.Client) error {
	var err error
	if sshClient == nil {
		err = os.MkdirAll(workspaceDir, 0755)
	} else {
		err = sshClient.Exec(fmt.Sprintf("mkdir -p %s", workspaceDir), nil)
	}

	return err
}

func (d *DockerClient) CreateProject(opts *CreateProjectOptions) error {
	// TODO: The image should be configurable
	err := d.PullImage("daytonaio/workspace-project", nil, opts.LogWriter)
	if err != nil {
		return err
	}

	err = d.cloneProjectRepository(opts)
	if err != nil {
		return err
	}

	builderType, err := detect.DetectProjectBuilderType(opts.Project, opts.ProjectDir, opts.SshClient)
	if err != nil {
		return err
	}

	switch builderType {
	case detect.BuilderTypeDevcontainer:
		_, err := d.createProjectFromDevcontainer(opts, true)
		return err
	case detect.BuilderTypeImage:
		return d.createProjectFromImage(opts)
	default:
		return fmt.Errorf("unknown builder type: %s", builderType)
	}
}

func (d *DockerClient) cloneProjectRepository(opts *CreateProjectOptions) error {
	ctx := context.Background()

	if opts.SshClient != nil {
		err := opts.SshClient.Exec(fmt.Sprintf("mkdir -p %s", opts.ProjectDir), nil)
		if err != nil {
			return err
		}
	} else {
		err := os.MkdirAll(opts.ProjectDir, 0755)
		if err != nil {
			return err
		}
	}

	var auth *http.BasicAuth
	if opts.Gpc != nil {
		auth = &http.BasicAuth{
			Username: opts.Gpc.Username,
			Password: opts.Gpc.Token,
		}
	}

	gitService := git.Service{
		ProjectDir: fmt.Sprintf("/workdir/%s-%s", opts.Project.WorkspaceId, opts.Project.Name),
	}

	cloneCmd := gitService.CloneRepositoryCmd(opts.Project, auth)

	c, err := d.apiClient.ContainerCreate(ctx, &container.Config{
		Image:      "daytonaio/workspace-project",
		Entrypoint: []string{"sleep"},
		Cmd:        []string{"infinity"},
		Env: []string{
			"GIT_SSL_NO_VERIFY=true",
		},
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: filepath.Dir(opts.ProjectDir),
				Target: "/workdir",
			},
		},
	}, nil, nil, fmt.Sprintf("git-clone-%s-%s", opts.Project.WorkspaceId, opts.Project.Name))
	if err != nil {
		return err
	}

	defer d.removeContainer(c.ID) // nolint:errcheck

	err = d.apiClient.ContainerStart(ctx, c.ID, container.StartOptions{})
	if err != nil {
		return err
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

	containerUser, err := d.updateContainerUserUidGid(c.ID, opts)

	res, err := d.ExecSync(c.ID, container.ExecOptions{
		User: containerUser,
		Cmd:  append([]string{"sh", "-c"}, strings.Join(cloneCmd, " ")),
	}, opts.LogWriter)
	if err != nil {
		return err
	}
	if res.ExitCode != 0 {
		return fmt.Errorf("git clone failed with exit code %d", res.ExitCode)
	}

	return nil
}

func (d *DockerClient) updateContainerUserUidGid(containerId string, opts *CreateProjectOptions) (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	containerUser := "daytona"
	newUid := currentUser.Uid
	newGid := currentUser.Gid

	if opts.SshClient != nil {
		newUid, newGid, err = opts.SshClient.GetUserUidGid()
		if err != nil {
			return "", err
		}
	}

	if newUid == "0" && newGid == "0" {
		containerUser = "root"
	}

	/*
		Patch UID and GID of the user cloning the repository
	*/
	if containerUser != "root" {
		_, err = d.ExecSync(containerId, container.ExecOptions{
			User: "root",
			Cmd:  []string{"sh", "-c", UPDATE_UID_GID_SCRIPT},
			Env: []string{
				fmt.Sprintf("REMOTE_USER=%s", containerUser),
				fmt.Sprintf("NEW_UID=%s", newUid),
				fmt.Sprintf("NEW_GID=%s", newGid),
			},
		}, opts.LogWriter)
		if err != nil {
			return "", err
		}
	}

	return containerUser, nil
}

const UPDATE_UID_GID_SCRIPT = `eval $(sed -n "s/${REMOTE_USER}:[^:]*:\([^:]*\):\([^:]*\):[^:]*:\([^:]*\).*/OLD_UID=\1;OLD_GID=\2;HOME_FOLDER=\3/p" /etc/passwd); \
eval $(sed -n "s/\([^:]*\):[^:]*:${NEW_UID}:.*/EXISTING_USER=\1/p" /etc/passwd); \
eval $(sed -n "s/\([^:]*\):[^:]*:${NEW_GID}:.*/EXISTING_GROUP=\1/p" /etc/group); \
if [ -z "$OLD_UID" ]; then \
	echo "Remote user not found in /etc/passwd ($REMOTE_USER)."; \
elif [ "$OLD_UID" = "$NEW_UID" -a "$OLD_GID" = "$NEW_GID" ]; then \
	echo "UIDs and GIDs are the same ($NEW_UID:$NEW_GID)."; \
elif [ "$OLD_UID" != "$NEW_UID" -a -n "$EXISTING_USER" ]; then \
	echo "User with UID exists ($EXISTING_USER=$NEW_UID)."; \
else \
	if [ "$OLD_GID" != "$NEW_GID" -a -n "$EXISTING_GROUP" ]; then \
		echo "Group with GID exists ($EXISTING_GROUP=$NEW_GID)."; \
		NEW_GID="$OLD_GID"; \
	fi; \
	echo "Updating UID:GID from $OLD_UID:$OLD_GID to $NEW_UID:$NEW_GID."; \
	sed -i -e "s/\(${REMOTE_USER}:[^:]*:\)[^:]*:[^:]*/\1${NEW_UID}:${NEW_GID}/" /etc/passwd; \
	if [ "$OLD_GID" != "$NEW_GID" ]; then \
		sed -i -e "s/\([^:]*:[^:]*:\)${OLD_GID}:/\1${NEW_GID}:/" /etc/group; \
	fi; \
	chown -R $NEW_UID:$NEW_GID $HOME_FOLDER; \
fi;`
