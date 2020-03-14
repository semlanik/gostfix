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
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	template "html/template"
	"log"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	common "git.semlanik.org/semlanik/gostfix/common"
	"git.semlanik.org/semlanik/gostfix/config"
)

func (s *Server) handleMailbox(w http.ResponseWriter, user, email string) {
	fmt.Fprint(w, s.templater.ExecuteIndex(&struct {
		Folders template.HTML
		MailNew template.HTML
		Version template.HTML
	}{
		MailNew: template.HTML(s.templater.ExecuteNewMail("")),
		Folders: "Folders",
		Version: common.Version,
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
	case "folderStat":
		s.handleFolderStat(w, r, user, emails[mailbox])
	case "statusLine":
		s.handleStatusLine(w, user, emails[mailbox])
	case "mailList":
		s.handleMailList(w, r, user, emails[mailbox])
	case "sendNewMail":
		s.handleNewMail(w, r, user, emails[mailbox])
	default:
		http.Redirect(w, r, "/m0", http.StatusTemporaryRedirect)
	}
}

func (s *Server) handleFolders(w http.ResponseWriter, user, email string) {
	folders := s.storage.GetFolders(email)

	out, err := json.Marshal(&struct {
		Folders []*common.Folder `json:"folders"`
		Html    string           `json:"html"`
	}{
		Folders: folders,
		Html:    s.templater.ExecuteFolders(s.storage.GetFolders(email)),
	})

	if err != nil {
		s.error(http.StatusInternalServerError, "Could not fetch folder list", w)
	}

	w.Write(out)
}

func (s *Server) handleFolderStat(w http.ResponseWriter, r *http.Request, user, email string) {
	unread, total, err := s.storage.GetEmailStats(user, email, s.extractFolder(email, r))
	if err != nil {
		s.error(http.StatusInternalServerError, "Couldn't read mailbox stat", w)
		return
	}

	out, err := json.Marshal(&struct {
		Total  int `json:"total"`
		Unread int `json:"unread"`
	}{
		Total:  total,
		Unread: unread,
	})

	if err != nil {
		s.error(http.StatusInternalServerError, "Couldn't parse mailbox stat", w)
		return
	}

	w.Write(out)
}

func (s *Server) handleMailList(w http.ResponseWriter, r *http.Request, user, email string) {
	folder := s.extractFolder(email, r)
	page, err := strconv.Atoi(r.FormValue("page"))

	if err != nil {
		page = 0
	}

	_, total, err := s.storage.GetEmailStats(user, email, folder)
	if err != nil {
		s.error(http.StatusInternalServerError, "Couldn't read email database", w)
		return
	}

	mailList, err := s.storage.GetMailList(user, email, folder, common.Frame{Skip: int32(50 * page), Limit: 50})

	if err != nil {
		s.error(http.StatusInternalServerError, "Couldn't read email database", w)
		return
	}

	out, err := json.Marshal(&struct {
		Total int    `json:"total"`
		Html  string `json:"html"`
	}{
		Total: total,
		Html:  s.templater.ExecuteMailList(mailList),
	})
	if err != nil {
		s.error(http.StatusInternalServerError, "Could not perform maillist", w)
		return
	}
	w.Write(out)
}

func (s *Server) handleStatusLine(w http.ResponseWriter, user, email string) {
	info, err := s.storage.GetUserInfo(user)
	if err != nil {
		s.error(http.StatusInternalServerError, "Could not read user info", w)
		return
	}

	type EmailIndexes struct {
		Index int
		Email string
	}
	emails, err := s.storage.GetEmails(user)
	emailsIndexes := []EmailIndexes{}

	k := 0
	for i, existingEmail := range emails {
		emailsIndexes = append(emailsIndexes, EmailIndexes{i, existingEmail})

		if existingEmail == email {
			k = i
		}
	}

	emailsIndexes = emailsIndexes[:k+copy(emailsIndexes[k:], emailsIndexes[k+1:])]
	if err != nil {
		s.error(http.StatusInternalServerError, "Could not read user info", w)
		return
	}

	emailHash := md5.Sum([]byte(strings.Trim(email, "\t ")))
	fmt.Fprint(w, s.templater.ExecuteStatusLine(&struct {
		Name          string
		Email         string
		EmailHash     string
		EmailsIndexes []EmailIndexes
	}{
		Name:          info.FullName,
		Email:         email,
		EmailHash:     hex.EncodeToString(emailHash[:]),
		EmailsIndexes: emailsIndexes,
	}))
}

func (s *Server) extractFolder(email string, r *http.Request) string {
	folder := r.FormValue("folder")
	folders := s.storage.GetFolders(email)
	ok := false
	for _, existFolder := range folders {
		if folder == existFolder.Name {
			ok = true
			break
		}
	}

	if !ok {
		folder = common.Inbox
	}

	return folder
}

func (s *Server) handleNewMail(w http.ResponseWriter, r *http.Request, user, email string) {
	to := r.FormValue("to")
	resultEmail := s.templater.ExecuteMail(&struct {
		From    string
		Subject string
		Date    template.HTML
		To      string
		Body    template.HTML
	}{
		From:    email,
		To:      to,
		Subject: r.FormValue("subject"),
		Date:    template.HTML(time.Now().Format(time.RFC1123Z)),
		Body:    template.HTML(r.FormValue("body")),
	})

	host := config.ConfigInstance().MyDomain
	server := host + ":25"
	_, token := s.extractAuth(w, r)
	auth := smtp.PlainAuth("token", user, token, host)

	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	client, err := smtp.Dial(server)
	if err != nil {
		s.error(http.StatusInternalServerError, "Unable to send message", w)
		log.Printf("Dial %s \n", err)
		return
	}

	err = client.StartTLS(tlsconfig)
	if err != nil {
		s.error(http.StatusInternalServerError, "Unable to send message", w)
		log.Printf("StartTLS %s \n", err)
		return
	}

	err = client.Auth(auth)
	if err != nil {
		s.error(http.StatusInternalServerError, "Unable to send message", w)
		log.Printf("Auth %s \n", err)
		return
	}

	err = client.Mail(email)
	if err != nil {
		s.error(http.StatusInternalServerError, "Unable to send message", w)
		log.Printf("Mail %s \n", err)
		return
	}

	err = client.Rcpt(to)
	if err != nil {
		s.error(http.StatusInternalServerError, "Unable to send message", w)
		log.Println(err)
		return
	}

	mailWriter, err := client.Data()
	if err != nil {
		s.error(http.StatusInternalServerError, "Unable to send message", w)
		log.Println(err)
		return
	}

	_, err = mailWriter.Write([]byte(resultEmail))
	if err != nil {
		s.error(http.StatusInternalServerError, "Unable to send message", w)
		log.Println(err)
		return
	}

	err = mailWriter.Close()
	if err != nil {
		s.error(http.StatusInternalServerError, "Unable to send message", w)
		log.Println(err)
		return
	}

	client.Quit()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte{0})
}
