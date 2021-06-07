// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package webhook

import (
	"testing"

	api "go.wandrs.dev/framework/modules/structs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDingTalkIssuesPayload(t *testing.T) {
	p := issueTestPayload()
	d := new(DingtalkPayload)
	p.Action = api.HookIssueOpened
	pl, err := d.Issue(p)
	require.NoError(t, err)
	require.NotNil(t, pl)
	assert.Equal(t, "#2 crash", pl.(*DingtalkPayload).ActionCard.Title)
	assert.Equal(t, "[test/repo] Issue opened: #2 crash by user1\r\n\r\n", pl.(*DingtalkPayload).ActionCard.Text)

	p.Action = api.HookIssueClosed
	pl, err = d.Issue(p)
	require.NoError(t, err)
	require.NotNil(t, pl)
	assert.Equal(t, "#2 crash", pl.(*DingtalkPayload).ActionCard.Title)
	assert.Equal(t, "[test/repo] Issue closed: #2 crash by user1\r\n\r\n", pl.(*DingtalkPayload).ActionCard.Text)
}
