package main

import (
	"flag"
	"log"
	"net/http"
	"text/template"

	// "github.com/gorilla/rpc/v2"
	// "github.com/gorilla/rpc/v2/json2"

	"github.com/domino14/gosports/channels"
	"github.com/domino14/gosports/wordwalls"
)

var addr = flag.String("addr", ":8080", "http service address")
var homeTempl = template.Must(template.ParseFiles("home.html"))

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	homeTempl.Execute(w, r.Host)
}

func main() {
	flag.Parse()
	go channels.Hub.Run(wordwalls.MessageHandler)
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, r.URL.Path[1:])
	})
	http.HandleFunc("/ws", channels.ServeWs)

	// s := rpc.NewServer()
	// s.RegisterCodec(json2.NewCodec(), "application/json")
	// s.RegisterService(new(wordwalls.WordwallsService), "")
	// http.Handle("/rpc", s)

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
