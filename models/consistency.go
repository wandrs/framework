// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"code.gitea.io/gitea/modules/setting"

	"github.com/stretchr/testify/assert"
	"xorm.io/builder"
)

// consistencyCheckable a type that can be tested for database consistency
type consistencyCheckable interface {
	checkForConsistency(t *testing.T)
}

// CheckConsistencyForAll test that the entire database is consistent
func CheckConsistencyForAll(t *testing.T) {
	CheckConsistencyFor(t,
		&User{},
		&Team{})
}

// CheckConsistencyFor test that all matching database entries are consistent
func CheckConsistencyFor(t *testing.T, beansToCheck ...interface{}) {
	for _, bean := range beansToCheck {
		sliceType := reflect.SliceOf(reflect.TypeOf(bean))
		sliceValue := reflect.MakeSlice(sliceType, 0, 10)

		ptrToSliceValue := reflect.New(sliceType)
		ptrToSliceValue.Elem().Set(sliceValue)

		assert.NoError(t, x.Table(bean).Find(ptrToSliceValue.Interface()))
		sliceValue = ptrToSliceValue.Elem()

		for i := 0; i < sliceValue.Len(); i++ {
			entity := sliceValue.Index(i).Interface()
			checkable, ok := entity.(consistencyCheckable)
			if !ok {
				t.Errorf("Expected %+v (of type %T) to be checkable for consistency",
					entity, entity)
			} else {
				checkable.checkForConsistency(t)
			}
		}
	}
}

// getCount get the count of database entries matching bean
func getCount(t *testing.T, e Engine, bean interface{}) int64 {
	count, err := e.Count(bean)
	assert.NoError(t, err)
	return count
}

// assertCount test the count of database entries matching bean
func assertCount(t *testing.T, bean interface{}, expected int) {
	assert.EqualValues(t, expected, getCount(t, x, bean),
		"Failed consistency test, the counted bean (of type %T) was %+v", bean, bean)
}

func (user *User) checkForConsistency(t *testing.T) {
	assertCount(t, &OrgUser{OrgID: user.ID}, user.NumMembers)
	assertCount(t, &Team{OrgID: user.ID}, user.NumTeams)
	assertCount(t, &Follow{UserID: user.ID}, user.NumFollowing)
	assertCount(t, &Follow{FollowID: user.ID}, user.NumFollowers)
	if user.Type != UserTypeOrganization {
		assert.EqualValues(t, 0, user.NumMembers)
		assert.EqualValues(t, 0, user.NumTeams)
	}
}

func (team *Team) checkForConsistency(t *testing.T) {
	assertCount(t, &TeamUser{TeamID: team.ID}, team.NumMembers)
}

// CountOrphanedObjects count subjects with have no existing refobject anymore
func CountOrphanedObjects(subject, refobject, joinCond string) (int64, error) {
	return x.Table("`"+subject+"`").
		Join("LEFT", refobject, joinCond).
		Where(builder.IsNull{"`" + refobject + "`.id"}).
		Count("id")
}

// DeleteOrphanedObjects delete subjects with have no existing refobject anymore
func DeleteOrphanedObjects(subject, refobject, joinCond string) error {
	subQuery := builder.Select("`"+subject+"`.id").
		From("`"+subject+"`").
		Join("LEFT", "`"+refobject+"`", joinCond).
		Where(builder.IsNull{"`" + refobject + "`.id"})
	sql, args, err := builder.Delete(builder.In("id", subQuery)).From("`" + subject + "`").ToSQL()
	if err != nil {
		return err
	}
	_, err = x.Exec(append([]interface{}{sql}, args...)...)
	return err
}

// CountWrongUserType count OrgUser who have wrong type
func CountWrongUserType() (int64, error) {
	return x.Where(builder.Eq{"type": 0}.And(builder.Neq{"num_teams": 0})).Count(new(User))
}

// FixWrongUserType fix OrgUser who have wrong type
func FixWrongUserType() (int64, error) {
	return x.Where(builder.Eq{"type": 0}.And(builder.Neq{"num_teams": 0})).Cols("type").NoAutoTime().Update(&User{Type: 1})
}

// CountBadSequences looks for broken sequences from recreate-table mistakes
func CountBadSequences() (int64, error) {
	if !setting.Database.UsePostgreSQL {
		return 0, nil
	}

	sess := x.NewSession()
	defer sess.Close()

	var sequences []string
	schema := sess.Engine().Dialect().URI().Schema

	sess.Engine().SetSchema("")
	if err := sess.Table("information_schema.sequences").Cols("sequence_name").Where("sequence_name LIKE 'tmp_recreate__%_id_seq%' AND sequence_catalog = ?", setting.Database.Name).Find(&sequences); err != nil {
		return 0, err
	}
	sess.Engine().SetSchema(schema)

	return int64(len(sequences)), nil
}

// FixBadSequences fixes for broken sequences from recreate-table mistakes
func FixBadSequences() error {
	if !setting.Database.UsePostgreSQL {
		return nil
	}

	sess := x.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	var sequences []string
	schema := sess.Engine().Dialect().URI().Schema

	sess.Engine().SetSchema("")
	if err := sess.Table("information_schema.sequences").Cols("sequence_name").Where("sequence_name LIKE 'tmp_recreate__%_id_seq%' AND sequence_catalog = ?", setting.Database.Name).Find(&sequences); err != nil {
		return err
	}
	sess.Engine().SetSchema(schema)

	sequenceRegexp := regexp.MustCompile(`tmp_recreate__(\w+)_id_seq.*`)

	for _, sequence := range sequences {
		tableName := sequenceRegexp.FindStringSubmatch(sequence)[1]
		newSequenceName := tableName + "_id_seq"
		if _, err := sess.Exec(fmt.Sprintf("ALTER SEQUENCE `%s` RENAME TO `%s`", sequence, newSequenceName)); err != nil {
			return err
		}
		if _, err := sess.Exec(fmt.Sprintf("SELECT setval('%s', COALESCE((SELECT MAX(id)+1 FROM `%s`), 1), false)", newSequenceName, tableName)); err != nil {
			return err
		}
	}

	return sess.Commit()
}
