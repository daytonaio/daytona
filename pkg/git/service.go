// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/target/project"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/capability"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"gopkg.in/ini.v1"
)

var MapStatus map[git.StatusCode]project.Status = map[git.StatusCode]project.Status{
	git.Unmodified:         project.Unmodified,
	git.Untracked:          project.Untracked,
	git.Modified:           project.Modified,
	git.Added:              project.Added,
	git.Deleted:            project.Deleted,
	git.Renamed:            project.Renamed,
	git.Copied:             project.Copied,
	git.UpdatedButUnmerged: project.UpdatedButUnmerged,
}

type IGitService interface {
	CloneRepository(repo *gitprovider.GitRepository, auth *http.BasicAuth) error
	CloneRepositoryCmd(repo *gitprovider.GitRepository, auth *http.BasicAuth) []string
	RepositoryExists() (bool, error)
	SetGitConfig(userData *gitprovider.GitUser, providerConfig *gitprovider.GitProviderConfig) error
	GetGitStatus() (*project.GitStatus, error)
}

type Service struct {
	ProjectDir        string
	GitConfigFileName string
	LogWriter         io.Writer
	OpenRepository    *git.Repository
}

func (s *Service) CloneRepository(repo *gitprovider.GitRepository, auth *http.BasicAuth) error {
	cloneOptions := &git.CloneOptions{
		URL:             repo.Url,
		SingleBranch:    true,
		InsecureSkipTLS: true,
		Auth:            auth,
	}

	if s.LogWriter != nil {
		cloneOptions.Progress = s.LogWriter
	}

	// Azure DevOps requires capabilities multi_ack / multi_ack_detailed,
	// which are not fully implemented and by default are included in
	// transport.UnsupportedCapabilities.
	//
	// This can be removed once go-git implements the git v2 protocol.
	transport.UnsupportedCapabilities = []capability.Capability{
		capability.ThinPack,
	}

	cloneOptions.ReferenceName = plumbing.ReferenceName("refs/heads/" + repo.Branch)

	_, err := git.PlainClone(s.ProjectDir, false, cloneOptions)
	if err != nil {
		return err
	}

	if repo.Target == gitprovider.CloneTargetCommit {
		r, err := git.PlainOpen(s.ProjectDir)
		if err != nil {
			return err
		}

		w, err := r.Worktree()
		if err != nil {
			return err
		}

		err = w.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(repo.Sha),
		})
		if err != nil {
			return err
		}
	}

	return err
}

func (s *Service) CloneRepositoryCmd(repo *gitprovider.GitRepository, auth *http.BasicAuth) []string {
	cloneCmd := []string{"git", "clone", "--single-branch", "--branch", fmt.Sprintf("\"%s\"", repo.Branch)}
	cloneUrl := repo.Url

	// Default to https protocol if not specified
	if !strings.Contains(cloneUrl, "://") {
		cloneUrl = fmt.Sprintf("https://%s", cloneUrl)
	}

	if auth != nil {
		cloneUrl = fmt.Sprintf("%s://%s:%s@%s", strings.Split(cloneUrl, "://")[0], auth.Username, auth.Password, strings.SplitN(cloneUrl, "://", 2)[1])
	}

	cloneCmd = append(cloneCmd, cloneUrl, s.ProjectDir)

	if repo.Target == gitprovider.CloneTargetCommit {
		cloneCmd = append(cloneCmd, "&&", "cd", s.ProjectDir)
		cloneCmd = append(cloneCmd, "&&", "git", "checkout", repo.Sha)
	}

	return cloneCmd
}

func (s *Service) RepositoryExists() (bool, error) {
	_, err := os.Stat(filepath.Join(s.ProjectDir, ".git"))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *Service) SetGitConfig(userData *gitprovider.GitUser, providerConfig *gitprovider.GitProviderConfig) error {
	gitConfigFileName := s.GitConfigFileName

	var gitConfigContent []byte
	gitConfigContent, err := os.ReadFile(gitConfigFileName)
	if err != nil {
		gitConfigContent = []byte{}
	}

	cfg, err := ini.Load(gitConfigContent)
	if err != nil {
		return err
	}

	if !cfg.HasSection("credential") {
		_, err := cfg.NewSection("credential")
		if err != nil {
			return err
		}
	}

	_, err = cfg.Section("credential").NewKey("helper", "/usr/local/bin/daytona git-cred")
	if err != nil {
		return err
	}

	if !cfg.HasSection("safe") {
		_, err := cfg.NewSection("safe")
		if err != nil {
			return err
		}
	}
	_, err = cfg.Section("safe").NewKey("directory", s.ProjectDir)
	if err != nil {
		return err
	}

	if userData != nil {
		if !cfg.HasSection("user") {
			_, err := cfg.NewSection("user")
			if err != nil {
				return err
			}
		}

		_, err := cfg.Section("user").NewKey("name", userData.Name)
		if err != nil {
			return err
		}

		_, err = cfg.Section("user").NewKey("email", userData.Email)
		if err != nil {
			return err
		}
	}

	if err := s.setSigningConfig(cfg, providerConfig, userData); err != nil {
		return err
	}

	var buf bytes.Buffer
	_, err = cfg.WriteTo(&buf)
	if err != nil {
		return err
	}

	return os.WriteFile(gitConfigFileName, buf.Bytes(), 0644)
}

func (s *Service) setSigningConfig(cfg *ini.File, providerConfig *gitprovider.GitProviderConfig, userData *gitprovider.GitUser) error {
	if providerConfig == nil || providerConfig.SigningMethod == nil || providerConfig.SigningKey == nil {
		return nil
	}

	if !cfg.HasSection("user") {
		_, err := cfg.NewSection("user")
		if err != nil {
			return err
		}
	}

	_, err := cfg.Section("user").NewKey("signingkey", *providerConfig.SigningKey)
	if err != nil {
		return err
	}

	if !cfg.HasSection("commit") {
		_, err := cfg.NewSection("commit")
		if err != nil {
			return err
		}
	}

	switch *providerConfig.SigningMethod {
	case gitprovider.SigningMethodGPG:
		_, err := cfg.Section("commit").NewKey("gpgSign", "true")
		if err != nil {
			return err
		}
	case gitprovider.SigningMethodSSH:
		err := s.configureAllowedSigners(userData.Email, *providerConfig.SigningKey)
		if err != nil {
			return err
		}

		if !cfg.HasSection("gpg") {
			_, err := cfg.NewSection("gpg")
			if err != nil {
				return err
			}
		}
		_, err = cfg.Section("gpg").NewKey("format", "ssh")
		if err != nil {
			return err
		}

		if !cfg.HasSection("gpg \"ssh\"") {
			_, err := cfg.NewSection("gpg \"ssh\"")
			if err != nil {
				return err
			}
		}

		allowedSignersFile := filepath.Join(os.Getenv("HOME"), ".ssh/allowed_signers")
		_, err = cfg.Section("gpg \"ssh\"").NewKey("allowedSignersFile", allowedSignersFile)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) configureAllowedSigners(email, sshKey string) error {
	homeDir := os.Getenv("HOME")
	sshDir := filepath.Join(homeDir, ".ssh")
	allowedSignersFile := filepath.Join(sshDir, "allowed_signers")

	err := os.MkdirAll(sshDir, 0700)
	if err != nil {
		return fmt.Errorf("failed to create SSH directory: %w", err)
	}

	entry := fmt.Sprintf("%s namespaces=\"git\" %s\n", email, sshKey)

	existingContent, err := os.ReadFile(allowedSignersFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read allowed_signers file: %w", err)
	}

	newContent := string(existingContent) + entry

	err = os.WriteFile(allowedSignersFile, []byte(newContent), 0600)
	if err != nil {
		return fmt.Errorf("failed to write to allowed_signers file: %w", err)
	}

	return nil
}

func (s *Service) isBranchPublished() (bool, error) {
	upstream, err := s.getUpstreamBranch()
	if err != nil {
		return false, err
	}
	return upstream != "", nil
}

func (s *Service) getUpstreamBranch() (string, error) {
	cmd := exec.Command("git", "-C", s.ProjectDir, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", nil
	}

	return strings.TrimSpace(string(out)), nil
}

func (s *Service) getAheadBehindInfo() (int, int, error) {
	upstream, err := s.getUpstreamBranch()
	if err != nil {
		return 0, 0, err
	}
	if upstream == "" {
		return 0, 0, nil
	}

	cmd := exec.Command("git", "-C", s.ProjectDir, "rev-list", "--left-right", "--count", fmt.Sprintf("%s...HEAD", upstream))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, 0, nil
	}

	return parseAheadBehind(out)
}

func parseAheadBehind(output []byte) (int, int, error) {
	counts := strings.Split(strings.TrimSpace(string(output)), "\t")
	if len(counts) != 2 {
		return 0, 0, nil
	}

	ahead, err := strconv.Atoi(counts[1])
	if err != nil {
		return 0, 0, nil
	}

	behind, err := strconv.Atoi(counts[0])
	if err != nil {
		return 0, 0, nil
	}

	return ahead, behind, nil
}

func (s *Service) GetGitStatus() (*project.GitStatus, error) {
	repo, err := git.PlainOpen(s.ProjectDir)
	if err != nil {
		return nil, err
	}

	ref, err := repo.Head()
	if err != nil {
		return nil, err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	status, err := worktree.Status()
	if err != nil {
		return nil, err
	}

	files := []*project.FileStatus{}
	for path, file := range status {
		files = append(files, &project.FileStatus{
			Name:     path,
			Extra:    file.Extra,
			Staging:  MapStatus[file.Staging],
			Worktree: MapStatus[file.Worktree],
		})
	}

	branchPublished, err := s.isBranchPublished()
	if err != nil {
		return nil, err
	}

	ahead, behind, err := s.getAheadBehindInfo()
	if err != nil {
		return nil, err
	}

	return &project.GitStatus{
		CurrentBranch:   ref.Name().Short(),
		Files:           files,
		BranchPublished: branchPublished,
		Ahead:           ahead,
		Behind:          behind,
	}, nil
}
