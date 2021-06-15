// Copyright 2021 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package convert

import (
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestToCorrectPageSize(t *testing.T) {
	assert.EqualValues(t, 30, ToCorrectPageSize(0))
	assert.EqualValues(t, 30, ToCorrectPageSize(-10))
	assert.EqualValues(t, 20, ToCorrectPageSize(20))
	assert.EqualValues(t, 50, ToCorrectPageSize(100))
}
