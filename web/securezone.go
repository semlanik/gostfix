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

	"git.semlanik.org/semlanik/gostfix/common"
)

func (s *Server) handleSecureZone(w http.ResponseWriter, r *http.Request, user string) {
	if user == "" {
		log.Printf("User could not be empty. Invalid usage of handleMailRequest")
		panic(nil)
	}
	s.error(http.StatusNotImplemented, "Admin panel is not implemented", w)
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request, user string) {
	if user == "" {
		log.Printf("User could not be empty. Invalid usage of handleMailRequest")
		panic(nil)
	}

	switch r.Method {
	case "GET":
		info, err := s.storage.GetUserInfo(user)
		if err != nil {
			s.error(http.StatusInternalServerError, "Unable to obtain user information", w)
			return
		}
		fmt.Fprintf(w, s.templater.ExecuteSettings(&struct {
			Version  string
			FullName string
		}{common.Version, info.FullName}))
	case "PATCH":
		s.handleSettingsUpdate(w, r, user)
	}
}

func (s *Server) handleSettingsUpdate(w http.ResponseWriter, r *http.Request, user string) {
	if err := r.ParseForm(); err != nil {
		s.error(http.StatusUnauthorized, "Password entered is invalid", w)
		return
	}

	oldPassword := r.FormValue("oldPassword")
	if err := s.authenticator.CheckUser(user, oldPassword); err != nil {
		s.error(http.StatusUnauthorized, "Password entered is invalid", w)
		return
	}

	password := r.FormValue("password")
	fullName := r.FormValue("fullName")

	err := s.storage.UpdateUser(user, password, fullName)
	if err != nil {
		log.Println(err.Error())
		s.error(http.StatusInternalServerError, "Unable to update user data", w)
		return
	}
	w.Write([]byte{0})
}
