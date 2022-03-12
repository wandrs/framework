// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build vendor
// +build vendor

package main

// Libraries that are included to vendor utilities used during build.
// These libraries will not be included in a normal compilation.

import (
	// for embed
	// for cover merge
	// for vet
	// for swagger

	_ "code.gitea.io/gitea-vet"
	_ "github.com/go-swagger/go-swagger/cmd/swagger"
	_ "github.com/shurcooL/vfsgen"
	_ "golang.org/x/tools/cover"
)
