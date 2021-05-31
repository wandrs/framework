// Copyright 2021 The Gitea Authors.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package user

import (
	"net/http"

	"go.wandrs.dev/framework/models"
	"go.wandrs.dev/framework/modules/context"
)

// GetUserByParamsName get user by name
func GetUserByParamsName(ctx *context.APIContext, name string) *models.User {
	username := ctx.Params(name)
	user, err := models.GetUserByName(username)
	if err != nil {
		if models.IsErrUserNotExist(err) {
			if redirectUserID, err := models.LookupUserRedirect(username); err == nil {
				context.RedirectToUser(ctx.Context, username, redirectUserID)
			} else {
				ctx.NotFound("GetUserByName", err)
			}
		} else {
			ctx.Error(http.StatusInternalServerError, "GetUserByName", err)
		}
		return nil
	}
	return user
}

// GetUserByParams returns user whose name is presented in URL (":username").
func GetUserByParams(ctx *context.APIContext) *models.User {
	return GetUserByParamsName(ctx, ":username")
}
