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

package scanner

import (
	"fmt"
	"log"
	"os"

	"git.semlanik.org/semlanik/gostfix/common"
	config "git.semlanik.org/semlanik/gostfix/config"
	db "git.semlanik.org/semlanik/gostfix/db"
	utils "git.semlanik.org/semlanik/gostfix/utils"
	fsnotify "github.com/fsnotify/fsnotify"
)

const (
	SignalReconfigure = iota
	SignalStop
)

type MailScanner struct {
	watcher       *fsnotify.Watcher
	emailMaps     map[string]string
	storage       *db.Storage
	signalChannel chan int
}

func NewMailScanner() (ms *MailScanner) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
		return
	}
	storage, err := db.NewStorage()
	if err != nil {
		log.Fatal(err)
		return
	}

	if !utils.DirectoryExists(config.ConfigInstance().AttachmentsPath) {
		err = os.Mkdir(config.ConfigInstance().AttachmentsPath, 0755)
		if err != nil {
			log.Fatal(err)
			return
		}
	}

	ms = &MailScanner{
		watcher:       watcher,
		storage:       storage,
		signalChannel: make(chan int),
	}

	return
}

func (ms *MailScanner) Reconfigure() {
	ms.signalChannel <- SignalReconfigure
}

func (ms *MailScanner) checkEmailRegistred(email string) bool {
	emails, err := ms.storage.GetAllEmails()

	if err != nil {
		return false
	}

	for _, e := range emails {
		if email == e {
			return true
		}
	}

	return false
}

func (ms *MailScanner) reconfigure() {
	log.Printf("Reconfiguring mail scanner")
	var err error
	ms.emailMaps, err = ms.storage.ReadEmailMaps()
	if err != nil {
		log.Fatal(err.Error())
	}

	for mailbox, mailPath := range ms.emailMaps {
		if !utils.FileExists(mailPath) {
			file, err := os.Create(mailPath)
			if err != nil {
				fmt.Printf("Unable to create mailbox for watching %s\n", err)
				continue
			}
			file.Close()
		}

		mails := ms.readMailFile(mailPath)
		for _, mail := range mails {
			ms.storage.SaveMail(mailbox, common.Inbox, mail, false)
		}
		log.Printf("New email for %s, emails read %d", mailPath, len(mails))

		err := ms.watcher.Add(mailPath)
		if err != nil {
			fmt.Printf("Unable to add mailbox for watching\n")
		} else {
			fmt.Printf("Add mail file %s for watching\n", mailPath)
		}
	}
}

func (ms *MailScanner) handleSignal(signal int) {
	switch signal {
	case SignalReconfigure:
		ms.reconfigure()
	}
}

func (ms *MailScanner) readMailFile(mailPath string) (mails []*common.Mail) {
	log.Println("Read mail file")
	defer log.Println("Exit read mail file")
	if !utils.FileExists(mailPath) {
		return nil
	}

	file, err := utils.OpenAndLockWait(mailPath)
	if err != nil {
		return nil
	}
	defer file.CloseAndUnlock()

	mails = parseFile(file)
	if len(mails) > 0 {
		file.Truncate(0)
	}

	return mails
}
