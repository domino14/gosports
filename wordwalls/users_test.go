package wordwalls

import (
	"testing"
)

const (
	tableName string = "footable"
	username  string = "cesar"
)

func TestJoin(t *testing.T) {
	gameStates.reset()
	users.reset()
	realm := toRealm(tableName)
	users.add(realm, username, stSitting, "id1")
	users.add(realm, username, stWatching, "id2")
	if len(users.userMap[realm]) != 1 {
		t.Errorf("Length should have been 1.")
	}
}

func TestJoinAndLeave(t *testing.T) {
	gameStates.reset()
	users.reset()
	realm := toRealm(tableName)
	users.add(realm, username, stSitting, "id1")
	users.add(realm, username, stWatching, "id2")
	users.remove(realm, username, "id2")
	if len(users.userMap[realm]) != 1 {
		t.Errorf("Length should have been 1.")
	}
}

func TestSingleJoinAndLeave(t *testing.T) {
	gameStates.reset()
	users.reset()
	realm := toRealm(tableName)
	users.add(realm, username, stSitting, "id1")
	users.remove(realm, username, "id1")
	if len(users.userMap[realm]) != 0 {
		t.Errorf("Length should have been 0.")
	}
}

func TestSingleJoinAndLeaveMismatchedIds(t *testing.T) {
	gameStates.reset()
	users.reset()
	realm := toRealm(tableName)
	users.add(realm, username, stSitting, "id1")
	users.remove(realm, username, "id2")
	// Mistake so don't remove it.
	if len(users.userMap[realm]) != 1 {
		t.Errorf("Length should have been 1.")
	}
}

func TestJoinLeaveAndStart(t *testing.T) {
	gameStates.reset()
	users.reset()
	realm := toRealm(tableName)
	users.add(realm, username, stSitting, "id1")
	users.add(realm, username, stWatching, "id2")
	users.remove(realm, username, "id2")
	users.wantsToPlay(realm, username)
	if len(users.userMap[realm]) != 1 {
		t.Errorf("Length should have been 1.")
	}
	if users.allowStart(realm) != true {
		t.Errorf("Should have been allowed to start")
	}
}
