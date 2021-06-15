// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"context"
	"fmt"
	"strings"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/models/migrations"
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/setting"
	"code.gitea.io/gitea/modules/storage"

	"github.com/urfave/cli/v2"
)

// CmdMigrateStorage represents the available migrate storage sub-command.
var CmdMigrateStorage = &cli.Command{
	Name:        "migrate-storage",
	Usage:       "Migrate the storage",
	Description: "This is a command for migrating storage.",
	Action:      runMigrateStorage,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "type",
			Aliases: []string{"t"},
			Value:   "",
			Usage:   "Kinds of files to migrate, currently only 'attachments' is supported",
		},
		&cli.StringFlag{
			Name:    "storage",
			Aliases: []string{"s"},
			Value:   "",
			Usage:   "New storage type: local (default) or minio",
		},
		&cli.StringFlag{
			Name:    "path",
			Aliases: []string{"p"},
			Value:   "",
			Usage:   "New storage placement if store is local (leave blank for default)",
		},
		&cli.StringFlag{
			Name:  "minio-endpoint",
			Value: "",
			Usage: "Minio storage endpoint",
		},
		&cli.StringFlag{
			Name:  "minio-access-key-id",
			Value: "",
			Usage: "Minio storage accessKeyID",
		},
		&cli.StringFlag{
			Name:  "minio-secret-access-key",
			Value: "",
			Usage: "Minio storage secretAccessKey",
		},
		&cli.StringFlag{
			Name:  "minio-bucket",
			Value: "",
			Usage: "Minio storage bucket",
		},
		&cli.StringFlag{
			Name:  "minio-location",
			Value: "",
			Usage: "Minio storage location to create bucket",
		},
		&cli.StringFlag{
			Name:  "minio-base-path",
			Value: "",
			Usage: "Minio storage basepath on the bucket",
		},
		&cli.BoolFlag{
			Name:  "minio-use-ssl",
			Usage: "Enable SSL for minio",
		},
	},
}

func migrateAvatars(dstStorage storage.ObjectStorage) error {
	return models.IterateUser(func(user *models.User) error {
		_, err := storage.Copy(dstStorage, user.CustomAvatarRelativePath(), storage.Avatars, user.CustomAvatarRelativePath())
		return err
	})
}

func runMigrateStorage(ctx *cli.Context) error {
	if err := initDB(); err != nil {
		return err
	}

	log.Trace("AppPath: %s", setting.AppPath)
	log.Trace("AppWorkPath: %s", setting.AppWorkPath)
	log.Trace("Custom path: %s", setting.CustomPath)
	log.Trace("Log path: %s", setting.LogRootPath)
	setting.InitDBConfig()

	if err := models.NewEngine(context.Background(), migrations.Migrate); err != nil {
		log.Fatal("Failed to initialize ORM engine: %v", err)
		return err
	}

	goCtx := context.Background()

	if err := storage.Init(); err != nil {
		return err
	}

	var dstStorage storage.ObjectStorage
	var err error
	switch strings.ToLower(ctx.String("storage")) {
	case "":
		fallthrough
	case string(storage.LocalStorageType):
		p := ctx.String("path")
		if p == "" {
			log.Fatal("Path must be given when storage is loal")
			return nil
		}
		dstStorage, err = storage.NewLocalStorage(
			goCtx,
			storage.LocalStorageConfig{
				Path: p,
			})
	case string(storage.MinioStorageType):
		dstStorage, err = storage.NewMinioStorage(
			goCtx,
			storage.MinioStorageConfig{
				Endpoint:        ctx.String("minio-endpoint"),
				AccessKeyID:     ctx.String("minio-access-key-id"),
				SecretAccessKey: ctx.String("minio-secret-access-key"),
				Bucket:          ctx.String("minio-bucket"),
				Location:        ctx.String("minio-location"),
				BasePath:        ctx.String("minio-base-path"),
				UseSSL:          ctx.Bool("minio-use-ssl"),
			})
	default:
		return fmt.Errorf("Unsupported storage type: %s", ctx.String("storage"))
	}
	if err != nil {
		return err
	}

	tp := strings.ToLower(ctx.String("type"))
	switch tp {
	case "avatars":
		if err := migrateAvatars(dstStorage); err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unsupported storage: %s", ctx.String("type"))
	}

	log.Warn("All files have been copied to the new placement but old files are still on the orignial placement.")

	return nil
}
