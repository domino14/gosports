package wordwalls

import (
	"log"
	"sync"

	"github.com/domino14/gosports/channels"
)

// This file will contain the global maps that represent game states and
// options. The maps are protected by sync.RWMutex because maps
// are not thread-safe by default.

// gameState represents the state for a single game
type gameState struct {
	scores map[string]int
}

// A map of table id to a GameOptions object.
type settingPopulation struct {
	sync.RWMutex
	options map[channels.Realm]*GameOptions
}

type userPopulation struct {
	sync.RWMutex
	userMap map[channels.Realm]map[string]UserState
}

type gamestatePopulation struct {
	sync.RWMutex
	listMap  map[channels.Realm]*WordList
	stateMap map[channels.Realm]*gameState
}

type UserState int

const (
	// A user who wants to play has clicked start. They can only click
	// start if they are sitting. If they are watching they have no
	// bearing on the game.
	stWantsToPlay UserState = iota
	stWatching
	stSitting
)

func (u UserState) String() string {
	switch u {
	case stWantsToPlay:
		return "WantsToPlay"
	case stWatching:
		return "Watching"
	case stSitting:
		return "Sitting"
	}
	return "UnknownState!"
}

var gameSettings settingPopulation
var users userPopulation
var gameStates gamestatePopulation

func init() {
	gameSettings.options = make(map[channels.Realm]*GameOptions)
	users.userMap = make(map[channels.Realm]map[string]UserState)
	gameStates.listMap = make(map[channels.Realm]*WordList)
	gameStates.stateMap = make(map[channels.Realm]*gameState)
}

// XXX: Is there a way to only mutex-lock specific maps?
// (There probably is)

func (gs *gamestatePopulation) set(table channels.Realm, list *WordList) {
	gs.Lock()
	defer gs.Unlock()
	gs.listMap[table] = list
	// Make a new scores map when we set a new list.
	gs.stateMap[table] = &gameState{}
	gs.stateMap[table].scores = make(map[string]int)
}

func (gs *gamestatePopulation) nextSet(table channels.Realm,
	numQuestions int) []Question {
	gs.Lock()
	defer gs.Unlock()
	list := gs.listMap[table]
	if list == nil {
		return nil
	}
	// Reset the scores. Setting to a new map should hopefully
	// GC the old one :P
	gs.stateMap[table].scores = make(map[string]int)
	return gs.listMap[table].nextSet(numQuestions)
}

// Check if guess is in answer hash. If it is, increase user score by 1.
// XXX: Nil pointer errors in gs.listMap[table] if we restart go server
// while game is going. We will need to save state between restarts
// somehow.
func (gs *gamestatePopulation) guess(data string, table channels.Realm,
	user string) *CorrectAnswer {
	gs.Lock()
	defer gs.Unlock()
	if answer, ok := gs.listMap[table].answerHash[data]; ok {
		ca := &CorrectAnswer{}
		ca.Answer = data
		ca.Idx = answer.Idx
		ca.User = user
		ca.Alphagram = answer.Alphagram
		delete(gs.listMap[table].answerHash, data)
		// nil value for int is 0 so this will work.
		gs.stateMap[table].scores[user] = gs.stateMap[table].scores[user] + 1
		ca.Score = gs.stateMap[table].scores[user]
		return ca
	}
	return nil
}

func (s *settingPopulation) set(table channels.Realm, options *GameOptions) {
	s.Lock()
	defer s.Unlock()
	s.options[table] = options
}

// XXX: does this have to be * for the mutex to work?
func (s settingPopulation) wordListID(table channels.Realm) int {
	s.RLock()
	defer s.RUnlock()
	return s.options[table].WordListID
}

func (s settingPopulation) numQuestions(table channels.Realm) int {
	s.RLock()
	defer s.RUnlock()
	return s.options[table].QuestionsToPull
}

func (s *settingPopulation) exists(table channels.Realm) bool {
	s.RLock()
	defer s.RUnlock()
	_, ok := s.options[table]
	return ok
}

func (u *userPopulation) add(table channels.Realm, username string,
	state UserState) {
	log.Printf("Adding user %s to table %s in state %v\n", username, table,
		state)
	u.Lock()
	defer u.Unlock()
	usersHere := u.userMap[table]
	if usersHere == nil {
		usersHere = make(map[string]UserState)
		u.userMap[table] = usersHere
	}
	usersHere[username] = state
}

func (u *userPopulation) remove(table channels.Realm, username string) {
	log.Printf("Removing user %s from table %s\n", username, table)
	u.Lock()
	defer u.Unlock()
	usersHere := u.userMap[table]
	if usersHere != nil {
		// XXX: Error check?
		delete(usersHere, username)
	}
}

// Only allow start if the users in this table meet the following conditions:
// - At least one is in state stWantsToPlay
// - All other users (can be none) are in state stWatching
func (u userPopulation) allowStart(table channels.Realm) bool {
	u.RLock()
	defer u.RUnlock()
	usersHere := u.userMap[table]
	numWantToPlay := 0
	numWatching := 0
	numTotal := len(usersHere)
	for _, state := range usersHere {
		if state == stWantsToPlay {
			numWantToPlay++
		} else if state == stWatching {
			numWatching++
		}
	}
	return numWantToPlay+numWatching == numTotal
}

func (u *userPopulation) wantsToPlay(table channels.Realm, username string) {
	log.Printf("[DEBUG] User %s wants to play on table %s\n", username, table)
	u.modifyState(table, username, stWantsToPlay)
}

func (u *userPopulation) watching(table channels.Realm, username string) {
	log.Printf("[DEBUG] User %s is watching on table %s\n", username, table)
	u.modifyState(table, username, stWatching)
}

func (u *userPopulation) sitting(table channels.Realm, username string) {
	log.Printf("[DEBUG] User %s is sitting on table %s\n", username, table)
	u.modifyState(table, username, stSitting)
}

func (u *userPopulation) modifyState(table channels.Realm, username string,
	state UserState) {
	u.Lock()
	defer u.Unlock()
	usersHere := u.userMap[table]
	usersHere[username] = state
}
