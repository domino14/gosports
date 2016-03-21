package wordwalls

// This file contains functions for interfacing with the webolith API,
// for when we need to get word list, etc info.

import (
	"bytes"
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
	// Post a json-encoded buffer to a path, return a body.
	Post(path string, buf []byte) ([]byte, error)
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

func (w Webolith) Post(path string, buf []byte) ([]byte, error) {
	webolithUrl := os.Getenv("WEBOLITH_URL")
	if webolithUrl == "" {
		log.Println("[ERROR] No webolith")
		return nil, fmt.Errorf("no webolith url.")
	}
	resp, err := http.Post(webolithUrl+path, "application/json",
		bytes.NewBuffer(buf))
	if err != nil {
		log.Println("[ERROR]", err)
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func getWordList(w WebolithCommunicator, wordListId int) *WordList {
	lId := strconv.Itoa(wordListId)
	body, err := w.Get("/base/api/wordlist/" + lId + "?action=continue")
	if err != nil {
		log.Println("[ERROR] getting", err)
		return nil
	}
	list := &WordList{}
	err = json.Unmarshal(body, list)
	if err != nil {
		log.Println("[ERROR] Unmarshalling list", err)
		return nil
	}
	return list
}

type fullQRequest struct {
	Questions []Question `json:"questions"`
	Lexicon   string     `json:"lexicon"`
}

func getFullQInfo(w WebolithCommunicator, questions []Question,
	lexicon string) ([]byte, error) {

	fqr := fullQRequest{Questions: questions, Lexicon: lexicon}
	qs, err := json.Marshal(fqr)
	if err != nil {
		log.Println("[ERROR] Marshalling in getFullQInfo", err)
		return nil, err
	}
	body, err := w.Post("/base/api/word_db/full_questions/", qs)
	if err != nil {
		log.Println("[ERROR] posting", err)
		return nil, err
	}
	return body, nil
}

func getGameOptions(w WebolithCommunicator, table channels.Realm) *GameOptions {

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

// Synchronize word list state with the API.
func syncWordList(w WebolithCommunicator, list *WordList) {

}
