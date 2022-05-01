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
	"log"

	"git.semlanik.org/semlanik/gostfix/common"
	"github.com/fsnotify/fsnotify"
)

func (ms *MailScanner) ServiceName() string {
	return "Mail Scanner"
}

func (ms *MailScanner) Run() {
	ms.reconfigure()

	for {
		select {
		case signal := <-ms.signalChannel:
			ms.handleSignal(signal)
		case event, ok := <-ms.watcher.Events:
			if !ok {
				log.Printf("Unable to read mail watcher event. Mail scanner restart is needed.")
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
					for _, mail := range mails {
						ms.storage.SaveMail(mailbox, common.Inbox, mail, false)
					}
					log.Printf("New email for %s, emails read %d", mailPath, len(mails))
				} else {
					log.Printf("Invalid path update triggered: %s", mailPath)
				}

			}
		case err, ok := <-ms.watcher.Errors:
			if !ok {
				log.Printf("Unable to read mail watcher errors. Mail scanner restart is needed.")
				return
			}
			log.Println("Mail watcher error:", err)
		}
	}
	defer ms.watcher.Close()
}

func (ms *MailScanner) Stop() {
	ms.signalChannel <- SignalStop
}

func (ms *MailScanner) Error() string {
	return ""
}
