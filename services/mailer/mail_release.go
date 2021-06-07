// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package mailer

import (
	"bytes"

	"go.wandrs.dev/framework/models"
	"go.wandrs.dev/framework/modules/base"
	"go.wandrs.dev/framework/modules/log"
	"go.wandrs.dev/framework/modules/markup"
	"go.wandrs.dev/framework/modules/markup/markdown"
	"go.wandrs.dev/framework/modules/setting"
	"go.wandrs.dev/framework/modules/translation"
)

const (
	tplNewReleaseMail base.TplName = "release"
)

// MailNewRelease send new release notify to all all repo watchers.
func MailNewRelease(rel *models.Release) {
	watcherIDList, err := models.GetRepoWatchersIDs(rel.RepoID)
	if err != nil {
		log.Error("GetRepoWatchersIDs(%d): %v", rel.RepoID, err)
		return
	}

	recipients, err := models.GetMaileableUsersByIDs(watcherIDList, false)
	if err != nil {
		log.Error("models.GetMaileableUsersByIDs: %v", err)
		return
	}

	langMap := make(map[string][]string)
	for _, user := range recipients {
		if user.ID != rel.PublisherID {
			langMap[user.Language] = append(langMap[user.Language], user.Email)
		}
	}

	for lang, tos := range langMap {
		mailNewRelease(lang, tos, rel)
	}
}

func mailNewRelease(lang string, tos []string, rel *models.Release) {
	locale := translation.NewLocale(lang)

	var err error
	rel.RenderedNote, err = markdown.RenderString(&markup.RenderContext{
		URLPrefix: rel.Repo.Link(),
		Metas:     rel.Repo.ComposeMetas(),
	}, rel.Note)
	if err != nil {
		log.Error("markdown.RenderString(%d): %v", rel.RepoID, err)
		return
	}

	subject := locale.Tr("mail.release.new.subject", rel.TagName, rel.Repo.FullName())
	mailMeta := map[string]interface{}{
		"Release":  rel,
		"Subject":  subject,
		"i18n":     locale,
		"Language": locale.Language(),
	}

	var mailBody bytes.Buffer

	// TODO: i18n templates?
	if err := bodyTemplates.ExecuteTemplate(&mailBody, string(tplNewReleaseMail), mailMeta); err != nil {
		log.Error("ExecuteTemplate [%s]: %v", string(tplNewReleaseMail)+"/body", err)
		return
	}

	msgs := make([]*Message, 0, len(tos))
	publisherName := rel.Publisher.DisplayName()
	relURL := "<" + rel.HTMLURL() + ">"
	for _, to := range tos {
		msg := NewMessageFrom([]string{to}, publisherName, setting.MailService.FromEmail, subject, mailBody.String())
		msg.Info = subject
		msg.SetHeader("Message-ID", relURL)
		msgs = append(msgs, msg)
	}

	SendAsyncs(msgs)
}
