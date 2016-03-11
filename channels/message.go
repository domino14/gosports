package channels

type MessageType string

const (
	ServerMT MessageType = "server" // Message from the server

	// Message types that should not be broadcast.
	PrivateMT MessageType = "pm"
)

type Message struct {
	Data    string      `json:"data"`
	Mtype   MessageType `json:"type"`
	rawdata []byte
	realm   Realm  // This will get copied from the subscription.
	From    string `json:"from"`
}

func (m *Message) Realm() Realm {
	return m.realm
}

func (m *Message) SetRealm(realm Realm) {
	m.realm = realm
}

type SocketMessageHandler interface {
	// HandleMessage must take in a message and perform some sort of
	// action with it.
	HandleMessage(m Message)
	// On the creation of a realm, run some optional initialization
	// code provided by the implementer of this interface.
	RealmCreation(realm Realm)
	// On the joining of a realm, run some code.
	RealmJoin(realm Realm, user string, connId string, firstUser bool)
	// On the leaving of a realm, run some code
	RealmLeave(realm Realm, user string, connId string)
}

// SocketMessageSender is an interface that will send messages from our
// server to one or more websockets.
type SocketMessageSender interface {
	// Broadcast a message to all available sockets.
	BroadcastMessage(realm Realm, mt MessageType, msg string)
	// Send a message to just one single player.
	SendMessage(realm Realm, mt MessageType, msg string, to string)
}
