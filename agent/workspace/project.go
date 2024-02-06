// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/daytonaio/daytona/agent/config"
	config_ssh_key "github.com/daytonaio/daytona/agent/config/ssh_key"
	"github.com/daytonaio/daytona/agent/event_bus"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/reactivex/rxgo/v2"
	log "github.com/sirupsen/logrus"
)

type Project struct {
	Repository Repository `json:"repository"`

	Workspace *Workspace     `json:"-"`
	Events    chan rxgo.Item `json:"-"`
}

type ProjectConfig struct {
	Projects []Project
}

type ProjectEventExtensionPayload struct {
	Extension Extension
	Info      string
}

type ExtensionInfo struct {
	Name string
	Info string
}

type ProjectContainerInfo struct {
	Created   string
	Started   string
	Finished  string
	IP        string
	IsRunning bool
}

type ProjectInfo struct {
	Name          string
	Available     bool
	ContainerInfo *ProjectContainerInfo
	Extensions    []ExtensionInfo
}

func (project Project) GetName() string {
	//	todo: project name must be unique
	return strings.ToLower(strings.TrimSuffix(path.Base(project.Repository.Url), ".git"))
}

func (project Project) GetContainerName() string {
	return project.Workspace.Name + "-" + project.GetName()
}

func (project Project) GetPath() string {
	return path.Join(project.Workspace.Cwd, fmt.Sprintf("%s-%s", project.Workspace.Name, project.GetName()))
}

func (project Project) GetDockerPath() string {
	return path.Join(project.GetPath(), "docker")
}

func (project Project) GetSetupPath() string {
	return path.Join(project.GetPath(), "setup")
}

func (project Project) GetWorkdirPath() string {
	return path.Join(project.GetPath(), "workspace") //	TODO: does it make sense to rename this to something else?
}

func (project Project) GetEnvVars() []string {
	return []string{
		"DAYTONA_WS_NAME=" + project.Workspace.Name,
		"DAYTONA_WS_DIR=" + project.GetName(),
		"DAYTONA_WS_PROJECT_NAME=" + project.GetName(),
		"DAYTONA_WS_PROJECT_REPOSITORY_URL=" + project.Repository.Url,
		//	todo: more vars here?
	}
}

func (project Project) Info() (*ProjectInfo, error) {
	extensions := []ExtensionInfo{}
	for _, extension := range project.Workspace.Extensions {
		info := extension.Info(project)
		if info != "" {
			extensions = append(extensions, ExtensionInfo{
				Name: extension.Name(),
				Info: info,
			})
		}
	}
	available := true
	info, err := project.GetContainerInfo()
	if err != nil {
		if client.IsErrNotFound(err) {
			log.Debug("Container not found, project is not running")
			available = false
		} else {
			return nil, err
		}
	}

	return &ProjectInfo{
		Name:          project.GetName(),
		Available:     available,
		ContainerInfo: info,
		Extensions:    extensions,
	}, nil
}

func (project Project) Start() error {
	event_bus.Publish(event_bus.Event{
		Name: event_bus.ProjectEventCloningRepo,
		Payload: event_bus.ProjectEventPayload{
			ProjectName:   project.GetName(),
			WorkspaceName: project.Workspace.Name,
		},
	})

	err := project.startContainer()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for _, extension := range project.Workspace.Extensions {
		wg.Add(1)
		log.WithFields(log.Fields{
			"project":   project.GetName(),
			"extension": extension.Name(),
		}).Info("Starting extension")

		event_bus.Publish(event_bus.Event{
			Name: event_bus.ProjectEventStartingExtension,
			Payload: event_bus.ProjectEventPayload{
				ProjectName:   project.GetName(),
				WorkspaceName: project.Workspace.Name,
				ExtensionName: extension.Name(),
			},
		})

		go func(extension Extension) {
			err := extension.Start(project)
			if err != nil {
				log.WithFields(log.Fields{
					"project":   project.GetName(),
					"extension": extension.Name(),
				}).Error(err)
			}
		}(extension)

		go func(extension Extension) {
			timeout := time.After(time.Duration(extension.LivenessProbeTimeout()) * time.Second)
			for {
				resultCh := make(chan bool)
				go func() {
					started, err := extension.LivenessProbe(project)
					if err != nil {
						log.WithFields(log.Fields{
							"project":   project.GetName(),
							"extension": extension.Name(),
						}).Error(err)
					}
					resultCh <- started
				}()

				select {
				case <-timeout:
					log.WithFields(log.Fields{
						"project":   project.GetName(),
						"extension": extension.Name(),
					}).Error("Liveness probe timeout reached")
					wg.Done()
					return
				case started := <-resultCh:
					if started {
						log.WithFields(log.Fields{
							"project":   project.GetName(),
							"extension": extension.Name(),
						}).Info("Started extension")

						wg.Done()
						return
					}
				}

				time.Sleep(1 * time.Second)
			}
		}(extension)
	}

	wg.Wait()

	event_bus.Publish(event_bus.Event{
		Name: event_bus.ProjectEventStarted,
		Payload: event_bus.ProjectEventPayload{
			ProjectName:   project.GetName(),
			WorkspaceName: project.Workspace.Name,
		},
	})

	return nil
}

func (project Project) Stop() error {
	event_bus.Publish(event_bus.Event{
		Name: event_bus.ProjectEventStopping,
		Payload: event_bus.ProjectEventPayload{
			ProjectName:   project.GetName(),
			WorkspaceName: project.Workspace.Name,
		},
	})

	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	err = cli.ContainerStop(ctx, project.GetContainerName(), container.StopOptions{})
	if err != nil {
		return err
	}

	// make sure container is running
	//	TODO: timeout
	for {
		inspect, err := cli.ContainerInspect(ctx, project.GetContainerName())
		if err != nil {
			return err
		}

		if !inspect.State.Running {
			break
		}

		time.Sleep(1 * time.Second)
	}
	var wg sync.WaitGroup

	wg.Wait()

	event_bus.Publish(event_bus.Event{
		Name: event_bus.ProjectEventStopped,
		Payload: event_bus.ProjectEventPayload{
			ProjectName:   project.GetName(),
			WorkspaceName: project.Workspace.Name,
		},
	})

	return nil
}

func (project Project) Remove(force bool) error {
	event_bus.Publish(event_bus.Event{
		Name: event_bus.ProjectEventRemoving,
		Payload: event_bus.ProjectEventPayload{
			ProjectName:   project.GetName(),
			WorkspaceName: project.Workspace.Name,
		},
	})

	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	err = cli.ContainerRemove(ctx, project.GetContainerName(), types.ContainerRemoveOptions{
		Force:         force,
		RemoveVolumes: true,
	})
	if err != nil && !client.IsErrNotFound(err) {
		return err
	}

	err = cli.VolumeRemove(ctx, project.getDockerVolumeName(), force)
	if err != nil && !client.IsErrNotFound(err) {
		return err
	}

	err = os.RemoveAll(project.GetPath())
	if err != nil {
		return err
	}

	event_bus.Publish(event_bus.Event{
		Name: event_bus.ProjectEventRemoved,
		Payload: event_bus.ProjectEventPayload{
			ProjectName:   project.GetName(),
			WorkspaceName: project.Workspace.Name,
		},
	})

	return nil
}

func (project Project) initialize() error {
	log.Info("Initializing project: ", project.GetName())
	event_bus.Publish(event_bus.Event{
		Name: event_bus.ProjectEventInitializing,
		Payload: event_bus.ProjectEventPayload{
			ProjectName:   project.GetName(),
			WorkspaceName: project.Workspace.Name,
		},
	})

	err := os.MkdirAll(project.GetDockerPath(), 0755)
	if err != nil {
		return err
	}

	err = os.MkdirAll(project.GetWorkdirPath(), 0755)
	if err != nil {
		return err
	}

	err = project.cloneRepository()
	if err != nil {
		return err
	}

	for _, extension := range project.Workspace.Extensions {
		log.WithFields(log.Fields{
			"project":   project.GetName(),
			"extension": extension.Name(),
		}).Info("Preparing extension")

		event_bus.Publish(event_bus.Event{
			Name: event_bus.ProjectEventPreparingExtension,
			Payload: event_bus.ProjectEventPayload{
				ProjectName:   project.GetName(),
				WorkspaceName: project.Workspace.Name,
				ExtensionName: extension.Name(),
			},
		})

		err = extension.PreInit(project)
		if err != nil {
			return err
		}

		log.WithFields(log.Fields{
			"project":   project.GetName(),
			"extension": extension.Name(),
		}).Info("Extension prepared")
	}

	err = project.initContainer()
	if err != nil {
		return err
	}

	err = project.startContainer()
	if err != nil {
		return err
	}

	for _, extension := range project.Workspace.Extensions {
		log.WithFields(log.Fields{
			"project":   project.GetName(),
			"extension": extension.Name(),
		}).Info("Initializing extension")

		event_bus.Publish(event_bus.Event{
			Name: event_bus.ProjectEventInitializingExtension,
			Payload: event_bus.ProjectEventPayload{
				ProjectName:   project.GetName(),
				WorkspaceName: project.Workspace.Name,
				ExtensionName: extension.Name(),
			},
		})

		err = extension.Init(project)
		if err != nil {
			return err
		}

		log.WithFields(log.Fields{
			"project":   project.GetName(),
			"extension": extension.Name(),
		}).Info("Extension initialized")
	}

	if project.isDevcontainer() {
		return project.initDevcontainerProject()
	}

	event_bus.Publish(event_bus.Event{
		Name: event_bus.ProjectEventInitialized,
		Payload: event_bus.ProjectEventPayload{
			ProjectName:   project.GetName(),
			WorkspaceName: project.Workspace.Name,
		},
	})

	return nil
}

func (project Project) initContainer() error {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	config, err := config.GetConfig()
	if err != nil {
		return err
	}

	imageName := config.ProjectBaseImage
	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		return err
	}

	found := false
	for _, image := range images {
		for _, tag := range image.RepoTags {
			if strings.HasPrefix(tag, imageName) {
				found = true
				break
			}
		}
	}

	if !found {
		log.Info("Image not found, pulling...")
		responseBody, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
		if err != nil {
			return err
		}
		defer responseBody.Close()
		_, err = io.Copy(io.Discard, responseBody)
		if err != nil {
			return err
		}
		log.Info("Image pulled successfully")
	}

	var extensions []string
	for _, extension := range project.Workspace.Extensions {
		extensions = append(extensions, extension.Name())
	}
	extensionList := strings.Join(extensions, ", ")

	mounts := []mount.Mount{
		{
			Type:   mount.TypeVolume,
			Source: project.getDockerVolumeName(),
			Target: "/var/lib/docker",
		},
	}

	_, err = cli.ContainerCreate(ctx, &container.Config{
		Hostname: project.GetName(),
		Image:    imageName,
		Labels: map[string]string{
			"daytona.workspace.name":                   project.Workspace.Name,
			"daytona.workspace.cwd":                    project.Workspace.Cwd,
			"daytona.workspace.extensions":             extensionList,
			"daytona.workspace.project.name":           project.GetName(),
			"daytona.workspace.project.repository.url": project.Repository.Url,
			// todo: Add more properties here
		},
		Env: project.GetEnvVars(),
	}, &container.HostConfig{
		Privileged: true,
		Binds: []string{
			fmt.Sprintf("%s:/%s", project.GetWorkdirPath(), project.GetName()),
			project.GetSetupPath() + ":/setup",
			"/tmp/daytona:/tmp/daytona",
		},
		Mounts:      mounts,
		NetworkMode: container.NetworkMode(project.Workspace.Name),
	}, nil, nil, project.GetContainerName()) //	TODO: namespaced names
	if err != nil {
		return err
	}

	return nil
}

func (project Project) startContainer() error {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	err = cli.ContainerStart(ctx, project.GetContainerName(), types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	// make sure container is running
	//	TODO: timeout
	for {
		inspect, err := cli.ContainerInspect(ctx, project.GetContainerName())
		if err != nil {
			return err
		}

		if inspect.State.Running {
			break
		}

		time.Sleep(1 * time.Second)
	}

	//	copy daytona binary
	daytonaPath, err := os.Executable()
	if err != nil {
		return err
	}

	file, err := os.Open(daytonaPath)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(fileInfo, fileInfo.Name())
	if err != nil {
		return err
	}

	// Set the name of the file in the tar archive
	header.Name = "daytona"

	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(tw, file)
	if err != nil {
		return err
	}

	err = tw.Close()
	if err != nil {
		return err
	}

	// Copy the file content to the container
	err = cli.CopyToContainer(ctx, project.GetContainerName(), "/usr/local/bin", buf, types.CopyToContainerOptions{})
	if err != nil {
		return err
	}

	// start dockerd
	execConfig := types.ExecConfig{
		Tty:          true,
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{"dockerd"},
		User:         "root",
	}
	execResp, err := cli.ContainerExecCreate(ctx, project.GetContainerName(), execConfig)
	if err != nil {
		return err
	}

	log.Debug("Daytona binary copied to container")

	err = cli.ContainerExecStart(ctx, execResp.ID, types.ExecStartCheck{})
	if err != nil {
		return err
	}

	//	todo: wait for dockerd to start
	time.Sleep(3 * time.Second)

	return nil
}

func (project Project) cloneRepository() error {
	repo := project.Repository
	clonePath := project.GetWorkdirPath()

	existingRepo, err := git.PlainOpen(clonePath)
	if err != git.ErrRepositoryNotExists {
		return err
	}
	if existingRepo != nil {
		log.WithFields(log.Fields{
			"project": project.GetName(),
		}).Info("Repository " + repo.Url + " exists. Skipping clone.")
		return nil
	}

	privateKey, err := config_ssh_key.GetPrivateKey()

	cloneOptions := &git.CloneOptions{
		URL: repo.Url,
	}

	if err == nil && privateKey != nil {
		cloneOptions.Auth = &ssh.PublicKeys{
			User:   "git",
			Signer: *privateKey,
		}
		cloneOptions.URL = httpToSsh(repo.Url)
	}

	log.WithFields(log.Fields{
		"project": project.GetName(),
	}).Info("Cloning repository: " + cloneOptions.URL)

	// if strings.Contains(repo.Url, "github.com") || strings.Contains(repo.Url, "gitlab.com") || strings.Contains(repo.Url, "bitbucket.org") {
	// 	log.Debug("The Git URL contains known hostnames, no need to skip certificate verification")
	// } else {
	// 	cloneOptions.FetchOptions.RemoteCallbacks.CertificateCheckCallback = func(cert *git.Certificate, valid bool, hostname string) git.ErrorCode {
	// 		return 0
	// 	}
	// }

	// If the branch is equal to SHA, then we need to reset to the SHA
	if repo.Branch != nil && repo.Branch != repo.SHA {
		cloneOptions.ReferenceName = plumbing.ReferenceName("refs/heads/" + *repo.Branch)
		cloneOptions.SingleBranch = true
	}

	//	todo: repo - lookup commit by sha and checkout
	_, err = git.PlainClone(clonePath, false, cloneOptions)
	if err != nil {
		if strings.Contains(err.Error(), "403") {
			return errors.New("unauthorized")
		}
		if err == transport.ErrEmptyRemoteRepository {
			return project.initializeEmtpyRepository(clonePath)
		}
		return err
	}

	// if repo.Branch != nil && repo.Branch == repo.SHA {
	// 	oid, err := git.NewOid(*repo.SHA)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	commit, err := repository.LookupCommit(oid)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	tree, err := commit.Tree()
	// 	if err != nil {
	// 		return err
	// 	}

	// 	log.Info("\nChecking out " + *repo.SHA + "...\n")
	// 	err = repository.CheckoutTree(tree, &git.CheckoutOptions{
	// 		Strategy: git.CheckoutForce,
	// 	})
	// 	if err != nil {
	// 		return err
	// 	}

	// 	err = repository.SetHeadDetached(commit.Id())
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

func (p Project) GetContainerInfo() (*ProjectContainerInfo, error) {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	inspect, err := cli.ContainerInspect(ctx, p.GetContainerName())
	if err != nil {
		return nil, err
	}

	info := &ProjectContainerInfo{
		Created:   inspect.Created,
		Started:   inspect.State.StartedAt,
		Finished:  inspect.State.FinishedAt,
		IP:        inspect.NetworkSettings.Networks[p.Workspace.Name].IPAddress,
		IsRunning: inspect.State.Running,
	}

	return info, nil
}

func (p Project) initializeEmtpyRepository(clonePath string) error {
	repo := p.Repository

	log.WithFields(log.Fields{
		"project": p.GetName(),
	}).Info("Initializing empty repository: " + repo.Url)
	// Initialize emtpy repo
	_, err := git.PlainInit(clonePath, false)
	if err != nil {
		return err
	}

	repository, err := git.PlainOpen(clonePath)
	if err != nil {
		return err
	}

	_, err = repository.CreateRemote(&gitconfig.RemoteConfig{
		Name: "origin",
		URLs: []string{repo.Url},
	})
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"project": p.GetName(),
	}).Info("Initialized empty repository: " + repo.Url)

	return nil
}

func (p Project) getDockerVolumeName() string {
	return p.GetContainerName() + "-docker"
}

func httpToSsh(repoUrl string) string {
	if strings.HasPrefix(repoUrl, "https://") {
		source := strings.Split(strings.TrimPrefix(repoUrl, "https://"), "/")[0]
		repo := strings.TrimPrefix(repoUrl, "https://"+source+"/")
		repo = strings.TrimSuffix(repo, ".git")
		return fmt.Sprintf("git@%s:%s.git", source, repo)
	}
	return repoUrl
}
