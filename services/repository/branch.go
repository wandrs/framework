// Copyright 2021 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repository

import (
	"errors"

	"go.wandrs.dev/framework/models"
	"go.wandrs.dev/framework/modules/git"
	"go.wandrs.dev/framework/modules/log"
	repo_module "go.wandrs.dev/framework/modules/repository"
	pull_service "go.wandrs.dev/framework/services/pull"
)

// enmuerates all branch related errors
var (
	ErrBranchIsDefault   = errors.New("branch is default")
	ErrBranchIsProtected = errors.New("branch is protected")
)

// DeleteBranch delete branch
func DeleteBranch(doer *models.User, repo *models.Repository, gitRepo *git.Repository, branchName string) error {
	if branchName == repo.DefaultBranch {
		return ErrBranchIsDefault
	}

	isProtected, err := repo.IsProtectedBranch(branchName, doer)
	if err != nil {
		return err
	}

	if isProtected {
		return ErrBranchIsProtected
	}

	commit, err := gitRepo.GetBranchCommit(branchName)
	if err != nil {
		return err
	}

	if err := gitRepo.DeleteBranch(branchName, git.DeleteBranchOptions{
		Force: true,
	}); err != nil {
		return err
	}

	if err := pull_service.CloseBranchPulls(doer, repo.ID, branchName); err != nil {
		return err
	}

	// Don't return error below this
	if err := PushUpdate(
		&repo_module.PushUpdateOptions{
			RefFullName:  git.BranchPrefix + branchName,
			OldCommitID:  commit.ID.String(),
			NewCommitID:  git.EmptySHA,
			PusherID:     doer.ID,
			PusherName:   doer.Name,
			RepoUserName: repo.OwnerName,
			RepoName:     repo.Name,
		}); err != nil {
		log.Error("Update: %v", err)
	}

	if err := repo.AddDeletedBranch(branchName, commit.ID.String(), doer.ID); err != nil {
		log.Warn("AddDeletedBranch: %v", err)
	}

	return nil
}
