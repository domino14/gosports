package wordwalls

import (
	"log"
	"sync"

	"github.com/domino14/gosports/channels"
)

type UserState int

const (
	// A user who wants to play has clicked start. They can only click
	// start if they are sitting. If they are watching they have no
	// bearing on the game.
	stWantsToPlay UserState = iota
	stWatching
	stSitting
)

type userPopulation struct {
	sync.RWMutex
	userMap map[channels.Realm]map[string]UserState
}

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

var users userPopulation

func init() {
	users.userMap = make(map[channels.Realm]map[string]UserState)
}

func toRealm(tbl string) channels.Realm {
	return channels.Realm(tbl)
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
	log.Printf("Users here: %v, count: %d\n", u.userMap[table],
		len(u.userMap[table]))
}

// Only allow start if the users in this table meet the following conditions:
// - At least one is in state stWantsToPlay
// - All other users (can be none) are in state stWatching
func (u *userPopulation) allowStart(table channels.Realm) bool {
	u.RLock()
	defer u.RUnlock()
	log.Printf("In allow start, users here: %v, count: %d\n", u.userMap[table],
		len(u.userMap[table]))
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
	log.Printf("Want to play: %v, watching: %v, total: %v\n",
		numWantToPlay, numWatching, numTotal)
	return numWantToPlay > 0 && numWantToPlay+numWatching == numTotal
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
	if _, ok := usersHere[username]; ok {
		usersHere[username] = state
	} else {
		log.Printf("[ERROR] User %s not in table %s\n", username, table)
	}
}
