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

// processStart sends a request for a start to the Webolith api, given
// the table, and returns a processed response.
// XXX: This function is very similar to getGameOptions. Perhaps use an
// interface.
// func processStart(table string) *StartMessage {
// 	webolithUrl := os.Getenv("WEBOLITH_URL")
// 	if webolithUrl == "" {
// 		log.Println("[ERROR] No webolith")
// 		return nil
// 	}
// 	resp, err := http.Post(webolithUrl+"/wordwalls/api/start_game/"+table+"/",
// 		"application/json", nil)
// 	if err != nil {
// 		log.Println("[ERROR] processStart error", err)
// 		return nil
// 	}
// 	defer resp.Body.Close()
// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Println("[ERROR] Reading all", err)
// 		return nil
// 	}
// 	startMessage := &StartMessage{}
// 	err = json.Unmarshal(body, startMessage)
// 	log.Println("[DEBUG] Got body", string(body))
// 	if err != nil {
// 		log.Println("[ERROR] Unmarshalling", err)
// 		return nil
// 	}
// 	return startMessage
// }

// Gets the path from Webolith into a byte array
func getWebolithPath(path string) ([]byte, error) {
	webolithUrl := os.Getenv("WEBOLITH_URL")
	if webolithUrl == "" {
		log.Println("[ERROR] No webolith")
		return nil, fmt.Errorf("no webolith url.")
	}
	resp, err := http.Get(webolithUrl + path)
	if err != nil {
		log.Println("[ERROR] getGameOptions error", err)
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func getWordList(wordListId int) *WordList {
	lId := strconv.Itoa(wordListId)
	body, err := getWebolithPath("/base/api/wordlist/" + lId)
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

func getGameOptions(table channels.Realm) *GameOptions {

	body, err := getWebolithPath("/wordwalls/api/game_options/" +
		string(table) + "/")
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
