// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package mailer

import (
	"go.wandrs.dev/framework/models"
	"go.wandrs.dev/framework/modules/log"
)

// MailParticipantsComment sends new comment emails to repository watchers and mentioned people.
func MailParticipantsComment(c *models.Comment, opType models.ActionType, issue *models.Issue, mentions []*models.User) error {
	content := c.Content
	if c.Type == models.CommentTypePullPush {
		content = ""
	}
	if err := mailIssueCommentToParticipants(
		&mailCommentContext{
			Issue:      issue,
			Doer:       c.Poster,
			ActionType: opType,
			Content:    content,
			Comment:    c,
		}, mentions); err != nil {
		log.Error("mailIssueCommentToParticipants: %v", err)
	}
	return nil
}

// MailMentionsComment sends email to users mentioned in a code comment
func MailMentionsComment(pr *models.PullRequest, c *models.Comment, mentions []*models.User) (err error) {
	visited := make(map[int64]bool, len(mentions)+1)
	visited[c.Poster.ID] = true
	if err = mailIssueCommentBatch(
		&mailCommentContext{
			Issue:      pr.Issue,
			Doer:       c.Poster,
			ActionType: models.ActionCommentPull,
			Content:    c.Content,
			Comment:    c,
		}, mentions, visited, true); err != nil {
		log.Error("mailIssueCommentBatch: %v", err)
	}
	return nil
}
