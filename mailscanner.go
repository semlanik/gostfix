package main

import (
	"fmt"
	ioutil "io/ioutil"
	"log"

	fsnotify "github.com/fsnotify/fsnotify"
)

type MailScanner struct {
	watcher *fsnotify.Watcher
}

// func fileExists(filename string) bool {
// 	info, err := os.Stat(filename)
// 	if os.IsNotExist(err) {
// 		return false
// 	}
// 	return err == nil && !info.IsDir() && info != nil
// }

func NewMailScanner(mailPath string) (ms *MailScanner) {
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
		if fileExists(f.Name()) {
			fmt.Printf("Add mail file %s for watching\n", f.Name())
			watcher.Add(f.Name())
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
