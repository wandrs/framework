// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package setting

import (
	"path/filepath"

	"go.wandrs.dev/framework/models"
	"go.wandrs.dev/framework/modules/context"
	"go.wandrs.dev/framework/modules/repository"
	"go.wandrs.dev/framework/modules/setting"
	"go.wandrs.dev/framework/modules/util"
)

// AdoptOrDeleteRepository adopts or deletes a repository
func AdoptOrDeleteRepository(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("settings")
	ctx.Data["PageIsSettingsRepos"] = true
	allowAdopt := ctx.IsUserSiteAdmin() || setting.Repository.AllowAdoptionOfUnadoptedRepositories
	ctx.Data["allowAdopt"] = allowAdopt
	allowDelete := ctx.IsUserSiteAdmin() || setting.Repository.AllowDeleteOfUnadoptedRepositories
	ctx.Data["allowDelete"] = allowDelete

	dir := ctx.Query("id")
	action := ctx.Query("action")

	ctxUser := ctx.User
	root := filepath.Join(models.UserPath(ctxUser.LowerName))

	// check not a repo
	has, err := models.IsRepositoryExist(ctxUser, dir)
	if err != nil {
		ctx.ServerError("IsRepositoryExist", err)
		return
	}

	isDir, err := util.IsDir(filepath.Join(root, dir+".git"))
	if err != nil {
		ctx.ServerError("IsDir", err)
		return
	}
	if has || !isDir {
		// Fallthrough to failure mode
	} else if action == "adopt" && allowAdopt {
		if _, err := repository.AdoptRepository(ctxUser, ctxUser, models.CreateRepoOptions{
			Name:      dir,
			IsPrivate: true,
		}); err != nil {
			ctx.ServerError("repository.AdoptRepository", err)
			return
		}
		ctx.Flash.Success(ctx.Tr("repo.adopt_preexisting_success", dir))
	} else if action == "delete" && allowDelete {
		if err := repository.DeleteUnadoptedRepository(ctxUser, ctxUser, dir); err != nil {
			ctx.ServerError("repository.AdoptRepository", err)
			return
		}
		ctx.Flash.Success(ctx.Tr("repo.delete_preexisting_success", dir))
	}

	ctx.Redirect(setting.AppSubURL + "/user/settings/repos")
}
