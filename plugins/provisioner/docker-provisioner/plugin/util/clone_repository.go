package util

import (
	"errors"
	"fmt"
	"strings"

	config_ssh_key "github.com/daytonaio/daytona/agent/config/ssh_key"
	"github.com/daytonaio/daytona/agent/workspace"

	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	log "github.com/sirupsen/logrus"
)

func CloneRepository(project workspace.Project, clonePath string) error {
	repo := project.Repository

	existingRepo, err := git.PlainOpen(clonePath)
	if err != git.ErrRepositoryNotExists {
		return err
	}
	if existingRepo != nil {
		log.WithFields(log.Fields{
			"project": project.Name,
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
		"project": project.Name,
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
			return initializeEmtpyRepository(project, clonePath)
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

func httpToSsh(repoUrl string) string {
	if strings.HasPrefix(repoUrl, "https://") {
		source := strings.Split(strings.TrimPrefix(repoUrl, "https://"), "/")[0]
		repo := strings.TrimPrefix(repoUrl, "https://"+source+"/")
		repo = strings.TrimSuffix(repo, ".git")
		return fmt.Sprintf("git@%s:%s.git", source, repo)
	}
	return repoUrl
}

func initializeEmtpyRepository(project workspace.Project, clonePath string) error {
	repo := project.Repository

	log.WithFields(log.Fields{
		"project": project.Name,
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
		"project": project.Name,
	}).Info("Initialized empty repository: " + repo.Url)

	return nil
}
