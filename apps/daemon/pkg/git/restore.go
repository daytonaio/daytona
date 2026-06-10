// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/format/index"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type restoreItem struct {
	path string
	hash plumbing.Hash
	mode filemode.FileMode
}

// Restore restores working tree files and/or unstages changes, mirroring
// `git restore`. When neither staged nor worktree is set the working tree is
// restored. The staged side uses go-git; the working-tree side is materialized
// blob-by-blob (go-git has no path-scoped worktree checkout).
func (s *Service) Restore(files []string, staged, worktree *bool, source string) error {
	if len(files) == 0 {
		return fmt.Errorf("at least one path is required")
	}

	restoreStaged, restoreWorktree := resolveRestoreTargets(staged, worktree)
	if !restoreStaged && !restoreWorktree {
		return fmt.Errorf("at least one of staged or worktree must be true")
	}

	repo, err := git.PlainOpen(s.WorkDir)
	if err != nil {
		return err
	}

	if source == "" {
		return s.restoreFromHead(repo, files, restoreStaged, restoreWorktree)
	}
	return s.restoreFromSource(repo, source, files, restoreStaged, restoreWorktree)
}

func (s *Service) restoreFromHead(repo *git.Repository, files []string, restoreStaged, restoreWorktree bool) error {
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	// Staged restore from HEAD (optionally with worktree) is handled by go-git.
	if restoreStaged {
		return worktree.Restore(&git.RestoreOptions{Staged: true, Worktree: restoreWorktree, Files: files})
	}

	// Worktree-only restore from the index: go-git can't do this directly.
	idx, err := repo.Storer.Index()
	if err != nil {
		return err
	}
	items, err := matchIndex(idx, files)
	if err != nil {
		return err
	}
	return s.materializeAll(repo, items)
}

func (s *Service) restoreFromSource(repo *git.Repository, source string, files []string, restoreStaged, restoreWorktree bool) error {
	tree, err := resolveTree(repo, source)
	if err != nil {
		return err
	}

	available := treeItems(tree)

	if restoreStaged {
		if err := s.restoreStagedFromTree(repo, available, files); err != nil {
			return err
		}
	}

	if restoreWorktree {
		items, err := matchTreeItems(available, files)
		if err != nil {
			return err
		}
		if err := s.materializeAll(repo, items); err != nil {
			return err
		}
	}

	return nil
}

// restoreStagedFromTree rewrites the index entries for the given paths to the
// source tree's content (without moving HEAD).
func (s *Service) restoreStagedFromTree(repo *git.Repository, available map[string]restoreItem, files []string) error {
	idx, err := repo.Storer.Index()
	if err != nil {
		return err
	}

	for _, p := range files {
		for _, name := range indexNamesUnder(idx, p) {
			_, _ = idx.Remove(name)
		}
		for _, it := range matchAvailable(available, p) {
			entry := idx.Add(it.path)
			entry.Hash = it.hash
			entry.Mode = it.mode
		}
	}

	return repo.Storer.SetIndex(idx)
}

func (s *Service) materializeAll(repo *git.Repository, items []restoreItem) error {
	for _, item := range items {
		if err := s.materialize(repo, item); err != nil {
			return err
		}
	}
	return nil
}

// materialize writes a blob onto the working tree, honoring its mode (regular,
// executable, or symlink).
func (s *Service) materialize(repo *git.Repository, item restoreItem) error {
	blob, err := repo.BlobObject(item.hash)
	if err != nil {
		return err
	}
	reader, err := blob.Reader()
	if err != nil {
		return err
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	abs := filepath.Join(s.WorkDir, item.path)
	if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
		return err
	}

	if item.mode == filemode.Symlink {
		_ = os.Remove(abs)
		return os.Symlink(string(content), abs)
	}

	osMode, err := item.mode.ToOSFileMode()
	if err != nil {
		osMode = 0o644
	}
	if err := os.WriteFile(abs, content, osMode.Perm()); err != nil {
		return err
	}
	return os.Chmod(abs, osMode.Perm()) // re-assert mode (WriteFile honors umask)
}

func resolveTree(repo *git.Repository, rev string) (*object.Tree, error) {
	hash, err := repo.ResolveRevision(plumbing.Revision(rev))
	if err != nil {
		return nil, err
	}
	commit, err := repo.CommitObject(*hash)
	if err != nil {
		return nil, err
	}
	return commit.Tree()
}

func treeItems(tree *object.Tree) map[string]restoreItem {
	items := map[string]restoreItem{}
	_ = tree.Files().ForEach(func(f *object.File) error {
		items[f.Name] = restoreItem{path: f.Name, hash: f.Hash, mode: f.Mode}
		return nil
	})
	return items
}

// matchIndex / matchTreeItems resolve paths (exact file or directory prefix),
// erroring when a path matches nothing.
func matchIndex(idx *index.Index, paths []string) ([]restoreItem, error) {
	var items []restoreItem
	for _, p := range paths {
		matched := false
		for _, e := range idx.Entries {
			if e.Name == p || strings.HasPrefix(e.Name, p+"/") {
				items = append(items, restoreItem{path: e.Name, hash: e.Hash, mode: e.Mode})
				matched = true
			}
		}
		if !matched {
			return nil, fmt.Errorf("pathspec %q did not match any file(s) known to git", p)
		}
	}
	return items, nil
}

func matchTreeItems(available map[string]restoreItem, paths []string) ([]restoreItem, error) {
	var items []restoreItem
	for _, p := range paths {
		matched := matchAvailable(available, p)
		if len(matched) == 0 {
			return nil, fmt.Errorf("pathspec %q did not match any file(s) in source", p)
		}
		items = append(items, matched...)
	}
	return items, nil
}

func matchAvailable(available map[string]restoreItem, path string) []restoreItem {
	if it, ok := available[path]; ok {
		return []restoreItem{it}
	}
	var items []restoreItem
	for name, it := range available {
		if strings.HasPrefix(name, path+"/") {
			items = append(items, it)
		}
	}
	return items
}

func indexNamesUnder(idx *index.Index, path string) []string {
	var names []string
	for _, e := range idx.Entries {
		if e.Name == path || strings.HasPrefix(e.Name, path+"/") {
			names = append(names, e.Name)
		}
	}
	return names
}

func resolveRestoreTargets(staged, worktree *bool) (restoreStaged, restoreWorktree bool) {
	switch {
	case staged == nil && worktree == nil:
		return false, true
	case staged != nil && worktree == nil:
		return *staged, false
	case staged == nil && worktree != nil:
		return false, *worktree
	default:
		return *staged, *worktree
	}
}
