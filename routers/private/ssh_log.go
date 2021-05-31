// Copyright 2021 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package private

import (
	"net/http"

	"go.wandrs.dev/framework/modules/context"
	"go.wandrs.dev/framework/modules/log"
	"go.wandrs.dev/framework/modules/private"
	"go.wandrs.dev/framework/modules/setting"
	"go.wandrs.dev/framework/modules/web"
)

// SSHLog hook to response ssh log
func SSHLog(ctx *context.PrivateContext) {
	if !setting.EnableSSHLog {
		ctx.Status(http.StatusOK)
		return
	}

	opts := web.GetForm(ctx).(*private.SSHLogOption)

	if opts.IsError {
		log.Error("ssh: %v", opts.Message)
		ctx.Status(http.StatusOK)
		return
	}

	log.Debug("ssh: %v", opts.Message)
	ctx.Status(http.StatusOK)
}
