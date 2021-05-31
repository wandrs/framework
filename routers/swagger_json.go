// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package routers

import (
	"net/http"

	"go.wandrs.dev/framework/modules/base"
	"go.wandrs.dev/framework/modules/context"
	"go.wandrs.dev/framework/modules/log"
)

// tplSwaggerV1Json swagger v1 json template
const tplSwaggerV1Json base.TplName = "swagger/v1_json"

// SwaggerV1Json render swagger v1 json
func SwaggerV1Json(ctx *context.Context) {
	t := ctx.Render.TemplateLookup(string(tplSwaggerV1Json))
	ctx.Resp.Header().Set("Content-Type", "application/json")
	if err := t.Execute(ctx.Resp, ctx.Data); err != nil {
		log.Error("%v", err)
		ctx.Error(http.StatusInternalServerError)
	}
}
