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

	auth "git.semlanik.org/semlanik/gostfix/auth"
	common "git.semlanik.org/semlanik/gostfix/common"
	db "git.semlanik.org/semlanik/gostfix/db"
	utils "git.semlanik.org/semlanik/gostfix/utils"

	sessions "github.com/gorilla/sessions"
)

const (
	StateHeaderScan = iota
	StateBodyScan
	StateContentScan
)

const (
	AtLeastOneHeaderMask = 1 << iota
	FromHeaderMask
	DateHeaderMask
	ToHeaderMask
	AllHeaderMask = 15
)

const (
	CookieSessionToken = "gostfix_session"
)

func NewEmail() *common.Mail {
	return &common.Mail{
		Header: &common.MailHeader{},
		Body:   &common.MailBody{},
	}
}

type Server struct {
	authenticator *auth.Authenticator
	fileServer    http.Handler
	templater     *Templater
	sessionStore  *sessions.CookieStore
	storage       *db.Storage
}

func NewServer() *Server {

	storage, err := db.NewStorage()

	if err != nil {
		log.Fatalf("Unable to intialize mail storage %s", err)
		return nil
	}

	s := &Server{
		authenticator: auth.NewAuthenticator(),
		templater:     NewTemplater("data/templates"),
		fileServer:    http.FileServer(http.Dir("data")),
		sessionStore:  sessions.NewCookieStore(make([]byte, 32)),
		storage:       storage,
	}

	return s
}

func (s *Server) Run() {
	http.Handle("/", s)
	log.Fatal(http.ListenAndServe(":65200", nil))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Path)
	if utils.StartsWith(r.URL.Path, "/css/") ||
		utils.StartsWith(r.URL.Path, "/assets/") ||
		utils.StartsWith(r.URL.Path, "/js/") {
		s.fileServer.ServeHTTP(w, r)
	} else {
		switch r.URL.Path {
		case "/login":
			s.handleLogin(w, r)
		case "/logout":
			s.handleLogout(w, r)
		case "/messageDetails":
			s.handleMessageDetails(w, r)
		case "/statusLine":
			s.handleStatusLine(w, r)
		default:
			s.handleMailbox(w, r)
		}
	}
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	//Check passed in form login/password pair first
	if err := r.ParseForm(); err == nil {
		user := r.FormValue("user")
		password := r.FormValue("password")
		token, ok := s.authenticator.Authenticate(user, password)
		if ok {
			s.login(user, token, w, r)
			return
		}
	}

	//Check if user already logged in and entered login page accidently
	if s.authenticator.Verify(s.extractAuth(w, r)) {
		http.Redirect(w, r, "/mailbox", http.StatusTemporaryRedirect)
		return
	}

	//Otherwise make sure user logged out and show login page
	s.logout(w, r)
	fmt.Fprint(w, s.templater.ExecuteLogin(&struct {
		Version string
	}{common.Version}))
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	s.logout(w, r)
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}

func (s *Server) handleMessageDetails(w http.ResponseWriter, r *http.Request) {
	//TODO: Not implemented yet. Need database mail storage implemented first
	user, token := s.extractAuth(w, r)
	if !s.authenticator.Verify(user, token) {
		fmt.Fprint(w, "")
		return
	}

	messageId := r.FormValue("messageId")

	mail, err := s.storage.GetMail(user, messageId)
	if err != nil {
		s.error(http.StatusInternalServerError, "Unable to read message", w, r)
		return
	}

	s.storage.SetRead(user, messageId, true)
	fmt.Fprint(w, s.templater.ExecuteDetails(mail))
}

func (s *Server) handleStatusLine(w http.ResponseWriter, r *http.Request) {
	//TODO: Not implemented yet. Need database mail storage implemented first
	user, token := s.extractAuth(w, r)
	if !s.authenticator.Verify(user, token) {
		fmt.Fprint(w, "")
		return
	}

	fmt.Fprint(w, s.templater.ExecuteStatusLine(&struct {
		Name   string
		Read   int
		Unread int
	}{
		Name:   "No name", //TODO: read from database
		Read:   0,         //TODO: read from database
		Unread: 0,         //TODO: read from database
	}))
}

func (s *Server) handleMailbox(w http.ResponseWriter, r *http.Request) {
	user, token := s.extractAuth(w, r)
	if !s.authenticator.Verify(user, token) {
		s.logout(w, r)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	mailList, err := s.storage.MailList(user, user, "Inbox", common.Frame{Skip: 0, Limit: 0})

	if err != nil {
		s.error(http.StatusInternalServerError, "Couldn't read email database", w, r)
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

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	fmt.Println("logout")

	session, _ := s.sessionStore.Get(r, CookieSessionToken)
	session.Values["user"] = ""
	session.Values["token"] = ""
	session.Save(r, w)
}

func (s *Server) login(user, token string, w http.ResponseWriter, r *http.Request) {
	session, _ := s.sessionStore.Get(r, CookieSessionToken)
	session.Values["user"] = user
	session.Values["token"] = token
	session.Save(r, w)
	http.Redirect(w, r, "/mailbox", http.StatusTemporaryRedirect)
}

func (s *Server) error(code int, text string, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(code)
	fmt.Fprint(w, s.templater.ExecuteError(&struct {
		Code    int
		Text    string
		Version string
	}{
		Code:    code,
		Text:    "Unable to access your mailbox. Please contact Administrator.",
		Version: common.Version,
	}))
}

func (s *Server) extractAuth(w http.ResponseWriter, r *http.Request) (user, token string) {
	session, err := s.sessionStore.Get(r, CookieSessionToken)
	if err != nil {
		log.Printf("Unable to read user session %s\n", err)
		return
	}
	user, _ = session.Values["user"].(string)
	token, _ = session.Values["token"].(string)

	return
}
