// Copyright 2014 The Gogs Authors. All rights reserved.
// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sso

import (
	"net/http"
	"strings"
	"time"

	"go.wandrs.dev/framework/models"
	"go.wandrs.dev/framework/modules/log"
	"go.wandrs.dev/framework/modules/timeutil"
	"go.wandrs.dev/framework/modules/web/middleware"
)

// Ensure the struct implements the interface.
var (
	_ SingleSignOn = &OAuth2{}
)

// CheckOAuthAccessToken returns uid of user from oauth token
func CheckOAuthAccessToken(accessToken string) int64 {
	// JWT tokens require a "."
	if !strings.Contains(accessToken, ".") {
		return 0
	}
	token, err := models.ParseOAuth2Token(accessToken)
	if err != nil {
		log.Trace("ParseOAuth2Token: %v", err)
		return 0
	}
	var grant *models.OAuth2Grant
	if grant, err = models.GetOAuth2GrantByID(token.GrantID); err != nil || grant == nil {
		return 0
	}
	if token.Type != models.TypeAccessToken {
		return 0
	}
	if token.ExpiresAt < time.Now().Unix() || token.IssuedAt > time.Now().Unix() {
		return 0
	}
	return grant.UserID
}

// OAuth2 implements the SingleSignOn interface and authenticates requests
// (API requests only) by looking for an OAuth token in query parameters or the
// "Authorization" header.
type OAuth2 struct{}

// Init does nothing as the OAuth2 implementation does not need to allocate any resources
func (o *OAuth2) Init() error {
	return nil
}

// Free does nothing as the OAuth2 implementation does not have to release any resources
func (o *OAuth2) Free() error {
	return nil
}

// userIDFromToken returns the user id corresponding to the OAuth token.
func (o *OAuth2) userIDFromToken(req *http.Request, store DataStore) int64 {
	_ = req.ParseForm()

	// Check access token.
	tokenSHA := req.Form.Get("token")
	if len(tokenSHA) == 0 {
		tokenSHA = req.Form.Get("access_token")
	}
	if len(tokenSHA) == 0 {
		// Well, check with header again.
		auHead := req.Header.Get("Authorization")
		if len(auHead) > 0 {
			auths := strings.Fields(auHead)
			if len(auths) == 2 && (auths[0] == "token" || strings.ToLower(auths[0]) == "bearer") {
				tokenSHA = auths[1]
			}
		}
	}
	if len(tokenSHA) == 0 {
		return 0
	}

	// Let's see if token is valid.
	if strings.Contains(tokenSHA, ".") {
		uid := CheckOAuthAccessToken(tokenSHA)
		if uid != 0 {
			store.GetData()["IsApiToken"] = true
		}
		return uid
	}
	t, err := models.GetAccessTokenBySHA(tokenSHA)
	if err != nil {
		if !models.IsErrAccessTokenNotExist(err) && !models.IsErrAccessTokenEmpty(err) {
			log.Error("GetAccessTokenBySHA: %v", err)
		}
		return 0
	}
	t.UpdatedUnix = timeutil.TimeStampNow()
	if err = models.UpdateAccessToken(t); err != nil {
		log.Error("UpdateAccessToken: %v", err)
	}
	store.GetData()["IsApiToken"] = true
	return t.UID
}

// IsEnabled returns true as this plugin is enabled by default and its not possible
// to disable it from settings.
func (o *OAuth2) IsEnabled() bool {
	return true
}

// VerifyAuthData extracts the user ID from the OAuth token in the query parameters
// or the "Authorization" header and returns the corresponding user object for that ID.
// If verification is successful returns an existing user object.
// Returns nil if verification fails.
func (o *OAuth2) VerifyAuthData(req *http.Request, w http.ResponseWriter, store DataStore, sess SessionStore) *models.User {
	if !models.HasEngine {
		return nil
	}

	if middleware.IsInternalPath(req) || !middleware.IsAPIPath(req) && !isAttachmentDownload(req) {
		return nil
	}

	id := o.userIDFromToken(req, store)
	if id <= 0 {
		return nil
	}
	log.Trace("OAuth2 Authorization: Found token for user[%d]", id)

	user, err := models.GetUserByID(id)
	if err != nil {
		if !models.IsErrUserNotExist(err) {
			log.Error("GetUserByName: %v", err)
		}
		return nil
	}

	log.Trace("OAuth2 Authorization: Logged in user %-v", user)
	return user
}
