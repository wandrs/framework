// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package migrations

import (
	"fmt"

	"go.wandrs.dev/framework/modules/timeutil"

	"xorm.io/xorm"
)

func addLanguageStats(x *xorm.Engine) error {
	// LanguageStat see models/repo_language_stats.go
	type LanguageStat struct {
		ID          int64 `xorm:"pk autoincr"`
		RepoID      int64 `xorm:"UNIQUE(s) INDEX NOT NULL"`
		CommitID    string
		IsPrimary   bool
		Language    string             `xorm:"VARCHAR(30) UNIQUE(s) INDEX NOT NULL"`
		Percentage  float32            `xorm:"NUMERIC(5,2) NOT NULL DEFAULT 0"`
		Color       string             `xorm:"-"`
		CreatedUnix timeutil.TimeStamp `xorm:"INDEX CREATED"`
	}

	type RepoIndexerType int

	// RepoIndexerStatus see models/repo_stats_indexer.go
	type RepoIndexerStatus struct {
		ID          int64           `xorm:"pk autoincr"`
		RepoID      int64           `xorm:"INDEX(s)"`
		CommitSha   string          `xorm:"VARCHAR(40)"`
		IndexerType RepoIndexerType `xorm:"INDEX(s) NOT NULL DEFAULT 0"`
	}

	if err := x.Sync2(new(LanguageStat)); err != nil {
		return fmt.Errorf("Sync2: %v", err)
	}
	if err := x.Sync2(new(RepoIndexerStatus)); err != nil {
		return fmt.Errorf("Sync2: %v", err)
	}
	return nil
}
