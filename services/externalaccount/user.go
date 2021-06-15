// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package externalaccount

import (
	"code.gitea.io/gitea/models"

	"github.com/markbates/goth"
)

// LinkAccountToUser link the gothUser to the user
func LinkAccountToUser(user *models.User, gothUser goth.User) error {
	loginSource, err := models.GetActiveOAuth2LoginSourceByName(gothUser.Provider)
	if err != nil {
		return err
	}

	externalLoginUser := &models.ExternalLoginUser{
		ExternalID:        gothUser.UserID,
		UserID:            user.ID,
		LoginSourceID:     loginSource.ID,
		RawData:           gothUser.RawData,
		Provider:          gothUser.Provider,
		Email:             gothUser.Email,
		Name:              gothUser.Name,
		FirstName:         gothUser.FirstName,
		LastName:          gothUser.LastName,
		NickName:          gothUser.NickName,
		Description:       gothUser.Description,
		AvatarURL:         gothUser.AvatarURL,
		Location:          gothUser.Location,
		AccessToken:       gothUser.AccessToken,
		AccessTokenSecret: gothUser.AccessTokenSecret,
		RefreshToken:      gothUser.RefreshToken,
		ExpiresAt:         gothUser.ExpiresAt,
	}

	return models.LinkExternalToUser(user, externalLoginUser)
}
