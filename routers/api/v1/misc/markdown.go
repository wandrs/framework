// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package misc

import (
	"net/http"
	"strings"

	"code.gitea.io/gitea/modules/context"
	"code.gitea.io/gitea/modules/markup"
	"code.gitea.io/gitea/modules/markup/markdown"
	"code.gitea.io/gitea/modules/setting"
	api "code.gitea.io/gitea/modules/structs"
	"code.gitea.io/gitea/modules/util"
	"code.gitea.io/gitea/modules/web"

	"mvdan.cc/xurls/v2"
)

// Markdown render markdown document to HTML
func Markdown(ctx *context.APIContext) {
	// swagger:operation POST /markdown miscellaneous renderMarkdown
	// ---
	// summary: Render a markdown document as HTML
	// parameters:
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/MarkdownOption"
	// consumes:
	// - application/json
	// produces:
	//     - text/html
	// responses:
	//   "200":
	//     "$ref": "#/responses/MarkdownRender"
	//   "422":
	//     "$ref": "#/responses/validationError"

	form := web.GetForm(ctx).(*api.MarkdownOption)

	if ctx.HasAPIError() {
		ctx.Error(http.StatusUnprocessableEntity, "", ctx.GetErrMsg())
		return
	}

	if len(form.Text) == 0 {
		_, _ = ctx.Write([]byte(""))
		return
	}

	switch form.Mode {
	case "comment":
		fallthrough
	case "gfm":
		urlPrefix := form.Context
		meta := map[string]string{}
		if !strings.HasPrefix(setting.AppSubURL+"/", urlPrefix) {
			// check if urlPrefix is already set to a URL
			linkRegex, _ := xurls.StrictMatchingScheme("https?://")
			m := linkRegex.FindStringIndex(urlPrefix)
			if m == nil {
				urlPrefix = util.URLJoin(setting.AppURL, form.Context)
			}
		}
		if form.Mode == "gfm" {
			meta["mode"] = "document"
		}

		if err := markdown.Render(&markup.RenderContext{
			URLPrefix: urlPrefix,
			Metas:     meta,
			IsWiki:    form.Wiki,
		}, strings.NewReader(form.Text), ctx.Resp); err != nil {
			ctx.InternalServerError(err)
			return
		}
	default:
		if err := markdown.RenderRaw(&markup.RenderContext{
			URLPrefix: form.Context,
		}, strings.NewReader(form.Text), ctx.Resp); err != nil {
			ctx.InternalServerError(err)
			return
		}
	}
}

// MarkdownRaw render raw markdown HTML
func MarkdownRaw(ctx *context.APIContext) {
	// swagger:operation POST /markdown/raw miscellaneous renderMarkdownRaw
	// ---
	// summary: Render raw markdown as HTML
	// parameters:
	//     - name: body
	//       in: body
	//       description: Request body to render
	//       required: true
	//       schema:
	//         type: string
	// consumes:
	//     - text/plain
	// produces:
	//     - text/html
	// responses:
	//   "200":
	//     "$ref": "#/responses/MarkdownRender"
	//   "422":
	//     "$ref": "#/responses/validationError"
	defer ctx.Req.Body.Close()
	if err := markdown.RenderRaw(&markup.RenderContext{}, ctx.Req.Body, ctx.Resp); err != nil {
		ctx.InternalServerError(err)
		return
	}
}
