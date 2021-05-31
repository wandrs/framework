// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package code

import (
	"io/ioutil"
	"testing"

	"go.wandrs.dev/framework/models"
	"go.wandrs.dev/framework/modules/util"

	"github.com/stretchr/testify/assert"
)

func TestBleveIndexAndSearch(t *testing.T) {
	models.PrepareTestEnv(t)

	dir, err := ioutil.TempDir("", "bleve.index")
	assert.NoError(t, err)
	if err != nil {
		assert.Fail(t, "Unable to create temporary directory")
		return
	}
	defer util.RemoveAll(dir)

	idx, _, err := NewBleveIndexer(dir)
	if err != nil {
		assert.Fail(t, "Unable to create bleve indexer Error: %v", err)
		if idx != nil {
			idx.Close()
		}
		return
	}
	defer idx.Close()

	testIndexer("beleve", t, idx)
}
