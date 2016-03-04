package wordwalls

import (
	"testing"
)

const (
	tableName string = "footable"
	username  string = "cesar"
)

func TestJoin(t *testing.T) {
	realm := toRealm(tableName)
	users.add(realm, username, stSitting)
	users.add(realm, username, stWatching)
	if len(users.userMap[realm]) != 1 {
		t.Errorf("Length should have been 1.")
	}
}

func TestJoinAndLeave(t *testing.T) {
	realm := toRealm(tableName)
	users.add(realm, username, stSitting)
	users.add(realm, username, stWatching)
	users.remove(realm, username)
	if len(users.userMap[realm]) != 0 {
		t.Errorf("Length should have been 0.")
	}
}

func TestJoinLeaveAndStart(t *testing.T) {
	realm := toRealm(tableName)
	users.add(realm, username, stSitting)
	users.add(realm, username, stWatching)
	users.remove(realm, username)
	users.wantsToPlay(realm, username)
	if len(users.userMap[realm]) != 0 {
		t.Errorf("Length should have been 0.")
	}
	if users.allowStart(realm) != false {
		t.Errorf("Should not be allowed to start")
	}
}
