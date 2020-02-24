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
	"bufio"
	"log"
	"os"
	"strings"

	config "../config"
	utils "../utils"
)

type Authenticator struct {
	mailMaps map[string]string //TODO: temporary here. Later should be part of mailscanner and never accessed from here
}

func NewAuthenticator() (a *Authenticator) {
	a = &Authenticator{
		mailMaps: readMailMaps(), //TODO: temporary here. Later should be part of mailscanner and never accessed from here
	}
	return
}

func (a *Authenticator) Authenticate(user, password string) (string, bool) {
	if !utils.RegExpUtilsInstance().EmailChecker.MatchString(user) {
		return "", false
	}
	_, ok := a.mailMaps[user]

	return "", ok
}

func (a *Authenticator) Verify(user, token string) bool {
	if !utils.RegExpUtilsInstance().EmailChecker.MatchString(user) {
		return false
	}
	_, ok := a.mailMaps[user]

	return ok
}

func (a *Authenticator) MailPath(user string) string { //TODO: temporary here. Later should be part of mailscanner and never accessed from here
	return a.mailMaps[user]
}

func readMailMaps() map[string]string { //TODO: temporary here. Later should be part of mailscanner and never accessed from here
	mailMaps := make(map[string]string)
	mapsFile := config.ConfigInstance().VMailboxMaps
	if !utils.FileExists(mapsFile) {
		return mailMaps
	}

	file, err := os.Open(mapsFile)
	if err != nil {
		log.Fatalf("Unable to open virtual mailbox maps %s\n", mapsFile)
	}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		mailPathPair := strings.Split(scanner.Text(), " ")
		if len(mailPathPair) != 2 {
			log.Printf("Invalid record in virtual mailbox maps %s", scanner.Text())
			continue
		}
		mailMaps[mailPathPair[0]] = mailPathPair[1]
	}

	return mailMaps
}
