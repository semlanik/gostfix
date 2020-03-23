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
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"git.semlanik.org/semlanik/gostfix/common"
	config "git.semlanik.org/semlanik/gostfix/config"
	db "git.semlanik.org/semlanik/gostfix/db"
	utils "git.semlanik.org/semlanik/gostfix/utils"
	fsnotify "github.com/fsnotify/fsnotify"
)

const (
	SignalReconfigure = iota
)

type MailScanner struct {
	watcher       *fsnotify.Watcher
	emailMaps     map[string]string
	storage       *db.Storage
	signalChannel chan int
	notifiers     []common.Notifier
	notifiersLock sync.Mutex
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
		notifiers:     []common.Notifier{},
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

func (ms *MailScanner) readEmailMaps() {
	registredEmails, err := ms.storage.GetAllEmails()
	if err != nil {
		log.Fatal(err)
		return
	}

	mailPath := config.ConfigInstance().VMailboxBase

	emailMaps := make(map[string]string)
	mapsFile := config.ConfigInstance().VMailboxMaps
	if !utils.FileExists(mapsFile) {
		log.Fatal("Could not read virtual mailbox maps")
		return
	}

	file, err := os.Open(mapsFile)
	if err != nil {
		log.Fatalf("Unable to open virtual mailbox maps %s\n", mapsFile)
	}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		emailMapPair := strings.Split(scanner.Text(), " ")
		if len(emailMapPair) != 2 {
			log.Printf("Invalid record in virtual mailbox maps %s\n", scanner.Text())
			continue
		}

		found := false
		email := emailMapPair[0]
		for _, registredEmail := range registredEmails {
			if email == registredEmail {
				found = true
			}
		}
		if !found {
			log.Fatalf("Found non-registred mailbox <%s> in mail maps. Database has inconsistancy.\n", email)
			return
		}
		emailMaps[email] = mailPath + "/" + emailMapPair[1]
	}

	for _, registredEmail := range registredEmails {
		if _, exists := emailMaps[registredEmail]; !exists {
			log.Fatalf("Found existing mailbox <%s> in database. Mail maps has inconsistancy.\n", registredEmail)
		}
	}
	ms.emailMaps = emailMaps

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

func (ms *MailScanner) Run() {
	go func() {
		ms.readEmailMaps()

		for {
			select {
			case signal := <-ms.signalChannel:
				switch signal {
				case SignalReconfigure:
					ms.readEmailMaps()
				}
			case event, ok := <-ms.watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("iNotify write")

					mailPath := event.Name
					mailbox := ""
					for k, v := range ms.emailMaps {
						if v == mailPath {
							mailbox = k
						}
					}

					if mailbox != "" {
						mails := ms.readMailFile(mailPath)
						if len(mails) > 0 {
							ms.notifyMailboxUpdate(mailbox)
						}
						for _, mail := range mails {
							ms.storage.SaveMail(mailbox, common.Inbox, mail, false)
							ms.notifyNewMail(mailbox, *mail)
						}
						log.Printf("New email for %s, emails read %d", mailPath, len(mails))
					} else {
						log.Printf("Invalid path update triggered: %s", mailPath)
					}

				}
			case err, ok := <-ms.watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()
}

func (ms *MailScanner) Stop() {
	defer ms.watcher.Close()
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

func (ms *MailScanner) RegisterNotifier(notifier common.Notifier) {
	if notifier != nil {
		ms.notifiersLock.Lock()
		defer ms.notifiersLock.Unlock()
		ms.notifiers = append(ms.notifiers, notifier)
	}
}

func (ms *MailScanner) notifyNewMail(email string, mail common.Mail) {
	ms.notifiersLock.Lock()
	defer ms.notifiersLock.Unlock()
	for _, notifier := range ms.notifiers {
		notifier.NotifyNewMail(email, mail)
	}
}

func (ms *MailScanner) notifyMailboxUpdate(email string) {
	ms.notifiersLock.Lock()
	defer ms.notifiersLock.Unlock()
	for _, notifier := range ms.notifiers {
		notifier.NotifyMaiboxUpdate(email)
	}
}
