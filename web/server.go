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
	ioutil "io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	common "../common"
	config "../config"
	utils "../utils"
	"github.com/gorilla/sessions"
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
	mailMaps     map[string]string //TODO: Temporary
	fileServer   http.Handler
	templater    *Templater
	sessionStore *sessions.CookieStore
}

func NewServer() *Server {
	return &Server{
		mailMaps:     readMailMaps(), //TODO: Temporary
		templater:    NewTemplater("./data/templates"),
		fileServer:   http.FileServer(http.Dir("./data")),
		sessionStore: sessions.NewCookieStore(make([]byte, 32)),
	}
}

func (s *Server) Run() {
	http.Handle("/", s)
	log.Fatal(http.ListenAndServe(":65200", nil))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Path)
	if utils.StartsWith(r.URL.Path, "/css/") ||
		utils.StartsWith(r.URL.Path, "/assets/") ||
		utils.StartsWith(r.URL.Path, "/js/") {
		s.fileServer.ServeHTTP(w, r)
	} else if r.URL.Path == "/login" {
		session, _ := s.sessionStore.Get(r, "token")
		user, ok := session.Values["user"].(string)
		if !ok || user == "" {
			user = ""
			if err := r.ParseForm(); err == nil {
				user = r.FormValue("user")
			}
		}

		_, ok = s.mailMaps[user]
		if utils.RegExpUtilsInstance().EmailChecker.MatchString(user) &&
			user != "" && ok {
			session.Values["user"] = user
			session.Save(r, w)
			http.Redirect(w, r, "/mailbox", http.StatusTemporaryRedirect)
			return
		}

		session.Values["user"] = ""
		session.Save(r, w)

		fmt.Printf("Actual user: %s, Actual map: %v\n", user, s.mailMaps)
		data, _ := ioutil.ReadFile("./data/templates/login.html")
		w.Write(data)
	} else if r.URL.Path == "/logout" {
		fmt.Printf("logout")
		session, _ := s.sessionStore.Get(r, "token")
		session.Values["user"] = ""
		session.Save(r, w)
		fmt.Printf("Reset user session: %s\n", session.Values["user"])
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	} else if r.URL.Path == "/messageDetails" {
		fmt.Fprint(w, s.templater.ExecuteDetails(""))
	} else {
		session, _ := s.sessionStore.Get(r, "token")
		user, ok := session.Values["user"].(string)

		fmt.Printf("User session: %s\n", user)

		if !ok || utils.RegExpUtilsInstance().EmailChecker.FindString(user) != user || user == "" {
			fmt.Print("Invalid user")
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		mailRelPath, ok := s.mailMaps[user]
		if !ok {
			fmt.Print("Invalid user")
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		mailPath := config.ConfigInstance().VMailboxBase + "/" + mailRelPath
		if !utils.FileExists(mailPath) {
			fmt.Printf("logout")
			session, _ := s.sessionStore.Get(r, "token")
			session.Values["user"] = ""
			session.Save(r, w)
			fmt.Printf("Reset user session: %s\n", session.Values["user"])

			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, s.templater.ExecuteError(&Error{
				Code:   http.StatusInternalServerError,
				String: "Unable to access your mailbox. Please contact Administrator.",
			}))
			return
		}

		file, err := utils.OpenAndLockWait(mailPath)
		if err != nil {
			fmt.Printf("logout")
			session, _ := s.sessionStore.Get(r, "token")
			session.Values["user"] = ""
			session.Save(r, w)
			fmt.Printf("Reset user session: %s\n", session.Values["user"])

			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, s.templater.ExecuteError(&Error{
				Code:   http.StatusInternalServerError,
				String: "Unable to access your mailbox. Please contact Administrator.",
			}))
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
						// fmt.Printf("Actual headers: %d\n", mandatoryHeaders)
						if mandatoryHeaders == AllHeaderMask {
							emails = append(emails, email)
						}
						email = NewEmail()
						state = StateHeaderScan
						mandatoryHeaders = 0
					} else {
						// fmt.Printf("Still in body scan\n")
						continue
					}
				} else {
					// fmt.Printf("Empty line in state %d\n", state)
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

func readMailMaps() map[string]string { //TODO: Temporary
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
