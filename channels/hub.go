// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package channels

import (
	"encoding/json"
	"log"
)

type subscription struct {
	conn  *connection
	realm Realm
}

type Realm string

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type hub struct {
	// Registered connections.
	realms map[Realm]map[*connection]bool

	// Inbound messages from the connections.
	broadcast chan Message

	// Register requests from the connections.
	register chan *subscription

	// Unregister requests from connections.
	unregister chan *subscription

	// A handler of messages.
	handler SocketMessageHandler
}

// The global, singleton Hub object. This manages all our connections.
var Hub = hub{
	broadcast:  make(chan Message),
	register:   make(chan *subscription),
	unregister: make(chan *subscription),
	realms:     make(map[Realm]map[*connection]bool),
}

func BroadcastMessage(realm Realm, mt MessageType, msg string) {
	Hub.broadcastMessage(realm, mt, msg)
}

func (h *hub) broadcastMessage(realm Realm, mt MessageType, msg string) {
	var msgWrapper Message
	msgWrapper = Message{
		Data:  msg,
		Mtype: mt,
		realm: realm,
	}
	rawdata, err := json.Marshal(msgWrapper)
	if err != nil {
		log.Println("[ERROR] JSON encoding - broadcasting message", err)
		return
	}
	msgWrapper.rawdata = rawdata
	log.Println("[DEBUG] Writing a message, rawdata:",
		string(msgWrapper.rawdata))
	h.broadcast <- msgWrapper
}

func (h *hub) Run(handler SocketMessageHandler) {
	h.handler = handler
	for {
		select {
		case sub := <-h.register:
			connections := h.realms[sub.realm]
			newRoom := false
			if connections == nil {
				newRoom = true
				connections = make(map[*connection]bool)
				h.realms[sub.realm] = connections
				h.handler.RealmCreation(sub.realm)
			}
			h.handler.RealmJoin(sub.realm, sub.conn.username, newRoom)
			connections[sub.conn] = true
		case sub := <-h.unregister:
			connections := h.realms[sub.realm]
			h.handler.RealmLeave(sub.realm, sub.conn.username)
			if connections != nil {
				if _, ok := connections[sub.conn]; ok {
					log.Println("[DEBUG] Unregistering", sub.conn.username)
					delete(connections, sub.conn)
					close(sub.conn.send)
					if len(connections) == 0 {
						// Last person left the room.
						delete(h.realms, sub.realm)
					}
				}
			}
		case m := <-h.broadcast:
			connections := h.realms[m.realm]
			for c := range connections {
				select {
				case c.send <- m.rawdata:
				default:
					log.Println("[DEBUG] Disconnecting", c.username)
					close(c.send)
					delete(connections, c)
					if len(connections) == 0 {
						delete(h.realms, m.realm)
					}
				}
			}
		}
	}
}
