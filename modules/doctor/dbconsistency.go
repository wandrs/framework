// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package doctor

import (
	"context"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/models/migrations"
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/setting"
)

func checkDBConsistency(logger log.Logger, autofix bool) error {
	// make sure DB version is uptodate
	if err := models.NewEngine(context.Background(), migrations.EnsureUpToDate); err != nil {
		logger.Critical("Model version on the database does not match the current Gitea version. Model consistency will not be checked until the database is upgraded")
		return err
	}

	// TODO: function to recalc all counters

	if setting.Database.UsePostgreSQL {
		count, err := models.CountBadSequences()
		if err != nil {
			logger.Critical("Error: %v whilst checking sequence values", err)
			return err
		}
		if count > 0 {
			if autofix {
				err := models.FixBadSequences()
				if err != nil {
					logger.Critical("Error: %v whilst attempting to fix sequences", err)
					return err
				}
				logger.Info("%d sequences updated", count)
			} else {
				logger.Warn("%d sequences with incorrect values", count)
			}
		}
	}

	return nil
}

func init() {
	Register(&Check{
		Title:     "Check consistency of database",
		Name:      "check-db-consistency",
		IsDefault: false,
		Run:       checkDBConsistency,
		Priority:  3,
	})
}
