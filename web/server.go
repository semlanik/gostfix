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
	"bufio"
	"fmt"
	template "html/template"
	"log"
	"net/http"
	"strings"

	common "../common"
	utils "../utils"
)

const (
	StateHeaderScan = iota
	StateBodyScan
	StateContentScan
)

const (
	AtLeastOneHeaderMask = 1 << iota
	FromHeaderMask
	DateHeaderMask
	ToHeaderMask
	AllHeaderMask = 15
)

func NewEmail() *common.Mail {
	return &common.Mail{
		Header: &common.MailHeader{},
		Body: &common.MailBody{
			ContentType: "plain/text",
		},
	}
}

type Server struct {
	fileServer http.Handler
	templater  *Templater
	mailPath   string
}

func NewServer(mailPath string) *Server {
	return &Server{
		templater:  NewTemplater("templates"),
		fileServer: http.FileServer(http.Dir(".")),
		mailPath:   mailPath,
	}
}

func (s *Server) Run() {
	http.Handle("/", s)
	log.Fatal(http.ListenAndServe(":65200", nil))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Path)
	if strings.Index(r.URL.Path, "/css/") == 0 || strings.Index(r.URL.Path, "/assets/") == 0 {
		s.fileServer.ServeHTTP(w, r)
	} else {
		user := r.URL.Query().Get("user")

		if utils.RegExpUtilsInstance().UserChecker.FindString(user) != user || user == "" {
			fmt.Print("Invalid user")
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "401 - Access denied")
			return
		}

		// mailPath = config.mailPath + "/" + r.URL.Query().Get("user")
		mailPath := "tmp" + "/" + r.URL.Query().Get("user")
		if !utils.FileExists(mailPath) {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, "403 Unknown user")
			return
		}

		file, err := utils.OpenAndLockWait(mailPath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "500 Internal server error")
			return
		}

		defer file.CloseAndUnlock()

		scanner := bufio.NewScanner(file)
		activeBoundary := ""
		var previousHeader *string = nil
		var emails []*common.Mail
		mandatoryHeaders := 0
		email := NewEmail()
		state := StateHeaderScan
		for scanner.Scan() {
			if scanner.Text() == "" {
				if state == StateHeaderScan && mandatoryHeaders&AtLeastOneHeaderMask == AtLeastOneHeaderMask {
					boundaryCapture := utils.RegExpUtilsInstance().BoundaryFinder.FindStringSubmatch(email.Body.ContentType)
					if len(boundaryCapture) == 2 {
						activeBoundary = boundaryCapture[1]
					} else {
						activeBoundary = ""
					}
					state = StateBodyScan
					// fmt.Printf("--------------------------Start body scan content type:%s boundary: %s -------------------------\n", email.Body.ContentType, activeBoundary)
				} else if state == StateBodyScan {
					// fmt.Printf("--------------------------Previous email-------------------------\n%v\n", email)
					if activeBoundary == "" {
						previousHeader = nil
						activeBoundary = ""
						fmt.Printf("Actual headers: %d\n", mandatoryHeaders)
						if mandatoryHeaders == AllHeaderMask {
							emails = append(emails, email)
						}
						email = NewEmail()
						state = StateHeaderScan
						mandatoryHeaders = 0
					} else {
						fmt.Printf("Still in body scan\n")
						continue
					}
				} else {
					fmt.Printf("Empty line in state %d\n", state)
				}
			}

			if state == StateHeaderScan {
				capture := utils.RegExpUtilsInstance().HeaderFinder.FindStringSubmatch(scanner.Text())
				if len(capture) == 3 {
					// fmt.Printf("capture Header %s : %s\n", strings.ToLower(capture[0]), strings.ToLower(capture[1]))
					header := strings.ToLower(capture[1])
					mandatoryHeaders |= AtLeastOneHeaderMask
					switch header {
					case "from":
						previousHeader = &email.Header.From
						mandatoryHeaders |= FromHeaderMask
					case "to":
						previousHeader = &email.Header.To
						mandatoryHeaders |= ToHeaderMask
					case "cc":
						previousHeader = &email.Header.Cc
					case "bcc":
						previousHeader = &email.Header.Bcc
						mandatoryHeaders |= ToHeaderMask
					case "subject":
						previousHeader = &email.Header.Subject
					case "date":
						previousHeader = &email.Header.Date
						mandatoryHeaders |= DateHeaderMask
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

				capture = utils.RegExpUtilsInstance().FoldingFinder.FindStringSubmatch(scanner.Text())
				if len(capture) == 2 && previousHeader != nil {
					*previousHeader += capture[1]
					continue
				}
			} else {
				// email.Body.Content += scanner.Text() + "\n"
				if activeBoundary != "" {
					capture := utils.RegExpUtilsInstance().BoundaryEndFinder.FindStringSubmatch(scanner.Text())
					if len(capture) == 2 {
						// fmt.Printf("capture Boundary End %s\n", capture[1])
						if activeBoundary == capture[1] {
							state = StateBodyScan
							activeBoundary = ""
						}

						continue
					}
					// capture = boundaryStartFinder.FindStringSubmatch(scanner.Text())
					// if len(capture) == 2 && activeBoundary == capture[1] {
					// 	// fmt.Printf("capture Boundary Start %s\n", capture[1])
					// 	state = StateContentScan
					// 	continue
					// }
				}
			}
		}

		if state == StateBodyScan && mandatoryHeaders == AllHeaderMask { //Finalize if body read till EOF
			// fmt.Printf("--------------------------Previous email-------------------------\n%v\n", email)

			previousHeader = nil
			activeBoundary = ""
			emails = append(emails, email)
			state = StateHeaderScan
		}

		fmt.Fprint(w, s.templater.ExecuteIndex(&Index{
			MailList: template.HTML(s.templater.ExecuteMailList(emails)),
			Folders:  "Folders",
			Version:  common.Version,
		}))
	}
}
