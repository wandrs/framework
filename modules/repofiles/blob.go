// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repofiles

import (
	"go.wandrs.dev/framework/models"
	"go.wandrs.dev/framework/modules/git"
	"go.wandrs.dev/framework/modules/setting"
	api "go.wandrs.dev/framework/modules/structs"
)

// GetBlobBySHA get the GitBlobResponse of a repository using a sha hash.
func GetBlobBySHA(repo *models.Repository, sha string) (*api.GitBlobResponse, error) {
	gitRepo, err := git.OpenRepository(repo.RepoPath())
	if err != nil {
		return nil, err
	}
	defer gitRepo.Close()
	gitBlob, err := gitRepo.GetBlob(sha)
	if err != nil {
		return nil, err
	}
	content := ""
	if gitBlob.Size() <= setting.API.DefaultMaxBlobSize {
		content, err = gitBlob.GetBlobContentBase64()
		if err != nil {
			return nil, err
		}
	}
	return &api.GitBlobResponse{
		SHA:      gitBlob.ID.String(),
		URL:      repo.APIURL() + "/git/blobs/" + gitBlob.ID.String(),
		Size:     gitBlob.Size(),
		Encoding: "base64",
		Content:  content,
	}, nil
}
