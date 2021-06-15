// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package test

import (
	scontext "context"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/context"
	"code.gitea.io/gitea/modules/web/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/unrolled/render"
)

// MockContext mock context for unit tests
func MockContext(t *testing.T, path string) *context.Context {
	var resp = &mockResponseWriter{}
	var ctx = context.Context{
		Render: &mockRender{},
		Data:   make(map[string]interface{}),
		Flash: &middleware.Flash{
			Values: make(url.Values),
		},
		Resp:   context.NewResponse(resp),
		Locale: &mockLocale{},
	}

	requestURL, err := url.Parse(path)
	assert.NoError(t, err)
	var req = &http.Request{
		URL:  requestURL,
		Form: url.Values{},
	}

	chiCtx := chi.NewRouteContext()
	req = req.WithContext(scontext.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
	ctx.Req = context.WithContext(req, &ctx)
	return &ctx
}

// LoadUser load a user into a test context.
func LoadUser(t *testing.T, ctx *context.Context, userID int64) {
	ctx.User = models.AssertExistsAndLoadBean(t, &models.User{ID: userID}).(*models.User)
}

type mockLocale struct{}

func (l mockLocale) Language() string {
	return "en"
}

func (l mockLocale) Tr(s string, _ ...interface{}) string {
	return s
}

type mockResponseWriter struct {
	httptest.ResponseRecorder
	size int
}

func (rw *mockResponseWriter) Write(b []byte) (int, error) {
	rw.size += len(b)
	return rw.ResponseRecorder.Write(b)
}

func (rw *mockResponseWriter) Status() int {
	return rw.ResponseRecorder.Code
}

func (rw *mockResponseWriter) Written() bool {
	return rw.ResponseRecorder.Code > 0
}

func (rw *mockResponseWriter) Size() int {
	return rw.size
}

func (rw *mockResponseWriter) Push(target string, opts *http.PushOptions) error {
	return nil
}

type mockRender struct {
}

func (tr *mockRender) TemplateLookup(tmpl string) *template.Template {
	return nil
}

func (tr *mockRender) HTML(w io.Writer, status int, _ string, _ interface{}, _ ...render.HTMLOptions) error {
	if resp, ok := w.(http.ResponseWriter); ok {
		resp.WriteHeader(status)
	}
	return nil
}
