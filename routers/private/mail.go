// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package private

import (
	"fmt"
	"net/http"
	"strconv"

	"go.wandrs.dev/framework/models"
	"go.wandrs.dev/framework/modules/context"
	"go.wandrs.dev/framework/modules/log"
	"go.wandrs.dev/framework/modules/private"
	"go.wandrs.dev/framework/modules/setting"
	"go.wandrs.dev/framework/services/mailer"

	jsoniter "github.com/json-iterator/go"
)

// SendEmail pushes messages to mail queue
//
// It doesn't wait before each message will be processed
func SendEmail(ctx *context.PrivateContext) {
	if setting.MailService == nil {
		ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"err": "Mail service is not enabled.",
		})
		return
	}

	var mail private.Email
	rd := ctx.Req.Body
	defer rd.Close()
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	if err := json.NewDecoder(rd).Decode(&mail); err != nil {
		log.Error("%v", err)
		ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"err": err,
		})
		return
	}

	var emails []string
	if len(mail.To) > 0 {
		for _, uname := range mail.To {
			user, err := models.GetUserByName(uname)
			if err != nil {
				err := fmt.Sprintf("Failed to get user information: %v", err)
				log.Error(err)
				ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
					"err": err,
				})
				return
			}

			if user != nil && len(user.Email) > 0 {
				emails = append(emails, user.Email)
			}
		}
	} else {
		err := models.IterateUser(func(user *models.User) error {
			if len(user.Email) > 0 {
				emails = append(emails, user.Email)
			}
			return nil
		})
		if err != nil {
			err := fmt.Sprintf("Failed to find users: %v", err)
			log.Error(err)
			ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
				"err": err,
			})
			return
		}
	}

	sendEmail(ctx, mail.Subject, mail.Message, emails)
}

func sendEmail(ctx *context.PrivateContext, subject, message string, to []string) {
	for _, email := range to {
		msg := mailer.NewMessage([]string{email}, subject, message)
		mailer.SendAsync(msg)
	}

	wasSent := strconv.Itoa(len(to))

	ctx.PlainText(http.StatusOK, []byte(wasSent))
}
