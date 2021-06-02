// Copyright 2015 The Gogs Authors. All rights reserved.
// Copyright 2018 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package convert

import (
	"go.wandrs.dev/framework/models"
	api "go.wandrs.dev/framework/modules/structs"
)

// ToEmail convert models.EmailAddress to api.Email
func ToEmail(email *models.EmailAddress) *api.Email {
	return &api.Email{
		Email:    email.Email,
		Verified: email.IsActivated,
		Primary:  email.IsPrimary,
	}
}

// ToOrganization convert models.User to api.Organization
func ToOrganization(org *models.User) *api.Organization {
	return &api.Organization{
		ID:                        org.ID,
		AvatarURL:                 org.AvatarLink(),
		UserName:                  org.Name,
		FullName:                  org.FullName,
		Description:               org.Description,
		Website:                   org.Website,
		Location:                  org.Location,
		Visibility:                org.Visibility.String(),
		RepoAdminChangeTeamAccess: org.RepoAdminChangeTeamAccess,
	}
}

// ToTeam convert models.Team to api.Team
func ToTeam(team *models.Team) *api.Team {
	if team == nil {
		return nil
	}

	return &api.Team{
		ID:          team.ID,
		Name:        team.Name,
		Description: team.Description,
		Permission:  team.Authorize.String(),
	}
}

// ToOAuth2Application convert from models.OAuth2Application to api.OAuth2Application
func ToOAuth2Application(app *models.OAuth2Application) *api.OAuth2Application {
	return &api.OAuth2Application{
		ID:           app.ID,
		Name:         app.Name,
		ClientID:     app.ClientID,
		ClientSecret: app.ClientSecret,
		RedirectURIs: app.RedirectURIs,
		Created:      app.CreatedUnix.AsTime(),
	}
}
