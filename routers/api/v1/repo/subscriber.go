// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	"net/http"

	"go.wandrs.dev/framework/modules/context"
	"go.wandrs.dev/framework/modules/convert"
	api "go.wandrs.dev/framework/modules/structs"
	"go.wandrs.dev/framework/routers/api/v1/utils"
)

// ListSubscribers list a repo's subscribers (i.e. watchers)
func ListSubscribers(ctx *context.APIContext) {
	// swagger:operation GET /repos/{owner}/{repo}/subscribers repository repoListSubscribers
	// ---
	// summary: List a repo's watchers
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the repo
	//   type: string
	//   required: true
	// - name: repo
	//   in: path
	//   description: name of the repo
	//   type: string
	//   required: true
	// - name: page
	//   in: query
	//   description: page number of results to return (1-based)
	//   type: integer
	// - name: limit
	//   in: query
	//   description: page size of results
	//   type: integer
	// responses:
	//   "200":
	//     "$ref": "#/responses/UserList"

	subscribers, err := ctx.Repo.Repository.GetWatchers(utils.GetListOptions(ctx))
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "GetWatchers", err)
		return
	}
	users := make([]*api.User, len(subscribers))
	for i, subscriber := range subscribers {
		users[i] = convert.ToUser(subscriber, ctx.User)
	}
	ctx.JSON(http.StatusOK, users)
}
