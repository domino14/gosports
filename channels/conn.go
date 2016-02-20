// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package channels

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
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
	// XXX: Later check Aerolith.org or something. We will be on different
	// ports but we should still accept.
	// (or maybe socket.aerolith.org?)
	CheckOrigin: func(r *http.Request) bool {
		log.Println("[DEBUG] In CheckOrigin, will return true.")
		return true
	},
}

// connection is an middleman between the websocket connection and the hub.
type connection struct {
	// The websocket connection.
	ws *websocket.Conn
	// Buffered channel of outbound messages.
	send     chan []byte
	username string
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
			log.Println("error unmarshaling: ", err)
			break
		}
		// Save raw data for this message for further processing, but append
		// username.
		// XXX: we'll end up unmarshalling twice. We should re-think this later.
		m.room = s.room
		m.From = s.conn.username
		// Remarshal to m.rawdata
		rawdata, err := json.Marshal(m)
		if err != nil {
			log.Println("Error re-marshalling: ", err)
			break
		}
		log.Println("[DEBUG] raw:", string(rawdata))
		m.rawdata = rawdata
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

func validateWsRequest(v url.Values, now int64) error {
	secretKey := os.Getenv("SECRET_KEY")
	realm := v.Get("realm")
	user := v.Get("user")
	timestamp := v.Get("expire")
	token := v.Get("_token")
	// Convert token to an array of bytes. Assume token is hex encoded.

	if secretKey == "" {
		// This should be a panic but let's not go overboard.
		// Maybe it's a too many open files issue.
		log.Println("[ERROR] Secret key missing!")
		return fmt.Errorf("No secret key in environment.")
	}
	// Convert timestamp to an int.
	ts_int, err := strconv.Atoi(timestamp)
	if err != nil {
		return err
	}
	if int64(ts_int) < now {
		return fmt.Errorf("your token has expired (ts = %v, now = %v)",
			ts_int, now)
	}
	if realm == "" {
		return fmt.Errorf("no realm was specified")
	}
	if user == "" {
		return fmt.Errorf("no user was specified")
	}
	tokenHex, err := hex.DecodeString(token)
	if err != nil {
		return err
	}

	// Reconstruct signed string.
	ss := fmt.Sprintf("expire=%v&realm=%v&user=%v", timestamp, realm, user)
	log.Println("[DEBUG] Signing:", ss)
	mac := hmac.New(sha1.New, []byte(secretKey))
	mac.Write([]byte(ss))
	expectedMac := mac.Sum(nil)
	if !hmac.Equal(expectedMac, tokenHex) {
		return fmt.Errorf(
			"token signature was not correct (got %x, expected %x)",
			expectedMac, tokenHex)
	}
	return nil
}

// serveWs handles websocket requests from the peer.
func ServeWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	qvals := r.URL.Query()
	err = validateWsRequest(qvals, time.Now().Unix())
	if err != nil {
		log.Println("[ERROR] Got an error:", err)
		return // should write a rejection
	}
	c := &connection{send: make(chan []byte, 256), ws: ws,
		username: qvals.Get("user")}
	s := &subscription{conn: c, room: qvals.Get("realm")}
	log.Println("[DEBUG] Made new connection", c)
	Hub.register <- s
	go s.writePump()
	s.readPump()
}
