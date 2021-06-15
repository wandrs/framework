// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package settings

import (
	"net/http"

	"code.gitea.io/gitea/modules/context"
	"code.gitea.io/gitea/modules/setting"
	api "code.gitea.io/gitea/modules/structs"
)

// GetGeneralUISettings returns instance's global settings for ui
func GetGeneralUISettings(ctx *context.APIContext) {
	// swagger:operation GET /settings/ui settings getGeneralUISettings
	// ---
	// summary: Get instance's global settings for ui
	// produces:
	// - application/json
	// responses:
	//   "200":
	//     "$ref": "#/responses/GeneralUISettings"
	ctx.JSON(http.StatusOK, api.GeneralUISettings{
		DefaultTheme:     setting.UI.DefaultTheme,
		AllowedReactions: setting.UI.Reactions,
	})
}

// GetGeneralAPISettings returns instance's global settings for api
func GetGeneralAPISettings(ctx *context.APIContext) {
	// swagger:operation GET /settings/api settings getGeneralAPISettings
	// ---
	// summary: Get instance's global settings for api
	// produces:
	// - application/json
	// responses:
	//   "200":
	//     "$ref": "#/responses/GeneralAPISettings"
	ctx.JSON(http.StatusOK, api.GeneralAPISettings{
		MaxResponseItems:   setting.API.MaxResponseItems,
		DefaultPagingNum:   setting.API.DefaultPagingNum,
		DefaultMaxBlobSize: setting.API.DefaultMaxBlobSize,
	})
}
