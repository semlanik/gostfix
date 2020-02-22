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
	"bufio"
	"fmt"
	template "html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	unix "golang.org/x/sys/unix"
)

const (
	StateHeaderScan = iota
	StateBodyScan
	StateContentScan
)

const (
	HeaderRegExp        = "^([\x21-\x7E^:]+):(.*)"
	FoldingRegExp       = "^\\s+(.*)"
	BoundaryStartRegExp = "^--(.*)"
	BoundaryEndRegExp   = "^--(.*)--$"
	BoundaryRegExp      = "boundary=\"(.*)\""
	UserRegExp          = "^[a-zA-Z][\\w0-9\\._]*"
)

// type Email struct {
// 	From        string
// 	To          string
// 	Cc          string
// 	Bcc         string
// 	Date        string
// 	Subject     string
// 	ContentType string
// 	Body        string
// }

func NewEmail() *Mail {
	return &Mail{
		Header: &MailHeader{},
		Body: &MailBody{
			ContentType: "plain/text",
		},
	}
}

type GofixEngine struct {
	templater   *Templater
	fileServer  http.Handler
	userChecker *regexp.Regexp
	scanner     *MailScanner
	mailPath    string
}

func NewGofixEngine(mailPath string) (e *GofixEngine) {
	e = &GofixEngine{
		templater:  NewTemplater("templates"),
		fileServer: http.FileServer(http.Dir("./")),
		scanner:    NewMailScanner(mailPath),
		mailPath:   mailPath,
	}

	var err error = nil
	e.userChecker, err = regexp.Compile(UserRegExp)
	if err != nil {
		log.Fatal("Could not compile user checker regex")
	}
	return
}

func (e *GofixEngine) Run() {
	defer e.scanner.watcher.Close()
	e.scanner.Run()
	http.Handle("/", e)
	log.Fatal(http.ListenAndServe(":65200", nil))
}

func (e *GofixEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Path)
	switch r.URL.Path {
	case "/css/styles.css":
		e.fileServer.ServeHTTP(w, r)
	default:
		{
			user := r.URL.Query().Get("user")

			if e.userChecker.FindString(user) != user || user == "" {
				fmt.Print("Invalid user")
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, "401 - Access denied")
				return
			}

			state := StateHeaderScan
			headerFinder, err := regexp.Compile(HeaderRegExp)
			if err != nil {
				log.Fatalf("Invalid regexp %s\n", err)
			}

			foldingFinder, err := regexp.Compile(FoldingRegExp)
			if err != nil {
				log.Fatalf("Invalid regexp %s\n", err)
			}

			boundaryStartFinder, err := regexp.Compile(BoundaryStartRegExp)
			if err != nil {
				log.Fatalf("Invalid regexp %s\n", err)
			}

			boundaryEndFinder, err := regexp.Compile(BoundaryEndRegExp)
			if err != nil {
				log.Fatalf("Invalid regexp %s\n", err)
			}

			boundaryFinder, err := regexp.Compile(BoundaryRegExp)

			if !fileExists(e.mailPath + "/" + r.URL.Query().Get("user")) {
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprint(w, "403 Unknown user")
				return
			}

			file, _ := os.Open(e.mailPath + "/" + r.URL.Query().Get("user"))
			scanner := bufio.NewScanner(file)
			activeBoundary := ""
			var previousHeader *string = nil
			var emails []*Mail
			email := NewEmail()
			for scanner.Scan() {
				if scanner.Text() == "" {
					if state == StateHeaderScan {
						boundaryCapture := boundaryFinder.FindStringSubmatch(email.Body.ContentType)
						if len(boundaryCapture) == 2 {
							activeBoundary = boundaryCapture[1]
						} else {
							activeBoundary = ""
						}
						state = StateBodyScan
						// fmt.Printf("--------------------------Start body scan content type:%s boundary: %s -------------------------\n", email.Body.ContentType, activeBoundary)
					} else if state == StateBodyScan {
						// fmt.Printf("--------------------------Previous email-------------------------\n%v\n", email)

						previousHeader = nil
						activeBoundary = ""
						emails = append(emails, email)
						email = NewEmail()
						state = StateHeaderScan
					} else {
						// fmt.Printf("Empty line in state %d\n", state)
					}
				}

				if state == StateHeaderScan {
					capture := headerFinder.FindStringSubmatch(scanner.Text())
					if len(capture) == 3 {
						// fmt.Printf("capture Header %s : %s\n", strings.ToLower(capture[0]), strings.ToLower(capture[1]))
						header := strings.ToLower(capture[1])
						switch header {
						case "from":
							previousHeader = &email.Header.From
						case "to":
							previousHeader = &email.Header.To
						case "cc":
							previousHeader = &email.Header.Cc
						case "bcc":
							previousHeader = &email.Header.Bcc
						case "subject":
							previousHeader = &email.Header.Subject
						case "date":
							previousHeader = &email.Header.Date
						case "content-type":
							previousHeader = &email.Body.ContentType
						default:
							previousHeader = nil
						}
						if previousHeader != nil {
							*previousHeader += capture[2]
						}
						continue
					}

					capture = foldingFinder.FindStringSubmatch(scanner.Text())
					if len(capture) == 2 && previousHeader != nil {
						*previousHeader += capture[1]
						continue
					}
				} else {
					// email.Body.Content += scanner.Text() + "\n"
					if activeBoundary != "" {
						capture := boundaryEndFinder.FindStringSubmatch(scanner.Text())
						if len(capture) == 2 {
							// fmt.Printf("capture Boundary End %s\n", capture[1])
							if activeBoundary == capture[1] {
								state = StateBodyScan
							}

							continue
						}
						capture = boundaryStartFinder.FindStringSubmatch(scanner.Text())
						if len(capture) == 2 {
							// fmt.Printf("capture Boundary Start %s\n", capture[1])
							state = StateContentScan
							continue
						}
					}
				}
			}

			if state == StateBodyScan { //Finalize if body read till EOF
				// fmt.Printf("--------------------------Previous email-------------------------\n%v\n", email)

				previousHeader = nil
				activeBoundary = ""
				emails = append(emails, email)
				state = StateHeaderScan
			}

			content := template.HTML(e.templater.ExecuteMailList(emails))

			fmt.Fprint(w, e.templater.ExecuteIndex(content))
		}
	}
}

func openAndLockMailFile() {
	file, err := os.OpenFile("/home/vmail/semlanik.org/ci", os.O_RDWR, 0)
	if err != nil {
		log.Fatalf("Error to open /home/vmail/semlanik.org/ci %s", err)
	}
	defer file.Close()

	lk := &unix.Flock_t{
		Type: unix.F_WRLCK,
	}
	err = unix.FcntlFlock(file.Fd(), unix.F_SETLKW, lk)
	lk.Type = unix.F_UNLCK

	if err != nil {
		log.Fatalf("Error to set lock %s", err)
	}
	defer unix.FcntlFlock(file.Fd(), unix.F_SETLKW, lk)

	fmt.Printf("Succesfully locked PID: %d", lk.Pid)
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil && !info.IsDir() && info != nil
}

func main() {
	mailPath := "./"
	if len(os.Args) >= 2 {
		mailPath = os.Args[1]
	}
	engine := NewGofixEngine(mailPath)
	engine.Run()
}
