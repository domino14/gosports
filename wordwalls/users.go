package wordwalls

import (
	"log"
	"sync"

	"github.com/domino14/gosports/channels"
)

type UserState int

type UserInfo struct {
	connIds map[string]bool
	state   UserState
}

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
	userMap map[channels.Realm]map[string]*UserInfo
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
	users.reset()
}

func toRealm(tbl string) channels.Realm {
	return channels.Realm(tbl)
}

func (u *userPopulation) reset() {
	users.userMap = make(map[channels.Realm]map[string]*UserInfo)
}

func (u *userPopulation) add(table channels.Realm, username string,
	state UserState, connId string) {
	log.Printf("Adding user %s to table %s in state %v, connId %s\n", username,
		table, state, connId)
	u.Lock()
	defer u.Unlock()
	usersHere := u.userMap[table]
	if usersHere == nil {
		usersHere = make(map[string]*UserInfo)
		u.userMap[table] = usersHere
	}
	uInfo := usersHere[username]
	if uInfo == nil {
		uInfo = &UserInfo{}
		uInfo.connIds = make(map[string]bool)
	}
	uInfo.connIds[connId] = true
	uInfo.state = state
}

func (u *userPopulation) remove(table channels.Realm, username string,
	connId string) {
	log.Printf("Removing user %s from table %s, connId %s\n", username, table,
		connId)
	u.Lock()
	defer u.Unlock()
	usersHere := u.userMap[table]
	if usersHere != nil {
		uInfo := usersHere[username]
		if uInfo != nil {
			delete(uInfo.connIds, connId)
			if len(uInfo.connIds) == 0 {
				delete(usersHere, username)
			}
		}
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
	for _, user := range usersHere {
		if user.state == stWantsToPlay {
			numWantToPlay++
		} else if user.state == stWatching {
			numWatching++
		}
	}
	log.Printf("Want to play: %v, watching: %v, total: %v\n",
		numWantToPlay, numWatching, numTotal)
	allow := numWantToPlay > 0 && numWantToPlay+numWatching == numTotal
	log.Printf("Allow returning: %v", allow)
	return allow
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
	if uInfo, ok := usersHere[username]; ok {
		uInfo.state = state
	} else {
		log.Printf("[ERROR] User %s not in table %s (%v)\n", username, table,
			usersHere)
	}
}
