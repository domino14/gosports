package wordwalls

// This file contains functions for interfacing with the webolith API,
// for when we need to get word list, etc info.

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/domino14/gosports/channels"
)

type WebolithCommunicator interface {
	// Get a path and return a body.
	Get(path string) ([]byte, error)
}

type Webolith struct{}

func (w Webolith) Get(path string) ([]byte, error) {
	webolithUrl := os.Getenv("WEBOLITH_URL")
	if webolithUrl == "" {
		log.Println("[ERROR] No webolith")
		return nil, fmt.Errorf("no webolith url.")
	}
	resp, err := http.Get(webolithUrl + path)
	if err != nil {
		log.Println("[ERROR]", err)
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func getWordList(wordListId int, w WebolithCommunicator) *WordList {
	lId := strconv.Itoa(wordListId)
	body, err := w.Get("/base/api/wordlist/" + lId)
	if err != nil {
		log.Println("[ERROR] getting", err)
		return nil
	}
	log.Println("[DEBUG]", string(body))
	list := &WordList{}
	err = json.Unmarshal(body, list)
	if err != nil {
		log.Println("[ERROR] Unmarshalling list", err)
		return nil
	}
	return list
}

func getGameOptions(table channels.Realm, w WebolithCommunicator) *GameOptions {

	body, err := w.Get("/wordwalls/api/game_options/" + string(table) + "/")
	if err != nil {
		log.Println("[ERROR] getting", err)
		return nil
	}
	gameOptions := &GameOptions{}
	log.Println("[DEBUG] Going to unmarshal", string(body))
	err = json.Unmarshal(body, gameOptions)
	if err != nil {
		log.Println("[ERROR] Unmarshalling options", err)
		return nil
	}
	return gameOptions
}
