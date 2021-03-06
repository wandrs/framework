// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cron

import (
	"context"
	"time"

	"go.wandrs.dev/framework/models"
)

func registerDeleteInactiveUsers() {
	RegisterTaskFatal("delete_inactive_accounts", &OlderThanConfig{
		BaseConfig: BaseConfig{
			Enabled:    false,
			RunAtStart: false,
			Schedule:   "@annually",
		},
		OlderThan: 0 * time.Second,
	}, func(ctx context.Context, _ *models.User, config Config) error {
		olderThanConfig := config.(*OlderThanConfig)
		return models.DeleteInactiveUsers(ctx, olderThanConfig.OlderThan)
	})
}

func initExtendedTasks() {
	registerDeleteInactiveUsers()
}
