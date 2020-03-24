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

package auth

import (
	"log"

	db "git.semlanik.org/semlanik/gostfix/db"
	utils "git.semlanik.org/semlanik/gostfix/utils"
	uuid "github.com/google/uuid"
)

type Authenticator struct {
	storage *db.Storage
}

func NewAuthenticator() (a *Authenticator) {
	storage, err := db.NewStorage()

	if err != nil {
		log.Fatalf("Unable to intialize user storage %s", err)
		return nil
	}

	a = &Authenticator{
		storage: storage,
	}
	return
}

func (a *Authenticator) Authenticate(user, password string) (string, bool) {
	if !utils.RegExpUtilsInstance().EmailChecker.MatchString(user) {
		return "", false
	}

	if a.storage.CheckUser(user, password) != nil {
		return "", false
	}

	token := uuid.New().String()
	a.storage.AddToken(user, token)
	return token, true
}

func (a *Authenticator) Verify(user, token string) bool {
	if !utils.RegExpUtilsInstance().EmailChecker.MatchString(user) {
		return false
	}

	return a.storage.CheckToken(user, token) == nil
}
