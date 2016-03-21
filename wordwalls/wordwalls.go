package wordwalls

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/domino14/gosports/channels"
)

type GameType string

const (
	Challenge     GameType = "challenge"
	Regular       GameType = "regular"
	CountdownTime int      = 3
)

const (
	FailureSettingsDoNotExist = "SETTINGS_DONT_EXIST"
	FailureNotAllowed         = "START_NOT_ALLOWED"
	FailureNullWordList       = "NULL_WORD_LIST"
	FailureQuestionInfo       = "QUESTION_INFO"
	FailureGameGoing          = "GAME_GOING"
)

type GameOptions struct {
	QuestionsToPull  int    `json:"questionsToPull"`
	AnswersThisRound int    `json:"numAnswersThisRound"`
	GameType         string `json:"gameType"`
	ChallengeId      int    `json:"challengeId"`
	TimerSecs        int    `json:"timerSecs"`
	QualifyForAward  bool   `json:"qualifyForAward"`
	WordListID       int    `json:"_word_list_id"`
}

// CorrectAnswer encodes the index of the answer, the answer, the user
// who got it, and the user's score.
type CorrectAnswer struct {
	Answer    string `json:"answer"`
	Alphagram string `json:"alphagram"`
	User      string `json:"user"`
	Idx       int    `json:"idx"`
	Score     int    `json:"score"`
}

type wwMessageHandler struct {
	webolith WebolithCommunicator
	sender   channels.SocketMessageSender
}

type wwMessageSender struct{}

type MessageType string

type Answer struct {
	Alphagram string `json:"a"`
	Idx       int    `json:"i"`
}

const (
	ChatMT MessageType = "chat"
	// The data field would be the actual command, for example "start".
	TableMT MessageType = "tableCmd"
	GuessMT MessageType = "guess"

	CountdownMT channels.MessageType = "countdown"
	QuestionsMT channels.MessageType = "questions"
	TimerMT     channels.MessageType = "timer"
	GameOverMT  channels.MessageType = "gameover"
	ScoreMT     channels.MessageType = "score"
	FailMT      channels.MessageType = "fail"
)

func (s wwMessageSender) BroadcastMessage(realm channels.Realm,
	mt channels.MessageType, msg string) {
	channels.BroadcastMessage(realm, mt, msg)
}
func (s wwMessageSender) SendMessage(realm channels.Realm,
	mt channels.MessageType, msg string, to string) {
}

func (m wwMessageHandler) HandleMessage(msg channels.Message) {
	log.Println("[DEBUG] Got a message", msg.Data, msg.Mtype,
		msg.From, msg.Realm())
	switch MessageType(msg.Mtype) {
	case TableMT:
		log.Println("[DEBUG] Got a table command.")
		handleTableMessage(msg.Data, msg.Realm(), msg.From, m.webolith,
			m.sender)

	case GuessMT:
		handleGuess(msg.Data, msg.Realm(), msg.From, m.sender)
	}
}

var MessageHandler wwMessageHandler

func init() {
	// Create a new WebolithCommunicator for the message handler.
	MessageHandler.webolith = &Webolith{}
	MessageHandler.sender = &wwMessageSender{}
}

// On the creation of a new realm (table) get info about the game options
// for it.
// RealmCreation should only get called once (concurrently) due to the
// channels in the hub.
func (m wwMessageHandler) RealmCreation(table channels.Realm) {
	state := gameStates.createState(table)
	state.setOptions(getGameOptions(m.webolith, table))
	log.Println("[DEBUG] In realm creation. Game settings is now", state.options)
}

// On the deletion of a realm, clean up any timers, end games, save
// lists in progress, etc.
func (m wwMessageHandler) RealmDeletion(table channels.Realm) {
	gameStates.cleanlyStopGame(table, m.webolith)
	log.Println("[DEBUG] In realm deletion.")
}

// On joining a table, set users for this table.
// firstUser is true if this is the first user to join.
func (m wwMessageHandler) RealmJoin(table channels.Realm, user string,
	connId string, firstUser bool) {
	state := stWatching
	if firstUser {
		state = stSitting
	}
	users.add(table, user, state, connId)
}

// On leaving a table, remove from the list of users for this table.
func (m wwMessageHandler) RealmLeave(table channels.Realm, user string,
	connId string) {
	users.remove(table, user, connId)
	users.RLock()
	log.Printf("After RealmLeave: %v\n", users.userMap[table])
	users.RUnlock()
}

func handleTableMessage(data string, table channels.Realm, user string,
	wc WebolithCommunicator, sender channels.SocketMessageSender) {

	switch data {
	case "start":
		handleStart(table, user, wc, sender)
	}
}

// handle a Start message. We set a lock when someone clicks Start
// to prevent race conditions.
func handleStart(table channels.Realm, user string, wc WebolithCommunicator,
	sender channels.SocketMessageSender) {
	log.Println("[DEBUG] In handleStart....")
	sendFail := func(errorCode string) {
		sender.BroadcastMessage(table, FailMT, errorCode)
	}
	// XXX: Set a lock for this table on start.
	st := gameStates.getState(table)
	st.Lock()
	defer st.Unlock()

	if st.options == nil {
		// XXX: This should always be really quick but maybe once in a
		// while it'll fail; should prompt user to try again
		log.Println("[ERROR] Settings for this table do not yet exist!")
		sendFail(FailureSettingsDoNotExist)
		return
	}
	users.wantsToPlay(table, user)
	if !users.allowStart(table) {
		log.Println("[DEBUG] Start not yet allowed.")
		sendFail(FailureNotAllowed)
		return
	}

	if st.going != GameDone {
		log.Println("[DEBUG] This game is going or about to start.")
		sendFail(FailureGameGoing)
		return
	}
	st.going = GameInitializing
	wordList := getWordList(wc, st.options.WordListID)
	if wordList == nil {
		log.Println("[ERROR] Got nil word list, error!")
		sendFail(FailureNullWordList)
		return
	}
	st.setList(wordList)
	qToSend := st.nextQuestionSet(st.options.QuestionsToPull)
	// Turn the raw alphagrams into full question objects.
	fullQResponse, err := getFullQInfo(wc, qToSend, wordList.Lexicon)
	if err != nil {
		log.Println("[ERROR] Error getting full Q response!", err)
		sendFail(FailureQuestionInfo)
		return
	}
	log.Println("[DEBUG] Got full Q response:", string(fullQResponse))
	countdown := time.NewTimer(time.Second * time.Duration(CountdownTime))
	st.setCountdownTimer(countdown)
	st.going = GameCountingDown
	sender.BroadcastMessage(table, CountdownMT, strconv.Itoa(CountdownTime))
	// Countdown before starting game.
	// We should not accept guesses until the game has started.
	go handleGameTimer(table, countdown, string(fullQResponse), sender)
	log.Println("[DEBUG] Leaving start, mutex should unlock.")
}

func handleGameTimer(table channels.Realm, countdown *time.Timer,
	questionsToSend string, sender channels.SocketMessageSender) {

	<-countdown.C
	log.Println("[DEBUG] Finished counting down! About to send qs...")
	st := gameStates.getState(table)
	st.Lock()
	defer st.Unlock()
	st.going = GameStarted
	sender.BroadcastMessage(table, QuestionsMT, questionsToSend)
	sender.BroadcastMessage(table, TimerMT, strconv.Itoa(st.options.TimerSecs))
	// Start another nested goroutine here for game over. This looks
	// messy, but it seems easy enough to do.
	gameOver := time.NewTimer(time.Second * time.Duration(st.options.TimerSecs))
	st.setGameTimer(gameOver)
	go func() {
		<-gameOver.C
		log.Println("[DEBUG] This game is over!")
		st := gameStates.getState(table)
		st.Lock()
		defer st.Unlock()
		st.going = GameDone
		sender.BroadcastMessage(table, GameOverMT, "")
	}()
}

func handleGuess(data string, table channels.Realm, user string,
	sender channels.SocketMessageSender) {

	if gameStates.getGameGoing(table) != GameStarted {
		log.Println("[DEBUG] Got a guess when game had not started.")
		return
	}

	answer := gameStates.guess(data, table, user)
	if answer == nil {
		return
	}
	// Otherwise, broadcast correct guess and score!
	msg, err := json.Marshal(answer)
	if err != nil {
		log.Println("[ERROR] Marshalling answer", answer, err)
	}
	sender.BroadcastMessage(table, ScoreMT, string(msg))
}

type Alphagram struct {
	Alphagram string `json:"alphagram"`
	Idx       int    `json:"idx"`
}
