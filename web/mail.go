/*
 * MIT License
 *
 * Copyright (c) 2020 Alexey Edelev <semlanik@gmail.com>
 *
 * This file is part of gostfix project https://git.semlanik.org/semlanik/gostfix
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this
 * software and associated documentation files (the "Software"), to deal in the Software
 * without restriction, including without limitation the rights to use, copy, modify,
 * merge, publish, distribute, sublicense, and/or sell copies of the Software, and
 * to permit persons to whom the Software is furnished to do so, subject to the following
 * conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies
 * or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED,
 * INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR
 * PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE
 * FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR
 * OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
 * DEALINGS IN THE SOFTWARE.
 */

package web

import (
	"fmt"
	template "html/template"
	"log"
	"net/http"
	"strings"

	"git.semlanik.org/semlanik/gostfix/common"
)

func (s *Server) handleMailRequest(w http.ResponseWriter, r *http.Request) {
	user, token := s.extractAuth(w, r)
	if !s.authenticator.Verify(user, token) {
		s.error(http.StatusUnauthorized, "You are not allowed to access this function", w)
		return
	}

	mailId := r.FormValue("mailId")

	if mailId == "" {
		s.error(http.StatusBadRequest, "Invalid mail id requested", w)
		return
	}

	switch r.URL.Path {
	case "/mail":
		s.handleMailDetails(w, user, mailId)
	case "/setRead":
		s.handleSetRead(w, r, user, mailId)
	case "/remove":
		s.handleRemove(w, user, mailId)
	case "/restore":
		s.handleRestore(w, user, mailId)
	case "/delete":
		s.handleDelete(w, user, mailId)
	}
}

func (s *Server) handleMailDetails(w http.ResponseWriter, user, mailId string) {
	mail, err := s.storage.GetMail(user, mailId)
	if err != nil {
		s.error(http.StatusInternalServerError, "Unable to read mail", w)
		return
	}

	text := mail.Mail.Body.RichText
	if text == "" {
		text = strings.Replace(mail.Mail.Body.PlainText, "\n", "</br>", -1)
	}

	s.storage.SetRead(user, mailId, true)
	fmt.Fprint(w, s.templater.ExecuteDetails(&struct {
		From    string
		To      string
		Subject string
		Text    template.HTML
		MailId  string
		Read    bool
		Trash   bool
	}{
		From:    mail.Mail.Header.From,
		To:      mail.Mail.Header.To,
		Subject: mail.Mail.Header.Subject,
		Text:    template.HTML(text),
		MailId:  mailId,
		Read:    false,
		Trash: mail.Trash ||
			mail.Folder == common.Trash, //TODO: Legacy for old databases remove soon
	}))
}

func (s *Server) handleSetRead(w http.ResponseWriter, r *http.Request, user, mailId string) {
	read := r.FormValue("read") == "true"
	s.storage.SetRead(user, mailId, read)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte{})
}

func (s *Server) handleRemove(w http.ResponseWriter, user, mailId string) {
	err := s.storage.MoveMail(user, mailId, common.Trash)
	if err != nil {
		s.error(http.StatusInternalServerError, "Could not move email to trash", w)
	}
}

func (s *Server) handleRestore(w http.ResponseWriter, user, mailId string) {
	err := s.storage.RestoreMail(user, mailId)
	if err != nil {
		s.error(http.StatusInternalServerError, "Could not move email to trash", w)
	}
}

func (s *Server) handleDelete(w http.ResponseWriter, user, mailId string) {
	log.Printf("Delete mail")
	err := s.storage.DeleteMail(user, mailId)
	if err != nil {
		s.error(http.StatusInternalServerError, "Could not delete email", w)
	}
}