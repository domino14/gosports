package wordwalls

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
	// Needs a mutex, or we do this in Redis.
	AnswerHash          map[string]*Alphagram `json:"answerHash"`
	QuestionsToPull     int                   `json:"questionsToPull"`
	QuizGoing           bool                  `json:"quizGoing"`
	QuizStartTime       int                   `json:"quizStartTime"`
	NumAnswersThisRound int                   `json:"numAnswersThisRound"`
	GameType            GameType              `json:"gameType"`
	ChallengeId         int                   `json:"challengeId"`
	TimerSecs           int                   `json:"timerSecs"`
	TimeRemaining       int                   `json:"timeRemaining"`
	QualifyForAward     bool                  `json:"qualifyForAward"`
	SaveName            string                `json:"saveName"`
}
