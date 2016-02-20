// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package channels

import "log"

type Message struct {
	Data    string `json:"data"`
	Mtype   string `json:"type"`
	rawdata []byte
	room    string // This will get copied from the subscription.
	From    string `json:"from"`
}

type subscription struct {
	conn *connection
	room string
}

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type hub struct {
	// Registered connections.
	rooms map[string]map[*connection]bool

	// Inbound messages from the connections.
	broadcast chan Message

	// Register requests from the connections.
	register chan *subscription

	// Unregister requests from connections.
	unregister chan *subscription
}

// The global, singleton Hub object. This manages all our connections.
var Hub = hub{
	broadcast:  make(chan Message),
	register:   make(chan *subscription),
	unregister: make(chan *subscription),
	rooms:      make(map[string]map[*connection]bool),
}

func (h *hub) Run() {
	for {
		select {
		case sub := <-h.register:
			connections := h.rooms[sub.room]
			if connections == nil {
				connections = make(map[*connection]bool)
				h.rooms[sub.room] = connections
			}
			connections[sub.conn] = true
		case sub := <-h.unregister:
			connections := h.rooms[sub.room]

			if connections != nil {
				if _, ok := connections[sub.conn]; ok {
					log.Println("[DEBUG] Unregistering", sub.conn.username)
					delete(connections, sub.conn)
					close(sub.conn.send)
					if len(connections) == 0 {
						// Last person left the room.
						delete(h.rooms, sub.room)
					}
				}
			}
		case m := <-h.broadcast:
			connections := h.rooms[m.room]
			for c := range connections {
				select {
				case c.send <- m.rawdata:
				default:
					log.Println("[DEBUG] Disconnecting", c.username)
					close(c.send)
					delete(connections, c)
					if len(connections) == 0 {
						delete(h.rooms, m.room)
					}
				}
			}
		}
	}
}
