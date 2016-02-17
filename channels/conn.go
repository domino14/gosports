// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package channels

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// connection is an middleman between the websocket connection and the hub.
type connection struct {
	// The websocket connection.
	ws *websocket.Conn
	// Buffered channel of outbound messages.
	send chan []byte
}

// readPump pumps messages from the websocket connection to the hub.
func (s *subscription) readPump() {
	defer func() {
		Hub.unregister <- s
		s.conn.ws.Close()
	}()
	s.conn.ws.SetReadLimit(maxMessageSize)
	s.conn.ws.SetReadDeadline(time.Now().Add(pongWait))
	s.conn.ws.SetPongHandler(func(string) error {
		s.conn.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, message, err := s.conn.ws.ReadMessage()
		if err != nil {
			/** XXX: Reenable when we update the websocket package. */

			/*			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
						log.Printf("error: %v", err)
					}*/
			log.Printf("error: %v", err)
			break
		}
		// Parse as JSON.
		var m Message
		err = json.Unmarshal(message, &m)
		if err != nil {
			log.Printf("error unmarshaling: %v", err)
			break
		}
		// Save raw data for this message for further processing.
		// XXX: we'll end up unmarshalling twice.
		m.rawdata = message
		m.room = s.room
		Hub.broadcast <- m
	}
}

// write writes a message with the given message type and payload.
func (s *subscription) write(mt int, payload []byte) error {
	s.conn.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return s.conn.ws.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (s *subscription) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		s.conn.ws.Close()
	}()
	for {
		select {
		case message, ok := <-s.conn.send:
			if !ok {
				s.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := s.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := s.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func ServeWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	// Get room from query params here. Later add a signature/timestamp
	// or session token for validating.
	room := r.URL.Query().Get("room")
	if len(room) == 0 {
		log.Println("Rejected, no room")
		return
	}

	c := &connection{send: make(chan []byte, 256), ws: ws}
	s := &subscription{conn: c, room: room}
	Hub.register <- s
	go s.writePump()
	s.readPump()
}
