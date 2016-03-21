package wordwalls

import (
	"log"
	"sync"
	"time"

	"github.com/domino14/gosports/channels"
)

// This file will contain the global maps that represent game states and
// options. The maps are protected by sync.RWMutex because maps
// are not thread-safe by default.
type gameGoingState int

const (
	GameInitializing = iota // Requesting all list info, etc.
	GameCountingDown        // In 3-second countdown
	GameStarted             // Game is going
	GameDone                // Game is finished or has never started
)

// gameState represents the state for a single game. This has a lock
// to protect its inner members.
type gameState struct {
	scores         map[string]int
	options        *GameOptions
	list           *WordList
	going          gameGoingState
	countdownTimer *time.Timer
	gameTimer      *time.Timer
	sync.RWMutex
}

// gamestatePopulation represents the states for all games. We protect
// the outer map here with a lock as well, but only when adding/removing
// new states.
type gamestatePopulation struct {
	sync.RWMutex
	stateMap map[channels.Realm]*gameState
}

var gameStates gamestatePopulation

func init() {
	gameStates.reset()
}

func (gs *gamestatePopulation) reset() {
	gameStates.stateMap = make(map[channels.Realm]*gameState)
}

func (gs *gamestatePopulation) createState(table channels.Realm) *gameState {
	gs.Lock()
	defer gs.Unlock()

	log.Println("[DEBUG] In createState: %v", table)
	state := &gameState{}
	// Start in the "Done" state.
	state.going = GameDone
	gs.stateMap[table] = state
	return state
}

// Send a save command to the API for the word list, prior
// to quitting, cancel timers as well.
func (gs *gamestatePopulation) cleanlyStopGame(table channels.Realm,
	w WebolithCommunicator) {

	st := gs.getState(table)

	st.Lock()
	defer st.Unlock()
	st.list.saveProgress(w)
	st.cancelTimers()
}

// Cleanly delete the state for this wordwalls game. We want to stop
// the game.
//
// We want to do this in such a way so that this locks the re-creation of
// this state until it is done deleting.
//
// XXX: TODO: A periodic function that deletes stale states. We won't
// delete them with a function here due to sync issues.

// func (gs *gamestatePopulation) deleteState(table channels.Realm) {
// 	gs.Lock()
// 	defer gs.Unlock()

// 	log.Println("[DEBUG] In deleteState: %v", table)
// 	state, ok := gs.stateMap[table]
// 	state.Lock()
// 	defer state.Unlock()
// 	if !ok {
// 		log.Println("[ERROR] Going to delete table from state map, but it "+
// 			"is not there!", table)
// 		return
// 	}
// 	state.cancelTimers()
// 	delete(gs.stateMap, table)

// }

func (gs *gamestatePopulation) getState(table channels.Realm) *gameState {
	gs.RLock()
	defer gs.RUnlock()
	return gs.stateMap[table]
}

func (s *gameState) setOptions(options *GameOptions) {
	s.Lock()
	defer s.Unlock()
	s.options = options
}

func (s *gameState) setList(list *WordList) {
	s.list = list
	// Make a new scores map too.
	s.scores = make(map[string]int)
}

func (s *gameState) nextQuestionSet(numQuestions int) []Question {
	if s.list == nil {
		return nil
	}
	// Reset the scores. Setting to a new map should hopefully
	// GC the old one :P
	s.scores = make(map[string]int)
	return s.list.nextSet(numQuestions)
}

// Check if guess is in answer hash. If it is, increase user score by 1.
// XXX: Nil pointer errors in gs.listMap[table] if we restart go server
// while game is going. We will need to save state between restarts
// somehow.
func (gs *gamestatePopulation) guess(data string, table channels.Realm,
	user string) *CorrectAnswer {
	state := gs.getState(table)
	return state.guess(data, user)
}

func (s *gameState) guess(data string, user string) *CorrectAnswer {
	s.Lock()
	defer s.Unlock()
	if answer, ok := s.list.answerHash[data]; ok {
		ca := &CorrectAnswer{}
		ca.Answer = data
		ca.Idx = answer.Idx
		ca.User = user
		ca.Alphagram = answer.Alphagram
		delete(s.list.answerHash, data)
		// nil value for int is 0 so this will work.
		s.scores[user] = s.scores[user] + 1
		ca.Score = s.scores[user]
		return ca
	}
	return nil
}

func (gs *gamestatePopulation) scores(table channels.Realm) map[string]int {
	state := gs.getState(table)
	return state.scores
}

func (gs *gamestatePopulation) timer(table channels.Realm) int {
	state := gs.getState(table)
	state.RLock()
	defer state.RUnlock()
	return state.options.TimerSecs
}

func (gs *gamestatePopulation) getGameGoing(table channels.Realm) gameGoingState {
	state := gs.getState(table)
	state.RLock()
	defer state.RUnlock()
	return state.going
}

// Since this timer is set while the game state is locked, we call it
// as a method on gameState
func (s *gameState) setCountdownTimer(ct *time.Timer) {
	s.countdownTimer = ct
}

func (s *gameState) setGameTimer(t *time.Timer) {
	s.gameTimer = t
}

func (s *gameState) cancelTimers() {
	if s.gameTimer != nil {
		cl1 := s.gameTimer.Stop()
		log.Printf("[DEBUG] Canceling gameTimer: %v", cl1)
	}
	if s.countdownTimer != nil {
		cl2 := s.countdownTimer.Stop()
		log.Printf("[DEBUG] Canceling countdownTimer: %v", cl2)
	}
}
