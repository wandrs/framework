// Copyright 2019 The Gitea Authors. All rights reserved.
// Copyright 2018 Jonas Franz. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package migrations

import (
	"context"
	"testing"
	"time"

	"go.wandrs.dev/framework/models"
	"go.wandrs.dev/framework/modules/graceful"
	"go.wandrs.dev/framework/modules/migrations/base"
	"go.wandrs.dev/framework/modules/structs"
	"go.wandrs.dev/framework/modules/util"

	"github.com/stretchr/testify/assert"
)

func TestGiteaUploadRepo(t *testing.T) {
	// FIXME: Since no accesskey or user/password will trigger rate limit of github, just skip
	t.Skip()

	models.PrepareTestEnv(t)

	user := models.AssertExistsAndLoadBean(t, &models.User{ID: 1}).(*models.User)

	var (
		downloader = NewGithubDownloaderV3(context.Background(), "https://github.com", "", "", "", "go-xorm", "builder")
		repoName   = "builder-" + time.Now().Format("2006-01-02-15-04-05")
		uploader   = NewGiteaLocalUploader(graceful.GetManager().HammerContext(), user, user.Name, repoName)
	)

	err := migrateRepository(downloader, uploader, base.MigrateOptions{
		CloneAddr:    "https://github.com/go-xorm/builder",
		RepoName:     repoName,
		AuthUsername: "",

		Wiki:         true,
		Issues:       true,
		Milestones:   true,
		Labels:       true,
		Releases:     true,
		Comments:     true,
		PullRequests: true,
		Private:      true,
		Mirror:       false,
	})
	assert.NoError(t, err)

	repo := models.AssertExistsAndLoadBean(t, &models.Repository{OwnerID: user.ID, Name: repoName}).(*models.Repository)
	assert.True(t, repo.HasWiki())
	assert.EqualValues(t, models.RepositoryReady, repo.Status)

	milestones, err := models.GetMilestones(models.GetMilestonesOption{
		RepoID: repo.ID,
		State:  structs.StateOpen,
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, len(milestones))

	milestones, err = models.GetMilestones(models.GetMilestonesOption{
		RepoID: repo.ID,
		State:  structs.StateClosed,
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 0, len(milestones))

	labels, err := models.GetLabelsByRepoID(repo.ID, "", models.ListOptions{})
	assert.NoError(t, err)
	assert.EqualValues(t, 11, len(labels))

	releases, err := models.GetReleasesByRepoID(repo.ID, models.FindReleasesOptions{
		ListOptions: models.ListOptions{
			PageSize: 10,
			Page:     0,
		},
		IncludeTags: true,
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 8, len(releases))

	releases, err = models.GetReleasesByRepoID(repo.ID, models.FindReleasesOptions{
		ListOptions: models.ListOptions{
			PageSize: 10,
			Page:     0,
		},
		IncludeTags: false,
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, len(releases))

	issues, err := models.Issues(&models.IssuesOptions{
		RepoIDs:  []int64{repo.ID},
		IsPull:   util.OptionalBoolFalse,
		SortType: "oldest",
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 14, len(issues))
	assert.NoError(t, issues[0].LoadDiscussComments())
	assert.EqualValues(t, 0, len(issues[0].Comments))

	pulls, _, err := models.PullRequests(repo.ID, &models.PullRequestsOptions{
		SortType: "oldest",
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 34, len(pulls))
	assert.NoError(t, pulls[0].LoadIssue())
	assert.NoError(t, pulls[0].Issue.LoadDiscussComments())
	assert.EqualValues(t, 2, len(pulls[0].Issue.Comments))
}
