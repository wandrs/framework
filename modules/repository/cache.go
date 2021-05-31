// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repository

import (
	"strings"

	"go.wandrs.dev/framework/models"
	"go.wandrs.dev/framework/modules/cache"
	"go.wandrs.dev/framework/modules/git"
	"go.wandrs.dev/framework/modules/setting"
)

func getRefName(fullRefName string) string {
	if strings.HasPrefix(fullRefName, git.TagPrefix) {
		return fullRefName[len(git.TagPrefix):]
	} else if strings.HasPrefix(fullRefName, git.BranchPrefix) {
		return fullRefName[len(git.BranchPrefix):]
	}
	return ""
}

// CacheRef cachhe last commit information of the branch or the tag
func CacheRef(repo *models.Repository, gitRepo *git.Repository, fullRefName string) error {
	if !setting.CacheService.LastCommit.Enabled {
		return nil
	}

	commit, err := gitRepo.GetCommit(fullRefName)
	if err != nil {
		return err
	}

	commitsCount, err := cache.GetInt64(repo.GetCommitsCountCacheKey(getRefName(fullRefName), true), commit.CommitsCount)
	if err != nil {
		return err
	}
	if commitsCount < setting.CacheService.LastCommit.CommitsCount {
		return nil
	}

	commitCache := git.NewLastCommitCache(repo.FullName(), gitRepo, setting.LastCommitCacheTTLSeconds, cache.GetCache())

	return commitCache.CacheCommit(commit)
}
