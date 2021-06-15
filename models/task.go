// Copyright 2019 Gitea. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	"code.gitea.io/gitea/modules/structs"
	"code.gitea.io/gitea/modules/timeutil"

	"xorm.io/builder"
)

// Task represents a task
type Task struct {
	ID             int64
	DoerID         int64 `xorm:"index"` // operator
	Doer           *User `xorm:"-"`
	OwnerID        int64 `xorm:"index"` // repo owner id, when creating, the repoID maybe zero
	Owner          *User `xorm:"-"`
	Type           structs.TaskType
	Status         structs.TaskStatus `xorm:"index"`
	StartTime      timeutil.TimeStamp
	EndTime        timeutil.TimeStamp
	PayloadContent string             `xorm:"TEXT"`
	Errors         string             `xorm:"TEXT"` // if task failed, saved the error reason
	Created        timeutil.TimeStamp `xorm:"created"`
}

// LoadDoer loads do user
func (task *Task) LoadDoer() error {
	if task.Doer != nil {
		return nil
	}

	var doer User
	has, err := x.ID(task.DoerID).Get(&doer)
	if err != nil {
		return err
	} else if !has {
		return ErrUserNotExist{
			UID: task.DoerID,
		}
	}
	task.Doer = &doer

	return nil
}

// LoadOwner loads owner user
func (task *Task) LoadOwner() error {
	if task.Owner != nil {
		return nil
	}

	var owner User
	has, err := x.ID(task.OwnerID).Get(&owner)
	if err != nil {
		return err
	} else if !has {
		return ErrUserNotExist{
			UID: task.OwnerID,
		}
	}
	task.Owner = &owner

	return nil
}

// UpdateCols updates some columns
func (task *Task) UpdateCols(cols ...string) error {
	_, err := x.ID(task.ID).Cols(cols...).Update(task)
	return err
}

// FindTaskOptions find all tasks
type FindTaskOptions struct {
	Status int
}

// ToConds generates conditions for database operation.
func (opts FindTaskOptions) ToConds() builder.Cond {
	cond := builder.NewCond()
	if opts.Status >= 0 {
		cond = cond.And(builder.Eq{"status": opts.Status})
	}
	return cond
}

// FindTasks find all tasks
func FindTasks(opts FindTaskOptions) ([]*Task, error) {
	tasks := make([]*Task, 0, 10)
	err := x.Where(opts.ToConds()).Find(&tasks)
	return tasks, err
}

// CreateTask creates a task on database
func CreateTask(task *Task) error {
	return createTask(x, task)
}

func createTask(e Engine, task *Task) error {
	_, err := e.Insert(task)
	return err
}

// FinishMigrateTask updates database when migrate task finished
func FinishMigrateTask(task *Task) error {
	task.Status = structs.TaskStatusFinished
	task.EndTime = timeutil.TimeStampNow()
	sess := x.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}
	if _, err := sess.ID(task.ID).Cols("status", "end_time").Update(task); err != nil {
		return err
	}

	return sess.Commit()
}
