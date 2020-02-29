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

package main

import (
	"git.semlanik.org/semlanik/gostfix/db"
	scanner "git.semlanik.org/semlanik/gostfix/scanner"
	web "git.semlanik.org/semlanik/gostfix/web"
)

type GofixEngine struct {
	scanner *scanner.MailScanner
	web     *web.Server
}

func NewGofixEngine() (e *GofixEngine) {
	e = &GofixEngine{
		scanner: scanner.NewMailScanner(),
		web:     web.NewServer(),
	}

	return
}

func (e *GofixEngine) Run() {
	defer e.scanner.Stop()
	e.scanner.Run()
	e.web.Run()
}

func main() {
	//Bad
	storage, _ := db.NewStorage()
	storage.AddUser("semlanik@semlanik.org", "test", "Alexey Edelev")
	storage.AddUser("junkmail@semlanik.org", "test", "Alexey Edelev")
	storage.AddUser("git@semlanik.org", "test", "Alexey Edelev")
	storage.AddEmail("semlanik@semlanik.org", "ci@semlanik.org")
	storage.AddEmail("semlanik@semlanik.org", "shopping@semlanik.org")
	storage.AddEmail("semlanik@semlanik.org", "junkmail@semlanik.org")
	storage.AddEmail("junkmail@semlanik.org", "qqqqq@semlanik.org")
	storage.AddEmail("junkmail@semlanik.org", "main@semlanik.org")
	engine := NewGofixEngine()
	engine.Run()
}
