// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package private includes all internal routes. The package name internal is ideal but Golang is not allowed, so we use private as package name instead.
package private

import (
	"net/http"
	"reflect"
	"strings"

	"go.wandrs.dev/binding"
	"go.wandrs.dev/framework/modules/context"
	"go.wandrs.dev/framework/modules/log"
	"go.wandrs.dev/framework/modules/private"
	"go.wandrs.dev/framework/modules/setting"
	"go.wandrs.dev/framework/modules/web"
)

// CheckInternalToken check internal token is set
func CheckInternalToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		tokens := req.Header.Get("Authorization")
		fields := strings.SplitN(tokens, " ", 2)
		if len(fields) != 2 || fields[0] != "Bearer" || fields[1] != setting.InternalToken {
			log.Debug("Forbidden attempt to access internal url: Authorization header: %s", tokens)
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		} else {
			next.ServeHTTP(w, req)
		}
	})
}

// bind binding an obj to a handler
func bind(obj interface{}) http.HandlerFunc {
	tp := reflect.TypeOf(obj)
	for tp.Kind() == reflect.Ptr {
		tp = tp.Elem()
	}
	return web.Wrap(func(ctx *context.PrivateContext) {
		theObj := reflect.New(tp).Interface() // create a new form obj for every request but not use obj directly
		binding.Bind(ctx.Req, theObj)
		web.SetForm(ctx, theObj)
	})
}

// Routes registers all internal APIs routes to web application.
// These APIs will be invoked by internal commands for example `gitea serv` and etc.
func Routes() *web.Route {
	r := web.NewRoute()
	r.Use(context.PrivateContexter())
	r.Use(CheckInternalToken)

	r.Post("/manager/shutdown", Shutdown)
	r.Post("/manager/restart", Restart)
	r.Post("/manager/flush-queues", bind(private.FlushOptions{}), FlushQueues)
	r.Post("/manager/pause-logging", PauseLogging)
	r.Post("/manager/resume-logging", ResumeLogging)
	r.Post("/manager/release-and-reopen-logging", ReleaseReopenLogging)
	r.Post("/manager/add-logger", bind(private.LoggerOptions{}), AddLogger)
	r.Post("/manager/remove-logger/{group}/{name}", RemoveLogger)
	r.Post("/mail/send", SendEmail)

	return r
}
