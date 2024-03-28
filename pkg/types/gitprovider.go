// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package types

type GitProvider struct {
	Id         string `json:"id"`
	Username   string `json:"username"`
	Token      string `json:"token"`
	BaseApiUrl string `json:"baseApiUrl"`
} // @name GitProvider

type GitUser struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
} // @name GitUser

type GitRepository struct {
	Id       string   `json:"id"`
	Url      string   `json:"url"`
	Name     string   `json:"name"`
	Branch   *string  `json:"branch,omitempty"`
	Sha      string   `json:"sha"`
	Owner    string   `json:"owner"`
	PrNumber *uint32  `json:"prNumber,omitempty"`
	Source   string   `json:"source"`
	Path     *string  `json:"path,omitempty"`
	GitUser  *GitUser `json:"-"`
} // @name GitRepository

type GitNamespace struct {
	Id   string `json:"id"`
	Name string `json:"name"`
} // @name GitNamespace

type GitBranch struct {
	Name string `json:"name"`
	SHA  string `json:"sha"`
} // @name GitBranch

type GitPullRequest struct {
	Name   string `json:"name"`
	Branch string `json:"branch"`
} // @name GitPullRequest
