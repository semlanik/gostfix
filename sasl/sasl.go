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

package sasl

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"git.semlanik.org/semlanik/gostfix/auth"
	"github.com/google/uuid"
)

type SaslServer struct {
	pid           int
	cuid          int
	authenticator *auth.Authenticator
}

const (
	Version = "VERSION"
	CPid    = "CPID"
	SPid    = "SPID"
	Cuid    = "CUID"
	Cookie  = "COOKIE"
	Mech    = "MECH"
	Done    = "DONE"
	Auth    = "AUTH"
	Fail    = "FAIL"
	Cont    = "CONT"
	Ok      = "OK"
)

const (
	ContinueStateNone = iota
	ContinueStateCredentials
)

func NewSaslServer() (*SaslServer, error) {
	authenticator, err := auth.NewAuthenticator()
	if err != nil {
		return nil, err
	}
	return &SaslServer{
		pid:           os.Getpid(),
		cuid:          0,
		authenticator: authenticator,
	}, nil
}

func (s *SaslServer) Run() {
	go func() {
		l, err := net.Listen("tcp", "127.0.0.1:65201")
		if err != nil {
			log.Fatalf("Coulf not start SASL server: %s\n", err)
			return
		}
		defer l.Close()

		log.Printf("Listen sasl on: %s\n", l.Addr().String())

		for {
			conn, err := l.Accept()
			s.cuid++
			if err != nil {
				log.Println("Error accepting: ", err.Error())
				continue
			}
			go s.handleRequest(conn)
		}
	}()
}

func (s *SaslServer) handleRequest(conn net.Conn) {
	connectionReader := bufio.NewReader(conn)
	continueState := ContinueStateNone
	for {
		fullbuf, err := connectionReader.ReadString('\n')

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Printf("Read error %s\n", err)
			break
		}

		currentMessage := fullbuf

		ids := strings.Split(currentMessage, "\t")
		if len(ids) < 2 {
			break
		}

		switch ids[0] {
		case Version:
			if len(ids) < 3 {
				break
			}

			if major, err := strconv.Atoi(ids[1]); err != nil || major != 1 {
				break
			}

			cookieUuid := uuid.New()
			fmt.Fprintf(conn, "%s\t%d\t%d\n", Version, 1, 2)
			fmt.Fprintf(conn, "%s\t%s\t%s\n", Mech, "PLAIN", "plaintext")
			fmt.Fprintf(conn, "%s\t%s\t%s\n", Mech, "LOGIN", "plaintext")

			fmt.Fprintf(conn, "%s\t%d\n", SPid, s.pid)
			fmt.Fprintf(conn, "%s\t%d\n", Cuid, s.cuid)

			fmt.Fprintf(conn, "%s\t%s\n", Cookie, hex.EncodeToString(cookieUuid[:]))
			fmt.Fprintf(conn, "%s\n", Done)
		case Auth:
			for _, authId := range ids {
				if strings.Index(authId, "resp=") == 0 {
					login, err := s.checkCredentials(authId[5:])
					if err != nil {
						fmt.Fprintf(conn, "%s\t%s\treason=%s\n", Fail, ids[1], err.Error())
					} else {
						fmt.Fprintf(conn, "%s\t%s\tuser=%s\n", Ok, ids[1], login)
					}
					continueState = ContinueStateNone
					return
				}
			}

			fmt.Fprintf(conn, "%s\t%s\t%s\n", Cont, ids[1], base64.StdEncoding.EncodeToString([]byte("Username:")))
			continueState = ContinueStateCredentials
		case Cont:
			if len(ids) < 2 {
				break
			}

			if continueState == ContinueStateCredentials {
				if len(ids) < 3 {
					fmt.Fprintf(conn, "%s\t%s\treason=%s\n", Fail, ids[1], "invalid base64 data")
					return
				}

				login, err := s.checkCredentials(ids[2])
				if err != nil {
					fmt.Fprintf(conn, "%s\t%s\treason=%s\n", Fail, ids[1], err.Error())
				} else {
					fmt.Fprintf(conn, "%s\t%s\tuser=%s\n", Ok, ids[1], login)
				}
				continueState = ContinueStateNone
			} else {
				fmt.Fprintf(conn, "%s\t%s\treason=%s\n", Fail, ids[1], "invalid user or password")
			}
		}
	}
	conn.Close()
}

func (s *SaslServer) checkCredentials(credentialsBase64 string) (string, error) {
	credentials, err := base64.StdEncoding.DecodeString(credentialsBase64)
	if err != nil {
		return "", errors.New("invalid base64 data")
	}

	credentialList := bytes.Split(credentials, []byte{0})
	if len(credentialList) < 3 {
		return "", errors.New("invalid user or password")
	}

	identity := string(credentialList[0])
	login := string(credentialList[1])
	password := string(credentialList[2])
	if identity == "token" {
		if s.authenticator.Verify(login, password) {
			return login, nil
		}
	} else {
		if err := s.authenticator.CheckUser(login, password); err == nil {
			return login, nil
		}
	}

	return "", errors.New("invalid user or password")
}
