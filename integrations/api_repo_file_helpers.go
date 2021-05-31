// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package integrations

import (
	"go.wandrs.dev/framework/models"
	"go.wandrs.dev/framework/modules/repofiles"
	api "go.wandrs.dev/framework/modules/structs"
)

func createFileInBranch(user *models.User, repo *models.Repository, treePath, branchName, content string) (*api.FileResponse, error) {
	opts := &repofiles.UpdateRepoFileOptions{
		OldBranch: branchName,
		TreePath:  treePath,
		Content:   content,
		IsNewFile: true,
		Author:    nil,
		Committer: nil,
	}
	return repofiles.CreateOrUpdateRepoFile(repo, user, opts)
}

func createFile(user *models.User, repo *models.Repository, treePath string) (*api.FileResponse, error) {
	return createFileInBranch(user, repo, treePath, repo.DefaultBranch, "This is a NEW file")
}
