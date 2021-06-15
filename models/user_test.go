// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"code.gitea.io/gitea/modules/setting"
	"code.gitea.io/gitea/modules/util"

	"github.com/stretchr/testify/assert"
)

func TestUserIsPublicMember(t *testing.T) {
	assert.NoError(t, PrepareTestDatabase())

	tt := []struct {
		uid      int64
		orgid    int64
		expected bool
	}{
		{2, 3, true},
		{4, 3, false},
		{5, 6, true},
		{5, 7, false},
	}
	for _, v := range tt {
		t.Run(fmt.Sprintf("UserId%dIsPublicMemberOf%d", v.uid, v.orgid), func(t *testing.T) {
			testUserIsPublicMember(t, v.uid, v.orgid, v.expected)
		})
	}
}

func testUserIsPublicMember(t *testing.T, uid, orgID int64, expected bool) {
	user, err := GetUserByID(uid)
	assert.NoError(t, err)
	assert.Equal(t, expected, user.IsPublicMember(orgID))
}

func TestIsUserOrgOwner(t *testing.T) {
	assert.NoError(t, PrepareTestDatabase())

	tt := []struct {
		uid      int64
		orgid    int64
		expected bool
	}{
		{2, 3, true},
		{4, 3, false},
		{5, 6, true},
		{5, 7, true},
	}
	for _, v := range tt {
		t.Run(fmt.Sprintf("UserId%dIsOrgOwnerOf%d", v.uid, v.orgid), func(t *testing.T) {
			testIsUserOrgOwner(t, v.uid, v.orgid, v.expected)
		})
	}
}

func testIsUserOrgOwner(t *testing.T, uid, orgID int64, expected bool) {
	user, err := GetUserByID(uid)
	assert.NoError(t, err)
	assert.Equal(t, expected, user.IsUserOrgOwner(orgID))
}

func TestGetUserEmailsByNames(t *testing.T) {
	assert.NoError(t, PrepareTestDatabase())

	// ignore none active user email
	assert.Equal(t, []string{"user8@example.com"}, GetUserEmailsByNames([]string{"user8", "user9"}))
	assert.Equal(t, []string{"user8@example.com", "user5@example.com"}, GetUserEmailsByNames([]string{"user8", "user5"}))

	assert.Equal(t, []string{"user8@example.com"}, GetUserEmailsByNames([]string{"user8", "user7"}))
}

func TestCanCreateOrganization(t *testing.T) {
	assert.NoError(t, PrepareTestDatabase())

	admin := AssertExistsAndLoadBean(t, &User{ID: 1}).(*User)
	assert.True(t, admin.CanCreateOrganization())

	user := AssertExistsAndLoadBean(t, &User{ID: 2}).(*User)
	assert.True(t, user.CanCreateOrganization())
	// Disable user create organization permission.
	user.AllowCreateOrganization = false
	assert.False(t, user.CanCreateOrganization())

	setting.Admin.DisableRegularOrgCreation = true
	user.AllowCreateOrganization = true
	assert.True(t, admin.CanCreateOrganization())
	assert.False(t, user.CanCreateOrganization())
}

func TestSearchUsers(t *testing.T) {
	assert.NoError(t, PrepareTestDatabase())
	testSuccess := func(opts *SearchUserOptions, expectedUserOrOrgIDs []int64) {
		users, _, err := SearchUsers(opts)
		assert.NoError(t, err)
		if assert.Len(t, users, len(expectedUserOrOrgIDs)) {
			for i, expectedID := range expectedUserOrOrgIDs {
				assert.EqualValues(t, expectedID, users[i].ID)
			}
		}
	}

	// test orgs
	testOrgSuccess := func(opts *SearchUserOptions, expectedOrgIDs []int64) {
		opts.Type = UserTypeOrganization
		testSuccess(opts, expectedOrgIDs)
	}

	testOrgSuccess(&SearchUserOptions{OrderBy: "id ASC", ListOptions: ListOptions{Page: 1, PageSize: 2}},
		[]int64{3, 6})

	testOrgSuccess(&SearchUserOptions{OrderBy: "id ASC", ListOptions: ListOptions{Page: 2, PageSize: 2}},
		[]int64{7, 17})

	testOrgSuccess(&SearchUserOptions{OrderBy: "id ASC", ListOptions: ListOptions{Page: 3, PageSize: 2}},
		[]int64{19, 25})

	testOrgSuccess(&SearchUserOptions{OrderBy: "id ASC", ListOptions: ListOptions{Page: 4, PageSize: 2}},
		[]int64{26})

	testOrgSuccess(&SearchUserOptions{ListOptions: ListOptions{Page: 5, PageSize: 2}},
		[]int64{})

	// test users
	testUserSuccess := func(opts *SearchUserOptions, expectedUserIDs []int64) {
		opts.Type = UserTypeIndividual
		testSuccess(opts, expectedUserIDs)
	}

	testUserSuccess(&SearchUserOptions{OrderBy: "id ASC", ListOptions: ListOptions{Page: 1}},
		[]int64{1, 2, 4, 5, 8, 9, 10, 11, 12, 13, 14, 15, 16, 18, 20, 21, 24, 27, 28, 29, 30})

	testUserSuccess(&SearchUserOptions{ListOptions: ListOptions{Page: 1}, IsActive: util.OptionalBoolFalse},
		[]int64{9})

	testUserSuccess(&SearchUserOptions{OrderBy: "id ASC", ListOptions: ListOptions{Page: 1}, IsActive: util.OptionalBoolTrue},
		[]int64{1, 2, 4, 5, 8, 10, 11, 12, 13, 14, 15, 16, 18, 20, 21, 24, 28, 29, 30})

	testUserSuccess(&SearchUserOptions{Keyword: "user1", OrderBy: "id ASC", ListOptions: ListOptions{Page: 1}, IsActive: util.OptionalBoolTrue},
		[]int64{1, 10, 11, 12, 13, 14, 15, 16, 18})

	// order by name asc default
	testUserSuccess(&SearchUserOptions{Keyword: "user1", ListOptions: ListOptions{Page: 1}, IsActive: util.OptionalBoolTrue},
		[]int64{1, 10, 11, 12, 13, 14, 15, 16, 18})
}

func TestDeleteUser(t *testing.T) {
	test := func(userID int64) {
		assert.NoError(t, PrepareTestDatabase())
		user := AssertExistsAndLoadBean(t, &User{ID: userID}).(*User)

		orgUsers := make([]*OrgUser, 0, 10)
		assert.NoError(t, x.Find(&orgUsers, &OrgUser{UID: userID}))
		for _, orgUser := range orgUsers {
			if err := RemoveOrgUser(orgUser.OrgID, orgUser.UID); err != nil {
				assert.True(t, IsErrLastOrgOwner(err))
				return
			}
		}
		assert.NoError(t, DeleteUser(user))
		AssertNotExistsBean(t, &User{ID: userID})
	}
	test(2)
	test(4)
	test(8)
	test(11)

	org := AssertExistsAndLoadBean(t, &User{ID: 3}).(*User)
	assert.Error(t, DeleteUser(org))
}

func TestEmailNotificationPreferences(t *testing.T) {
	assert.NoError(t, PrepareTestDatabase())
	for _, test := range []struct {
		expected string
		userID   int64
	}{
		{EmailNotificationsEnabled, 1},
		{EmailNotificationsEnabled, 2},
		{EmailNotificationsOnMention, 3},
		{EmailNotificationsOnMention, 4},
		{EmailNotificationsEnabled, 5},
		{EmailNotificationsEnabled, 6},
		{EmailNotificationsDisabled, 7},
		{EmailNotificationsEnabled, 8},
		{EmailNotificationsOnMention, 9},
	} {
		user := AssertExistsAndLoadBean(t, &User{ID: test.userID}).(*User)
		assert.Equal(t, test.expected, user.EmailNotifications())

		// Try all possible settings
		assert.NoError(t, user.SetEmailNotifications(EmailNotificationsEnabled))
		assert.Equal(t, EmailNotificationsEnabled, user.EmailNotifications())

		assert.NoError(t, user.SetEmailNotifications(EmailNotificationsOnMention))
		assert.Equal(t, EmailNotificationsOnMention, user.EmailNotifications())

		assert.NoError(t, user.SetEmailNotifications(EmailNotificationsDisabled))
		assert.Equal(t, EmailNotificationsDisabled, user.EmailNotifications())
	}
}

func TestHashPasswordDeterministic(t *testing.T) {
	b := make([]byte, 16)
	u := &User{}
	algos := []string{"argon2", "pbkdf2", "scrypt", "bcrypt"}
	for j := 0; j < len(algos); j++ {
		u.PasswdHashAlgo = algos[j]
		for i := 0; i < 50; i++ {
			// generate a random password
			rand.Read(b)
			pass := string(b)

			// save the current password in the user - hash it and store the result
			u.SetPassword(pass)
			r1 := u.Passwd

			// run again
			u.SetPassword(pass)
			r2 := u.Passwd

			assert.NotEqual(t, r1, r2)
			assert.True(t, u.ValidatePassword(pass))
		}
	}
}

func BenchmarkHashPassword(b *testing.B) {
	// BenchmarkHashPassword ensures that it takes a reasonable amount of time
	// to hash a password - in order to protect from brute-force attacks.
	pass := "password1337"
	u := &User{Passwd: pass}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		u.SetPassword(pass)
	}
}

func TestNewGitSig(t *testing.T) {
	users := make([]*User, 0, 20)
	sess := x.NewSession()
	defer sess.Close()
	sess.Find(&users)

	for _, user := range users {
		sig := user.NewGitSig()
		assert.NotContains(t, sig.Name, "<")
		assert.NotContains(t, sig.Name, ">")
		assert.NotContains(t, sig.Name, "\n")
		assert.NotEqual(t, len(strings.TrimSpace(sig.Name)), 0)
	}
}

func TestDisplayName(t *testing.T) {
	users := make([]*User, 0, 20)
	sess := x.NewSession()
	defer sess.Close()
	sess.Find(&users)

	for _, user := range users {
		displayName := user.DisplayName()
		assert.Equal(t, strings.TrimSpace(displayName), displayName)
		if len(strings.TrimSpace(user.FullName)) == 0 {
			assert.Equal(t, user.Name, displayName)
		}
		assert.NotEqual(t, len(strings.TrimSpace(displayName)), 0)
	}
}

func TestCreateUser(t *testing.T) {
	user := &User{
		Name:               "GiteaBot",
		Email:              "GiteaBot@gitea.io",
		Passwd:             ";p['////..-++']",
		IsAdmin:            false,
		Theme:              setting.UI.DefaultTheme,
		MustChangePassword: false,
	}

	assert.NoError(t, CreateUser(user))

	assert.NoError(t, DeleteUser(user))
}

func TestCreateUserInvalidEmail(t *testing.T) {
	user := &User{
		Name:               "GiteaBot",
		Email:              "GiteaBot@gitea.io\r\n",
		Passwd:             ";p['////..-++']",
		IsAdmin:            false,
		Theme:              setting.UI.DefaultTheme,
		MustChangePassword: false,
	}

	err := CreateUser(user)
	assert.Error(t, err)
	assert.True(t, IsErrEmailInvalid(err))
}

func TestCreateUser_Issue5882(t *testing.T) {
	// Init settings
	_ = setting.Admin

	passwd := ".//.;1;;//.,-=_"

	tt := []struct {
		user               *User
		disableOrgCreation bool
	}{
		{&User{Name: "GiteaBot", Email: "GiteaBot@gitea.io", Passwd: passwd, MustChangePassword: false}, false},
		{&User{Name: "GiteaBot2", Email: "GiteaBot2@gitea.io", Passwd: passwd, MustChangePassword: false}, true},
	}

	setting.Service.DefaultAllowCreateOrganization = true

	for _, v := range tt {
		setting.Admin.DisableRegularOrgCreation = v.disableOrgCreation

		assert.NoError(t, CreateUser(v.user))

		u, err := GetUserByEmail(v.user.Email)
		assert.NoError(t, err)

		assert.Equal(t, !u.AllowCreateOrganization, v.disableOrgCreation)

		assert.NoError(t, DeleteUser(v.user))
	}
}

func TestGetUserIDsByNames(t *testing.T) {
	assert.NoError(t, PrepareTestDatabase())

	// ignore non existing
	IDs, err := GetUserIDsByNames([]string{"user1", "user2", "none_existing_user"}, true)
	assert.NoError(t, err)
	assert.Equal(t, []int64{1, 2}, IDs)

	// ignore non existing
	IDs, err = GetUserIDsByNames([]string{"user1", "do_not_exist"}, false)
	assert.Error(t, err)
	assert.Equal(t, []int64(nil), IDs)
}

func TestGetMaileableUsersByIDs(t *testing.T) {
	assert.NoError(t, PrepareTestDatabase())

	results, err := GetMaileableUsersByIDs([]int64{1, 4}, false)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	if len(results) > 1 {
		assert.Equal(t, results[0].ID, 1)
	}

	results, err = GetMaileableUsersByIDs([]int64{1, 4}, true)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	if len(results) > 2 {
		assert.Equal(t, results[0].ID, 1)
		assert.Equal(t, results[1].ID, 4)
	}
}
