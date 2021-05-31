// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package migrations

import (
	"context"

	"go.wandrs.dev/framework/models"
	"go.wandrs.dev/framework/modules/log"
	"go.wandrs.dev/framework/modules/structs"
)

// UpdateMigrationPosterID updates all migrated repositories' issues and comments posterID
func UpdateMigrationPosterID(ctx context.Context) error {
	for _, gitService := range structs.SupportedFullGitService {
		select {
		case <-ctx.Done():
			log.Warn("UpdateMigrationPosterID aborted before %s", gitService.Name())
			return models.ErrCancelledf("during UpdateMigrationPosterID before %s", gitService.Name())
		default:
		}
		if err := updateMigrationPosterIDByGitService(ctx, gitService); err != nil {
			log.Error("updateMigrationPosterIDByGitService failed: %v", err)
		}
	}
	return nil
}

func updateMigrationPosterIDByGitService(ctx context.Context, tp structs.GitServiceType) error {
	provider := tp.Name()
	if len(provider) == 0 {
		return nil
	}

	const batchSize = 100
	var start int
	for {
		select {
		case <-ctx.Done():
			log.Warn("UpdateMigrationPosterIDByGitService(%s) cancelled", tp.Name())
			return nil
		default:
		}

		users, err := models.FindExternalUsersByProvider(models.FindExternalUserOptions{
			Provider: provider,
			Start:    start,
			Limit:    batchSize,
		})
		if err != nil {
			return err
		}

		for _, user := range users {
			select {
			case <-ctx.Done():
				log.Warn("UpdateMigrationPosterIDByGitService(%s) cancelled", tp.Name())
				return nil
			default:
			}
			externalUserID := user.ExternalID
			if err := models.UpdateMigrationsByType(tp, externalUserID, user.UserID); err != nil {
				log.Error("UpdateMigrationsByType type %s external user id %v to local user id %v failed: %v", tp.Name(), user.ExternalID, user.UserID, err)
			}
		}

		if len(users) < batchSize {
			break
		}
		start += len(users)
	}
	return nil
}
