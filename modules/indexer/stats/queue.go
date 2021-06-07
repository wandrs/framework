// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package stats

import (
	"fmt"

	"go.wandrs.dev/framework/models"
	"go.wandrs.dev/framework/modules/graceful"
	"go.wandrs.dev/framework/modules/log"
	"go.wandrs.dev/framework/modules/queue"
)

// statsQueue represents a queue to handle repository stats updates
var statsQueue queue.UniqueQueue

// handle passed PR IDs and test the PRs
func handle(data ...queue.Data) {
	for _, datum := range data {
		opts := datum.(int64)
		if err := indexer.Index(opts); err != nil {
			log.Error("stats queue indexer.Index(%d) failed: %v", opts, err)
		}
	}
}

func initStatsQueue() error {
	statsQueue = queue.CreateUniqueQueue("repo_stats_update", handle, int64(0)).(queue.UniqueQueue)
	if statsQueue == nil {
		return fmt.Errorf("Unable to create repo_stats_update Queue")
	}

	go graceful.GetManager().RunWithShutdownFns(statsQueue.Run)

	return nil
}

// UpdateRepoIndexer update a repository's entries in the indexer
func UpdateRepoIndexer(repo *models.Repository) error {
	if err := statsQueue.Push(repo.ID); err != nil {
		if err != queue.ErrAlreadyInQueue {
			return err
		}
		log.Debug("Repo ID: %d already queued", repo.ID)
	}
	return nil
}
