package wordwalls

import (
	"fmt"
	"log"
)

type GameType string

const (
	Challenge GameType = "challenge"
	Regular   GameType = "regular"
)

type Alphagram struct {
	Alphagram string `json:"alphagram"`
	Idx       int    `json:"idx"`
}

type GameState struct {
	// Needs a mutex, or we do this in Redis... or some channel trickery.
	AnswerHash          map[string]*Alphagram `json:"answerHash"`
	QuestionsToPull     int                   `json:"questionsToPull"`
	QuizGoing           bool                  `json:"quizGoing"`
	QuizStartTime       int                   `json:"quizStartTime"`
	NumAnswersThisRound int                   `json:"numAnswersThisRound"`
	GameType            GameType              `json:"gameType",omitempty`
	ChallengeId         int                   `json:"challengeId",omitempty`
	TimerSecs           int                   `json:"timerSecs",omitempty`
	TimeRemaining       int                   `json:"timeRemaining",omitempty`
	QualifyForAward     bool                  `json:"qualifyForAward",omitempty`
	SaveName            string                `json:"saveName",omitempty`
}

func (gs *GameState) init() {

}

// startQuiz starts a quiz.
func (gs *GameState) startQuiz() error {
	if gs.QuizGoing {
		log.Println("[DEBUG] Quiz is going, cannot start.")
		return fmt.Errorf("Cannot start quiz, since it's already going.")
	}

	return nil
}
