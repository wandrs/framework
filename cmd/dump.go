// Copyright 2014 The Gogs Authors. All rights reserved.
// Copyright 2016 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"go.wandrs.dev/framework/models"
	"go.wandrs.dev/framework/modules/log"
	"go.wandrs.dev/framework/modules/setting"
	"go.wandrs.dev/framework/modules/storage"
	"go.wandrs.dev/framework/modules/util"
	"go.wandrs.dev/session"

	jsoniter "github.com/json-iterator/go"
	archiver "github.com/mholt/archiver/v3"
	"github.com/urfave/cli/v2"
)

func addFile(w archiver.Writer, filePath string, absPath string, verbose bool) error {
	if verbose {
		log.Info("Adding file %s\n", filePath)
	}
	file, err := os.Open(absPath)
	if err != nil {
		return err
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	return w.Write(archiver.File{
		FileInfo: archiver.FileInfo{
			FileInfo:   fileInfo,
			CustomName: filePath,
		},
		ReadCloser: file,
	})
}

func isSubdir(upper string, lower string) (bool, error) {
	if relPath, err := filepath.Rel(upper, lower); err != nil {
		return false, err
	} else if relPath == "." || !strings.HasPrefix(relPath, ".") {
		return true, nil
	}
	return false, nil
}

type outputType struct {
	Enum     []string
	Default  string
	selected string
}

func (o outputType) Join() string {
	return strings.Join(o.Enum, ", ")
}

func (o *outputType) Set(value string) error {
	for _, enum := range o.Enum {
		if enum == value {
			o.selected = value
			return nil
		}
	}

	return fmt.Errorf("allowed values are %s", o.Join())
}

func (o outputType) String() string {
	if o.selected == "" {
		return o.Default
	}
	return o.selected
}

var outputTypeEnum = &outputType{
	Enum:    []string{"zip", "tar", "tar.gz", "tar.xz", "tar.bz2"},
	Default: "zip",
}

// CmdDump represents the available dump sub-command.
var CmdDump = &cli.Command{
	Name:  "dump",
	Usage: "Dump Gitea files and database",
	Description: `Dump compresses all related files and database into zip file.
It can be used for backup and capture Gitea server image to send to maintainer`,
	Action: runDump,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "file",
			Aliases: []string{"f"},
			Value:   fmt.Sprintf("gitea-dump-%d.zip", time.Now().Unix()),
			Usage:   "Name of the dump file which will be created. Supply '-' for stdout. See type for available types.",
		},
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"V"},
			Usage:   "Show process details",
		},
		&cli.StringFlag{
			Name:    "tempdir",
			Aliases: []string{"t"},
			Value:   os.TempDir(),
			Usage:   "Temporary dir path",
		},
		&cli.StringFlag{
			Name:    "database",
			Aliases: []string{"d"},
			Usage:   "Specify the database SQL syntax",
		},
		&cli.BoolFlag{
			Name:    "skip-repository",
			Aliases: []string{"R"},
			Usage:   "Skip the repository dumping",
		},
		&cli.BoolFlag{
			Name:    "skip-log",
			Aliases: []string{"L"},
			Usage:   "Skip the log dumping",
		},
		&cli.BoolFlag{
			Name:  "skip-custom-dir",
			Usage: "Skip custom directory",
		},
		&cli.BoolFlag{
			Name:  "skip-lfs-data",
			Usage: "Skip LFS data",
		},
		&cli.BoolFlag{
			Name:  "skip-attachment-data",
			Usage: "Skip attachment data",
		},
		&cli.GenericFlag{
			Name:  "type",
			Value: outputTypeEnum,
			Usage: fmt.Sprintf("Dump output format: %s", outputTypeEnum.Join()),
		},
	},
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	log.Fatal(format, args...)
}

func runDump(ctx *cli.Context) error {
	var file *os.File
	fileName := ctx.String("file")
	if fileName == "-" {
		file = os.Stdout
		err := log.DelLogger("console")
		if err != nil {
			fatal("Deleting default logger failed. Can not write to stdout: %v", err)
		}
	}
	setting.NewContext()
	// make sure we are logging to the console no matter what the configuration tells us do to
	if _, err := setting.Cfg.Section("log").NewKey("MODE", "console"); err != nil {
		fatal("Setting logging mode to console failed: %v", err)
	}
	if _, err := setting.Cfg.Section("log.console").NewKey("STDERR", "true"); err != nil {
		fatal("Setting console logger to stderr failed: %v", err)
	}
	if !setting.InstallLock {
		log.Error("Is '%s' really the right config path?\n", setting.CustomConf)
		return fmt.Errorf("gitea is not initialized")
	}
	setting.NewServices() // cannot access session settings otherwise

	err := models.SetEngine()
	if err != nil {
		return err
	}

	if err := storage.Init(); err != nil {
		return err
	}

	if file == nil {
		file, err = os.Create(fileName)
		if err != nil {
			fatal("Unable to open %s: %v", fileName, err)
		}
	}
	defer file.Close()

	absFileName, err := filepath.Abs(fileName)
	if err != nil {
		return err
	}

	verbose := ctx.Bool("verbose")
	outType := ctx.String("type")
	var iface interface{}
	if fileName == "-" {
		iface, err = archiver.ByExtension(fmt.Sprintf(".%s", outType))
	} else {
		iface, err = archiver.ByExtension(fileName)
	}
	if err != nil {
		fatal("Unable to get archiver for extension: %v", err)
	}

	w, _ := iface.(archiver.Writer)
	if err := w.Create(file); err != nil {
		fatal("Creating archiver.Writer failed: %v", err)
	}
	defer w.Close()

	tmpDir := ctx.String("tempdir")
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		fatal("Path does not exist: %s", tmpDir)
	}

	dbDump, err := os.CreateTemp(tmpDir, "gitea-db.sql")
	if err != nil {
		fatal("Failed to create tmp file: %v", err)
	}
	defer func() {
		if err := util.Remove(dbDump.Name()); err != nil {
			log.Warn("Unable to remove temporary file: %s: Error: %v", dbDump.Name(), err)
		}
	}()

	targetDBType := ctx.String("database")
	if len(targetDBType) > 0 && targetDBType != setting.Database.Type {
		log.Info("Dumping database %s => %s...", setting.Database.Type, targetDBType)
	} else {
		log.Info("Dumping database...")
	}

	if err := models.DumpDatabase(dbDump.Name(), targetDBType); err != nil {
		fatal("Failed to dump database: %v", err)
	}

	if err := addFile(w, "gitea-db.sql", dbDump.Name(), verbose); err != nil {
		fatal("Failed to include gitea-db.sql: %v", err)
	}

	if len(setting.CustomConf) > 0 {
		log.Info("Adding custom configuration file from %s", setting.CustomConf)
		if err := addFile(w, "app.ini", setting.CustomConf, verbose); err != nil {
			fatal("Failed to include specified app.ini: %v", err)
		}
	}

	if ctx.IsSet("skip-custom-dir") && ctx.Bool("skip-custom-dir") {
		log.Info("Skiping custom directory")
	} else {
		customDir, err := os.Stat(setting.CustomPath)
		if err == nil && customDir.IsDir() {
			if is, _ := isSubdir(setting.AppDataPath, setting.CustomPath); !is {
				if err := addRecursiveExclude(w, "custom", setting.CustomPath, []string{absFileName}, verbose); err != nil {
					fatal("Failed to include custom: %v", err)
				}
			} else {
				log.Info("Custom dir %s is inside data dir %s, skipped", setting.CustomPath, setting.AppDataPath)
			}
		} else {
			log.Info("Custom dir %s doesn't exist, skipped", setting.CustomPath)
		}
	}

	isExist, err := util.IsExist(setting.AppDataPath)
	if err != nil {
		log.Error("Unable to check if %s exists. Error: %v", setting.AppDataPath, err)
	}
	if isExist {
		log.Info("Packing data directory...%s", setting.AppDataPath)

		var excludes []string
		if setting.Cfg.Section("session").Key("PROVIDER").Value() == "file" {
			var opts session.Options
			json := jsoniter.ConfigCompatibleWithStandardLibrary
			if err = json.Unmarshal([]byte(setting.SessionConfig.ProviderConfig), &opts); err != nil {
				return err
			}
			excludes = append(excludes, opts.ProviderConfig)
		}

		excludes = append(excludes, setting.LogRootPath)
		excludes = append(excludes, absFileName)
		if err := addRecursiveExclude(w, "data", setting.AppDataPath, excludes, verbose); err != nil {
			fatal("Failed to include data directory: %v", err)
		}
	}

	// Doesn't check if LogRootPath exists before processing --skip-log intentionally,
	// ensuring that it's clear the dump is skipped whether the directory's initialized
	// yet or not.
	if ctx.IsSet("skip-log") && ctx.Bool("skip-log") {
		log.Info("Skip dumping log files")
	} else {
		isExist, err := util.IsExist(setting.LogRootPath)
		if err != nil {
			log.Error("Unable to check if %s exists. Error: %v", setting.LogRootPath, err)
		}
		if isExist {
			if err := addRecursiveExclude(w, "log", setting.LogRootPath, []string{absFileName}, verbose); err != nil {
				fatal("Failed to include log: %v", err)
			}
		}
	}

	if fileName != "-" {
		if err = w.Close(); err != nil {
			_ = util.Remove(fileName)
			fatal("Failed to save %s: %v", fileName, err)
		}

		if err := os.Chmod(fileName, 0o600); err != nil {
			log.Info("Can't change file access permissions mask to 0600: %v", err)
		}
	}

	if fileName != "-" {
		log.Info("Finish dumping in file %s", fileName)
	} else {
		log.Info("Finish dumping to stdout")
	}

	return nil
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// addRecursiveExclude zips absPath to specified insidePath inside writer excluding excludeAbsPath
func addRecursiveExclude(w archiver.Writer, insidePath, absPath string, excludeAbsPath []string, verbose bool) error {
	absPath, err := filepath.Abs(absPath)
	if err != nil {
		return err
	}
	dir, err := os.Open(absPath)
	if err != nil {
		return err
	}
	defer dir.Close()

	files, err := dir.Readdir(0)
	if err != nil {
		return err
	}
	for _, file := range files {
		currentAbsPath := path.Join(absPath, file.Name())
		currentInsidePath := path.Join(insidePath, file.Name())
		if file.IsDir() {
			if !contains(excludeAbsPath, currentAbsPath) {
				if err := addFile(w, currentInsidePath, currentAbsPath, false); err != nil {
					return err
				}
				if err = addRecursiveExclude(w, currentInsidePath, currentAbsPath, excludeAbsPath, verbose); err != nil {
					return err
				}
			}
		} else {
			if err = addFile(w, currentInsidePath, currentAbsPath, verbose); err != nil {
				return err
			}
		}
	}
	return nil
}
