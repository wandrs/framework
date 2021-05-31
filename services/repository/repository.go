// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repository

import (
	"fmt"

	"go.wandrs.dev/framework/models"
	"go.wandrs.dev/framework/modules/log"
	"go.wandrs.dev/framework/modules/notification"
	repo_module "go.wandrs.dev/framework/modules/repository"
	cfg "go.wandrs.dev/framework/modules/setting"
	pull_service "go.wandrs.dev/framework/services/pull"
)

// CreateRepository creates a repository for the user/organization.
func CreateRepository(doer, owner *models.User, opts models.CreateRepoOptions) (*models.Repository, error) {
	repo, err := repo_module.CreateRepository(doer, owner, opts)
	if err != nil {
		// No need to rollback here we should do this in CreateRepository...
		return nil, err
	}

	notification.NotifyCreateRepository(doer, owner, repo)

	return repo, nil
}

// AdoptRepository adopts pre-existing repository files for the user/organization.
func AdoptRepository(doer, owner *models.User, opts models.CreateRepoOptions) (*models.Repository, error) {
	repo, err := repo_module.AdoptRepository(doer, owner, opts)
	if err != nil {
		// No need to rollback here we should do this in AdoptRepository...
		return nil, err
	}

	notification.NotifyCreateRepository(doer, owner, repo)

	return repo, nil
}

// DeleteUnadoptedRepository adopts pre-existing repository files for the user/organization.
func DeleteUnadoptedRepository(doer, owner *models.User, name string) error {
	return repo_module.DeleteUnadoptedRepository(doer, owner, name)
}

// ForkRepository forks a repository
func ForkRepository(doer, u *models.User, oldRepo *models.Repository, name, desc string) (*models.Repository, error) {
	repo, err := repo_module.ForkRepository(doer, u, oldRepo, name, desc)
	if err != nil {
		return nil, err
	}

	notification.NotifyForkRepository(doer, oldRepo, repo)

	return repo, nil
}

// DeleteRepository deletes a repository for a user or organization.
func DeleteRepository(doer *models.User, repo *models.Repository) error {
	if err := pull_service.CloseRepoBranchesPulls(doer, repo); err != nil {
		log.Error("CloseRepoBranchesPulls failed: %v", err)
	}

	// If the repo itself has webhooks, we need to trigger them before deleting it...
	notification.NotifyDeleteRepository(doer, repo)

	err := models.DeleteRepository(doer, repo.OwnerID, repo.ID)
	return err
}

// PushCreateRepo creates a repository when a new repository is pushed to an appropriate namespace
func PushCreateRepo(authUser, owner *models.User, repoName string) (*models.Repository, error) {
	if !authUser.IsAdmin {
		if owner.IsOrganization() {
			if ok, err := owner.CanCreateOrgRepo(authUser.ID); err != nil {
				return nil, err
			} else if !ok {
				return nil, fmt.Errorf("cannot push-create repository for org")
			}
		} else if authUser.ID != owner.ID {
			return nil, fmt.Errorf("cannot push-create repository for another user")
		}
	}

	repo, err := CreateRepository(authUser, owner, models.CreateRepoOptions{
		Name:      repoName,
		IsPrivate: cfg.Repository.DefaultPushCreatePrivate,
	})
	if err != nil {
		return nil, err
	}

	return repo, nil
}

// NewContext start repository service
func NewContext() error {
	return initPushQueue()
}
