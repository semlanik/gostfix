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
	"log"
	"os"
	"os/signal"
	"syscall"

	sasl "git.semlanik.org/semlanik/gostfix/sasl"
	scanner "git.semlanik.org/semlanik/gostfix/scanner"
	service "git.semlanik.org/semlanik/gostfix/service"
	web "git.semlanik.org/semlanik/gostfix/web"
	"github.com/pkg/profile"
)

type GostfixEngine struct {
	services []service.NanoService
	stats    []*service.NanoServiceStats
}

func NewGostfixEngine() (e *GostfixEngine) {
	mailScanner := scanner.NewMailScanner()
	saslService, err := sasl.NewSaslServer()
	if err != nil {
		log.Fatalf("Unable to intialize sasl server %s\n", err)
	}

	webServer := web.NewServer(mailScanner, e)

	e = &GostfixEngine{}
	e.services = append(e.services, saslService, mailScanner, webServer)
	e.stats = make([]*service.NanoServiceStats, len(e.services))
	return
}

func (e *GostfixEngine) Run() {
	for i, s := range e.services {
		if e.stats[i] == nil {
			e.stats[i] = &service.NanoServiceStats{}
		}
		go func(s service.NanoService, stats *service.NanoServiceStats) {
			stats.Name = s.ServiceName()
			stats.Status = service.NanoServiceStatus_NanoServiceRunning
			log.Printf("Running %s", s.ServiceName())
			s.Run()
			stats.Status = service.NanoServiceStatus_NanoServiceStopped
			log.Printf("%s is stopped", s.ServiceName())
		}(s, e.stats[i])
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)
	go func() {
		sig := <-sigs
		log.Printf("Exit by signal %v", sig)
		done <- true
	}()

	log.Printf("Server is running succesfully")
	<-done

	log.Printf("Server is stopping")
	for i, s := range e.services {
		if e.stats[i] != nil && e.stats[i].Status == service.NanoServiceStatus_NanoServiceRunning {
			e.stats[i] = &service.NanoServiceStats{}
			s.Stop()
		}
	}
	log.Printf("Server shutdown")
}

func (e *GostfixEngine) ReadStats() []*service.NanoServiceStats {
	return e.stats
}

func main() {
	defer profile.Start().Stop()
	engine := NewGostfixEngine()
	engine.Run()
}
