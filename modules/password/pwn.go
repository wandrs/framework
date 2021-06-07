// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package password

import (
	"context"

	"go.wandrs.dev/framework/modules/setting"

	"go.jolheiser.com/pwn"
)

// IsPwned checks whether a password has been pwned
// NOTE: This func returns true if it encounters an error under the assumption that you ALWAYS want to check against
// HIBP, so not getting a response should block a password until it can be verified.
func IsPwned(ctx context.Context, password string) (bool, error) {
	if !setting.PasswordCheckPwn {
		return false, nil
	}

	client := pwn.New(pwn.WithContext(ctx))
	count, err := client.CheckPassword(password, true)
	if err != nil {
		return true, err
	}

	return count > 0, nil
}
