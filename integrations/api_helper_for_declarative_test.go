// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package integrations

import (
	"fmt"
	"net/http"
	"testing"

	api "code.gitea.io/gitea/modules/structs"
)

type APITestContext struct {
	Session      *TestSession
	Token        string
	Username     string
	ExpectedCode int
}

func NewAPITestContext(t *testing.T, username string) APITestContext {
	session := loginUser(t, username)
	token := getTokenForLoggedInUser(t, session)
	return APITestContext{
		Session:  session,
		Token:    token,
		Username: username,
	}
}

func doAPICreateOrganization(ctx APITestContext, options *api.CreateOrgOption, callback ...func(*testing.T, api.Organization)) func(t *testing.T) {
	return func(t *testing.T) {
		url := fmt.Sprintf("/api/v1/orgs?token=%s", ctx.Token)

		req := NewRequestWithJSON(t, "POST", url, &options)
		if ctx.ExpectedCode != 0 {
			ctx.Session.MakeRequest(t, req, ctx.ExpectedCode)
			return
		}
		resp := ctx.Session.MakeRequest(t, req, http.StatusCreated)

		var contents api.Organization
		DecodeJSON(t, resp, &contents)
		if len(callback) > 0 {
			callback[0](t, contents)
		}
	}
}

func doAPICreateOrganizationTeam(ctx APITestContext, orgName string, options *api.CreateTeamOption, callback ...func(*testing.T, api.Team)) func(t *testing.T) {
	return func(t *testing.T) {
		url := fmt.Sprintf("/api/v1/orgs/%s/teams?token=%s", orgName, ctx.Token)

		req := NewRequestWithJSON(t, "POST", url, &options)
		if ctx.ExpectedCode != 0 {
			ctx.Session.MakeRequest(t, req, ctx.ExpectedCode)
			return
		}
		resp := ctx.Session.MakeRequest(t, req, http.StatusCreated)

		var contents api.Team
		DecodeJSON(t, resp, &contents)
		if len(callback) > 0 {
			callback[0](t, contents)
		}
	}
}

func doAPIAddUserToOrganizationTeam(ctx APITestContext, teamID int64, username string) func(t *testing.T) {
	return func(t *testing.T) {
		url := fmt.Sprintf("/api/v1/teams/%d/members/%s?token=%s", teamID, username, ctx.Token)

		req := NewRequest(t, "PUT", url)
		if ctx.ExpectedCode != 0 {
			ctx.Session.MakeRequest(t, req, ctx.ExpectedCode)
			return
		}
		ctx.Session.MakeRequest(t, req, http.StatusNoContent)
	}
}
