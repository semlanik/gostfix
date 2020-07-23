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
	"html/template"
	"log"
	"net/http"
	"strconv"

	auth "git.semlanik.org/semlanik/gostfix/auth"
	common "git.semlanik.org/semlanik/gostfix/common"
	"git.semlanik.org/semlanik/gostfix/config"
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

type Server struct {
	authenticator     *auth.Authenticator
	fileServer        http.Handler
	attachmentsServer http.Handler
	templater         *Templater
	sessionStore      *sessions.CookieStore
	storage           *db.Storage
	Notifier          *webNotifier
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
		Notifier:          NewWebNotifier(),
		scanner:           scanner,
	}

	return s
}

func (s *Server) Run() {
	http.Handle("/", s)
	log.Fatal(http.ListenAndServe(":"+config.ConfigInstance().WebPort, nil))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Path)
	if utils.StartsWith(r.URL.Path, "/css/") ||
		utils.StartsWith(r.URL.Path, "/assets/") ||
		utils.StartsWith(r.URL.Path, "/js/") {
		s.fileServer.ServeHTTP(w, r)
	} else if utils.StartsWith(r.URL.Path, "/attachment") {
		s.attachmentsServer.ServeHTTP(w, r)
	} else if cap := utils.RegExpUtilsInstance().MailboxFinder.FindStringSubmatch(r.URL.Path); len(cap) == 3 {
		user, token := s.extractAuth(w, r)
		if !s.authenticator.Verify(user, token) {
			s.logout(w, r)
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		mailbox, err := strconv.Atoi(cap[1])
		if err != nil || mailbox < 0 {
			http.Redirect(w, r, "/m0", http.StatusTemporaryRedirect)
			return
		}

		path := cap[2]

		s.handleMailboxRequest(path, user, mailbox, w, r)
	} else {
		switch r.URL.Path {
		case "/login":
			s.handleLogin(w, r)
		case "/logout":
			s.handleLogout(w, r)
		case "/register":
			s.handleRegister(w, r)
		case "/checkEmail":
			s.handleCheckEmail(w, r)
		case "/mail":
			fallthrough
		case "/setRead":
			fallthrough
		case "/remove":
			fallthrough
		case "/restore":
			fallthrough
		case "/delete":
			s.handleMailRequest(w, r)
		case "/settings":
			fallthrough
		case "/update":
			fallthrough
		case "/admin":
			s.handleSecureZone(w, r)
		default:
			http.Redirect(w, r, "/m0", http.StatusTemporaryRedirect)
		}
	}
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	// if session, err := s.sessionStore.Get(r, CookieSessionToken); err == nil && session.Values["user"] != nil && session.Values["token"] != nil {
	// 	http.Redirect(w, r, "/m0", http.StatusTemporaryRedirect)
	// 	return
	// }
	if !config.ConfigInstance().RegistrationEnabled {
		s.error(http.StatusNotImplemented, "Registration is disabled on this server", w)
		return
	}

	if err := r.ParseForm(); err == nil {
		user := r.FormValue("user")
		password := r.FormValue("password")
		fullName := r.FormValue("fullName")
		if user != "" && password != "" && fullName != "" {
			ok, email := s.checkEmail(user)
			if ok && len(password) < 128 && len(fullName) < 128 && utils.RegExpUtilsInstance().FullNameChecker.MatchString(fullName) {
				err := s.storage.AddUser(email, password, fullName)
				if err != nil {
					log.Println(err.Error())
					s.error(http.StatusInternalServerError, "Unable to create user", w)
					return
				}

				s.scanner.Reconfigure()
				token, _ := s.authenticator.Login(email, password)
				s.login(email, token, w, r)
				return
			}
		}
	}

	fmt.Fprint(w, s.templater.ExecuteRegister(&struct {
		Version string
		Domain  string
	}{common.Version, config.ConfigInstance().MyDomain}))
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	//Check passed in form login/password pair first
	if err := r.ParseForm(); err == nil {
		user := r.FormValue("user")
		password := r.FormValue("password")
		token, ok := s.authenticator.Login(user, password)
		if ok {
			s.login(user, token, w, r)
			return
		}
	}

	//Check if user already logged in and entered login page accidently
	if s.authenticator.Verify(s.extractAuth(w, r)) {
		http.Redirect(w, r, "/m0", http.StatusTemporaryRedirect)
		return
	}

	var signupTemplate template.HTML
	if config.ConfigInstance().RegistrationEnabled {
		signupTemplate = template.HTML(s.templater.ExecuteSignup(""))
	} else {
		signupTemplate = ""
	}

	//Otherwise make sure user logged out and show login page
	s.logout(w, r)
	fmt.Fprint(w, s.templater.ExecuteLogin(&struct {
		Version string
		Signup  template.HTML
	}{common.Version, signupTemplate}))
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	s.logout(w, r)
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}

func (s *Server) handleCheckEmail(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err == nil {
		if ok, _ := s.checkEmail(r.FormValue("user")); ok {
			w.Write([]byte{0})
			return
		}
		s.error(http.StatusNotAcceptable, "Email exists", w)
		return
	}
	s.error(http.StatusBadRequest, "Invalid arguments", w)
	return
}

func (s *Server) checkEmail(user string) (bool, string) {
	email := user + "@" + config.ConfigInstance().MyDomain
	return utils.RegExpUtilsInstance().EmailChecker.MatchString(email) && !s.storage.CheckEmailExists(email), email
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	session, err := s.sessionStore.Get(r, CookieSessionToken)
	if err == nil {
		if session.Values["user"] != nil && session.Values["token"] != nil {
			s.authenticator.Logout(session.Values["user"].(string), session.Values["token"].(string))
		}
		session.Values["user"] = ""
		session.Values["token"] = ""
		session.Save(r, w)
	}
}

func (s *Server) login(user, token string, w http.ResponseWriter, r *http.Request) {
	session, _ := s.sessionStore.Get(r, CookieSessionToken)
	session.Values["user"] = user
	session.Values["token"] = token
	session.Save(r, w)
	http.Redirect(w, r, "/m0", http.StatusTemporaryRedirect)
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
