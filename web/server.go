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
	"log"
	"net/http"
	"strings"

	auth "git.semlanik.org/semlanik/gostfix/auth"
	common "git.semlanik.org/semlanik/gostfix/common"
	"git.semlanik.org/semlanik/gostfix/config"
	db "git.semlanik.org/semlanik/gostfix/db"

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

type Server struct {
	authenticator     *auth.Authenticator
	fileServer        http.Handler
	attachmentsServer http.Handler
	templater         *Templater
	sessionStore      *sessions.CookieStore
	storage           *db.Storage
	notifier          *webNotifier
	scanner           common.Scanner
}

func NewServer(scanner common.Scanner) *Server {

	storage, err := db.NewStorage()

	if err != nil {
		log.Fatalf("Unable to intialize mail storage %s", err)
		return nil
	}

	authenticator, err := auth.NewAuthenticator()
	if err != nil {
		log.Fatalf("Unable to intialize authenticator %s", err)
		return nil
	}
	s := &Server{
		authenticator:     authenticator,
		templater:         NewTemplater("data/templates"),
		fileServer:        http.FileServer(http.Dir("data")),
		attachmentsServer: http.StripPrefix("/attachment/", http.FileServer(http.Dir(config.ConfigInstance().AttachmentsPath))),
		sessionStore:      sessions.NewCookieStore(make([]byte, 32)),
		storage:           storage,
		notifier:          NewWebNotifier(),
		scanner:           scanner,
	}

	s.notifier.server = s
	s.storage.RegisterNotifier(s.notifier)

	return s
}

func (s *Server) Run() {
	http.Handle("/", s)
	log.Fatal(http.ListenAndServe(":"+config.ConfigInstance().WebPort, nil))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)
	urlParts := strings.Split(r.URL.Path, "/")[1:]
	if len(urlParts) == 0 || urlParts[0] == "" {
		http.Redirect(w, r, "/m/0", http.StatusTemporaryRedirect)
		return
	}

	switch urlParts[0] {
	case "css":
		fallthrough
	case "assets":
		fallthrough
	case "js":
		s.fileServer.ServeHTTP(w, r)
	case "login":
		s.handleLogin(w, r)
	case "logout":
		s.handleLogout(w, r)
	case "register":
		s.handleRegister(w, r)
	case "checkEmail":
		s.handleCheckEmail(w, r)
	default:
		s.handleSecure(w, r, urlParts)
	}
}

func (s *Server) handleSecure(w http.ResponseWriter, r *http.Request, urlParts []string) {
	user, token := s.extractAuth(w, r)
	if !s.authenticator.Verify(user, token) {
		if r.Method == "GET" && urlParts[0] == "m" {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		} else {
			s.error(http.StatusUnauthorized, "You are not allowed to access this function", w)
		}
		return
	}

	switch urlParts[0] {
	case "m":
		s.handleMailboxRequest(w, r, user, urlParts)
	case "attachment":
		if len(urlParts) == 2 {
			s.handleAttachment(w, r, user, urlParts[1])
		} else {
			s.error(http.StatusBadRequest, "Invalid attachments request", w)
		}
	case "mail":
		if len(urlParts) == 2 {
			s.handleMailRequest(w, r, user, urlParts[1])
		}
	case "settings":
		s.handleSettings(w, r, user)
	case "admin":
		s.handleSecureZone(w, r, user)
	default:
		http.Redirect(w, r, "/m/0", http.StatusTemporaryRedirect)
	}
}

func (s *Server) handleAttachment(w http.ResponseWriter, r *http.Request, user, attachment string) {
	if user == "" {
		log.Printf("User could not be empty. Invalid usage of handleMailRequest")
		panic(nil)
	}

	if r.Method != "GET" {
		s.error(http.StatusNotImplemented, "You only may download attachments", w)
		return
	}

	if !s.storage.CheckAttachment(user, attachment) {
		s.error(http.StatusNotFound, "Attachment not found", w)
		return
	}

	s.attachmentsServer.ServeHTTP(w, r)
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

func (s *Server) error(code int, text string, w http.ResponseWriter) {
	w.WriteHeader(code)
	fmt.Fprint(w, s.templater.ExecuteError(&struct {
		Code    int
		Text    string
		Version string
	}{
		Code:    code,
		Text:    text,
		Version: common.Version,
	}))
}
