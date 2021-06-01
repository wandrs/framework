// Copyright 2014 The Gogs Authors. All rights reserved.
// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package user

import (
	"net/http"

	"go.wandrs.dev/framework/models"
	"go.wandrs.dev/framework/modules/base"
	"go.wandrs.dev/framework/modules/context"
	"go.wandrs.dev/framework/modules/setting"
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

	if setting.Service.EnableUserHeatmap {
		data, err := models.GetUserHeatmapDataByUserTeam(ctxUser, ctx.Org.Team, ctx.User)
		if err != nil {
			ctx.ServerError("GetUserHeatmapDataByUserTeam", err)
			return
		}
		ctx.Data["HeatmapData"] = data
	}

	var err error
	var mirrors []*models.Repository
	if ctxUser.IsOrganization() {
		var env models.AccessibleReposEnvironment
		if ctx.Org.Team != nil {
			env = ctxUser.AccessibleTeamReposEnv(ctx.Org.Team)
		} else {
			env, err = ctxUser.AccessibleReposEnv(ctx.User.ID)
			if err != nil {
				ctx.ServerError("AccessibleReposEnv", err)
				return
			}
		}
		mirrors, err = env.MirrorRepos()
		if err != nil {
			ctx.ServerError("env.MirrorRepos", err)
			return
		}
	} else {
		mirrors, err = ctxUser.GetMirrorRepositories()
		if err != nil {
			ctx.ServerError("GetMirrorRepositories", err)
			return
		}
	}
	ctx.Data["MaxShowRepoNum"] = setting.UI.User.RepoPagingNum

	if err := models.MirrorRepositoryList(mirrors).LoadAttributes(); err != nil {
		ctx.ServerError("MirrorRepositoryList.LoadAttributes", err)
		return
	}
	ctx.Data["MirrorCount"] = len(mirrors)
	ctx.Data["Mirrors"] = mirrors

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
