package main

import (
	"encoding/json"
	"sync"
	//"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
)

const pfx = "/hanabi/"

func main() {
	port := flag.Int("port", 9001, "port to listen on")
	flag.Parse()
	serveStr := fmt.Sprintf(":%v", *port)
	log.Printf("Serving at localhost%v", serveStr)
	server := NewServer()
	http.HandleFunc(pfx, server.Handle)
	log.Fatal(http.ListenAndServe(serveStr, nil))
}

type Server struct {
	state ServerState
}

type ServerState struct {
	Games        map[string]*Game
	GamesMapLock sync.Mutex // Lock that guards the mapping, not the Games.
}

// Get a game. Acquires GamesMapLock. Can return nil.
func (s *ServerState) lookupGame(name string) *Game {
	s.GamesMapLock.Lock()
	defer s.GamesMapLock.Unlock()
	game, _ := s.Games[name]
	return game
}

func NewServer() *Server {
	return &Server{
		state: ServerState{
			Games: make(map[string]*Game),
		},
	}
}

func (s *Server) Handle(w http.ResponseWriter, req *http.Request) {
	log.Printf("Request to %v", req.URL.Path)
	dec := json.NewDecoder(req.Body)
	var response interface{}
	var err error
	if strings.HasPrefix(req.URL.Path, pfx+"start-game") {
		var v StartGameRequest
		if err = dec.Decode(&v); handleErr(err, w) {
			return
		}
		response, err = StartGame(&s.state, v)
	} else if strings.HasPrefix(req.URL.Path, pfx+"join-game") {
		var v JoinGameRequest
		if err = dec.Decode(&v); handleErr(err, w) {
			return
		}
		response = JoinGame(&s.state, v)
	} else {
		w.WriteHeader(404)
		return
	}
	if err == nil {
		writeJson(w, response)
	} else {
		handleErr(err, w)
	}
}

func okResponse() []byte {
	return []byte("{status: \"ok\"}")
}

func writeJson(w http.ResponseWriter, obj interface{}) {
	respStr, err := json.Marshal(obj)
	if err != nil {
		log.Printf("Error during JSON marshal: %v\n", err)
		return
	}
	w.Write([]byte(respStr))
}

// true if there was an error that we handled
func handleErr(err error, w http.ResponseWriter) bool {
	if err != nil {
		log.Printf("Error: %v", err.Error())
		w.WriteHeader(500)
		writeJson(w, struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		}{
			Status:  "error",
			Message: err.Error(),
		})
		return true
	}
	return false
}
