// Copyright 2021 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package integrations

import (
	"bytes"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"go.wandrs.dev/framework/models"
	"go.wandrs.dev/framework/modules/avatar"

	"github.com/stretchr/testify/assert"
)

func TestUserAvatar(t *testing.T) {
	onGiteaRun(t, func(t *testing.T, u *url.URL) {
		user2 := models.AssertExistsAndLoadBean(t, &models.User{ID: 2}).(*models.User) // owner of the repo3, is an org

		seed := user2.Email
		if len(seed) == 0 {
			seed = user2.Name
		}

		img, err := avatar.RandomImage([]byte(seed))
		if err != nil {
			assert.NoError(t, err)
			return
		}

		session := loginUser(t, "user2")
		csrf := GetCSRF(t, session, "/user/settings")

		imgData := &bytes.Buffer{}

		body := &bytes.Buffer{}

		//Setup multi-part
		writer := multipart.NewWriter(body)
		writer.WriteField("source", "local")
		part, err := writer.CreateFormFile("avatar", "avatar-for-testuseravatar.png")
		if err != nil {
			assert.NoError(t, err)
			return
		}

		if err := png.Encode(imgData, img); err != nil {
			assert.NoError(t, err)
			return
		}

		if _, err := io.Copy(part, imgData); err != nil {
			assert.NoError(t, err)
			return
		}

		if err := writer.Close(); err != nil {
			assert.NoError(t, err)
			return
		}

		req := NewRequestWithBody(t, "POST", "/user/settings/avatar", body)
		req.Header.Add("X-Csrf-Token", csrf)
		req.Header.Add("Content-Type", writer.FormDataContentType())

		session.MakeRequest(t, req, http.StatusFound)

		user2 = models.AssertExistsAndLoadBean(t, &models.User{ID: 2}).(*models.User) // owner of the repo3, is an org

		req = NewRequest(t, "GET", user2.AvatarLink())
		resp := session.MakeRequest(t, req, http.StatusFound)
		location := resp.Header().Get("Location")
		if !strings.HasPrefix(location, "/avatars") {
			assert.Fail(t, "Avatar location is not local: %s", location)
		}
		req = NewRequest(t, "GET", location)
		session.MakeRequest(t, req, http.StatusOK)

		// Can't test if the response matches because the image is regened on upload but checking that this at least doesn't give a 404 should be enough.
	})
}
