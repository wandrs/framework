// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package integrations

import (
	"net/http"
	"path"
	"testing"

	"go.wandrs.dev/framework/modules/setting"
	"go.wandrs.dev/framework/modules/test"

	"github.com/stretchr/testify/assert"
)

func TestLinksNoLogin(t *testing.T) {
	defer prepareTestEnv(t)()

	var links = []string{
		"/explore/repos",
		"/explore/repos?q=test&tab=",
		"/explore/users",
		"/explore/users?q=test&tab=",
		"/explore/organizations",
		"/explore/organizations?q=test&tab=",
		"/",
		"/user/sign_up",
		"/user/login",
		"/user/forgot_password",
		"/api/swagger",
		"/user2/repo1",
		"/user2/repo1/projects",
		"/user2/repo1/projects/1",
		"/assets/img/404.png",
		"/assets/img/500.png",
	}

	for _, link := range links {
		req := NewRequest(t, "GET", link)
		MakeRequest(t, req, http.StatusOK)
	}
}

func TestRedirectsNoLogin(t *testing.T) {
	defer prepareTestEnv(t)()

	var redirects = map[string]string{
		"/user2/repo1/commits/master":                "/user2/repo1/commits/branch/master",
		"/user2/repo1/src/master":                    "/user2/repo1/src/branch/master",
		"/user2/repo1/src/master/file.txt":           "/user2/repo1/src/branch/master/file.txt",
		"/user2/repo1/src/master/directory/file.txt": "/user2/repo1/src/branch/master/directory/file.txt",
		"/user/avatar/Ghost/-1":                      "/assets/img/avatar_default.png",
		"/api/v1/swagger":                            "/api/swagger",
	}
	for link, redirectLink := range redirects {
		req := NewRequest(t, "GET", link)
		resp := MakeRequest(t, req, http.StatusFound)
		assert.EqualValues(t, path.Join(setting.AppSubURL, redirectLink), test.RedirectURL(resp))
	}
}

func TestNoLoginNotExist(t *testing.T) {
	defer prepareTestEnv(t)()

	var links = []string{
		"/user5/repo4/projects",
		"/user5/repo4/projects/3",
	}

	for _, link := range links {
		req := NewRequest(t, "GET", link)
		MakeRequest(t, req, http.StatusNotFound)
	}
}

func testLinksAsUser(userName string, t *testing.T) {
	var links = []string{
		"/explore/users",
		"/explore/users?q=test&tab=",
		"/explore/organizations",
		"/explore/organizations?q=test&tab=",
		"/",
		"/user/forgot_password",
		"/api/swagger",
		"/org/create",
		"/user2",
		"/user2?tab=stars",
		"/user2?tab=activity",
		"/user/settings",
		"/user/settings/account",
		"/user/settings/security",
		"/user/settings/security/two_factor/enroll",
		"/user/settings/keys",
		"/user/settings/organization",
	}

	session := loginUser(t, userName)
	for _, link := range links {
		req := NewRequest(t, "GET", link)
		session.MakeRequest(t, req, http.StatusOK)
	}

	reqAPI := NewRequestf(t, "GET", "/api/v1/users/%s/repos", userName)
	respAPI := MakeRequest(t, reqAPI, http.StatusOK)
}

func TestLinksLogin(t *testing.T) {
	defer prepareTestEnv(t)()

	testLinksAsUser("user2", t)
}
