// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package issue

import (
	"go.wandrs.dev/framework/models"
	"go.wandrs.dev/framework/modules/notification"
)

// ChangeContent changes issue content, as the given user.
func ChangeContent(issue *models.Issue, doer *models.User, content string) (err error) {
	oldContent := issue.Content

	if err := issue.ChangeContent(doer, content); err != nil {
		return err
	}

	notification.NotifyIssueChangeContent(doer, issue, oldContent)

	return nil
}
