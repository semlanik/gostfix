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
	"fmt"
	"log"
	"net/http"
	"sync"

	"git.semlanik.org/semlanik/gostfix/common"
	"github.com/gorilla/websocket"
)

type websocketChannel struct {
	connection *websocket.Conn
	channel    chan *common.Mail
}

type webNotifier struct {
	notifiers     map[string]*websocketChannel
	notifiersLock sync.Mutex
}

func NewWebNotifier() *webNotifier {
	return &webNotifier{
		notifiers: make(map[string]*websocketChannel),
	}
}

func (wn *webNotifier) NotifyMaiboxUpdate(email string) {
	if channel, ok := wn.getNotifier(email); ok {
		channel.channel <- &common.Mail{} //TODO: Dummy notificator for now, later need to make separate interface to handle this
	}
}

func (wn *webNotifier) NotifyNewMail(email string, m common.Mail) {
	// if channel, ok := wn.getNotifier(email); ok {
	// 	channel.channel <- &m
	// }
	//TODO: this functionality needs JS support to create new mails from templates
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (wn *webNotifier) handleNotifierRequest(w http.ResponseWriter, r *http.Request, email string) {
	fmt.Printf("New web socket session start %s\n", email)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}

	c := &websocketChannel{
		connection: conn,
		channel:    make(chan *common.Mail, 10),
	}
	wn.addNotifier(email, c)

	conn.SetCloseHandler(func(code int, text string) error {
		fmt.Printf("Web socket session end %s\n", email)
		wn.removeNotifier(email)
		conn.Close()
		return nil
	})

	go wn.handleNotifications(c)
}

func (wn *webNotifier) handleNotifications(c *websocketChannel) {
	//Do nothing for now
	for {
		select {
		case newMail := <-c.channel:
			err := c.connection.WriteJSON(newMail)
			if err != nil {
				log.Println(err.Error())
				return
			}
		}
	}
}

func (wn *webNotifier) getNotifier(email string) (channel *websocketChannel, ok bool) {
	wn.notifiersLock.Lock()
	defer wn.notifiersLock.Unlock()
	channel, ok = wn.notifiers[email]
	return
}

func (wn *webNotifier) addNotifier(email string, channel *websocketChannel) {
	wn.notifiersLock.Lock()
	defer wn.notifiersLock.Unlock()
	wn.notifiers[email] = channel
}

func (wn *webNotifier) removeNotifier(email string) {
	wn.notifiersLock.Lock()
	defer wn.notifiersLock.Unlock()
	delete(wn.notifiers, email)
}
