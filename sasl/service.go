/*
 * MIT License
 *
 * Copyright (c) 2022 Alexey Edelev <semlanik@gmail.com>
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
	"log"
	"net"

	"git.semlanik.org/semlanik/gostfix/config"
)

func (s *SaslServer) ServiceName() string {
	return "SASL Server"
}

func (s *SaslServer) Run() {
	l, err := net.Listen("tcp", "127.0.0.1:"+config.ConfigInstance().SASLPort)
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
}

func (s *SaslServer) Stop() {
	// TODO: Make possible to stop SASL
}

func (s *SaslServer) Error() string {
	return "" // TODO: return last error in the service
}
