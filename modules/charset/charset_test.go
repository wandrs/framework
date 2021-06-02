// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package charset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveBOMIfPresent(t *testing.T) {
	res := RemoveBOMIfPresent([]byte{0xc3, 0xa1, 0xc3, 0xa9, 0xc3, 0xad, 0xc3, 0xb3, 0xc3, 0xba})
	assert.Equal(t, []byte{0xc3, 0xa1, 0xc3, 0xa9, 0xc3, 0xad, 0xc3, 0xb3, 0xc3, 0xba}, res)

	res = RemoveBOMIfPresent([]byte{0xef, 0xbb, 0xbf, 0xc3, 0xa1, 0xc3, 0xa9, 0xc3, 0xad, 0xc3, 0xb3, 0xc3, 0xba})
	assert.Equal(t, []byte{0xc3, 0xa1, 0xc3, 0xa9, 0xc3, 0xad, 0xc3, 0xb3, 0xc3, 0xba}, res)
}
