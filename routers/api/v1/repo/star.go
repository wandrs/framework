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

// ListStargazers list a repository's stargazers
func ListStargazers(ctx *context.APIContext) {
	// swagger:operation GET /repos/{owner}/{repo}/stargazers repository repoListStargazers
	// ---
	// summary: List a repo's stargazers
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

	stargazers, err := ctx.Repo.Repository.GetStargazers(utils.GetListOptions(ctx))
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "GetStargazers", err)
		return
	}
	users := make([]*api.User, len(stargazers))
	for i, stargazer := range stargazers {
		users[i] = convert.ToUser(stargazer, ctx.User)
	}
	ctx.JSON(http.StatusOK, users)
}
