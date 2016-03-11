package wordwalls

import (
	"log"
	"sync"

	"github.com/domino14/gosports/channels"
)

// This file will contain the global maps that represent game states and
// options. The maps are protected by sync.RWMutex because maps
// are not thread-safe by default.

// gameState represents the state for a single game. This has a lock
// to protect its inner members.
type gameState struct {
	scores  map[string]int
	options *GameOptions
	list    *WordList
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
	state := &gameState{}
	gs.stateMap[table] = state
	return state
}

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

func (gs *gamestatePopulation) setList(table channels.Realm, list *WordList) {
	state := gs.getState(table)
	state.setList(list)
}

func (s *gameState) setList(list *WordList) {
	s.Lock()
	defer s.Unlock()
	s.list = list
	// Make a new scores map too.
	s.scores = make(map[string]int)
}

func (gs *gamestatePopulation) nextSet(table channels.Realm,
	numQuestions int) []Question {

	state := gs.getState(table)
	return state.nextQuestionSet(numQuestions)
}

func (s *gameState) nextQuestionSet(numQuestions int) []Question {
	s.Lock()
	defer s.Unlock()
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
	log.Printf("List is %v", s.list)
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

// XXX: does this have to be * for the mutex to work?
func (gs *gamestatePopulation) wordListID(table channels.Realm) int {
	state := gs.getState(table)
	state.RLock()
	defer state.RUnlock()
	return state.options.WordListID
}

func (gs *gamestatePopulation) numQuestions(table channels.Realm) int {
	state := gs.getState(table)
	state.RLock()
	defer state.RUnlock()
	return state.options.QuestionsToPull
}

func (gs *gamestatePopulation) exists(table channels.Realm) bool {
	state := gs.getState(table)
	state.RLock()
	defer state.RUnlock()
	return state.options != nil
}
