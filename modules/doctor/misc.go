// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package doctor

import (
	"fmt"
	"os/exec"

	"go.wandrs.dev/framework/modules/log"
	"go.wandrs.dev/framework/modules/setting"
)

func checkScriptType(logger log.Logger, autofix bool) error {
	path, err := exec.LookPath(setting.ScriptType)
	if err != nil {
		logger.Critical("ScriptType \"%q\" is not on the current PATH. Error: %v", setting.ScriptType, err)
		return fmt.Errorf("ScriptType \"%q\" is not on the current PATH. Error: %v", setting.ScriptType, err)
	}
	logger.Info("ScriptType %s is on the current PATH at %s", setting.ScriptType, path)
	return nil
}
func init() {
	Register(&Check{
		Title:     "Check if SCRIPT_TYPE is available",
		Name:      "script-type",
		IsDefault: false,
		Run:       checkScriptType,
		Priority:  5,
	})
}
