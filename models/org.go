// Copyright 2014 The Gogs Authors. All rights reserved.
// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	"fmt"
	"strings"

	"go.wandrs.dev/framework/modules/log"
	"go.wandrs.dev/framework/modules/setting"
	"go.wandrs.dev/framework/modules/storage"
	"go.wandrs.dev/framework/modules/structs"

	"xorm.io/builder"
	"xorm.io/xorm"
)

// IsOwnedBy returns true if given user is in the owner team.
func (org *User) IsOwnedBy(uid int64) (bool, error) {
	return IsOrganizationOwner(org.ID, uid)
}

// IsOrgMember returns true if given user is member of organization.
func (org *User) IsOrgMember(uid int64) (bool, error) {
	return IsOrganizationMember(org.ID, uid)
}

func (org *User) getTeam(e Engine, name string) (*Team, error) {
	return getTeam(e, org.ID, name)
}

// GetTeam returns named team of organization.
func (org *User) GetTeam(name string) (*Team, error) {
	return org.getTeam(x, name)
}

func (org *User) getOwnerTeam(e Engine) (*Team, error) {
	return org.getTeam(e, ownerTeamName)
}

// GetOwnerTeam returns owner team of organization.
func (org *User) GetOwnerTeam() (*Team, error) {
	return org.getOwnerTeam(x)
}

func (org *User) getTeams(e Engine) error {
	if org.Teams != nil {
		return nil
	}
	return e.
		Where("org_id=?", org.ID).
		OrderBy("CASE WHEN name LIKE '" + ownerTeamName + "' THEN '' ELSE name END").
		Find(&org.Teams)
}

// GetTeams returns paginated teams that belong to organization.
func (org *User) GetTeams(opts *SearchTeamOptions) error {
	if opts.Page != 0 {
		return org.getTeams(opts.getPaginatedSession())
	}

	return org.getTeams(x)
}

// GetMembers returns all members of organization.
func (org *User) GetMembers() (err error) {
	org.Members, org.MembersIsPublic, err = FindOrgMembers(&FindOrgMembersOpts{
		OrgID: org.ID,
	})
	return
}

// FindOrgMembersOpts represensts find org members condtions
type FindOrgMembersOpts struct {
	ListOptions
	OrgID      int64
	PublicOnly bool
}

// CountOrgMembers counts the organization's members
func CountOrgMembers(opts FindOrgMembersOpts) (int64, error) {
	sess := x.Where("org_id=?", opts.OrgID)
	if opts.PublicOnly {
		sess.And("is_public = ?", true)
	}
	return sess.Count(new(OrgUser))
}

// FindOrgMembers loads organization members according conditions
func FindOrgMembers(opts *FindOrgMembersOpts) (UserList, map[int64]bool, error) {
	ous, err := GetOrgUsersByOrgID(opts)
	if err != nil {
		return nil, nil, err
	}

	ids := make([]int64, len(ous))
	idsIsPublic := make(map[int64]bool, len(ous))
	for i, ou := range ous {
		ids[i] = ou.UID
		idsIsPublic[ou.UID] = ou.IsPublic
	}

	users, err := GetUsersByIDs(ids)
	if err != nil {
		return nil, nil, err
	}
	return users, idsIsPublic, nil
}

// AddMember adds new member to organization.
func (org *User) AddMember(uid int64) error {
	return AddOrgUser(org.ID, uid)
}

// RemoveMember removes member from organization.
func (org *User) RemoveMember(uid int64) error {
	return RemoveOrgUser(org.ID, uid)
}

// CreateOrganization creates record of a new organization.
func CreateOrganization(org, owner *User) (err error) {
	if !owner.CanCreateOrganization() {
		return ErrUserNotAllowedCreateOrg{}
	}

	if err = IsUsableUsername(org.Name); err != nil {
		return err
	}

	isExist, err := IsUserExist(0, org.Name)
	if err != nil {
		return err
	} else if isExist {
		return ErrUserAlreadyExist{org.Name}
	}

	org.LowerName = strings.ToLower(org.Name)
	if org.Rands, err = GetUserSalt(); err != nil {
		return err
	}
	if org.Salt, err = GetUserSalt(); err != nil {
		return err
	}
	org.UseCustomAvatar = true
	org.NumTeams = 1
	org.NumMembers = 1
	org.Type = UserTypeOrganization

	sess := x.NewSession()
	defer sess.Close()
	if err = sess.Begin(); err != nil {
		return err
	}

	if err = deleteUserRedirect(sess, org.Name); err != nil {
		return err
	}

	if _, err = sess.Insert(org); err != nil {
		return fmt.Errorf("insert organization: %v", err)
	}
	if err = org.generateRandomAvatar(sess); err != nil {
		return fmt.Errorf("generate random avatar: %v", err)
	}

	// Add initial creator to organization and owner team.
	if _, err = sess.Insert(&OrgUser{
		UID:   owner.ID,
		OrgID: org.ID,
	}); err != nil {
		return fmt.Errorf("insert org-user relation: %v", err)
	}

	// Create default owner team.
	t := &Team{
		OrgID:      org.ID,
		LowerName:  strings.ToLower(ownerTeamName),
		Name:       ownerTeamName,
		Authorize:  AccessModeOwner,
		NumMembers: 1,
	}
	if _, err = sess.Insert(t); err != nil {
		return fmt.Errorf("insert owner team: %v", err)
	}

	if _, err = sess.Insert(&TeamUser{
		UID:    owner.ID,
		OrgID:  org.ID,
		TeamID: t.ID,
	}); err != nil {
		return fmt.Errorf("insert team-user relation: %v", err)
	}

	return sess.Commit()
}

// GetOrgByName returns organization by given name.
func GetOrgByName(name string) (*User, error) {
	if len(name) == 0 {
		return nil, ErrOrgNotExist{0, name}
	}
	u := &User{
		LowerName: strings.ToLower(name),
		Type:      UserTypeOrganization,
	}
	has, err := x.Get(u)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrOrgNotExist{0, name}
	}
	return u, nil
}

// CountOrganizations returns number of organizations.
func CountOrganizations() int64 {
	count, _ := x.
		Where("type=1").
		Count(new(User))
	return count
}

// DeleteOrganization completely and permanently deletes everything of organization.
func DeleteOrganization(org *User) (err error) {
	if !org.IsOrganization() {
		return fmt.Errorf("%s is a user not an organization", org.Name)
	}

	sess := x.NewSession()
	defer sess.Close()

	if err = sess.Begin(); err != nil {
		return err
	}

	if err = deleteOrg(sess, org); err != nil {
		return fmt.Errorf("deleteOrg: %v", err)
	}

	return sess.Commit()
}

func deleteOrg(e *xorm.Session, u *User) error {
	if err := deleteBeans(e,
		&Team{OrgID: u.ID},
		&OrgUser{OrgID: u.ID},
		&TeamUser{OrgID: u.ID},
	); err != nil {
		return fmt.Errorf("deleteBeans: %v", err)
	}

	if _, err := e.ID(u.ID).Delete(new(User)); err != nil {
		return fmt.Errorf("Delete: %v", err)
	}

	if len(u.Avatar) > 0 {
		avatarPath := u.CustomAvatarRelativePath()
		if err := storage.Avatars.Delete(avatarPath); err != nil {
			return fmt.Errorf("Failed to remove %s: %v", avatarPath, err)
		}
	}

	return nil
}

// ________                ____ ___
// \_____  \_______  ____ |    |   \______ ___________
//  /   |   \_  __ \/ ___\|    |   /  ___// __ \_  __ \
// /    |    \  | \/ /_/  >    |  /\___ \\  ___/|  | \/
// \_______  /__|  \___  /|______//____  >\___  >__|
//         \/     /_____/              \/     \/

// OrgUser represents an organization-user relation.
type OrgUser struct {
	ID       int64 `xorm:"pk autoincr"`
	UID      int64 `xorm:"INDEX UNIQUE(s)"`
	OrgID    int64 `xorm:"INDEX UNIQUE(s)"`
	IsPublic bool  `xorm:"INDEX"`
}

func isOrganizationOwner(e Engine, orgID, uid int64) (bool, error) {
	ownerTeam, err := getOwnerTeam(e, orgID)
	if err != nil {
		if IsErrTeamNotExist(err) {
			log.Error("Organization does not have owner team: %d", orgID)
			return false, nil
		}
		return false, err
	}
	return isTeamMember(e, orgID, ownerTeam.ID, uid)
}

// IsOrganizationOwner returns true if given user is in the owner team.
func IsOrganizationOwner(orgID, uid int64) (bool, error) {
	return isOrganizationOwner(x, orgID, uid)
}

// IsOrganizationMember returns true if given user is member of organization.
func IsOrganizationMember(orgID, uid int64) (bool, error) {
	return isOrganizationMember(x, orgID, uid)
}

func isOrganizationMember(e Engine, orgID, uid int64) (bool, error) {
	return e.
		Where("uid=?", uid).
		And("org_id=?", orgID).
		Table("org_user").
		Exist()
}

// IsPublicMembership returns true if given user public his/her membership.
func IsPublicMembership(orgID, uid int64) (bool, error) {
	return x.
		Where("uid=?", uid).
		And("org_id=?", orgID).
		And("is_public=?", true).
		Table("org_user").
		Exist()
}

func getOrgsByUserID(sess *xorm.Session, userID int64, showAll bool) ([]*User, error) {
	orgs := make([]*User, 0, 10)
	if !showAll {
		sess.And("`org_user`.is_public=?", true)
	}
	return orgs, sess.
		And("`org_user`.uid=?", userID).
		Join("INNER", "`org_user`", "`org_user`.org_id=`user`.id").
		Asc("`user`.name").
		Find(&orgs)
}

// GetOrgsByUserID returns a list of organizations that the given user ID
// has joined.
func GetOrgsByUserID(userID int64, showAll bool) ([]*User, error) {
	sess := x.NewSession()
	defer sess.Close()
	return getOrgsByUserID(sess, userID, showAll)
}

func getOwnedOrgsByUserID(sess *xorm.Session, userID int64) ([]*User, error) {
	orgs := make([]*User, 0, 10)
	return orgs, sess.
		Join("INNER", "`team_user`", "`team_user`.org_id=`user`.id").
		Join("INNER", "`team`", "`team`.id=`team_user`.team_id").
		Where("`team_user`.uid=?", userID).
		And("`team`.authorize=?", AccessModeOwner).
		Asc("`user`.name").
		Find(&orgs)
}

// HasOrgVisible tells if the given user can see the given org
func HasOrgVisible(org, user *User) bool {
	return hasOrgVisible(x, org, user)
}

func hasOrgVisible(e Engine, org, user *User) bool {
	// Not SignedUser
	if user == nil {
		return org.Visibility == structs.VisibleTypePublic
	}

	if user.IsAdmin {
		return true
	}

	if (org.Visibility == structs.VisibleTypePrivate || user.IsRestricted) && !org.hasMemberWithUserID(e, user.ID) {
		return false
	}
	return true
}

// HasOrgsVisible tells if the given user can see at least one of the orgs provided
func HasOrgsVisible(orgs []*User, user *User) bool {
	if len(orgs) == 0 {
		return false
	}

	for _, org := range orgs {
		if HasOrgVisible(org, user) {
			return true
		}
	}
	return false
}

// GetOwnedOrgsByUserID returns a list of organizations are owned by given user ID.
func GetOwnedOrgsByUserID(userID int64) ([]*User, error) {
	sess := x.NewSession()
	defer sess.Close()
	return getOwnedOrgsByUserID(sess, userID)
}

// GetOwnedOrgsByUserIDDesc returns a list of organizations are owned by
// given user ID, ordered descending by the given condition.
func GetOwnedOrgsByUserIDDesc(userID int64, desc string) ([]*User, error) {
	return getOwnedOrgsByUserID(x.Desc(desc), userID)
}

// GetOrgsCanCreateRepoByUserID returns a list of organizations where given user ID
// are allowed to create repos.
func GetOrgsCanCreateRepoByUserID(userID int64) ([]*User, error) {
	orgs := make([]*User, 0, 10)

	return orgs, x.Where(builder.In("id", builder.Select("`user`.id").From("`user`").
		Join("INNER", "`team_user`", "`team_user`.org_id = `user`.id").
		Join("INNER", "`team`", "`team`.id = `team_user`.team_id").
		Where(builder.Eq{"`team_user`.uid": userID}).
		And(builder.Eq{"`team`.authorize": AccessModeOwner}.Or(builder.Eq{"`team`.can_create_org_repo": true})))).
		Asc("`user`.name").
		Find(&orgs)
}

// GetOrgUsersByUserID returns all organization-user relations by user ID.
func GetOrgUsersByUserID(uid int64, opts *SearchOrganizationsOptions) ([]*OrgUser, error) {
	ous := make([]*OrgUser, 0, 10)
	sess := x.
		Join("LEFT", "`user`", "`org_user`.org_id=`user`.id").
		Where("`org_user`.uid=?", uid)
	if !opts.All {
		// Only show public organizations
		sess.And("is_public=?", true)
	}

	if opts.PageSize != 0 {
		sess = opts.setSessionPagination(sess)
	}

	err := sess.
		Asc("`user`.name").
		Find(&ous)
	return ous, err
}

// GetOrgUsersByOrgID returns all organization-user relations by organization ID.
func GetOrgUsersByOrgID(opts *FindOrgMembersOpts) ([]*OrgUser, error) {
	return getOrgUsersByOrgID(x, opts)
}

func getOrgUsersByOrgID(e Engine, opts *FindOrgMembersOpts) ([]*OrgUser, error) {
	sess := e.Where("org_id=?", opts.OrgID)
	if opts.PublicOnly {
		sess.And("is_public = ?", true)
	}
	if opts.ListOptions.PageSize > 0 {
		sess = opts.setSessionPagination(sess)

		ous := make([]*OrgUser, 0, opts.PageSize)
		return ous, sess.Find(&ous)
	}

	var ous []*OrgUser
	return ous, sess.Find(&ous)
}

// ChangeOrgUserStatus changes public or private membership status.
func ChangeOrgUserStatus(orgID, uid int64, public bool) error {
	ou := new(OrgUser)
	has, err := x.
		Where("uid=?", uid).
		And("org_id=?", orgID).
		Get(ou)
	if err != nil {
		return err
	} else if !has {
		return nil
	}

	ou.IsPublic = public
	_, err = x.ID(ou.ID).Cols("is_public").Update(ou)
	return err
}

// AddOrgUser adds new user to given organization.
func AddOrgUser(orgID, uid int64) error {
	isAlreadyMember, err := IsOrganizationMember(orgID, uid)
	if err != nil || isAlreadyMember {
		return err
	}

	sess := x.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	ou := &OrgUser{
		UID:      uid,
		OrgID:    orgID,
		IsPublic: setting.Service.DefaultOrgMemberVisible,
	}

	if _, err := sess.Insert(ou); err != nil {
		if err := sess.Rollback(); err != nil {
			log.Error("AddOrgUser: sess.Rollback: %v", err)
		}
		return err
	} else if _, err = sess.Exec("UPDATE `user` SET num_members = num_members + 1 WHERE id = ?", orgID); err != nil {
		if err := sess.Rollback(); err != nil {
			log.Error("AddOrgUser: sess.Rollback: %v", err)
		}
		return err
	}

	return sess.Commit()
}

func removeOrgUser(sess *xorm.Session, orgID, userID int64) error {
	ou := new(OrgUser)

	has, err := sess.
		Where("uid=?", userID).
		And("org_id=?", orgID).
		Get(ou)
	if err != nil {
		return fmt.Errorf("get org-user: %v", err)
	} else if !has {
		return nil
	}

	org, err := getUserByID(sess, orgID)
	if err != nil {
		return fmt.Errorf("GetUserByID [%d]: %v", orgID, err)
	}

	// Check if the user to delete is the last member in owner team.
	if isOwner, err := isOrganizationOwner(sess, orgID, userID); err != nil {
		return err
	} else if isOwner {
		t, err := org.getOwnerTeam(sess)
		if err != nil {
			return err
		}
		if t.NumMembers == 1 {
			if err := t.getMembers(sess); err != nil {
				return err
			}
			if t.Members[0].ID == userID {
				return ErrLastOrgOwner{UID: userID}
			}
		}
	}

	if _, err := sess.ID(ou.ID).Delete(ou); err != nil {
		return err
	} else if _, err = sess.Exec("UPDATE `user` SET num_members=num_members-1 WHERE id=?", orgID); err != nil {
		return err
	}

	// Delete member in his/her teams.
	teams, err := getUserOrgTeams(sess, org.ID, userID)
	if err != nil {
		return err
	}
	for _, t := range teams {
		if err = removeTeamMember(sess, t, userID); err != nil {
			return err
		}
	}

	return nil
}

// RemoveOrgUser removes user from given organization.
func RemoveOrgUser(orgID, userID int64) error {
	sess := x.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}
	if err := removeOrgUser(sess, orgID, userID); err != nil {
		return err
	}
	return sess.Commit()
}

func (org *User) getUserTeams(e Engine, userID int64, cols ...string) ([]*Team, error) {
	teams := make([]*Team, 0, org.NumTeams)
	return teams, e.
		Where("`team_user`.org_id = ?", org.ID).
		Join("INNER", "team_user", "`team_user`.team_id = team.id").
		Join("INNER", "`user`", "`user`.id=team_user.uid").
		And("`team_user`.uid = ?", userID).
		Asc("`user`.name").
		Cols(cols...).
		Find(&teams)
}

func (org *User) getUserTeamIDs(e Engine, userID int64) ([]int64, error) {
	teamIDs := make([]int64, 0, org.NumTeams)
	return teamIDs, e.
		Table("team").
		Cols("team.id").
		Where("`team_user`.org_id = ?", org.ID).
		Join("INNER", "team_user", "`team_user`.team_id = team.id").
		And("`team_user`.uid = ?", userID).
		Find(&teamIDs)
}

// GetUserTeamIDs returns of all team IDs of the organization that user is member of.
func (org *User) GetUserTeamIDs(userID int64) ([]int64, error) {
	return org.getUserTeamIDs(x, userID)
}

// GetUserTeams returns all teams that belong to user,
// and that the user has joined.
func (org *User) GetUserTeams(userID int64) ([]*Team, error) {
	return org.getUserTeams(x, userID)
}
