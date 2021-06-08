// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package routers

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go.wandrs.dev/framework/models"
	"go.wandrs.dev/framework/modules/base"
	"go.wandrs.dev/framework/modules/context"
	"go.wandrs.dev/framework/modules/generate"
	"go.wandrs.dev/framework/modules/graceful"
	"go.wandrs.dev/framework/modules/log"
	"go.wandrs.dev/framework/modules/setting"
	"go.wandrs.dev/framework/modules/templates"
	"go.wandrs.dev/framework/modules/translation"
	"go.wandrs.dev/framework/modules/user"
	"go.wandrs.dev/framework/modules/util"
	"go.wandrs.dev/framework/modules/web"
	"go.wandrs.dev/framework/modules/web/middleware"
	"go.wandrs.dev/framework/services/forms"
	"go.wandrs.dev/session"

	"gopkg.in/ini.v1"
)

const (
	// tplInstall template for installation page
	tplInstall     base.TplName = "install"
	tplPostInstall base.TplName = "post-install"
)

// InstallInit prepare for rendering installation page
func InstallInit(next http.Handler) http.Handler {
	var rnd = templates.HTMLRenderer()

	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if setting.InstallLock {
			resp.Header().Add("Refresh", "1; url="+setting.AppURL+"user/login")
			_ = rnd.HTML(resp, 200, string(tplPostInstall), nil)
			return
		}
		var locale = middleware.Locale(resp, req)
		var startTime = time.Now()
		var ctx = context.Context{
			Resp:    context.NewResponse(resp),
			Flash:   &middleware.Flash{},
			Locale:  locale,
			Render:  rnd,
			Session: session.GetSession(req),
			Data: map[string]interface{}{
				"Title":         locale.Tr("install.install"),
				"PageIsInstall": true,
				"DbOptions":     setting.SupportedDatabases,
				"i18n":          locale,
				"Language":      locale.Language(),
				"Lang":          locale.Language(),
				"AllLangs":      translation.AllLangs(),
				"CurrentURL":    setting.AppSubURL + req.URL.RequestURI(),
				"PageStartTime": startTime,
				"TmplLoadTimes": func() string {
					return time.Since(startTime).String()
				},
				"PasswordHashAlgorithms": models.AvailableHashAlgorithms,
			},
		}
		for _, lang := range translation.AllLangs() {
			if lang.Lang == locale.Language() {
				ctx.Data["LangName"] = lang.Name
				break
			}
		}
		ctx.Req = context.WithContext(req, &ctx)
		next.ServeHTTP(resp, ctx.Req)
	})
}

// Install render installation page
func Install(ctx *context.Context) {
	form := forms.InstallForm{}

	// Database settings
	form.DbHost = setting.Database.Host
	form.DbUser = setting.Database.User
	form.DbPasswd = setting.Database.Passwd
	form.DbName = setting.Database.Name
	form.DbPath = setting.Database.Path
	form.DbSchema = setting.Database.Schema
	form.Charset = setting.Database.Charset

	var curDBOption = "MySQL"
	switch setting.Database.Type {
	case "postgres":
		curDBOption = "PostgreSQL"
	case "mssql":
		curDBOption = "MSSQL"
	case "sqlite3":
		if setting.EnableSQLite3 {
			curDBOption = "SQLite3"
		}
	}

	ctx.Data["CurDbOption"] = curDBOption

	// Application general settings
	form.AppName = setting.AppName

	// Note(unknown): it's hard for Windows users change a running user,
	// 	so just use current one if config says default.
	if setting.IsWindows && setting.RunUser == "git" {
		form.RunUser = user.CurrentUsername()
	} else {
		form.RunUser = setting.RunUser
	}

	form.Domain = setting.Domain
	form.HTTPPort = setting.HTTPPort
	form.AppURL = setting.AppURL
	form.LogRootPath = setting.LogRootPath

	// E-mail service settings
	if setting.MailService != nil {
		form.SMTPHost = setting.MailService.Host
		form.SMTPFrom = setting.MailService.From
		form.SMTPUser = setting.MailService.User
	}
	form.RegisterConfirm = setting.Service.RegisterEmailConfirm
	form.MailNotify = setting.Service.EnableNotifyMail

	// Server and other services settings
	form.OfflineMode = setting.OfflineMode
	form.DisableGravatar = setting.DisableGravatar
	form.EnableFederatedAvatar = setting.EnableFederatedAvatar
	form.EnableOpenIDSignIn = setting.Service.EnableOpenIDSignIn
	form.EnableOpenIDSignUp = setting.Service.EnableOpenIDSignUp
	form.DisableRegistration = setting.Service.DisableRegistration
	form.AllowOnlyExternalRegistration = setting.Service.AllowOnlyExternalRegistration
	form.EnableCaptcha = setting.Service.EnableCaptcha
	form.RequireSignInView = setting.Service.RequireSignInView
	form.DefaultKeepEmailPrivate = setting.Service.DefaultKeepEmailPrivate
	form.DefaultAllowCreateOrganization = setting.Service.DefaultAllowCreateOrganization
	form.NoReplyAddress = setting.Service.NoReplyAddress
	form.PasswordAlgorithm = setting.PasswordHashAlgo

	middleware.AssignForm(form, ctx.Data)
	ctx.HTML(http.StatusOK, tplInstall)
}

// InstallPost response for submit install items
func InstallPost(ctx *context.Context) {
	form := *web.GetForm(ctx).(*forms.InstallForm)
	var err error
	ctx.Data["CurDbOption"] = form.DbType

	if ctx.HasError() {
		if ctx.HasValue("Err_SMTPUser") {
			ctx.Data["Err_SMTP"] = true
		}
		if ctx.HasValue("Err_AdminName") ||
			ctx.HasValue("Err_AdminPasswd") ||
			ctx.HasValue("Err_AdminEmail") {
			ctx.Data["Err_Admin"] = true
		}

		ctx.HTML(http.StatusOK, tplInstall)
		return
	}

	if _, err = exec.LookPath("git"); err != nil {
		ctx.RenderWithErr(ctx.Tr("install.test_git_failed", err), tplInstall, &form)
		return
	}

	// Pass basic check, now test configuration.
	// Test database setting.

	setting.Database.Type = setting.GetDBTypeByName(form.DbType)
	setting.Database.Host = form.DbHost
	setting.Database.User = form.DbUser
	setting.Database.Passwd = form.DbPasswd
	setting.Database.Name = form.DbName
	setting.Database.Schema = form.DbSchema
	setting.Database.SSLMode = form.SSLMode
	setting.Database.Charset = form.Charset
	setting.Database.Path = form.DbPath

	setting.PasswordHashAlgo = form.PasswordAlgorithm

	if (setting.Database.Type == "sqlite3") &&
		len(setting.Database.Path) == 0 {
		ctx.Data["Err_DbPath"] = true
		ctx.RenderWithErr(ctx.Tr("install.err_empty_db_path"), tplInstall, &form)
		return
	}

	// Set test engine.
	if err = models.NewTestEngine(); err != nil {
		if strings.Contains(err.Error(), `Unknown database type: sqlite3`) {
			ctx.Data["Err_DbType"] = true
			ctx.RenderWithErr(ctx.Tr("install.sqlite3_not_available", "https://docs.gitea.io/en-us/install-from-binary/"), tplInstall, &form)
		} else {
			ctx.Data["Err_DbSetting"] = true
			ctx.RenderWithErr(ctx.Tr("install.invalid_db_setting", err), tplInstall, &form)
		}
		return
	}

	// Test log root path.
	form.LogRootPath = strings.ReplaceAll(form.LogRootPath, "\\", "/")
	if err = os.MkdirAll(form.LogRootPath, os.ModePerm); err != nil {
		ctx.Data["Err_LogRootPath"] = true
		ctx.RenderWithErr(ctx.Tr("install.invalid_log_root_path", err), tplInstall, &form)
		return
	}

	currentUser, match := setting.IsRunUserMatchCurrentUser(form.RunUser)
	if !match {
		ctx.Data["Err_RunUser"] = true
		ctx.RenderWithErr(ctx.Tr("install.run_user_not_match", form.RunUser, currentUser), tplInstall, &form)
		return
	}

	// Check logic loophole between disable self-registration and no admin account.
	if form.DisableRegistration && len(form.AdminName) == 0 {
		ctx.Data["Err_Services"] = true
		ctx.Data["Err_Admin"] = true
		ctx.RenderWithErr(ctx.Tr("install.no_admin_and_disable_registration"), tplInstall, form)
		return
	}

	// Check admin user creation
	if len(form.AdminName) > 0 {
		// Ensure AdminName is valid
		if err := models.IsUsableUsername(form.AdminName); err != nil {
			ctx.Data["Err_Admin"] = true
			ctx.Data["Err_AdminName"] = true
			if models.IsErrNameReserved(err) {
				ctx.RenderWithErr(ctx.Tr("install.err_admin_name_is_reserved"), tplInstall, form)
				return
			} else if models.IsErrNamePatternNotAllowed(err) {
				ctx.RenderWithErr(ctx.Tr("install.err_admin_name_pattern_not_allowed"), tplInstall, form)
				return
			}
			ctx.RenderWithErr(ctx.Tr("install.err_admin_name_is_invalid"), tplInstall, form)
			return
		}
		// Check Admin email
		if len(form.AdminEmail) == 0 {
			ctx.Data["Err_Admin"] = true
			ctx.Data["Err_AdminEmail"] = true
			ctx.RenderWithErr(ctx.Tr("install.err_empty_admin_email"), tplInstall, form)
			return
		}
		// Check admin password.
		if len(form.AdminPasswd) == 0 {
			ctx.Data["Err_Admin"] = true
			ctx.Data["Err_AdminPasswd"] = true
			ctx.RenderWithErr(ctx.Tr("install.err_empty_admin_password"), tplInstall, form)
			return
		}
		if form.AdminPasswd != form.AdminConfirmPasswd {
			ctx.Data["Err_Admin"] = true
			ctx.Data["Err_AdminPasswd"] = true
			ctx.RenderWithErr(ctx.Tr("form.password_not_match"), tplInstall, form)
			return
		}
	}

	if form.AppURL[len(form.AppURL)-1] != '/' {
		form.AppURL += "/"
	}

	// Save settings.
	cfg := ini.Empty()
	isFile, err := util.IsFile(setting.CustomConf)
	if err != nil {
		log.Error("Unable to check if %s is a file. Error: %v", setting.CustomConf, err)
	}
	if isFile {
		// Keeps custom settings if there is already something.
		if err = cfg.Append(setting.CustomConf); err != nil {
			log.Error("Failed to load custom conf '%s': %v", setting.CustomConf, err)
		}
	}
	cfg.Section("database").Key("DB_TYPE").SetValue(setting.Database.Type)
	cfg.Section("database").Key("HOST").SetValue(setting.Database.Host)
	cfg.Section("database").Key("NAME").SetValue(setting.Database.Name)
	cfg.Section("database").Key("USER").SetValue(setting.Database.User)
	cfg.Section("database").Key("PASSWD").SetValue(setting.Database.Passwd)
	cfg.Section("database").Key("SCHEMA").SetValue(setting.Database.Schema)
	cfg.Section("database").Key("SSL_MODE").SetValue(setting.Database.SSLMode)
	cfg.Section("database").Key("CHARSET").SetValue(setting.Database.Charset)
	cfg.Section("database").Key("PATH").SetValue(setting.Database.Path)
	cfg.Section("database").Key("LOG_SQL").SetValue("false") // LOG_SQL is rarely helpful

	cfg.Section("").Key("APP_NAME").SetValue(form.AppName)
	cfg.Section("").Key("RUN_USER").SetValue(form.RunUser)
	cfg.Section("server").Key("SSH_DOMAIN").SetValue(form.Domain)
	cfg.Section("server").Key("DOMAIN").SetValue(form.Domain)
	cfg.Section("server").Key("HTTP_PORT").SetValue(form.HTTPPort)
	cfg.Section("server").Key("ROOT_URL").SetValue(form.AppURL)

	if len(strings.TrimSpace(form.SMTPHost)) > 0 {
		cfg.Section("mailer").Key("ENABLED").SetValue("true")
		cfg.Section("mailer").Key("HOST").SetValue(form.SMTPHost)
		cfg.Section("mailer").Key("FROM").SetValue(form.SMTPFrom)
		cfg.Section("mailer").Key("USER").SetValue(form.SMTPUser)
		cfg.Section("mailer").Key("PASSWD").SetValue(form.SMTPPasswd)
	} else {
		cfg.Section("mailer").Key("ENABLED").SetValue("false")
	}
	cfg.Section("service").Key("REGISTER_EMAIL_CONFIRM").SetValue(fmt.Sprint(form.RegisterConfirm))
	cfg.Section("service").Key("ENABLE_NOTIFY_MAIL").SetValue(fmt.Sprint(form.MailNotify))

	cfg.Section("server").Key("OFFLINE_MODE").SetValue(fmt.Sprint(form.OfflineMode))
	cfg.Section("picture").Key("DISABLE_GRAVATAR").SetValue(fmt.Sprint(form.DisableGravatar))
	cfg.Section("picture").Key("ENABLE_FEDERATED_AVATAR").SetValue(fmt.Sprint(form.EnableFederatedAvatar))
	cfg.Section("openid").Key("ENABLE_OPENID_SIGNIN").SetValue(fmt.Sprint(form.EnableOpenIDSignIn))
	cfg.Section("openid").Key("ENABLE_OPENID_SIGNUP").SetValue(fmt.Sprint(form.EnableOpenIDSignUp))
	cfg.Section("service").Key("DISABLE_REGISTRATION").SetValue(fmt.Sprint(form.DisableRegistration))
	cfg.Section("service").Key("ALLOW_ONLY_EXTERNAL_REGISTRATION").SetValue(fmt.Sprint(form.AllowOnlyExternalRegistration))
	cfg.Section("service").Key("ENABLE_CAPTCHA").SetValue(fmt.Sprint(form.EnableCaptcha))
	cfg.Section("service").Key("REQUIRE_SIGNIN_VIEW").SetValue(fmt.Sprint(form.RequireSignInView))
	cfg.Section("service").Key("DEFAULT_KEEP_EMAIL_PRIVATE").SetValue(fmt.Sprint(form.DefaultKeepEmailPrivate))
	cfg.Section("service").Key("DEFAULT_ALLOW_CREATE_ORGANIZATION").SetValue(fmt.Sprint(form.DefaultAllowCreateOrganization))
	cfg.Section("service").Key("DEFAULT_ENABLE_TIMETRACKING").SetValue(fmt.Sprint(form.DefaultEnableTimetracking))
	cfg.Section("service").Key("NO_REPLY_ADDRESS").SetValue(fmt.Sprint(form.NoReplyAddress))

	cfg.Section("").Key("RUN_MODE").SetValue("prod")

	cfg.Section("session").Key("PROVIDER").SetValue("file")

	cfg.Section("log").Key("MODE").SetValue("console")
	cfg.Section("log").Key("LEVEL").SetValue(setting.LogLevel.String())
	cfg.Section("log").Key("ROOT_PATH").SetValue(form.LogRootPath)
	cfg.Section("log").Key("ROUTER").SetValue("console")

	cfg.Section("security").Key("INSTALL_LOCK").SetValue("true")
	var secretKey string
	if secretKey, err = generate.NewSecretKey(); err != nil {
		ctx.RenderWithErr(ctx.Tr("install.secret_key_failed", err), tplInstall, &form)
		return
	}
	cfg.Section("security").Key("SECRET_KEY").SetValue(secretKey)
	if len(form.PasswordAlgorithm) > 0 {
		cfg.Section("security").Key("PASSWORD_HASH_ALGO").SetValue(form.PasswordAlgorithm)
	}

	err = os.MkdirAll(filepath.Dir(setting.CustomConf), os.ModePerm)
	if err != nil {
		ctx.RenderWithErr(ctx.Tr("install.save_config_failed", err), tplInstall, &form)
		return
	}

	if err = cfg.SaveTo(setting.CustomConf); err != nil {
		ctx.RenderWithErr(ctx.Tr("install.save_config_failed", err), tplInstall, &form)
		return
	}

	// Re-read settings
	PostInstallInit(ctx)

	// Create admin account
	if len(form.AdminName) > 0 {
		u := &models.User{
			Name:     form.AdminName,
			Email:    form.AdminEmail,
			Passwd:   form.AdminPasswd,
			IsAdmin:  true,
			IsActive: true,
		}
		if err = models.CreateUser(u); err != nil {
			if !models.IsErrUserAlreadyExist(err) {
				setting.InstallLock = false
				ctx.Data["Err_AdminName"] = true
				ctx.Data["Err_AdminEmail"] = true
				ctx.RenderWithErr(ctx.Tr("install.invalid_admin_setting", err), tplInstall, &form)
				return
			}
			log.Info("Admin account already exist")
			u, _ = models.GetUserByName(u.Name)
		}

		days := 86400 * setting.LogInRememberDays
		ctx.SetCookie(setting.CookieUserName, u.Name, days)

		ctx.SetSuperSecureCookie(base.EncodeMD5(u.Rands+u.Passwd),
			setting.CookieRememberName, u.Name, days)

		// Auto-login for admin
		if err = ctx.Session.Set("uid", u.ID); err != nil {
			ctx.RenderWithErr(ctx.Tr("install.save_config_failed", err), tplInstall, &form)
			return
		}
		if err = ctx.Session.Set("uname", u.Name); err != nil {
			ctx.RenderWithErr(ctx.Tr("install.save_config_failed", err), tplInstall, &form)
			return
		}

		if err = ctx.Session.Release(); err != nil {
			ctx.RenderWithErr(ctx.Tr("install.save_config_failed", err), tplInstall, &form)
			return
		}
	}

	log.Info("First-time run install finished!")

	ctx.Flash.Success(ctx.Tr("install.install_success"))

	ctx.Header().Add("Refresh", "1; url="+setting.AppURL+"user/login")
	ctx.HTML(http.StatusOK, tplPostInstall)

	// Now get the http.Server from this request and shut it down
	// NB: This is not our hammerable graceful shutdown this is http.Server.Shutdown
	srv := ctx.Value(http.ServerContextKey).(*http.Server)
	go func() {
		if err := srv.Shutdown(graceful.GetManager().HammerContext()); err != nil {
			log.Error("Unable to shutdown the install server! Error: %v", err)
		}
	}()
}
