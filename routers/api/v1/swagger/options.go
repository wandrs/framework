// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package swagger

import (
	api "go.wandrs.dev/framework/modules/structs"
)

// not actually a response, just a hack to get go-swagger to include definitions
// of the various XYZOption structs

// parameterBodies
// swagger:response parameterBodies
type swaggerParameterBodies struct {
	// in:body
	CreateEmailOption api.CreateEmailOption
	// in:body
	DeleteEmailOption api.DeleteEmailOption

	// in:body
	MarkdownOption api.MarkdownOption

	// in:body
	CreateOrgOption api.CreateOrgOption
	// in:body
	EditOrgOption api.EditOrgOption

	// in:body
	CreateTeamOption api.CreateTeamOption
	// in:body
	EditTeamOption api.EditTeamOption

	// in:body
	CreateUserOption api.CreateUserOption

	// in:body
	EditUserOption api.EditUserOption

	// in:body
	CreateOAuth2ApplicationOptions api.CreateOAuth2ApplicationOptions
}
