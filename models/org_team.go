// Copyright 2018 The Gitea Authors. All rights reserved.
// Copyright 2016 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"go.wandrs.dev/framework/modules/log"

	"xorm.io/builder"
	"xorm.io/xorm"
)

const ownerTeamName = "Owners"

// Team represents a organization team.
type Team struct {
	ID          int64 `xorm:"pk autoincr"`
	OrgID       int64 `xorm:"INDEX"`
	LowerName   string
	Name        string
	Description string
	Authorize   AccessMode
	Members     []*User `xorm:"-"`
	NumMembers  int
}

// SearchTeamOptions holds the search options
type SearchTeamOptions struct {
	ListOptions
	UserID      int64
	Keyword     string
	OrgID       int64
	IncludeDesc bool
}

// SearchMembersOptions holds the search options
type SearchMembersOptions struct {
	ListOptions
}

// SearchTeam search for teams. Caller is responsible to check permissions.
func SearchTeam(opts *SearchTeamOptions) ([]*Team, int64, error) {
	if opts.Page <= 0 {
		opts.Page = 1
	}
	if opts.PageSize == 0 {
		// Default limit
		opts.PageSize = 10
	}

	cond := builder.NewCond()

	if len(opts.Keyword) > 0 {
		lowerKeyword := strings.ToLower(opts.Keyword)
		var keywordCond builder.Cond = builder.Like{"lower_name", lowerKeyword}
		if opts.IncludeDesc {
			keywordCond = keywordCond.Or(builder.Like{"LOWER(description)", lowerKeyword})
		}
		cond = cond.And(keywordCond)
	}

	cond = cond.And(builder.Eq{"org_id": opts.OrgID})

	sess := x.NewSession()
	defer sess.Close()

	count, err := sess.
		Where(cond).
		Count(new(Team))
	if err != nil {
		return nil, 0, err
	}

	sess = sess.Where(cond)
	if opts.PageSize == -1 {
		opts.PageSize = int(count)
	} else {
		sess = sess.Limit(opts.PageSize, (opts.Page-1)*opts.PageSize)
	}

	teams := make([]*Team, 0, opts.PageSize)
	if err = sess.
		OrderBy("lower_name").
		Find(&teams); err != nil {
		return nil, 0, err
	}

	return teams, count, nil
}

// ColorFormat provides a basic color format for a Team
func (t *Team) ColorFormat(s fmt.State) {
	log.ColorFprintf(s, "%d:%s (OrgID: %d) %-v",
		log.NewColoredIDValue(t.ID),
		t.Name,
		log.NewColoredIDValue(t.OrgID),
		t.Authorize)
}

// HasWriteAccess returns true if team has at least write level access mode.
func (t *Team) HasWriteAccess() bool {
	return t.Authorize >= AccessModeWrite
}

// IsOwnerTeam returns true if team is owner team.
func (t *Team) IsOwnerTeam() bool {
	return t.Name == ownerTeamName
}

// IsMember returns true if given user is a member of team.
func (t *Team) IsMember(userID int64) bool {
	isMember, err := IsTeamMember(t.OrgID, t.ID, userID)
	if err != nil {
		log.Error("IsMember: %v", err)
		return false
	}
	return isMember
}

func (t *Team) getMembers(e Engine) (err error) {
	t.Members, err = getTeamMembers(e, t.ID)
	return err
}

// GetMembers returns paginated members in team of organization.
func (t *Team) GetMembers(opts *SearchMembersOptions) (err error) {
	if opts.Page == 0 {
		return t.getMembers(x)
	}

	return t.getMembers(opts.getPaginatedSession())
}

// AddMember adds new membership of the team to the organization,
// the user will have membership to the organization automatically when needed.
func (t *Team) AddMember(userID int64) error {
	return AddTeamMember(t, userID)
}

// RemoveMember removes member from team of organization.
func (t *Team) RemoveMember(userID int64) error {
	return RemoveTeamMember(t, userID)
}

// IsUsableTeamName tests if a name could be as team name
func IsUsableTeamName(name string) error {
	switch name {
	case "new":
		return ErrNameReserved{name}
	default:
		return nil
	}
}

// NewTeam creates a record of new team.
// It's caller's responsibility to assign organization ID.
func NewTeam(t *Team) (err error) {
	if len(t.Name) == 0 {
		return errors.New("empty team name")
	}

	if err = IsUsableTeamName(t.Name); err != nil {
		return err
	}

	has, err := x.ID(t.OrgID).Get(new(User))
	if err != nil {
		return err
	}
	if !has {
		return ErrOrgNotExist{t.OrgID, ""}
	}

	t.LowerName = strings.ToLower(t.Name)
	has, err = x.
		Where("org_id=?", t.OrgID).
		And("lower_name=?", t.LowerName).
		Get(new(Team))
	if err != nil {
		return err
	}
	if has {
		return ErrTeamAlreadyExist{t.OrgID, t.LowerName}
	}

	sess := x.NewSession()
	defer sess.Close()
	if err = sess.Begin(); err != nil {
		return err
	}

	if _, err = sess.Insert(t); err != nil {
		errRollback := sess.Rollback()
		if errRollback != nil {
			log.Error("NewTeam sess.Rollback: %v", errRollback)
		}
		return err
	}

	// Update organization number of teams.
	if _, err = sess.Exec("UPDATE `user` SET num_teams=num_teams+1 WHERE id = ?", t.OrgID); err != nil {
		errRollback := sess.Rollback()
		if errRollback != nil {
			log.Error("NewTeam sess.Rollback: %v", errRollback)
		}
		return err
	}
	return sess.Commit()
}

func getTeam(e Engine, orgID int64, name string) (*Team, error) {
	t := &Team{
		OrgID:     orgID,
		LowerName: strings.ToLower(name),
	}
	has, err := e.Get(t)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrTeamNotExist{orgID, 0, name}
	}
	return t, nil
}

// GetTeam returns team by given team name and organization.
func GetTeam(orgID int64, name string) (*Team, error) {
	return getTeam(x, orgID, name)
}

// GetTeamIDsByNames returns a slice of team ids corresponds to names.
func GetTeamIDsByNames(orgID int64, names []string, ignoreNonExistent bool) ([]int64, error) {
	ids := make([]int64, 0, len(names))
	for _, name := range names {
		u, err := GetTeam(orgID, name)
		if err != nil {
			if ignoreNonExistent {
				continue
			} else {
				return nil, err
			}
		}
		ids = append(ids, u.ID)
	}
	return ids, nil
}

// getOwnerTeam returns team by given team name and organization.
func getOwnerTeam(e Engine, orgID int64) (*Team, error) {
	return getTeam(e, orgID, ownerTeamName)
}

func getTeamByID(e Engine, teamID int64) (*Team, error) {
	t := new(Team)
	has, err := e.ID(teamID).Get(t)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrTeamNotExist{0, teamID, ""}
	}
	return t, nil
}

// GetTeamByID returns team by given ID.
func GetTeamByID(teamID int64) (*Team, error) {
	return getTeamByID(x, teamID)
}

// GetTeamNamesByID returns team's lower name from a list of team ids.
func GetTeamNamesByID(teamIDs []int64) ([]string, error) {
	if len(teamIDs) == 0 {
		return []string{}, nil
	}

	var teamNames []string
	err := x.Table("team").
		Select("lower_name").
		In("id", teamIDs).
		Asc("name").
		Find(&teamNames)

	return teamNames, err
}

// UpdateTeam updates information of team.
func UpdateTeam(t *Team, authChanged, includeAllChanged bool) (err error) {
	if len(t.Name) == 0 {
		return errors.New("empty team name")
	}

	if len(t.Description) > 255 {
		t.Description = t.Description[:255]
	}

	sess := x.NewSession()
	defer sess.Close()
	if err = sess.Begin(); err != nil {
		return err
	}

	t.LowerName = strings.ToLower(t.Name)
	has, err := sess.
		Where("org_id=?", t.OrgID).
		And("lower_name=?", t.LowerName).
		And("id!=?", t.ID).
		Get(new(Team))
	if err != nil {
		return err
	} else if has {
		return ErrTeamAlreadyExist{t.OrgID, t.LowerName}
	}

	if _, err = sess.ID(t.ID).Cols("name", "lower_name", "description",
		"can_create_org_repo", "authorize", "includes_all_repositories").Update(t); err != nil {
		return fmt.Errorf("update: %v", err)
	}

	return sess.Commit()
}

// DeleteTeam deletes given team.
// It's caller's responsibility to assign organization ID.
func DeleteTeam(t *Team) error {
	sess := x.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	if err := t.getMembers(sess); err != nil {
		return err
	}

	// Delete team-user.
	if _, err := sess.
		Where("org_id=?", t.OrgID).
		Where("team_id=?", t.ID).
		Delete(new(TeamUser)); err != nil {
		return err
	}

	// Delete team.
	if _, err := sess.ID(t.ID).Delete(new(Team)); err != nil {
		return err
	}
	// Update organization number of teams.
	if _, err := sess.Exec("UPDATE `user` SET num_teams=num_teams-1 WHERE id=?", t.OrgID); err != nil {
		return err
	}

	return sess.Commit()
}

// ___________                    ____ ___
// \__    ___/___ _____    _____ |    |   \______ ___________
//   |    |_/ __ \\__  \  /     \|    |   /  ___// __ \_  __ \
//   |    |\  ___/ / __ \|  Y Y  \    |  /\___ \\  ___/|  | \/
//   |____| \___  >____  /__|_|  /______//____  >\___  >__|
//              \/     \/      \/             \/     \/

// TeamUser represents an team-user relation.
type TeamUser struct {
	ID     int64 `xorm:"pk autoincr"`
	OrgID  int64 `xorm:"INDEX"`
	TeamID int64 `xorm:"UNIQUE(s)"`
	UID    int64 `xorm:"UNIQUE(s)"`
}

func isTeamMember(e Engine, orgID, teamID, userID int64) (bool, error) {
	return e.
		Where("org_id=?", orgID).
		And("team_id=?", teamID).
		And("uid=?", userID).
		Table("team_user").
		Exist()
}

// IsTeamMember returns true if given user is a member of team.
func IsTeamMember(orgID, teamID, userID int64) (bool, error) {
	return isTeamMember(x, orgID, teamID, userID)
}

func getTeamUsersByTeamID(e Engine, teamID int64) ([]*TeamUser, error) {
	teamUsers := make([]*TeamUser, 0, 10)
	return teamUsers, e.
		Where("team_id=?", teamID).
		Find(&teamUsers)
}

func getTeamMembers(e Engine, teamID int64) (_ []*User, err error) {
	teamUsers, err := getTeamUsersByTeamID(e, teamID)
	if err != nil {
		return nil, fmt.Errorf("get team-users: %v", err)
	}
	members := make([]*User, len(teamUsers))
	for i, teamUser := range teamUsers {
		member, err := getUserByID(e, teamUser.UID)
		if err != nil {
			return nil, fmt.Errorf("get user '%d': %v", teamUser.UID, err)
		}
		members[i] = member
	}
	sort.Slice(members, func(i, j int) bool {
		return members[i].DisplayName() < members[j].DisplayName()
	})
	return members, nil
}

// GetTeamMembers returns all members in given team of organization.
func GetTeamMembers(teamID int64) ([]*User, error) {
	return getTeamMembers(x, teamID)
}

func getUserTeams(e Engine, userID int64, listOptions ListOptions) (teams []*Team, err error) {
	sess := e.
		Join("INNER", "team_user", "team_user.team_id = team.id").
		Where("team_user.uid=?", userID)
	if listOptions.Page != 0 {
		sess = listOptions.setSessionPagination(sess)
	}
	return teams, sess.Find(&teams)
}

func getUserOrgTeams(e Engine, orgID, userID int64) (teams []*Team, err error) {
	return teams, e.
		Join("INNER", "team_user", "team_user.team_id = team.id").
		Where("team.org_id = ?", orgID).
		And("team_user.uid=?", userID).
		Find(&teams)
}

func getUserRepoTeams(e Engine, orgID, userID, repoID int64) (teams []*Team, err error) {
	return teams, e.
		Join("INNER", "team_user", "team_user.team_id = team.id").
		Join("INNER", "team_repo", "team_repo.team_id = team.id").
		Where("team.org_id = ?", orgID).
		And("team_user.uid=?", userID).
		And("team_repo.repo_id=?", repoID).
		Find(&teams)
}

// GetUserOrgTeams returns all teams that user belongs to in given organization.
func GetUserOrgTeams(orgID, userID int64) ([]*Team, error) {
	return getUserOrgTeams(x, orgID, userID)
}

// GetUserTeams returns all teams that user belongs across all organizations.
func GetUserTeams(userID int64, listOptions ListOptions) ([]*Team, error) {
	return getUserTeams(x, userID, listOptions)
}

// AddTeamMember adds new membership of given team to given organization,
// the user will have membership to given organization automatically when needed.
func AddTeamMember(team *Team, userID int64) error {
	isAlreadyMember, err := IsTeamMember(team.OrgID, team.ID, userID)
	if err != nil || isAlreadyMember {
		return err
	}

	if err := AddOrgUser(team.OrgID, userID); err != nil {
		return err
	}

	sess := x.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	if _, err := sess.Insert(&TeamUser{
		UID:    userID,
		OrgID:  team.OrgID,
		TeamID: team.ID,
	}); err != nil {
		return err
	} else if _, err := sess.Incr("num_members").ID(team.ID).Update(new(Team)); err != nil {
		return err
	}

	team.NumMembers++

	return sess.Commit()
}

func removeTeamMember(e *xorm.Session, team *Team, userID int64) error {
	isMember, err := isTeamMember(e, team.OrgID, team.ID, userID)
	if err != nil || !isMember {
		return err
	}

	// Check if the user to delete is the last member in owner team.
	if team.IsOwnerTeam() && team.NumMembers == 1 {
		return ErrLastOrgOwner{UID: userID}
	}

	team.NumMembers--

	if _, err := e.Delete(&TeamUser{
		UID:    userID,
		OrgID:  team.OrgID,
		TeamID: team.ID,
	}); err != nil {
		return err
	} else if _, err = e.
		ID(team.ID).
		Cols("num_members").
		Update(team); err != nil {
		return err
	}

	// Check if the user is a member of any team in the organization.
	if count, err := e.Count(&TeamUser{
		UID:   userID,
		OrgID: team.OrgID,
	}); err != nil {
		return err
	} else if count == 0 {
		return removeOrgUser(e, team.OrgID, userID)
	}

	return nil
}

// RemoveTeamMember removes member from given team of given organization.
func RemoveTeamMember(team *Team, userID int64) error {
	sess := x.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}
	if err := removeTeamMember(sess, team, userID); err != nil {
		return err
	}
	return sess.Commit()
}

// IsUserInTeams returns if a user in some teams
func IsUserInTeams(userID int64, teamIDs []int64) (bool, error) {
	return isUserInTeams(x, userID, teamIDs)
}

func isUserInTeams(e Engine, userID int64, teamIDs []int64) (bool, error) {
	return e.Where("uid=?", userID).In("team_id", teamIDs).Exist(new(TeamUser))
}

// UsersInTeamsCount counts the number of users which are in userIDs and teamIDs
func UsersInTeamsCount(userIDs, teamIDs []int64) (int64, error) {
	var ids []int64
	if err := x.In("uid", userIDs).In("team_id", teamIDs).
		Table("team_user").
		Cols("uid").GroupBy("uid").Find(&ids); err != nil {
		return 0, err
	}
	return int64(len(ids)), nil
}
