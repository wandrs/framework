// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package admin

import (
	"go.wandrs.dev/framework/modules/context"
	api "go.wandrs.dev/framework/modules/structs"
	"go.wandrs.dev/framework/modules/web"
	"go.wandrs.dev/framework/routers/api/v1/repo"
	"go.wandrs.dev/framework/routers/api/v1/user"
)

// CreateRepo api for creating a repository
func CreateRepo(ctx *context.APIContext) {
	// swagger:operation POST /admin/users/{username}/repos admin adminCreateRepo
	// ---
	// summary: Create a repository on behalf of a user
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: username
	//   in: path
	//   description: username of the user. This user will own the created repository
	//   type: string
	//   required: true
	// - name: repository
	//   in: body
	//   required: true
	//   schema: { "$ref": "#/definitions/CreateRepoOption" }
	// responses:
	//   "201":
	//     "$ref": "#/responses/Repository"
	//   "403":
	//     "$ref": "#/responses/forbidden"
	//   "404":
	//     "$ref": "#/responses/notFound"
	//   "409":
	//     "$ref": "#/responses/error"
	//   "422":
	//     "$ref": "#/responses/validationError"
	form := web.GetForm(ctx).(*api.CreateRepoOption)
	owner := user.GetUserByParams(ctx)
	if ctx.Written() {
		return
	}

	repo.CreateUserRepo(ctx, owner, *form)
}
