/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package git

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
)

// Git represents a git repository.
type Git struct {
	repo *git.Repository
}

// New opens the git repository in current directory.
func New() (*Git, error) {
	repo, err := git.PlainOpenWithOptions("", &git.PlainOpenOptions{DetectDotGit: true})
	return &Git{repo: repo}, err
}

// Revision returns the current revision (HEAD) of the repository in the format used by go pseudo versions.
func (g *Git) Revision() (string, time.Time, error) {
	commitRef, err := g.repo.Head()
	if err != nil {
		return "", time.Time{}, err
	}
	commit, err := g.repo.CommitObject(commitRef.Hash())
	if err != nil {
		return "", time.Time{}, err
	}
	return commitRef.Hash().String()[:12], commit.Committer.When, nil
}

// ParsedBranchName returns the name of the current branch.
// Special characters are replaced with "-", and the name is lowercased and trimmed to 49 characters.
// This makes sure that the branch name is usable as a GCP image name.
func (g *Git) ParsedBranchName() (string, error) {
	commitRef, err := g.repo.Head()
	if err != nil {
		return "", err
	}

	rxp, err := regexp.Compile("[^a-zA-Z0-9-]+")
	if err != nil {
		return "", err
	}

	branch := strings.ToLower(rxp.ReplaceAllString(commitRef.Name().Short(), "-"))
	if len(branch) > 49 {
		branch = branch[:49]
	}

	return strings.TrimSuffix(branch, "-"), nil
}

// BranchName of current HEAD.
func (g *Git) BranchName() (string, error) {
	commitRef, err := g.repo.Head()
	if err != nil {
		return "", err
	}
	return commitRef.Name().Short(), nil
}

// Path returns the path of the git repository.
func (g *Git) Path() (string, error) {
	worktree, err := g.repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}
	return worktree.Filesystem.Root(), nil
}
