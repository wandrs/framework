// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.package models

package integrations

import (
	"fmt"
	"net/http"
	"testing"

	"go.wandrs.dev/framework/models"
	"go.wandrs.dev/framework/modules/setting"
	api "go.wandrs.dev/framework/modules/structs"

	"github.com/stretchr/testify/assert"
)

type SearchResults struct {
	OK   bool        `json:"ok"`
	Data []*api.User `json:"data"`
}

func TestAPIUserSearchLoggedIn(t *testing.T) {
	defer prepareTestEnv(t)()
	adminUsername := "user1"
	session := loginUser(t, adminUsername)
	token := getTokenForLoggedInUser(t, session)
	query := "user2"
	req := NewRequestf(t, "GET", "/api/v1/users/search?token=%s&q=%s", token, query)
	resp := session.MakeRequest(t, req, http.StatusOK)

	var results SearchResults
	DecodeJSON(t, resp, &results)
	assert.NotEmpty(t, results.Data)
	for _, user := range results.Data {
		assert.Contains(t, user.UserName, query)
		assert.NotEmpty(t, user.Email)
	}
}

func TestAPIUserSearchNotLoggedIn(t *testing.T) {
	defer prepareTestEnv(t)()
	query := "user2"
	req := NewRequestf(t, "GET", "/api/v1/users/search?q=%s", query)
	resp := MakeRequest(t, req, http.StatusOK)

	var results SearchResults
	DecodeJSON(t, resp, &results)
	assert.NotEmpty(t, results.Data)
	var modelUser *models.User
	for _, user := range results.Data {
		assert.Contains(t, user.UserName, query)
		modelUser = models.AssertExistsAndLoadBean(t, &models.User{ID: user.ID}).(*models.User)
		if modelUser.KeepEmailPrivate {
			assert.EqualValues(t, fmt.Sprintf("%s@%s", modelUser.LowerName, setting.Service.NoReplyAddress), user.Email)
		} else {
			assert.EqualValues(t, modelUser.Email, user.Email)
		}
	}
}
