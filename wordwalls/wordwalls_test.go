package wordwalls

import (
	"fmt"
	"io/ioutil"
	"log"
	"testing"
	"time"

	"github.com/domino14/gosports/channels"
	"github.com/satori/go.uuid"
)

const (
	tablenum string = "123456"
)

type MockWebolithCommunicator struct{}

func (m MockWebolithCommunicator) Get(path string) ([]byte, error) {
	if path == "/wordwalls/api/game_options/123456/" {
		return []byte(`{"numAnswersThisRound": 0, "qualifyForAward": true,
        "gameType": "challenge", "challengeId": 43643, "timerSecs": 270,
        "_word_list_id": 22447, "questionsToPull": 50}`), nil
	} else if path == "/base/api/wordlist/22447?action=continue" {
		return ioutil.ReadFile("test_files/get_list_response.json")
	}
	return nil, fmt.Errorf("Path not supported: %s", path)
}

func (m MockWebolithCommunicator) Post(path string, buf []byte) ([]byte, error) {
	if path == "/base/api/word_db/full_questions/" {
		return ioutil.ReadFile("test_files/full_q_response.json")
	}
	return nil, fmt.Errorf("Path not supported: %s", path)
}

type MockMessageSender struct{}

func (s MockMessageSender) BroadcastMessage(realm channels.Realm,
	mt channels.MessageType, msg string) {
	log.Printf("[INFO] Mock broadcast of message: %s, sent to realm: %s, "+
		"message type: %s", msg, realm, mt)
}

func (s MockMessageSender) SendMessage(realm channels.Realm,
	mt channels.MessageType, msg string, to string) {
	log.Printf("[INFO] Mock send of message: %s, sent to realm: %s, "+
		"message type: %s, to user: %s", msg, realm, mt, to)
}

func TestMockBehavior(t *testing.T) {
	realm := toRealm(tablenum)
	// Set mock so we don't connect to external API.
	MessageHandler.webolith = &MockWebolithCommunicator{}
	MessageHandler.RealmCreation(realm)
	gs := gameStates.getState(realm)
	if gs.options.ChallengeId != 43643 {
		t.Errorf("Game state was not correctly set.")
	}
}

/////////////
/// Helper functions.

func joinSitting(users []string, realm channels.Realm) {
	doneCh := make(chan string, len(users))
	for _, user := range users {
		go func(user string) {
			// Set "firstUser" to true to make user sit.
			MessageHandler.RealmJoin(realm, user, uuid.NewV4().String(), true)
			doneCh <- user
		}(user)
	}
	for i := 0; i < len(users); i++ {
		// Drain the channel.
		log.Printf("[DEBUG] Draining %s\n", <-doneCh)
	}
}

func requestStart(users []string, realm channels.Realm) { // All request start.
	doneCh := make(chan string, len(users))
	for _, user := range users {
		go func(user string) {
			msg := channels.Message{
				Data:  "start",
				Mtype: channels.MessageType(TableMT),
				From:  user,
			}
			msg.SetRealm(realm)
			MessageHandler.HandleMessage(msg)
			doneCh <- user
		}(user)
	}
	for i := 0; i < len(users); i++ {
		// Drain the channel.
		log.Printf("[DEBUG] Draining %s\n", <-doneCh)
	}
}

func guessWords(users []string, realm channels.Realm) {
	countdown := time.NewTimer(time.Second * (time.Duration(CountdownTime) + 1))
	<-countdown.C
	words := []string{"ROAMINGS", "APIMANIA", "OPPUGNER", "LATHERED",
		"MIDWIVES", "COLORFUL", "SEAWEEDS", "LIKEABLY", "HALTERED", "BLINDAGE",
		"REEDITED", "SETIFORM", "BOUNDARY", "TRENCHES", "FINAGLED", "MOTHBALL",
		"BARRACKS", "TERAPHIM", "MOUCHING", "UPCOMING", "UNFETTER", "FLUTTERY",
		"VERITIES", "SUNDECKS", "SEESAWED", "RIVIERES", "EUPHRASY", "GHARIALS",
		"SHAMMING", "KIDNAPER", "FOILISTS", "CHURNERS", "QUINTALS", "EXPENSES",
		"KIPPERER", "DECENTLY", "CAMISOLE", "IMMESHED", "UNSTACKS", "CUMQUATS",
		"BEQUESTS", "ORGANISM", "HUGGIEST", "EPONYMIC", "HITHERTO", "VOLPLANE",
		"HUISACHE", "RIFFRAFF", "QUOINING", "ANTIHERO", "OUTSMOKE", "SEAWARES",
		"STOPPLED"}
	// Have all users guess all words.
	doneCh := make(chan string, len(users)*len(words))
	for _, word := range words {
		for _, user := range users {

			go func(user string, word string) {
				msg := channels.Message{
					Data:  word,
					Mtype: channels.MessageType(GuessMT),
					From:  user,
				}
				msg.SetRealm(realm)
				MessageHandler.HandleMessage(msg)
				doneCh <- user
			}(user, word)

		}
	}
	for i := 0; i < len(users)*len(words); i++ {
		<-doneCh
	}

}

// Test a whole game with 4 players. Use concurrency.
func TestSimpleGame(t *testing.T) {
	gameStates.reset()
	users.reset()
	realm := toRealm(tablenum)
	// Set mock so we don't connect to external API.
	MessageHandler.webolith = &MockWebolithCommunicator{}
	MessageHandler.sender = &MockMessageSender{}

	MessageHandler.RealmCreation(realm)
	userlist := []string{"cesar", "messi", "xavi", "iniesta"}
	joinSitting(userlist, realm)
	log.Printf("[DEBUG] About to start.")
	requestStart(userlist, realm)
	startAllowed := users.allowStart(realm)
	if !startAllowed {
		t.Fatalf("Should have been allowed to start")
	}
	guessWords(userlist, realm)

	scores := gameStates.scores(realm)
	log.Printf("Scores: %v", scores)
	sum := 0
	for _, score := range scores {
		sum += score
	}
	if sum != 53 {
		t.Errorf("Score total should have been 53.")
	}

}

func TestSameUserStart(t *testing.T) {
	gameStates.reset()
	users.reset()
	realm := toRealm(tablenum)
	// Set mock so we don't connect to external API.
	MessageHandler.webolith = &MockWebolithCommunicator{}
	MessageHandler.sender = &MockMessageSender{}

	MessageHandler.RealmCreation(realm)
	userlist := []string{"cesar", "cesar", "cesar", "cesar"}
	joinSitting(userlist, realm)
	log.Printf("[DEBUG] About to start.")
	requestStart(userlist, realm)
	startAllowed := users.allowStart(realm)
	if !startAllowed {
		t.Fatalf("Should have been allowed to start")
	}
	guessWords(userlist, realm)

	scores := gameStates.scores(realm)
	log.Printf("Scores: %v", scores)
	if scores["cesar"] != 53 {
		t.Errorf("Score for cesar should have been 53.")
	}
}
