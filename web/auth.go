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

	"git.semlanik.org/semlanik/gostfix/common"
	"git.semlanik.org/semlanik/gostfix/config"
	"git.semlanik.org/semlanik/gostfix/utils"
)

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	// if session, err := s.sessionStore.Get(r, CookieSessionToken); err == nil && session.Values["user"] != nil && session.Values["token"] != nil {
	// 	http.Redirect(w, r, "/m0", http.StatusTemporaryRedirect)
	// 	return
	// }
	if !config.ConfigInstance().RegistrationEnabled {
		s.error(http.StatusNotImplemented, "Registration is disabled on this server", w)
		return
	}

	//Check if user already logged in and entered register page accidently
	if s.authenticator.Verify(s.extractAuth(w, r)) {
		http.Redirect(w, r, "/m/0", http.StatusTemporaryRedirect)
		return
	}

	switch r.Method {
	case "GET":
		fmt.Fprint(w, s.templater.ExecuteRegister(&struct {
			Version string
			Domain  string
		}{common.Version, config.ConfigInstance().MyDomain}))
		return
	case "POST":
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

	s.error(http.StatusNotImplemented, "Invalid registration handling", w)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		//Check if user already logged in and entered login page accidently
		if s.authenticator.Verify(s.extractAuth(w, r)) {
			http.Redirect(w, r, "/m/0", http.StatusTemporaryRedirect)
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
	case "POST":
		//Check passed in form login/password pair first
		user := r.FormValue("user")
		password := r.FormValue("password")
		token, ok := s.authenticator.Login(user, password)
		if ok {
			s.login(user, token, w, r)
			return
		}
	}
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
	http.Redirect(w, r, "/m/0", http.StatusTemporaryRedirect)
}
