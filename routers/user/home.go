// Copyright 2014 The Gogs Authors. All rights reserved.
// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package user

import (
	"net/http"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/base"
	"code.gitea.io/gitea/modules/context"
	"code.gitea.io/gitea/modules/setting"
)

const (
	tplDashboard  base.TplName = "user/dashboard/dashboard"
	tplIssues     base.TplName = "user/dashboard/issues"
	tplMilestones base.TplName = "user/dashboard/milestones"
	tplProfile    base.TplName = "user/profile"
)

// getDashboardContextUser finds out which context user dashboard is being viewed as .
func getDashboardContextUser(ctx *context.Context) *models.User {
	ctxUser := ctx.User
	orgName := ctx.Params(":org")
	if len(orgName) > 0 {
		ctxUser = ctx.Org.Organization
		ctx.Data["Teams"] = ctx.Org.Organization.Teams
	}
	ctx.Data["ContextUser"] = ctxUser

	if err := ctx.User.GetOrganizations(&models.SearchOrganizationsOptions{All: true}); err != nil {
		ctx.ServerError("GetOrganizations", err)
		return nil
	}
	ctx.Data["Orgs"] = ctx.User.Orgs

	return ctxUser
}

// Dashboard render the dashboard page
func Dashboard(ctx *context.Context) {
	ctxUser := getDashboardContextUser(ctx)
	if ctx.Written() {
		return
	}

	ctx.Data["Title"] = ctxUser.DisplayName() + " - " + ctx.Tr("dashboard")
	ctx.Data["PageIsDashboard"] = true
	ctx.Data["PageIsNews"] = true
	ctx.Data["SearchLimit"] = setting.UI.User.RepoPagingNum

	if ctx.Written() {
		return
	}
	ctx.HTML(http.StatusOK, tplDashboard)
}

// Email2User show user page via email
func Email2User(ctx *context.Context) {
	u, err := models.GetUserByEmail(ctx.Query("email"))
	if err != nil {
		if models.IsErrUserNotExist(err) {
			ctx.NotFound("GetUserByEmail", err)
		} else {
			ctx.ServerError("GetUserByEmail", err)
		}
		return
	}
	ctx.Redirect(setting.AppSubURL + "/user/" + u.Name)
}
