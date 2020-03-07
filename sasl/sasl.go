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
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type SaslServer struct {
	pid  int
	cuid int
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

func NewSaslServer() *SaslServer {
	return &SaslServer{
		pid:  os.Getpid(),
		cuid: 0,
	}
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
			continue
		}

		if err != nil {
			fmt.Printf("Read error %s\n", err)
		}

		currentMessage := fullbuf
		if strings.Index(currentMessage, Version) == 0 {
			versionIds := strings.Split(currentMessage, "\t")

			if len(versionIds) < 3 {
				break
			}

			if major, err := strconv.Atoi(versionIds[1]); err != nil || major != 1 {
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
		} else if strings.Index(currentMessage, Auth) == 0 {
			authIds := strings.Split(currentMessage, "\t")
			if len(authIds) < 2 {
				break
			}
			fmt.Fprintf(conn, "%s\t%s\t%s\n", Cont, authIds[1], base64.StdEncoding.EncodeToString([]byte("Username:")))
			continueState = ContinueStateCredentials
		} else if strings.Index(currentMessage, Cont) == 0 {
			contIds := strings.Split(currentMessage, "\t")
			if len(contIds) < 2 {
				break
			}

			if continueState == ContinueStateCredentials {
				if len(contIds) < 3 {
					fmt.Fprintf(conn, "%s\t%s\treason=%s\n", Fail, contIds[1], "invalid base64 data")
					return
				}

				credentials, err := base64.StdEncoding.DecodeString(contIds[2])
				if err != nil {
					fmt.Fprintf(conn, "%s\t%s\treason=%s\n", Fail, contIds[1], "invalid base64 data")
					return
				}

				credentialList := bytes.Split(credentials, []byte{0})
				if len(credentialList) < 3 {
					fmt.Fprintf(conn, "%s\t%s\treason=%s\n", Fail, contIds[1], "invalid user or password")
					return
				}

				// identity := credentialList[0]
				login := credentialList[1]
				// password := credentialList[2]
				//TODO: Use auth here
				// if login != "semlanik@semlanik.org" || password != "test" {
				if true {
					fmt.Fprintf(conn, "%s\t%s\treason=%s\n", Fail, contIds[1], "invalid user or password")
					return
				}

				fmt.Fprintf(conn, "%s\t%s\tuser=%s\n", Ok, contIds[1], login)
				continueState = ContinueStateNone
			} else {
				fmt.Fprintf(conn, "%s\t%s\treason=%s\n", Fail, contIds[1], "invalid user or password")
			}
		}
	}
	conn.Close()
}
