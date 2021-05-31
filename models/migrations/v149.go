// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package migrations

import (
	"fmt"

	"go.wandrs.dev/framework/modules/timeutil"

	"xorm.io/xorm"
)

func addCreatedAndUpdatedToMilestones(x *xorm.Engine) error {
	type Milestone struct {
		CreatedUnix timeutil.TimeStamp `xorm:"INDEX created"`
		UpdatedUnix timeutil.TimeStamp `xorm:"INDEX updated"`
	}

	if err := x.Sync2(new(Milestone)); err != nil {
		return fmt.Errorf("Sync2: %v", err)
	}
	return nil
}
