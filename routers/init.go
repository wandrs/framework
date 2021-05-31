// Copyright 2016 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package routers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.wandrs.dev/framework/models"
	"go.wandrs.dev/framework/models/migrations"
	"go.wandrs.dev/framework/modules/auth/sso"
	"go.wandrs.dev/framework/modules/cache"
	"go.wandrs.dev/framework/modules/cron"
	"go.wandrs.dev/framework/modules/eventsource"
	"go.wandrs.dev/framework/modules/git"
	"go.wandrs.dev/framework/modules/highlight"
	code_indexer "go.wandrs.dev/framework/modules/indexer/code"
	issue_indexer "go.wandrs.dev/framework/modules/indexer/issues"
	stats_indexer "go.wandrs.dev/framework/modules/indexer/stats"
	"go.wandrs.dev/framework/modules/log"
	"go.wandrs.dev/framework/modules/markup"
	"go.wandrs.dev/framework/modules/markup/external"
	repo_migrations "go.wandrs.dev/framework/modules/migrations"
	"go.wandrs.dev/framework/modules/notification"
	"go.wandrs.dev/framework/modules/setting"
	"go.wandrs.dev/framework/modules/ssh"
	"go.wandrs.dev/framework/modules/storage"
	"go.wandrs.dev/framework/modules/svg"
	"go.wandrs.dev/framework/modules/task"
	"go.wandrs.dev/framework/modules/translation"
	"go.wandrs.dev/framework/services/mailer"
	mirror_service "go.wandrs.dev/framework/services/mirror"
	pull_service "go.wandrs.dev/framework/services/pull"
	"go.wandrs.dev/framework/services/repository"
	"go.wandrs.dev/framework/services/webhook"
)

func checkRunMode() {
	switch setting.RunMode {
	case "dev", "test":
		git.Debug = true
	default:
		git.Debug = false
	}
	log.Info("Run Mode: %s", strings.Title(setting.RunMode))
}

// NewServices init new services
func NewServices() {
	setting.NewServices()
	if err := storage.Init(); err != nil {
		log.Fatal("storage init failed: %v", err)
	}
	if err := repository.NewContext(); err != nil {
		log.Fatal("repository init failed: %v", err)
	}
	mailer.NewContext()
	_ = cache.NewContext()
	notification.NewContext()
}

// In case of problems connecting to DB, retry connection. Eg, PGSQL in Docker Container on Synology
func initDBEngine(ctx context.Context) (err error) {
	log.Info("Beginning ORM engine initialization.")
	for i := 0; i < setting.Database.DBConnectRetries; i++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("Aborted due to shutdown:\nin retry ORM engine initialization")
		default:
		}
		log.Info("ORM engine initialization attempt #%d/%d...", i+1, setting.Database.DBConnectRetries)
		if err = models.NewEngine(ctx, migrations.Migrate); err == nil {
			break
		} else if i == setting.Database.DBConnectRetries-1 {
			return err
		}
		log.Error("ORM engine initialization attempt #%d/%d failed. Error: %v", i+1, setting.Database.DBConnectRetries, err)
		log.Info("Backing off for %d seconds", int64(setting.Database.DBConnectBackoff/time.Second))
		time.Sleep(setting.Database.DBConnectBackoff)
	}
	models.HasEngine = true
	return nil
}

// PreInstallInit preloads the configuration to check if we need to run install
func PreInstallInit(ctx context.Context) bool {
	setting.NewContext()
	if !setting.InstallLock {
		log.Trace("AppPath: %s", setting.AppPath)
		log.Trace("AppWorkPath: %s", setting.AppWorkPath)
		log.Trace("Custom path: %s", setting.CustomPath)
		log.Trace("Log path: %s", setting.LogRootPath)
		log.Trace("Preparing to run install page")
		translation.InitLocales()
		if setting.EnableSQLite3 {
			log.Info("SQLite3 Supported")
		}
		setting.InitDBConfig()
		svg.Init()
	}

	return !setting.InstallLock
}

// PostInstallInit rereads the settings and starts up the database
func PostInstallInit(ctx context.Context) {
	setting.NewContext()
	setting.InitDBConfig()
	if setting.InstallLock {
		if err := initDBEngine(ctx); err == nil {
			log.Info("ORM engine initialization successful!")
		} else {
			log.Fatal("ORM engine initialization failed: %v", err)
		}
		svg.Init()
	}
}

// GlobalInit is for global configuration reload-able.
func GlobalInit(ctx context.Context) {
	setting.NewContext()
	if !setting.InstallLock {
		log.Fatal("Gitea is not installed")
	}

	if err := git.Init(ctx); err != nil {
		log.Fatal("Git module init failed: %v", err)
	}
	setting.CheckLFSVersion()
	log.Trace("AppPath: %s", setting.AppPath)
	log.Trace("AppWorkPath: %s", setting.AppWorkPath)
	log.Trace("Custom path: %s", setting.CustomPath)
	log.Trace("Log path: %s", setting.LogRootPath)
	checkRunMode()

	// Setup i18n
	translation.InitLocales()

	NewServices()

	highlight.NewContext()
	external.RegisterRenderers()
	markup.Init()

	if setting.EnableSQLite3 {
		log.Info("SQLite3 Supported")
	} else if setting.Database.UseSQLite3 {
		log.Fatal("SQLite3 is set in settings but NOT Supported")
	}
	if err := initDBEngine(ctx); err == nil {
		log.Info("ORM engine initialization successful!")
	} else {
		log.Fatal("ORM engine initialization failed: %v", err)
	}

	if err := models.InitOAuth2(); err != nil {
		log.Fatal("Failed to initialize OAuth2 support: %v", err)
	}

	models.NewRepoContext()

	// Booting long running goroutines.
	cron.NewContext()
	issue_indexer.InitIssueIndexer(false)
	code_indexer.Init()
	if err := stats_indexer.Init(); err != nil {
		log.Fatal("Failed to initialize repository stats indexer queue: %v", err)
	}
	mirror_service.InitSyncMirrors()
	webhook.InitDeliverHooks()
	if err := pull_service.Init(); err != nil {
		log.Fatal("Failed to initialize test pull requests queue: %v", err)
	}
	if err := task.Init(); err != nil {
		log.Fatal("Failed to initialize task scheduler: %v", err)
	}
	if err := repo_migrations.Init(); err != nil {
		log.Fatal("Failed to initialize repository migrations: %v", err)
	}
	eventsource.GetManager().Init()

	if setting.SSH.StartBuiltinServer {
		ssh.Listen(setting.SSH.ListenHost, setting.SSH.ListenPort, setting.SSH.ServerCiphers, setting.SSH.ServerKeyExchanges, setting.SSH.ServerMACs)
		log.Info("SSH server started on %s:%d. Cipher list (%v), key exchange algorithms (%v), MACs (%v)", setting.SSH.ListenHost, setting.SSH.ListenPort, setting.SSH.ServerCiphers, setting.SSH.ServerKeyExchanges, setting.SSH.ServerMACs)
	} else {
		ssh.Unused()
	}
	sso.Init()

	svg.Init()
}
