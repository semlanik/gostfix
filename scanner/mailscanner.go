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
	ioutil "io/ioutil"
	"log"

	utils "../utils"
	fsnotify "github.com/fsnotify/fsnotify"
)

type MailScanner struct {
	watcher *fsnotify.Watcher
}

func NewMailScanner(mailPath string) (ms *MailScanner) {
	fmt.Printf("Add mail folder %s for watching\n", mailPath)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	ms = &MailScanner{
		watcher: watcher,
	}

	files, err := ioutil.ReadDir(mailPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		fullPath := mailPath + "/" + f.Name()
		if utils.FileExists(fullPath) {
			fmt.Printf("Add mail file %s for watching\n", fullPath)
			watcher.Add(fullPath)
		}
	}

	return
}

func (ms *MailScanner) Run() {
	go func() {
		for {
			select {
			case event, ok := <-ms.watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("New email for", event.Name)
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
