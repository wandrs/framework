// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package issue

import (
	"go.wandrs.dev/framework/models"
	"go.wandrs.dev/framework/modules/notification"
)

// ChangeMilestoneAssign changes assignment of milestone for issue.
func ChangeMilestoneAssign(issue *models.Issue, doer *models.User, oldMilestoneID int64) (err error) {
	if err = models.ChangeMilestoneAssign(issue, doer, oldMilestoneID); err != nil {
		return
	}

	notification.NotifyIssueChangeMilestone(doer, issue, oldMilestoneID)

	return nil
}
