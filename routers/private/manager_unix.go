// +build !windows

// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package private

import (
	"net/http"

	"go.wandrs.dev/framework/modules/context"
	"go.wandrs.dev/framework/modules/graceful"
)

// Restart causes the server to perform a graceful restart
func Restart(ctx *context.PrivateContext) {
	graceful.GetManager().DoGracefulRestart()
	ctx.PlainText(http.StatusOK, []byte("success"))

}

// Shutdown causes the server to perform a graceful shutdown
func Shutdown(ctx *context.PrivateContext) {
	graceful.GetManager().DoGracefulShutdown()
	ctx.PlainText(http.StatusOK, []byte("success"))
}
