// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build gofuzz
// +build gofuzz

package fuzz

import (
	"bytes"
	"io"

	"go.wandrs.dev/framework/modules/markup"
	"go.wandrs.dev/framework/modules/markup/markdown"
)

// Contains fuzzing functions executed by
// fuzzing engine https://github.com/dvyukov/go-fuzz
//
// The function must return 1 if the fuzzer should increase priority of the given input during subsequent fuzzing
// (for example, the input is lexically correct and was parsed successfully).
// -1 if the input must not be added to corpus even if gives new coverage and 0 otherwise.

var renderContext = markup.RenderContext{
	URLPrefix: "https://example.com",
	Metas: map[string]string{
		"user": "go-gitea",
		"repo": "gitea",
	},
}

func FuzzMarkdownRenderRaw(data []byte) int {
	err := markdown.RenderRaw(&renderContext, bytes.NewReader(data), io.Discard)
	if err != nil {
		return 0
	}
	return 1
}

func FuzzMarkupPostProcess(data []byte) int {
	err := markup.PostProcess(&renderContext, bytes.NewReader(data), io.Discard)
	if err != nil {
		return 0
	}
	return 1
}
