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

	auth "../auth"
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

const (
	CookieSessionToken = "gostfix_session"
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
	authenticator *auth.Authenticator
	fileServer    http.Handler
	templater     *Templater
	sessionStore  *sessions.CookieStore
}

func NewServer() *Server {
	return &Server{
		authenticator: auth.NewAuthenticator(),
		templater:     NewTemplater("data/templates"),
		fileServer:    http.FileServer(http.Dir("data")),
		sessionStore:  sessions.NewCookieStore(make([]byte, 32)),
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
	} else {
		switch r.URL.Path {
		case "/login":
			s.handleLogin(w, r)
		case "/logout":
			s.handleLogout(w, r)
		case "/messageDetails":
			s.handleMessageDetails(w, r)
		case "/statusLine":
			s.handleStatusLine(w, r)
		default:
			s.handleMailbox(w, r)
		}
	}
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	//Check passed in form login/password pair first
	if err := r.ParseForm(); err == nil {
		user := r.FormValue("user")
		password := r.FormValue("password")
		token, ok := s.authenticator.Authenticate(user, password)
		if ok {
			s.login(user, token, w, r)
			return
		}
	}

	//Check if user already logged in and entered login page accidently
	if s.authenticator.Verify(s.extractAuth(w, r)) {
		http.Redirect(w, r, "/mailbox", http.StatusTemporaryRedirect)
		return
	}

	//Otherwise make sure user logged out and show login page
	s.logout(w, r)
	fmt.Fprint(w, s.templater.ExecuteLogin(&LoginTemplateData{
		common.Version,
	}))
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	s.logout(w, r)
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}

func (s *Server) handleMessageDetails(w http.ResponseWriter, r *http.Request) {
	//TODO: Not implemented yet. Need database mail storage implemented first
	user, token := s.extractAuth(w, r)
	if !s.authenticator.Verify(user, token) {
		fmt.Fprint(w, "")
		return
	}
	fmt.Fprint(w, s.templater.ExecuteDetails(""))
}

func (s *Server) handleStatusLine(w http.ResponseWriter, r *http.Request) {
	//TODO: Not implemented yet. Need database mail storage implemented first
	user, token := s.extractAuth(w, r)
	if !s.authenticator.Verify(user, token) {
		fmt.Fprint(w, "")
		return
	}

	fmt.Fprint(w, s.templater.ExecuteStatusLine(&StatusLineTemplateData{
		Name:   "No name", //TODO: read from database
		Read:   0,         //TODO: read from database
		Unread: 0,         //TODO: read from database
	}))
}

func (s *Server) handleMailbox(w http.ResponseWriter, r *http.Request) {
	user, token := s.extractAuth(w, r)
	if !s.authenticator.Verify(user, token) {
		s.logout(w, r)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	}

	mailPath := config.ConfigInstance().VMailboxBase + "/" + s.authenticator.MailPath(user)
	if !utils.FileExists(mailPath) {
		s.logout(w, r)
		s.error(http.StatusInternalServerError, "Unable to access your mailbox. Please contact Administrator.", w, r)
		return
	}

	file, err := utils.OpenAndLockWait(mailPath)
	if err != nil {
		s.logout(w, r)
		s.error(http.StatusInternalServerError, "Unable to access your mailbox. Please contact Administrator.", w, r)
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

	fmt.Fprint(w, s.templater.ExecuteIndex(&IndexTemplateData{
		MailList: template.HTML(s.templater.ExecuteMailList(emails)),
		Folders:  "Folders",
		Version:  common.Version,
	}))
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	fmt.Println("logout")

	session, _ := s.sessionStore.Get(r, CookieSessionToken)
	session.Values["user"] = ""
	session.Values["token"] = ""
	session.Save(r, w)
}

func (s *Server) login(user, token string, w http.ResponseWriter, r *http.Request) {
	session, _ := s.sessionStore.Get(r, CookieSessionToken)
	session.Values["user"] = user
	session.Values["token"] = token
	session.Save(r, w)
	http.Redirect(w, r, "/mailbox", http.StatusTemporaryRedirect)
}

func (s *Server) error(code int, text string, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprint(w, s.templater.ExecuteError(&ErrorTemplateData{
		Code: code,
		Text: "Unable to access your mailbox. Please contact Administrator.",
	}))
}

func (s *Server) extractAuth(w http.ResponseWriter, r *http.Request) (user, token string) {
	session, err := s.sessionStore.Get(r, CookieSessionToken)
	if err != nil {
		log.Printf("Unable to read user session %s\n", err)
		return
	}
	user, _ = session.Values["user"].(string)
	token, _ = session.Values["token"].(string)

	return
}
