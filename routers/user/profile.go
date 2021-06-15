// Copyright 2015 The Gogs Authors. All rights reserved.
// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package user

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/context"
	"code.gitea.io/gitea/modules/setting"
	"code.gitea.io/gitea/routers/org"
)

// GetUserByName get user by name
func GetUserByName(ctx *context.Context, name string) *models.User {
	user, err := models.GetUserByName(name)
	if err != nil {
		if models.IsErrUserNotExist(err) {
			if redirectUserID, err := models.LookupUserRedirect(name); err == nil {
				context.RedirectToUser(ctx, name, redirectUserID)
			} else {
				ctx.NotFound("GetUserByName", err)
			}
		} else {
			ctx.ServerError("GetUserByName", err)
		}
		return nil
	}
	return user
}

// GetUserByParams returns user whose name is presented in URL paramenter.
func GetUserByParams(ctx *context.Context) *models.User {
	return GetUserByName(ctx, ctx.Params(":username"))
}

// Profile render user's profile page
func Profile(ctx *context.Context) {
	uname := ctx.Params(":username")

	// Special handle for FireFox requests favicon.ico.
	if uname == "favicon.ico" {
		ctx.ServeFile(path.Join(setting.StaticRootPath, "public/img/favicon.png"))
		return
	}

	if strings.HasSuffix(uname, ".png") {
		ctx.Error(http.StatusNotFound)
		return
	}

	ctxUser := GetUserByName(ctx, uname)
	if ctx.Written() {
		return
	}

	if ctxUser.IsOrganization() {
		org.Home(ctx)
		return
	}

	// Show OpenID URIs
	openIDs, err := models.GetUserOpenIDs(ctxUser.ID)
	if err != nil {
		ctx.ServerError("GetUserOpenIDs", err)
		return
	}

	ctx.Data["Title"] = ctxUser.DisplayName()
	ctx.Data["PageIsUserProfile"] = true
	ctx.Data["Owner"] = ctxUser
	ctx.Data["OpenIDs"] = openIDs

	if len(ctxUser.Description) != 0 {
		ctx.Data["RenderedDescription"] = ctxUser.Description
	}

	showPrivate := ctx.IsSigned && (ctx.User.IsAdmin || ctx.User.ID == ctxUser.ID)

	orgs, err := models.GetOrgsByUserID(ctxUser.ID, showPrivate)
	if err != nil {
		ctx.ServerError("GetOrgsByUserIDDesc", err)
		return
	}

	ctx.Data["Orgs"] = orgs
	ctx.Data["HasOrgsVisible"] = models.HasOrgsVisible(orgs, ctx.User)

	tab := ctx.Query("tab")
	ctx.Data["TabName"] = tab

	page := ctx.QueryInt("page")
	if page <= 0 {
		page = 1
	}

	var (
		total int
	)

	ctx.Data["SortType"] = ctx.Query("sort")

	keyword := strings.Trim(ctx.Query("q"), " ")
	ctx.Data["Keyword"] = keyword
	switch tab {
	case "followers":
		items, err := ctxUser.GetFollowers(models.ListOptions{
			PageSize: setting.UI.User.RepoPagingNum,
			Page:     page,
		})
		if err != nil {
			ctx.ServerError("GetFollowers", err)
			return
		}
		ctx.Data["Cards"] = items

		total = ctxUser.NumFollowers
	case "following":
		items, err := ctxUser.GetFollowing(models.ListOptions{
			PageSize: setting.UI.User.RepoPagingNum,
			Page:     page,
		})
		if err != nil {
			ctx.ServerError("GetFollowing", err)
			return
		}
		ctx.Data["Cards"] = items

		total = ctxUser.NumFollowing
	}
	ctx.Data["Total"] = total

	pager := context.NewPagination(total, setting.UI.User.RepoPagingNum, page, 5)
	pager.SetDefaultParams(ctx)
	ctx.Data["Page"] = pager

	ctx.Data["ShowUserEmail"] = len(ctxUser.Email) > 0 && ctx.IsSigned && (!ctxUser.KeepEmailPrivate || ctxUser.ID == ctx.User.ID)

	ctx.HTML(http.StatusOK, tplProfile)
}

// Action response for follow/unfollow user request
func Action(ctx *context.Context) {
	u := GetUserByParams(ctx)
	if ctx.Written() {
		return
	}

	var err error
	switch ctx.Params(":action") {
	case "follow":
		err = models.FollowUser(ctx.User.ID, u.ID)
	case "unfollow":
		err = models.UnfollowUser(ctx.User.ID, u.ID)
	}

	if err != nil {
		ctx.ServerError(fmt.Sprintf("Action (%s)", ctx.Params(":action")), err)
		return
	}

	ctx.RedirectToFirst(ctx.Query("redirect_to"), u.HomeLink())
}
