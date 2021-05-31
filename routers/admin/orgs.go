// Copyright 2014 The Gogs Authors. All rights reserved.
// Copyright 2020 The Gitea Authors.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package admin

import (
	"go.wandrs.dev/framework/models"
	"go.wandrs.dev/framework/modules/base"
	"go.wandrs.dev/framework/modules/context"
	"go.wandrs.dev/framework/modules/setting"
	"go.wandrs.dev/framework/modules/structs"
	"go.wandrs.dev/framework/routers"
)

const (
	tplOrgs base.TplName = "admin/org/list"
)

// Organizations show all the organizations
func Organizations(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("admin.organizations")
	ctx.Data["PageIsAdmin"] = true
	ctx.Data["PageIsAdminOrganizations"] = true

	routers.RenderUserSearch(ctx, &models.SearchUserOptions{
		Type: models.UserTypeOrganization,
		ListOptions: models.ListOptions{
			PageSize: setting.UI.Admin.OrgPagingNum,
		},
		Visible: []structs.VisibleType{structs.VisibleTypePublic, structs.VisibleTypeLimited, structs.VisibleTypePrivate},
	}, tplOrgs)
}
