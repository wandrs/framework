// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package user

import (
	"fmt"
	"net/http"

	"go.wandrs.dev/framework/models"
	"go.wandrs.dev/framework/modules/context"
	"go.wandrs.dev/framework/modules/convert"
	"go.wandrs.dev/framework/modules/setting"
	api "go.wandrs.dev/framework/modules/structs"
	"go.wandrs.dev/framework/modules/web"
)

// ListEmails list all of the authenticated user's email addresses
// see https://github.com/gogits/go-gogs-client/wiki/Users-Emails#list-email-addresses-for-a-user
func ListEmails(ctx *context.APIContext) {
	// swagger:operation GET /user/emails user userListEmails
	// ---
	// summary: List the authenticated user's email addresses
	// produces:
	// - application/json
	// responses:
	//   "200":
	//     "$ref": "#/responses/EmailList"

	emails, err := models.GetEmailAddresses(ctx.User.ID)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "GetEmailAddresses", err)
		return
	}
	apiEmails := make([]*api.Email, len(emails))
	for i := range emails {
		apiEmails[i] = convert.ToEmail(emails[i])
	}
	ctx.JSON(http.StatusOK, &apiEmails)
}

// AddEmail add an email address
func AddEmail(ctx *context.APIContext) {
	// swagger:operation POST /user/emails user userAddEmail
	// ---
	// summary: Add email addresses
	// produces:
	// - application/json
	// parameters:
	// - name: options
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/CreateEmailOption"
	// parameters:
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/CreateEmailOption"
	// responses:
	//   '201':
	//     "$ref": "#/responses/EmailList"
	//   "422":
	//     "$ref": "#/responses/validationError"
	form := web.GetForm(ctx).(*api.CreateEmailOption)
	if len(form.Emails) == 0 {
		ctx.Error(http.StatusUnprocessableEntity, "", "Email list empty")
		return
	}

	emails := make([]*models.EmailAddress, len(form.Emails))
	for i := range form.Emails {
		emails[i] = &models.EmailAddress{
			UID:         ctx.User.ID,
			Email:       form.Emails[i],
			IsActivated: !setting.Service.RegisterEmailConfirm,
		}
	}

	if err := models.AddEmailAddresses(emails); err != nil {
		if models.IsErrEmailAlreadyUsed(err) {
			ctx.Error(http.StatusUnprocessableEntity, "", "Email address has been used: "+err.(models.ErrEmailAlreadyUsed).Email)
		} else if models.IsErrEmailInvalid(err) {
			errMsg := fmt.Sprintf("Email address %s invalid", err.(models.ErrEmailInvalid).Email)
			ctx.Error(http.StatusUnprocessableEntity, "", errMsg)
		} else {
			ctx.Error(http.StatusInternalServerError, "AddEmailAddresses", err)
		}
		return
	}

	apiEmails := make([]*api.Email, len(emails))
	for i := range emails {
		apiEmails[i] = convert.ToEmail(emails[i])
	}
	ctx.JSON(http.StatusCreated, &apiEmails)
}

// DeleteEmail delete email
func DeleteEmail(ctx *context.APIContext) {
	// swagger:operation DELETE /user/emails user userDeleteEmail
	// ---
	// summary: Delete email addresses
	// produces:
	// - application/json
	// parameters:
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/DeleteEmailOption"
	// responses:
	//   "204":
	//     "$ref": "#/responses/empty"
	//   "404":
	//     "$ref": "#/responses/notFound"
	form := web.GetForm(ctx).(*api.DeleteEmailOption)
	if len(form.Emails) == 0 {
		ctx.Status(http.StatusNoContent)
		return
	}

	emails := make([]*models.EmailAddress, len(form.Emails))
	for i := range form.Emails {
		emails[i] = &models.EmailAddress{
			Email: form.Emails[i],
			UID:   ctx.User.ID,
		}
	}

	if err := models.DeleteEmailAddresses(emails); err != nil {
		if models.IsErrEmailAddressNotExist(err) {
			ctx.Error(http.StatusNotFound, "DeleteEmailAddresses", err)
			return
		}
		ctx.Error(http.StatusInternalServerError, "DeleteEmailAddresses", err)
		return
	}
	ctx.Status(http.StatusNoContent)
}
