// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package indexer

import (
	"go.wandrs.dev/framework/models"
	"go.wandrs.dev/framework/modules/git"
	code_indexer "go.wandrs.dev/framework/modules/indexer/code"
	issue_indexer "go.wandrs.dev/framework/modules/indexer/issues"
	stats_indexer "go.wandrs.dev/framework/modules/indexer/stats"
	"go.wandrs.dev/framework/modules/log"
	"go.wandrs.dev/framework/modules/notification/base"
	"go.wandrs.dev/framework/modules/repository"
	"go.wandrs.dev/framework/modules/setting"
)

type indexerNotifier struct {
	base.NullNotifier
}

var (
	_ base.Notifier = &indexerNotifier{}
)

// NewNotifier create a new indexerNotifier notifier
func NewNotifier() base.Notifier {
	return &indexerNotifier{}
}

func (r *indexerNotifier) NotifyCreateIssueComment(doer *models.User, repo *models.Repository,
	issue *models.Issue, comment *models.Comment, mentions []*models.User) {
	if comment.Type == models.CommentTypeComment {
		if issue.Comments == nil {
			if err := issue.LoadDiscussComments(); err != nil {
				log.Error("LoadComments failed: %v", err)
				return
			}
		} else {
			issue.Comments = append(issue.Comments, comment)
		}

		issue_indexer.UpdateIssueIndexer(issue)
	}
}

func (r *indexerNotifier) NotifyNewIssue(issue *models.Issue, mentions []*models.User) {
	issue_indexer.UpdateIssueIndexer(issue)
}

func (r *indexerNotifier) NotifyNewPullRequest(pr *models.PullRequest, mentions []*models.User) {
	issue_indexer.UpdateIssueIndexer(pr.Issue)
}

func (r *indexerNotifier) NotifyUpdateComment(doer *models.User, c *models.Comment, oldContent string) {
	if c.Type == models.CommentTypeComment {
		var found bool
		if c.Issue.Comments != nil {
			for i := 0; i < len(c.Issue.Comments); i++ {
				if c.Issue.Comments[i].ID == c.ID {
					c.Issue.Comments[i] = c
					found = true
					break
				}
			}
		}

		if !found {
			if err := c.Issue.LoadDiscussComments(); err != nil {
				log.Error("LoadComments failed: %v", err)
				return
			}
		}

		issue_indexer.UpdateIssueIndexer(c.Issue)
	}
}

func (r *indexerNotifier) NotifyDeleteComment(doer *models.User, comment *models.Comment) {
	if comment.Type == models.CommentTypeComment {
		if err := comment.LoadIssue(); err != nil {
			log.Error("LoadIssue: %v", err)
			return
		}

		var found bool
		if comment.Issue.Comments != nil {
			for i := 0; i < len(comment.Issue.Comments); i++ {
				if comment.Issue.Comments[i].ID == comment.ID {
					comment.Issue.Comments = append(comment.Issue.Comments[:i], comment.Issue.Comments[i+1:]...)
					found = true
					break
				}
			}
		}

		if !found {
			if err := comment.Issue.LoadDiscussComments(); err != nil {
				log.Error("LoadComments failed: %v", err)
				return
			}
		}
		// reload comments to delete the old comment
		issue_indexer.UpdateIssueIndexer(comment.Issue)
	}
}

func (r *indexerNotifier) NotifyDeleteRepository(doer *models.User, repo *models.Repository) {
	issue_indexer.DeleteRepoIssueIndexer(repo)
	if setting.Indexer.RepoIndexerEnabled {
		code_indexer.DeleteRepoFromIndexer(repo)
	}
}

func (r *indexerNotifier) NotifyMigrateRepository(doer *models.User, u *models.User, repo *models.Repository) {
	issue_indexer.UpdateRepoIndexer(repo)
	if setting.Indexer.RepoIndexerEnabled && !repo.IsEmpty {
		code_indexer.UpdateRepoIndexer(repo)
	}
	if err := stats_indexer.UpdateRepoIndexer(repo); err != nil {
		log.Error("stats_indexer.UpdateRepoIndexer(%d) failed: %v", repo.ID, err)
	}
}

func (r *indexerNotifier) NotifyPushCommits(pusher *models.User, repo *models.Repository, opts *repository.PushUpdateOptions, commits *repository.PushCommits) {
	if setting.Indexer.RepoIndexerEnabled && opts.RefFullName == git.BranchPrefix+repo.DefaultBranch {
		code_indexer.UpdateRepoIndexer(repo)
	}
	if err := stats_indexer.UpdateRepoIndexer(repo); err != nil {
		log.Error("stats_indexer.UpdateRepoIndexer(%d) failed: %v", repo.ID, err)
	}
}

func (r *indexerNotifier) NotifySyncPushCommits(pusher *models.User, repo *models.Repository, opts *repository.PushUpdateOptions, commits *repository.PushCommits) {
	if setting.Indexer.RepoIndexerEnabled && opts.RefFullName == git.BranchPrefix+repo.DefaultBranch {
		code_indexer.UpdateRepoIndexer(repo)
	}
	if err := stats_indexer.UpdateRepoIndexer(repo); err != nil {
		log.Error("stats_indexer.UpdateRepoIndexer(%d) failed: %v", repo.ID, err)
	}
}

func (r *indexerNotifier) NotifyIssueChangeContent(doer *models.User, issue *models.Issue, oldContent string) {
	issue_indexer.UpdateIssueIndexer(issue)
}

func (r *indexerNotifier) NotifyIssueChangeTitle(doer *models.User, issue *models.Issue, oldTitle string) {
	issue_indexer.UpdateIssueIndexer(issue)
}

func (r *indexerNotifier) NotifyIssueChangeRef(doer *models.User, issue *models.Issue, oldRef string) {
	issue_indexer.UpdateIssueIndexer(issue)
}
