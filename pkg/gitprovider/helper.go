package gitprovider

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

func parseGitURLWithSSH(gitURL string) (*GitRepository, error) {
	re := regexp.MustCompile(`git@([\w\.]+):(.+?)/(.+?)(?:\.git)?$`)
	matches := re.FindStringSubmatch(gitURL)
	if len(matches) != 4 {
		return nil, errors.New("cannot parse git URL: " + gitURL)
	}

	source := matches[1]
	owner := matches[2]
	repo := matches[3]

	cloneURL := getCloneURL(source, owner, repo)

	return &GitRepository{
		Owner:  owner,
		Name:   repo,
		Source: source,
		Url:    cloneURL,
	}, nil
}

func getCloneURL(source, owner, repo string) string {
	return fmt.Sprintf("https://%s/%s/%s.git", source, owner, repo)
}

func parseGitComponents(gitURL string) (*GitRepository, error) {
	if strings.HasPrefix(gitURL, "git@") {
		return parseGitURLWithSSH(gitURL)
	}

	if !strings.HasPrefix(gitURL, "http") {
		return nil, errors.New("cannot parse git URL: " + gitURL)
	}

	u, err := url.Parse(gitURL)
	if err != nil {
		return nil, err
	}

	repo := &GitRepository{}

	path := strings.TrimPrefix(u.Path, "/")
	parts := strings.Split(path, "/")

	repo.Source = u.Host
	repo.Owner = parts[0]
	repo.Name = parts[1]
	branchPath := strings.Join(parts[2:], "/")
	repo.Path = &branchPath

	repo.Url = getCloneURL(repo.Source, repo.Owner, repo.Name)

	return repo, nil
}