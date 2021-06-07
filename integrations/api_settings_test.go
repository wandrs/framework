// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package integrations

import (
	"net/http"
	"testing"

	"go.wandrs.dev/framework/modules/setting"
	api "go.wandrs.dev/framework/modules/structs"

	"github.com/stretchr/testify/assert"
)

func TestAPIExposedSettings(t *testing.T) {
	defer prepareTestEnv(t)()

	ui := new(api.GeneralUISettings)
	req := NewRequest(t, "GET", "/api/v1/settings/ui")
	resp := MakeRequest(t, req, http.StatusOK)

	DecodeJSON(t, resp, &ui)
	assert.Len(t, ui.AllowedReactions, len(setting.UI.Reactions))
	assert.ElementsMatch(t, setting.UI.Reactions, ui.AllowedReactions)

	apiSettings := new(api.GeneralAPISettings)
	req = NewRequest(t, "GET", "/api/v1/settings/api")
	resp = MakeRequest(t, req, http.StatusOK)

	DecodeJSON(t, resp, &apiSettings)
	assert.EqualValues(t, &api.GeneralAPISettings{
		MaxResponseItems:   setting.API.MaxResponseItems,
		DefaultPagingNum:   setting.API.DefaultPagingNum,
		DefaultMaxBlobSize: setting.API.DefaultMaxBlobSize,
	}, apiSettings)
}
