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
	"strconv"

	common "git.semlanik.org/semlanik/gostfix/common"
)

func (s *Server) handleMailbox(w http.ResponseWriter, user, email string) {
	mailList, err := s.storage.MailList(user, email, "Inbox", common.Frame{Skip: 0, Limit: 50})

	if err != nil {
		s.error(http.StatusInternalServerError, "Couldn't read email database", w)
		return
	}

	fmt.Fprint(w, s.templater.ExecuteIndex(&struct {
		Folders  template.HTML
		MailList template.HTML
		Version  template.HTML
	}{
		MailList: template.HTML(s.templater.ExecuteMailList(mailList)),
		Folders:  "Folders",
		Version:  common.Version,
	}))
}

func (s *Server) handleMailboxRequest(path, user string, mailbox int, w http.ResponseWriter, r *http.Request) {
	log.Printf("Handle mailbox %s", path)
	emails, err := s.storage.GetEmails(user)

	if err != nil || len(emails) <= 0 {
		s.error(http.StatusInternalServerError, "Unable to access mailbox", w)
		return
	}

	if len(emails) <= mailbox {
		if path == "" {
			http.Redirect(w, r, "/m0", http.StatusTemporaryRedirect)
		} else {
			s.error(http.StatusInternalServerError, "Unable to access mailbox", w)
		}
		return
	}

	switch path {
	case "":
		s.handleMailbox(w, user, emails[mailbox])
	case "folders":
		s.handleFolders(w, user, emails[mailbox])
	case "statusLine":
		s.handleStatusLine(w, user, emails[mailbox])
	case "mailList":
		s.handleMailList(w, r, user, emails[mailbox])
	default:
		http.Redirect(w, r, "/m0", http.StatusTemporaryRedirect)
	}
}

func (s *Server) handleFolders(w http.ResponseWriter, user, email string) {
	fmt.Fprintf(w, s.templater.ExecuteFolders(s.storage.GetFolders(email)))
}

func (s *Server) handleMailList(w http.ResponseWriter, r *http.Request, user, email string) {
	folder := r.FormValue("folder")
	page, err := strconv.Atoi(r.FormValue("page"))
	if err != nil {
		page = 0
	}

	folders := s.storage.GetFolders(email)
	ok := false
	for _, existFolder := range folders {
		if folder == existFolder.Name {
			ok = true
			break
		}
	}

	if !ok {
		folder = "Inbox"
	}
	_, total, err := s.storage.GetEmailStats(user, email, folder)
	mailList, err := s.storage.MailList(user, email, folder, common.Frame{Skip: int32(50 * page), Limit: 50})

	if err != nil {
		s.error(http.StatusInternalServerError, "Couldn't read email database", w)
		return
	}

	out := fmt.Sprintf("{total: %d, html: \"%s\", }", total, s.templater.ExecuteMailList(mailList))
	fmt.Fprint(w, out)
}

func (s *Server) handleStatusLine(w http.ResponseWriter, user, email string) {
	info, err := s.storage.GetUserInfo(user)
	if err != nil {
		s.error(http.StatusInternalServerError, "Could not read user info", w)
		return
	}

	// unread, total, err := s.storage.GetEmailStats(user, email)
	// if err != nil {
	// 	s.error(http.StatusInternalServerError, "Could not read user stats", w)
	// 	return
	// }

	fmt.Fprint(w, s.templater.ExecuteStatusLine(&struct {
		Name  string
		Email string
	}{
		Name:  info.FullName,
		Email: email,
	}))
}
