// Copyright 2015 The Gogs Authors. All rights reserved.
// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	"fmt"
)

// ErrNotExist represents a non-exist error.
type ErrNotExist struct {
	ID int64
}

// IsErrNotExist checks if an error is an ErrNotExist
func IsErrNotExist(err error) bool {
	_, ok := err.(ErrNotExist)
	return ok
}

func (err ErrNotExist) Error() string {
	return fmt.Sprintf("record does not exist [id: %d]", err.ID)
}

// ErrNameReserved represents a "reserved name" error.
type ErrNameReserved struct {
	Name string
}

// IsErrNameReserved checks if an error is a ErrNameReserved.
func IsErrNameReserved(err error) bool {
	_, ok := err.(ErrNameReserved)
	return ok
}

func (err ErrNameReserved) Error() string {
	return fmt.Sprintf("name is reserved [name: %s]", err.Name)
}

// ErrNamePatternNotAllowed represents a "pattern not allowed" error.
type ErrNamePatternNotAllowed struct {
	Pattern string
}

// IsErrNamePatternNotAllowed checks if an error is an ErrNamePatternNotAllowed.
func IsErrNamePatternNotAllowed(err error) bool {
	_, ok := err.(ErrNamePatternNotAllowed)
	return ok
}

func (err ErrNamePatternNotAllowed) Error() string {
	return fmt.Sprintf("name pattern is not allowed [pattern: %s]", err.Pattern)
}

// ErrNameCharsNotAllowed represents a "character not allowed in name" error.
type ErrNameCharsNotAllowed struct {
	Name string
}

// IsErrNameCharsNotAllowed checks if an error is an ErrNameCharsNotAllowed.
func IsErrNameCharsNotAllowed(err error) bool {
	_, ok := err.(ErrNameCharsNotAllowed)
	return ok
}

func (err ErrNameCharsNotAllowed) Error() string {
	return fmt.Sprintf("User name is invalid [%s]: must be valid alpha or numeric or dash(-_) or dot characters", err.Name)
}

// ErrSSHDisabled represents an "SSH disabled" error.
type ErrSSHDisabled struct{}

// IsErrSSHDisabled checks if an error is a ErrSSHDisabled.
func IsErrSSHDisabled(err error) bool {
	_, ok := err.(ErrSSHDisabled)
	return ok
}

func (err ErrSSHDisabled) Error() string {
	return "SSH is disabled"
}

// ErrCancelled represents an error due to context cancellation
type ErrCancelled struct {
	Message string
}

// IsErrCancelled checks if an error is a ErrCancelled.
func IsErrCancelled(err error) bool {
	_, ok := err.(ErrCancelled)
	return ok
}

func (err ErrCancelled) Error() string {
	return "Cancelled: " + err.Message
}

// ErrCancelledf returns an ErrCancelled for the provided format and args
func ErrCancelledf(format string, args ...interface{}) error {
	return ErrCancelled{
		fmt.Sprintf(format, args...),
	}
}

//  ____ ___
// |    |   \______ ___________
// |    |   /  ___// __ \_  __ \
// |    |  /\___ \\  ___/|  | \/
// |______//____  >\___  >__|
//              \/     \/

// ErrUserAlreadyExist represents a "user already exists" error.
type ErrUserAlreadyExist struct {
	Name string
}

// IsErrUserAlreadyExist checks if an error is a ErrUserAlreadyExists.
func IsErrUserAlreadyExist(err error) bool {
	_, ok := err.(ErrUserAlreadyExist)
	return ok
}

func (err ErrUserAlreadyExist) Error() string {
	return fmt.Sprintf("user already exists [name: %s]", err.Name)
}

// ErrUserNotExist represents a "UserNotExist" kind of error.
type ErrUserNotExist struct {
	UID   int64
	Name  string
	KeyID int64
}

// IsErrUserNotExist checks if an error is a ErrUserNotExist.
func IsErrUserNotExist(err error) bool {
	_, ok := err.(ErrUserNotExist)
	return ok
}

func (err ErrUserNotExist) Error() string {
	return fmt.Sprintf("user does not exist [uid: %d, name: %s, keyid: %d]", err.UID, err.Name, err.KeyID)
}

// ErrUserRedirectNotExist represents a "UserRedirectNotExist" kind of error.
type ErrUserRedirectNotExist struct {
	Name string
}

// IsErrUserRedirectNotExist check if an error is an ErrUserRedirectNotExist.
func IsErrUserRedirectNotExist(err error) bool {
	_, ok := err.(ErrUserRedirectNotExist)
	return ok
}

func (err ErrUserRedirectNotExist) Error() string {
	return fmt.Sprintf("user redirect does not exist [name: %s]", err.Name)
}

// ErrUserProhibitLogin represents a "ErrUserProhibitLogin" kind of error.
type ErrUserProhibitLogin struct {
	UID  int64
	Name string
}

// IsErrUserProhibitLogin checks if an error is a ErrUserProhibitLogin
func IsErrUserProhibitLogin(err error) bool {
	_, ok := err.(ErrUserProhibitLogin)
	return ok
}

func (err ErrUserProhibitLogin) Error() string {
	return fmt.Sprintf("user is not allowed login [uid: %d, name: %s]", err.UID, err.Name)
}

// ErrUserInactive represents a "ErrUserInactive" kind of error.
type ErrUserInactive struct {
	UID  int64
	Name string
}

// IsErrUserInactive checks if an error is a ErrUserInactive
func IsErrUserInactive(err error) bool {
	_, ok := err.(ErrUserInactive)
	return ok
}

func (err ErrUserInactive) Error() string {
	return fmt.Sprintf("user is inactive [uid: %d, name: %s]", err.UID, err.Name)
}

// ErrEmailAlreadyUsed represents a "EmailAlreadyUsed" kind of error.
type ErrEmailAlreadyUsed struct {
	Email string
}

// IsErrEmailAlreadyUsed checks if an error is a ErrEmailAlreadyUsed.
func IsErrEmailAlreadyUsed(err error) bool {
	_, ok := err.(ErrEmailAlreadyUsed)
	return ok
}

func (err ErrEmailAlreadyUsed) Error() string {
	return fmt.Sprintf("e-mail already in use [email: %s]", err.Email)
}

// ErrEmailInvalid represents an error where the email address does not comply with RFC 5322
type ErrEmailInvalid struct {
	Email string
}

// IsErrEmailInvalid checks if an error is an ErrEmailInvalid
func IsErrEmailInvalid(err error) bool {
	_, ok := err.(ErrEmailInvalid)
	return ok
}

func (err ErrEmailInvalid) Error() string {
	return fmt.Sprintf("e-mail invalid [email: %s]", err.Email)
}

// ErrEmailAddressNotExist email address not exist
type ErrEmailAddressNotExist struct {
	Email string
}

// IsErrEmailAddressNotExist checks if an error is an ErrEmailAddressNotExist
func IsErrEmailAddressNotExist(err error) bool {
	_, ok := err.(ErrEmailAddressNotExist)
	return ok
}

func (err ErrEmailAddressNotExist) Error() string {
	return fmt.Sprintf("Email address does not exist [email: %s]", err.Email)
}

// ErrOpenIDAlreadyUsed represents a "OpenIDAlreadyUsed" kind of error.
type ErrOpenIDAlreadyUsed struct {
	OpenID string
}

// IsErrOpenIDAlreadyUsed checks if an error is a ErrOpenIDAlreadyUsed.
func IsErrOpenIDAlreadyUsed(err error) bool {
	_, ok := err.(ErrOpenIDAlreadyUsed)
	return ok
}

func (err ErrOpenIDAlreadyUsed) Error() string {
	return fmt.Sprintf("OpenID already in use [oid: %s]", err.OpenID)
}

// ErrUserOwnRepos represents a "UserOwnRepos" kind of error.
type ErrUserOwnRepos struct {
	UID int64
}

// IsErrUserOwnRepos checks if an error is a ErrUserOwnRepos.
func IsErrUserOwnRepos(err error) bool {
	_, ok := err.(ErrUserOwnRepos)
	return ok
}

func (err ErrUserOwnRepos) Error() string {
	return fmt.Sprintf("user still has ownership of repositories [uid: %d]", err.UID)
}

// ErrUserHasOrgs represents a "UserHasOrgs" kind of error.
type ErrUserHasOrgs struct {
	UID int64
}

// IsErrUserHasOrgs checks if an error is a ErrUserHasOrgs.
func IsErrUserHasOrgs(err error) bool {
	_, ok := err.(ErrUserHasOrgs)
	return ok
}

func (err ErrUserHasOrgs) Error() string {
	return fmt.Sprintf("user still has membership of organizations [uid: %d]", err.UID)
}

// ErrUserNotAllowedCreateOrg represents a "UserNotAllowedCreateOrg" kind of error.
type ErrUserNotAllowedCreateOrg struct{}

// IsErrUserNotAllowedCreateOrg checks if an error is an ErrUserNotAllowedCreateOrg.
func IsErrUserNotAllowedCreateOrg(err error) bool {
	_, ok := err.(ErrUserNotAllowedCreateOrg)
	return ok
}

func (err ErrUserNotAllowedCreateOrg) Error() string {
	return "user is not allowed to create organizations"
}

// ErrReachLimitOfRepo represents a "ReachLimitOfRepo" kind of error.
type ErrReachLimitOfRepo struct {
	Limit int
}

// IsErrReachLimitOfRepo checks if an error is a ErrReachLimitOfRepo.
func IsErrReachLimitOfRepo(err error) bool {
	_, ok := err.(ErrReachLimitOfRepo)
	return ok
}

func (err ErrReachLimitOfRepo) Error() string {
	return fmt.Sprintf("user has reached maximum limit of repositories [limit: %d]", err.Limit)
}

//  __      __.__ __   .__
// /  \    /  \__|  | _|__|
// \   \/\/   /  |  |/ /  |
//  \        /|  |    <|  |
//   \__/\  / |__|__|_ \__|
//        \/          \/

// ErrWikiAlreadyExist represents a "WikiAlreadyExist" kind of error.
type ErrWikiAlreadyExist struct {
	Title string
}

// IsErrWikiAlreadyExist checks if an error is an ErrWikiAlreadyExist.
func IsErrWikiAlreadyExist(err error) bool {
	_, ok := err.(ErrWikiAlreadyExist)
	return ok
}

func (err ErrWikiAlreadyExist) Error() string {
	return fmt.Sprintf("wiki page already exists [title: %s]", err.Title)
}

// ErrWikiReservedName represents a reserved name error.
type ErrWikiReservedName struct {
	Title string
}

// IsErrWikiReservedName checks if an error is an ErrWikiReservedName.
func IsErrWikiReservedName(err error) bool {
	_, ok := err.(ErrWikiReservedName)
	return ok
}

func (err ErrWikiReservedName) Error() string {
	return fmt.Sprintf("wiki title is reserved: %s", err.Title)
}

// ErrWikiInvalidFileName represents an invalid wiki file name.
type ErrWikiInvalidFileName struct {
	FileName string
}

// IsErrWikiInvalidFileName checks if an error is an ErrWikiInvalidFileName.
func IsErrWikiInvalidFileName(err error) bool {
	_, ok := err.(ErrWikiInvalidFileName)
	return ok
}

func (err ErrWikiInvalidFileName) Error() string {
	return fmt.Sprintf("Invalid wiki filename: %s", err.FileName)
}

// __________     ___.   .__  .__          ____  __.
// \______   \__ _\_ |__ |  | |__| ____   |    |/ _|____ ___.__.
//  |     ___/  |  \ __ \|  | |  |/ ___\  |      <_/ __ <   |  |
//  |    |   |  |  / \_\ \  |_|  \  \___  |    |  \  ___/\___  |
//  |____|   |____/|___  /____/__|\___  > |____|__ \___  > ____|
//                     \/             \/          \/   \/\/

// ErrKeyUnableVerify represents a "KeyUnableVerify" kind of error.
type ErrKeyUnableVerify struct {
	Result string
}

// IsErrKeyUnableVerify checks if an error is a ErrKeyUnableVerify.
func IsErrKeyUnableVerify(err error) bool {
	_, ok := err.(ErrKeyUnableVerify)
	return ok
}

func (err ErrKeyUnableVerify) Error() string {
	return fmt.Sprintf("Unable to verify key content [result: %s]", err.Result)
}

// ErrKeyNotExist represents a "KeyNotExist" kind of error.
type ErrKeyNotExist struct {
	ID int64
}

// IsErrKeyNotExist checks if an error is a ErrKeyNotExist.
func IsErrKeyNotExist(err error) bool {
	_, ok := err.(ErrKeyNotExist)
	return ok
}

func (err ErrKeyNotExist) Error() string {
	return fmt.Sprintf("public key does not exist [id: %d]", err.ID)
}

// ErrKeyAlreadyExist represents a "KeyAlreadyExist" kind of error.
type ErrKeyAlreadyExist struct {
	OwnerID     int64
	Fingerprint string
	Content     string
}

// IsErrKeyAlreadyExist checks if an error is a ErrKeyAlreadyExist.
func IsErrKeyAlreadyExist(err error) bool {
	_, ok := err.(ErrKeyAlreadyExist)
	return ok
}

func (err ErrKeyAlreadyExist) Error() string {
	return fmt.Sprintf("public key already exists [owner_id: %d, finger_print: %s, content: %s]",
		err.OwnerID, err.Fingerprint, err.Content)
}

// ErrKeyNameAlreadyUsed represents a "KeyNameAlreadyUsed" kind of error.
type ErrKeyNameAlreadyUsed struct {
	OwnerID int64
	Name    string
}

// IsErrKeyNameAlreadyUsed checks if an error is a ErrKeyNameAlreadyUsed.
func IsErrKeyNameAlreadyUsed(err error) bool {
	_, ok := err.(ErrKeyNameAlreadyUsed)
	return ok
}

func (err ErrKeyNameAlreadyUsed) Error() string {
	return fmt.Sprintf("public key already exists [owner_id: %d, name: %s]", err.OwnerID, err.Name)
}

// ErrGPGNoEmailFound represents a "ErrGPGNoEmailFound" kind of error.
type ErrGPGNoEmailFound struct {
	FailedEmails []string
}

// IsErrGPGNoEmailFound checks if an error is a ErrGPGNoEmailFound.
func IsErrGPGNoEmailFound(err error) bool {
	_, ok := err.(ErrGPGNoEmailFound)
	return ok
}

func (err ErrGPGNoEmailFound) Error() string {
	return fmt.Sprintf("none of the emails attached to the GPG key could be found: %v", err.FailedEmails)
}

// ErrGPGKeyParsing represents a "ErrGPGKeyParsing" kind of error.
type ErrGPGKeyParsing struct {
	ParseError error
}

// IsErrGPGKeyParsing checks if an error is a ErrGPGKeyParsing.
func IsErrGPGKeyParsing(err error) bool {
	_, ok := err.(ErrGPGKeyParsing)
	return ok
}

func (err ErrGPGKeyParsing) Error() string {
	return fmt.Sprintf("failed to parse gpg key %s", err.ParseError.Error())
}

// ErrGPGKeyNotExist represents a "GPGKeyNotExist" kind of error.
type ErrGPGKeyNotExist struct {
	ID int64
}

// IsErrGPGKeyNotExist checks if an error is a ErrGPGKeyNotExist.
func IsErrGPGKeyNotExist(err error) bool {
	_, ok := err.(ErrGPGKeyNotExist)
	return ok
}

func (err ErrGPGKeyNotExist) Error() string {
	return fmt.Sprintf("public gpg key does not exist [id: %d]", err.ID)
}

// ErrGPGKeyImportNotExist represents a "GPGKeyImportNotExist" kind of error.
type ErrGPGKeyImportNotExist struct {
	ID string
}

// IsErrGPGKeyImportNotExist checks if an error is a ErrGPGKeyImportNotExist.
func IsErrGPGKeyImportNotExist(err error) bool {
	_, ok := err.(ErrGPGKeyImportNotExist)
	return ok
}

func (err ErrGPGKeyImportNotExist) Error() string {
	return fmt.Sprintf("public gpg key import does not exist [id: %s]", err.ID)
}

// ErrGPGKeyIDAlreadyUsed represents a "GPGKeyIDAlreadyUsed" kind of error.
type ErrGPGKeyIDAlreadyUsed struct {
	KeyID string
}

// IsErrGPGKeyIDAlreadyUsed checks if an error is a ErrKeyNameAlreadyUsed.
func IsErrGPGKeyIDAlreadyUsed(err error) bool {
	_, ok := err.(ErrGPGKeyIDAlreadyUsed)
	return ok
}

func (err ErrGPGKeyIDAlreadyUsed) Error() string {
	return fmt.Sprintf("public key already exists [key_id: %s]", err.KeyID)
}

// ErrGPGKeyAccessDenied represents a "GPGKeyAccessDenied" kind of Error.
type ErrGPGKeyAccessDenied struct {
	UserID int64
	KeyID  int64
}

// IsErrGPGKeyAccessDenied checks if an error is a ErrGPGKeyAccessDenied.
func IsErrGPGKeyAccessDenied(err error) bool {
	_, ok := err.(ErrGPGKeyAccessDenied)
	return ok
}

// Error pretty-prints an error of type ErrGPGKeyAccessDenied.
func (err ErrGPGKeyAccessDenied) Error() string {
	return fmt.Sprintf("user does not have access to the key [user_id: %d, key_id: %d]",
		err.UserID, err.KeyID)
}

// ErrKeyAccessDenied represents a "KeyAccessDenied" kind of error.
type ErrKeyAccessDenied struct {
	UserID int64
	KeyID  int64
	Note   string
}

// IsErrKeyAccessDenied checks if an error is a ErrKeyAccessDenied.
func IsErrKeyAccessDenied(err error) bool {
	_, ok := err.(ErrKeyAccessDenied)
	return ok
}

func (err ErrKeyAccessDenied) Error() string {
	return fmt.Sprintf("user does not have access to the key [user_id: %d, key_id: %d, note: %s]",
		err.UserID, err.KeyID, err.Note)
}

//    _____                                   ___________     __
//   /  _  \   ____  ____  ____   ______ _____\__    ___/___ |  | __ ____   ____
//  /  /_\  \_/ ___\/ ___\/ __ \ /  ___//  ___/ |    | /  _ \|  |/ // __ \ /    \
// /    |    \  \__\  \__\  ___/ \___ \ \___ \  |    |(  <_> )    <\  ___/|   |  \
// \____|__  /\___  >___  >___  >____  >____  > |____| \____/|__|_ \\___  >___|  /
//         \/     \/    \/    \/     \/     \/                    \/    \/     \/

// ErrAccessTokenNotExist represents a "AccessTokenNotExist" kind of error.
type ErrAccessTokenNotExist struct {
	Token string
}

// IsErrAccessTokenNotExist checks if an error is a ErrAccessTokenNotExist.
func IsErrAccessTokenNotExist(err error) bool {
	_, ok := err.(ErrAccessTokenNotExist)
	return ok
}

func (err ErrAccessTokenNotExist) Error() string {
	return fmt.Sprintf("access token does not exist [sha: %s]", err.Token)
}

// ErrAccessTokenEmpty represents a "AccessTokenEmpty" kind of error.
type ErrAccessTokenEmpty struct{}

// IsErrAccessTokenEmpty checks if an error is a ErrAccessTokenEmpty.
func IsErrAccessTokenEmpty(err error) bool {
	_, ok := err.(ErrAccessTokenEmpty)
	return ok
}

func (err ErrAccessTokenEmpty) Error() string {
	return "access token is empty"
}

// ________                            .__                __  .__
// \_____  \_______  _________    ____ |__|____________ _/  |_|__| ____   ____
//  /   |   \_  __ \/ ___\__  \  /    \|  \___   /\__  \\   __\  |/  _ \ /    \
// /    |    \  | \/ /_/  > __ \|   |  \  |/    /  / __ \|  | |  (  <_> )   |  \
// \_______  /__|  \___  (____  /___|  /__/_____ \(____  /__| |__|\____/|___|  /
//         \/     /_____/     \/     \/         \/     \/                    \/

// ErrOrgNotExist represents a "OrgNotExist" kind of error.
type ErrOrgNotExist struct {
	ID   int64
	Name string
}

// IsErrOrgNotExist checks if an error is a ErrOrgNotExist.
func IsErrOrgNotExist(err error) bool {
	_, ok := err.(ErrOrgNotExist)
	return ok
}

func (err ErrOrgNotExist) Error() string {
	return fmt.Sprintf("org does not exist [id: %d, name: %s]", err.ID, err.Name)
}

// ErrLastOrgOwner represents a "LastOrgOwner" kind of error.
type ErrLastOrgOwner struct {
	UID int64
}

// IsErrLastOrgOwner checks if an error is a ErrLastOrgOwner.
func IsErrLastOrgOwner(err error) bool {
	_, ok := err.(ErrLastOrgOwner)
	return ok
}

func (err ErrLastOrgOwner) Error() string {
	return fmt.Sprintf("user is the last member of owner team [uid: %d]", err.UID)
}

//  _________ __                                __         .__
//  /   _____//  |_  ____ ________  _  _______ _/  |_  ____ |  |__
//  \_____  \\   __\/  _ \\____ \ \/ \/ /\__  \\   __\/ ___\|  |  \
//  /        \|  | (  <_> )  |_> >     /  / __ \|  | \  \___|   Y  \
//  /_______  /|__|  \____/|   __/ \/\_/  (____  /__|  \___  >___|  /
// \/             |__|                \/          \/     \/

// ErrStopwatchNotExist represents a "Stopwatch Not Exist" kind of error.
type ErrStopwatchNotExist struct {
	ID int64
}

// IsErrStopwatchNotExist checks if an error is a ErrStopwatchNotExist.
func IsErrStopwatchNotExist(err error) bool {
	_, ok := err.(ErrStopwatchNotExist)
	return ok
}

func (err ErrStopwatchNotExist) Error() string {
	return fmt.Sprintf("stopwatch does not exist [id: %d]", err.ID)
}

// ___________                     __              .______________.__
// \__    ___/___________    ____ |  | __ ____   __| _/\__    ___/|__| _____   ____
// |    |  \_  __ \__  \ _/ ___\|  |/ // __ \ / __ |   |    |   |  |/     \_/ __ \
// |    |   |  | \// __ \\  \___|    <\  ___// /_/ |   |    |   |  |  Y Y  \  ___/
// |____|   |__|  (____  /\___  >__|_ \\___  >____ |   |____|   |__|__|_|  /\___  >
// \/     \/     \/    \/     \/                     \/     \/

// ErrTrackedTimeNotExist represents a "TrackedTime Not Exist" kind of error.
type ErrTrackedTimeNotExist struct {
	ID int64
}

// IsErrTrackedTimeNotExist checks if an error is a ErrTrackedTimeNotExist.
func IsErrTrackedTimeNotExist(err error) bool {
	_, ok := err.(ErrTrackedTimeNotExist)
	return ok
}

func (err ErrTrackedTimeNotExist) Error() string {
	return fmt.Sprintf("tracked time does not exist [id: %d]", err.ID)
}

// .____                 .__           _________
// |    |    ____   ____ |__| ____    /   _____/ ____  __ _________   ____  ____
// |    |   /  _ \ / ___\|  |/    \   \_____  \ /  _ \|  |  \_  __ \_/ ___\/ __ \
// |    |__(  <_> ) /_/  >  |   |  \  /        (  <_> )  |  /|  | \/\  \__\  ___/
// |_______ \____/\___  /|__|___|  / /_______  /\____/|____/ |__|    \___  >___  >
//         \/    /_____/         \/          \/                          \/    \/

// ErrLoginSourceNotExist represents a "LoginSourceNotExist" kind of error.
type ErrLoginSourceNotExist struct {
	ID int64
}

// IsErrLoginSourceNotExist checks if an error is a ErrLoginSourceNotExist.
func IsErrLoginSourceNotExist(err error) bool {
	_, ok := err.(ErrLoginSourceNotExist)
	return ok
}

func (err ErrLoginSourceNotExist) Error() string {
	return fmt.Sprintf("login source does not exist [id: %d]", err.ID)
}

// ErrLoginSourceAlreadyExist represents a "LoginSourceAlreadyExist" kind of error.
type ErrLoginSourceAlreadyExist struct {
	Name string
}

// IsErrLoginSourceAlreadyExist checks if an error is a ErrLoginSourceAlreadyExist.
func IsErrLoginSourceAlreadyExist(err error) bool {
	_, ok := err.(ErrLoginSourceAlreadyExist)
	return ok
}

func (err ErrLoginSourceAlreadyExist) Error() string {
	return fmt.Sprintf("login source already exists [name: %s]", err.Name)
}

// ErrLoginSourceInUse represents a "LoginSourceInUse" kind of error.
type ErrLoginSourceInUse struct {
	ID int64
}

// IsErrLoginSourceInUse checks if an error is a ErrLoginSourceInUse.
func IsErrLoginSourceInUse(err error) bool {
	_, ok := err.(ErrLoginSourceInUse)
	return ok
}

func (err ErrLoginSourceInUse) Error() string {
	return fmt.Sprintf("login source is still used by some users [id: %d]", err.ID)
}

// ___________
// \__    ___/___ _____    _____
//   |    |_/ __ \\__  \  /     \
//   |    |\  ___/ / __ \|  Y Y  \
//   |____| \___  >____  /__|_|  /
//              \/     \/      \/

// ErrTeamAlreadyExist represents a "TeamAlreadyExist" kind of error.
type ErrTeamAlreadyExist struct {
	OrgID int64
	Name  string
}

// IsErrTeamAlreadyExist checks if an error is a ErrTeamAlreadyExist.
func IsErrTeamAlreadyExist(err error) bool {
	_, ok := err.(ErrTeamAlreadyExist)
	return ok
}

func (err ErrTeamAlreadyExist) Error() string {
	return fmt.Sprintf("team already exists [org_id: %d, name: %s]", err.OrgID, err.Name)
}

// ErrTeamNotExist represents a "TeamNotExist" error
type ErrTeamNotExist struct {
	OrgID  int64
	TeamID int64
	Name   string
}

// IsErrTeamNotExist checks if an error is a ErrTeamNotExist.
func IsErrTeamNotExist(err error) bool {
	_, ok := err.(ErrTeamNotExist)
	return ok
}

func (err ErrTeamNotExist) Error() string {
	return fmt.Sprintf("team does not exist [org_id %d, team_id %d, name: %s]", err.OrgID, err.TeamID, err.Name)
}

//
// Two-factor authentication
//

// ErrTwoFactorNotEnrolled indicates that a user is not enrolled in two-factor authentication.
type ErrTwoFactorNotEnrolled struct {
	UID int64
}

// IsErrTwoFactorNotEnrolled checks if an error is a ErrTwoFactorNotEnrolled.
func IsErrTwoFactorNotEnrolled(err error) bool {
	_, ok := err.(ErrTwoFactorNotEnrolled)
	return ok
}

func (err ErrTwoFactorNotEnrolled) Error() string {
	return fmt.Sprintf("user not enrolled in 2FA [uid: %d]", err.UID)
}

//  ____ ___        .__                    .___
// |    |   \______ |  |   _________     __| _/
// |    |   /\____ \|  |  /  _ \__  \   / __ |
// |    |  / |  |_> >  |_(  <_> ) __ \_/ /_/ |
// |______/  |   __/|____/\____(____  /\____ |
//           |__|                   \/      \/
//

// ErrUploadNotExist represents a "UploadNotExist" kind of error.
type ErrUploadNotExist struct {
	ID   int64
	UUID string
}

// IsErrUploadNotExist checks if an error is a ErrUploadNotExist.
func IsErrUploadNotExist(err error) bool {
	_, ok := err.(ErrAttachmentNotExist)
	return ok
}

func (err ErrUploadNotExist) Error() string {
	return fmt.Sprintf("attachment does not exist [id: %d, uuid: %s]", err.ID, err.UUID)
}

//  ___________         __                             .__    .____                 .__          ____ ___
//  \_   _____/__  ____/  |_  ___________  ____ _____  |  |   |    |    ____   ____ |__| ____   |    |   \______ ___________
//   |    __)_\  \/  /\   __\/ __ \_  __ \/    \\__  \ |  |   |    |   /  _ \ / ___\|  |/    \  |    |   /  ___// __ \_  __ \
//   |        \>    <  |  | \  ___/|  | \/   |  \/ __ \|  |__ |    |__(  <_> ) /_/  >  |   |  \ |    |  /\___ \\  ___/|  | \/
//  /_______  /__/\_ \ |__|  \___  >__|  |___|  (____  /____/ |_______ \____/\___  /|__|___|  / |______//____  >\___  >__|
//          \/      \/           \/           \/     \/               \/    /_____/         \/               \/     \/

// ErrExternalLoginUserAlreadyExist represents a "ExternalLoginUserAlreadyExist" kind of error.
type ErrExternalLoginUserAlreadyExist struct {
	ExternalID    string
	UserID        int64
	LoginSourceID int64
}

// IsErrExternalLoginUserAlreadyExist checks if an error is a ExternalLoginUserAlreadyExist.
func IsErrExternalLoginUserAlreadyExist(err error) bool {
	_, ok := err.(ErrExternalLoginUserAlreadyExist)
	return ok
}

func (err ErrExternalLoginUserAlreadyExist) Error() string {
	return fmt.Sprintf("external login user already exists [externalID: %s, userID: %d, loginSourceID: %d]", err.ExternalID, err.UserID, err.LoginSourceID)
}

// ErrExternalLoginUserNotExist represents a "ExternalLoginUserNotExist" kind of error.
type ErrExternalLoginUserNotExist struct {
	UserID        int64
	LoginSourceID int64
}

// IsErrExternalLoginUserNotExist checks if an error is a ExternalLoginUserNotExist.
func IsErrExternalLoginUserNotExist(err error) bool {
	_, ok := err.(ErrExternalLoginUserNotExist)
	return ok
}

func (err ErrExternalLoginUserNotExist) Error() string {
	return fmt.Sprintf("external login user link does not exists [userID: %d, loginSourceID: %d]", err.UserID, err.LoginSourceID)
}

// ____ ________________________________              .__          __                 __  .__
// |    |   \_____  \_   _____/\______   \ ____   ____ |__| _______/  |_____________ _/  |_|__| ____   ____
// |    |   //  ____/|    __)   |       _// __ \ / ___\|  |/  ___/\   __\_  __ \__  \\   __\  |/  _ \ /    \
// |    |  //       \|     \    |    |   \  ___// /_/  >  |\___ \  |  |  |  | \// __ \|  | |  (  <_> )   |  \
// |______/ \_______ \___  /    |____|_  /\___  >___  /|__/____  > |__|  |__|  (____  /__| |__|\____/|___|  /
// \/   \/            \/     \/_____/         \/                   \/                    \/

// ErrU2FRegistrationNotExist represents a "ErrU2FRegistrationNotExist" kind of error.
type ErrU2FRegistrationNotExist struct {
	ID int64
}

func (err ErrU2FRegistrationNotExist) Error() string {
	return fmt.Sprintf("U2F registration does not exist [id: %d]", err.ID)
}

// IsErrU2FRegistrationNotExist checks if an error is a ErrU2FRegistrationNotExist.
func IsErrU2FRegistrationNotExist(err error) bool {
	_, ok := err.(ErrU2FRegistrationNotExist)
	return ok
}

//  ________      _____          __  .__
//  \_____  \    /  _  \  __ ___/  |_|  |__
//   /   |   \  /  /_\  \|  |  \   __\  |  \
//  /    |    \/    |    \  |  /|  | |   Y  \
//  \_______  /\____|__  /____/ |__| |___|  /
//          \/         \/                 \/

// ErrOAuthClientIDInvalid will be thrown if client id cannot be found
type ErrOAuthClientIDInvalid struct {
	ClientID string
}

// IsErrOauthClientIDInvalid checks if an error is a ErrReviewNotExist.
func IsErrOauthClientIDInvalid(err error) bool {
	_, ok := err.(ErrOAuthClientIDInvalid)
	return ok
}

// Error returns the error message
func (err ErrOAuthClientIDInvalid) Error() string {
	return fmt.Sprintf("Client ID invalid [Client ID: %s]", err.ClientID)
}

// ErrOAuthApplicationNotFound will be thrown if id cannot be found
type ErrOAuthApplicationNotFound struct {
	ID int64
}

// IsErrOAuthApplicationNotFound checks if an error is a ErrReviewNotExist.
func IsErrOAuthApplicationNotFound(err error) bool {
	_, ok := err.(ErrOAuthApplicationNotFound)
	return ok
}

// Error returns the error message
func (err ErrOAuthApplicationNotFound) Error() string {
	return fmt.Sprintf("OAuth application not found [ID: %d]", err.ID)
}
