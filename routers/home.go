// Copyright 2014 The Gogs Authors. All rights reserved.
// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package routers

import (
	"bytes"
	"net/http"
	"strings"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/base"
	"code.gitea.io/gitea/modules/context"
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/setting"
	"code.gitea.io/gitea/modules/structs"
	"code.gitea.io/gitea/modules/util"
	"code.gitea.io/gitea/modules/web/middleware"
	"code.gitea.io/gitea/routers/user"
)

const (
	// tplHome home page template
	tplHome base.TplName = "home"
	// tplExploreUsers explore users page template
	tplExploreUsers base.TplName = "explore/users"
	// tplExploreOrganizations explore organizations page template
	tplExploreOrganizations base.TplName = "explore/organizations"
)

// Home render home page
func Home(ctx *context.Context) {
	if ctx.IsSigned {
		if !ctx.User.IsActive && setting.Service.RegisterEmailConfirm {
			ctx.Data["Title"] = ctx.Tr("auth.active_your_account")
			ctx.HTML(http.StatusOK, user.TplActivate)
		} else if !ctx.User.IsActive || ctx.User.ProhibitLogin {
			log.Info("Failed authentication attempt for %s from %s", ctx.User.Name, ctx.RemoteAddr())
			ctx.Data["Title"] = ctx.Tr("auth.prohibit_login")
			ctx.HTML(http.StatusOK, "user/auth/prohibit_login")
		} else if ctx.User.MustChangePassword {
			ctx.Data["Title"] = ctx.Tr("auth.must_change_password")
			ctx.Data["ChangePasscodeLink"] = setting.AppSubURL + "/user/change_password"
			middleware.SetRedirectToCookie(ctx.Resp, setting.AppSubURL+ctx.Req.URL.RequestURI())
			ctx.Redirect(setting.AppSubURL + "/user/settings/change_password")
		} else {
			user.Dashboard(ctx)
		}
		return
		// Check non-logged users landing page.
	} else if setting.LandingPageURL != setting.LandingPageHome {
		ctx.Redirect(setting.AppSubURL + string(setting.LandingPageURL))
		return
	}

	// Check auto-login.
	uname := ctx.GetCookie(setting.CookieUserName)
	if len(uname) != 0 {
		ctx.Redirect(setting.AppSubURL + "/user/login")
		return
	}

	ctx.Data["PageIsHome"] = true
	ctx.HTML(http.StatusOK, tplHome)
}

// RepoSearchOptions when calling search repositories
type RepoSearchOptions struct {
	OwnerID    int64
	Private    bool
	Restricted bool
	PageSize   int
	TplName    base.TplName
}

var (
	nullByte = []byte{0x00}
)

func isKeywordValid(keyword string) bool {
	return !bytes.Contains([]byte(keyword), nullByte)
}

// RenderUserSearch render user search page
func RenderUserSearch(ctx *context.Context, opts *models.SearchUserOptions, tplName base.TplName) {
	opts.Page = ctx.QueryInt("page")
	if opts.Page <= 1 {
		opts.Page = 1
	}

	var (
		users   []*models.User
		count   int64
		err     error
		orderBy models.SearchOrderBy
	)

	ctx.Data["SortType"] = ctx.Query("sort")
	switch ctx.Query("sort") {
	case "newest":
		orderBy = models.SearchOrderByIDReverse
	case "oldest":
		orderBy = models.SearchOrderByID
	case "recentupdate":
		orderBy = models.SearchOrderByRecentUpdated
	case "leastupdate":
		orderBy = models.SearchOrderByLeastUpdated
	case "reversealphabetically":
		orderBy = models.SearchOrderByAlphabeticallyReverse
	case "alphabetically":
		orderBy = models.SearchOrderByAlphabetically
	default:
		ctx.Data["SortType"] = "alphabetically"
		orderBy = models.SearchOrderByAlphabetically
	}

	opts.Keyword = strings.Trim(ctx.Query("q"), " ")
	opts.OrderBy = orderBy
	if len(opts.Keyword) == 0 || isKeywordValid(opts.Keyword) {
		users, count, err = models.SearchUsers(opts)
		if err != nil {
			ctx.ServerError("SearchUsers", err)
			return
		}
	}
	ctx.Data["Keyword"] = opts.Keyword
	ctx.Data["Total"] = count
	ctx.Data["Users"] = users
	ctx.Data["UsersTwoFaStatus"] = models.UserList(users).GetTwoFaStatus()
	ctx.Data["ShowUserEmail"] = setting.UI.ShowUserEmail

	pager := context.NewPagination(int(count), opts.PageSize, opts.Page, 5)
	pager.SetDefaultParams(ctx)
	ctx.Data["Page"] = pager

	ctx.HTML(http.StatusOK, tplName)
}

// ExploreUsers render explore users page
func ExploreUsers(ctx *context.Context) {
	if setting.Service.Explore.DisableUsersPage {
		ctx.Redirect(setting.AppSubURL + "/explore/repos")
		return
	}
	ctx.Data["Title"] = ctx.Tr("explore")
	ctx.Data["PageIsExplore"] = true
	ctx.Data["PageIsExploreUsers"] = true

	RenderUserSearch(ctx, &models.SearchUserOptions{
		Actor:       ctx.User,
		Type:        models.UserTypeIndividual,
		ListOptions: models.ListOptions{PageSize: setting.UI.ExplorePagingNum},
		IsActive:    util.OptionalBoolTrue,
		Visible:     []structs.VisibleType{structs.VisibleTypePublic, structs.VisibleTypeLimited, structs.VisibleTypePrivate},
	}, tplExploreUsers)
}

// ExploreOrganizations render explore organizations page
func ExploreOrganizations(ctx *context.Context) {
	ctx.Data["UsersIsDisabled"] = setting.Service.Explore.DisableUsersPage
	ctx.Data["Title"] = ctx.Tr("explore")
	ctx.Data["PageIsExplore"] = true
	ctx.Data["PageIsExploreOrganizations"] = true

	visibleTypes := []structs.VisibleType{structs.VisibleTypePublic}
	if ctx.User != nil {
		visibleTypes = append(visibleTypes, structs.VisibleTypeLimited, structs.VisibleTypePrivate)
	}

	RenderUserSearch(ctx, &models.SearchUserOptions{
		Actor:       ctx.User,
		Type:        models.UserTypeOrganization,
		ListOptions: models.ListOptions{PageSize: setting.UI.ExplorePagingNum},
		Visible:     visibleTypes,
	}, tplExploreOrganizations)
}

// NotFound render 404 page
func NotFound(ctx *context.Context) {
	ctx.Data["Title"] = "Page Not Found"
	ctx.NotFound("home.NotFound", nil)
}
