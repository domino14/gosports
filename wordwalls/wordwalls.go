package wordwalls

import (
	"encoding/json"
	"log"

	"github.com/domino14/gosports/channels"
)

type GameType string

const (
	Challenge GameType = "challenge"
	Regular   GameType = "regular"
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

type wwMessageHandler struct{}
type MessageType string

type Answer struct {
	Alphagram string `json:"a"`
	Idx       int    `json:"i"`
}

// type StartMessage struct {
// 	AnswerHash Answers `json:"answerHash"`
// 	// We don't care about Questions. This will just get passed on to
// 	// the user.
// 	Questions  interface{} `json:"questions"`
// 	QBegin     int         `json:"qbegin"`
// 	QEnd       int         `json:"qend"`
// 	QTotal     int         `json:"qtotal"`
// 	NumAnswers int         `json:"numAnswersThisRound"`
// }

const (
	ChatMT MessageType = "chat"
	// The data field would be the actual command, for example "start".
	TableMT MessageType = "tableCmd"
	GuessMT MessageType = "guess"
)

func (m wwMessageHandler) HandleMessage(msg channels.Message) {
	log.Println("[DEBUG] Got a message", msg.Data, msg.Mtype,
		msg.From, msg.Realm())
	switch MessageType(msg.Mtype) {
	case TableMT:
		log.Println("[DEBUG] Got a table command.")
		handleTableMessage(msg.Data, msg.Realm(), msg.From)

	case GuessMT:
		handleGuess(msg.Data, msg.Realm(), msg.From)
	}
}

// On the creation of a new realm (table) get info about the game options
// for it.
func (m wwMessageHandler) RealmCreation(table channels.Realm) {
	options := getGameOptions(table)
	gameSettings.set(table, options)
	log.Println("[DEBUG] Game settings is now", gameSettings.options[table])
}

// On joining a table, set users for this table.
// firstUser is true if this is the first user to join.
func (m wwMessageHandler) RealmJoin(table channels.Realm, user string,
	firstUser bool) {
	state := stWatching
	if firstUser {
		state = stSitting
	}
	users.add(table, user, state)
}

// On leaving a table, remove from the list of users for this table.
func (m wwMessageHandler) RealmLeave(table channels.Realm, user string) {
	users.remove(table, user)
}

func handleTableMessage(data string, table channels.Realm, user string) {
	switch data {
	case "start":
		if !gameSettings.exists(table) {
			// XXX: This should always be really quick but maybe once in a
			// while it'll fail; should prompt user to try again
			log.Println("[ERROR] Settings for this table do not yet exist!")
			return
		}
		users.wantsToPlay(table, user)
		if !users.allowStart(table) {
			log.Println("[DEBUG] Start not yet allowed.")
			return
		}
		// XXX: Check if the game has already started. We don't want to
		// do this twice. (This could be a race condition)
		wordList := getWordList(gameSettings.wordListID(table))
		if wordList == nil {
			log.Println("[ERROR] Got nil word list, error!")
			return
		}
		gameStates.set(table, wordList)
		qToSend := gameStates.nextSet(table, gameSettings.numQuestions(table))
		// and send questions
		res, err := json.Marshal(qToSend)
		if err != nil {
			log.Println("[ERROR] Error marshalling questions to send!", err)
			return
		}
		channels.BroadcastMessage(table, "questions", string(res))
	}
}

func handleGuess(data string, table channels.Realm, user string) {
	answer := gameStates.guess(data, table, user)
	if answer == nil {
		return
	}
	// Otherwise, broadcast correct guess and score!
	msg, err := json.Marshal(answer)
	if err != nil {
		log.Println("[ERROR] Marshalling answer", answer, err)
	}
	channels.BroadcastMessage(table, "score", string(msg))
}

var MessageHandler wwMessageHandler

type Alphagram struct {
	Alphagram string `json:"alphagram"`
	Idx       int    `json:"idx"`
}

// type GameState struct {
// 	// Needs a mutex, or we do this in Redis... or some channel trickery.
// 	AnswerHash          map[string]*Alphagram `json:"answerHash"`
// 	QuestionsToPull     int                   `json:"questionsToPull"`
// 	QuizGoing           bool                  `json:"quizGoing"`
// 	QuizStartTime       int                   `json:"quizStartTime"`
// 	NumAnswersThisRound int                   `json:"numAnswersThisRound"`
// 	GameType            GameType              `json:"gameType",omitempty`
// 	ChallengeId         int                   `json:"challengeId",omitempty`
// 	TimerSecs           int                   `json:"timerSecs",omitempty`
// 	TimeRemaining       int                   `json:"timeRemaining",omitempty`
// 	QualifyForAward     bool                  `json:"qualifyForAward",omitempty`
// 	SaveName            string                `json:"saveName",omitempty`
// }

// func (gs *GameState) init() {

// }

// // startQuiz starts a quiz.
// func (gs *GameState) startQuiz() error {
// 	if gs.QuizGoing {
// 		log.Println("[DEBUG] Quiz is going, cannot start.")
// 		return fmt.Errorf("Cannot start quiz, since it's already going.")
// 	}

// 	return nil
// }
