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

package setup

import (
	config "git.semlanik.org/semlanik/gostfix/config"
)

type Setup struct {
	sessionStore      *sessions.CookieStore
}

func NewSetup() *Setup {
	s := &Setup{
		sessionStore:      sessions.NewCookieStore(make([]byte, 32)),
	}
	return s
}

func (s *Setup) Run() {
	if !config.ConfigInstance().SetupEnabled {
		return
	}
	http.Handle("/", s)
	log.Fatal(http.ListenAndServe(":"+config.ConfigInstance().WebPort, nil))
}

func (s *Setup) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	urlParts := strings.Split(r.URL.Path, "/")[1:]
	if len(urlParts) == 0 || urlParts[0] == "" {
		// TODO: Welcome page with password
		return
	}
	
	if !checkPassword(r) {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	
	switch urlParts[1] {
	case "config":
		// TODO: Configure values
	case "save":
		// TODO: Confirm storing new config and reload server
	case "recovery":
		// TODO: Recovery mode, where user can resolve system inconsistency
	case "admin":
		// TODO: Admin panel handling
	}
}

func (s *Server) checkPassword(r *http.Request) bool {
	switch r.Method {
	case "GET":
		session, err := s.sessionStore.Get(r, CookieSessionToken)
		if err != nil {
			log.Printf("Unable to read user session %s\n", err)
			return false
		}
		setupPassword, _ = session.Values["setupPassword"].(string)
	case "POST":
		setupPassword := r.FormValue("setupPassword")
	}
	return masterPassword == config.ConfigInstance().SetupPassword;
}