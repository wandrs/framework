// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sso

import (
	"net/http"

	"go.wandrs.dev/framework/models"
	"go.wandrs.dev/framework/modules/session"
	"go.wandrs.dev/framework/modules/web/middleware"
)

// DataStore represents a data store
type DataStore middleware.DataStore

// SessionStore represents a session store
type SessionStore session.Store

// SingleSignOn represents a SSO authentication method (plugin) for HTTP requests.
type SingleSignOn interface {
	// Init should be called exactly once before using any of the other methods,
	// in order to allow the plugin to allocate necessary resources
	Init() error

	// Free should be called exactly once before application closes, in order to
	// give chance to the plugin to free any allocated resources
	Free() error

	// IsEnabled checks if the current SSO method has been enabled in settings.
	IsEnabled() bool

	// VerifyAuthData tries to verify the SSO authentication data contained in the request.
	// If verification is successful returns either an existing user object (with id > 0)
	// or a new user object (with id = 0) populated with the information that was found
	// in the authentication data (username or email).
	// Returns nil if verification fails.
	VerifyAuthData(http *http.Request, w http.ResponseWriter, store DataStore, sess SessionStore) *models.User
}
