// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cron

import (
	"context"

	"code.gitea.io/gitea/models"
)

func registerSyncExternalUsers() {
	RegisterTaskFatal("sync_external_users", &UpdateExistingConfig{
		BaseConfig: BaseConfig{
			Enabled:    true,
			RunAtStart: false,
			Schedule:   "@every 24h",
		},
		UpdateExisting: true,
	}, func(ctx context.Context, _ *models.User, config Config) error {
		realConfig := config.(*UpdateExistingConfig)
		return models.SyncExternalUsers(ctx, realConfig.UpdateExisting)
	})
}

func initBasicTasks() {
	registerSyncExternalUsers()
}
