// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package notify

import (
	"net/http"

	"go.wandrs.dev/framework/models"
	"go.wandrs.dev/framework/modules/context"
	api "go.wandrs.dev/framework/modules/structs"
)

// NewAvailable check if unread notifications exist
func NewAvailable(ctx *context.APIContext) {
	// swagger:operation GET /notifications/new notification notifyNewAvailable
	// ---
	// summary: Check if unread notifications exist
	// responses:
	//   "200":
	//     "$ref": "#/responses/NotificationCount"
	ctx.JSON(http.StatusOK, api.NotificationCount{New: models.CountUnread(ctx.User)})
}
