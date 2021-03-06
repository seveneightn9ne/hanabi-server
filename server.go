package main

import (
	"encoding/json"
	"sync"
	//"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
)

const pfx = "/hanabi/"

func main() {
	port := flag.Int("port", 9001, "port to listen on")
	flag.Parse()
	serveStr := fmt.Sprintf(":%v", *port)
	log.Printf("Serving at localhost%v", serveStr)
	server := NewServer()

	path := "/hanabi/start-game"
	http.HandleFunc(path, server.MakeHandler(path, StartGame, &StartGameRequest{}))
	path = "/hanabi/join-game"
	http.HandleFunc(path, server.MakeHandler(path, JoinGame, &JoinGameRequest{}))
	path = "/hanabi/get-state"
	http.HandleFunc(path, server.MakeHandler(path, GetState, &GetStateRequest{}))
	path = "/hanabi/move"
	http.HandleFunc(path, server.MakeHandler(path, MoveHandler, &MoveRequest{}))
	log.Fatal(http.ListenAndServe(serveStr, nil))
}

type Server struct {
	state ServerState
}

type ServerState struct {
	Games        map[string]*Game
	Sessions     map[SessionToken]*Game
	GamesMapLock sync.Mutex // Lock that guards the mappings, not the Games.
}

// Get a game. Acquires GamesMapLock. Can return nil.
func (s *ServerState) lookupGame(name string) *Game {
	s.GamesMapLock.Lock()
	defer s.GamesMapLock.Unlock()
	game, _ := s.Games[name]
	return game
}

func (s *ServerState) gameForSession(session SessionToken) *Game {
	s.GamesMapLock.Lock()
	defer s.GamesMapLock.Unlock()
	game, _ := s.Sessions[session]
	return game
}

func (s *ServerState) addSession(session SessionToken, game *Game) {
	s.GamesMapLock.Lock()
	defer s.GamesMapLock.Unlock()
	s.Sessions[session] = game
}

func NewServer() *Server {
	return &Server{
		state: ServerState{
			Games:    make(map[string]*Game),
			Sessions: make(map[SessionToken]*Game),
		},
	}
}

func (s *Server) MakeHandler(path string, f func(*ServerState, interface{}) interface{}, requestStruct interface{}) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		log.Printf("Request to %v", req.URL.Path)
		if req.Method != "POST" {
			err := fmt.Errorf("request type %v != POST", req.Method)
			handleErr(err, w)
			return
		}
		dec := json.NewDecoder(req.Body)
		var response interface{}
		var err error
		err = dec.Decode(&requestStruct)
		if err != nil {
			err = fmt.Errorf("error decoding request: %v", err)
			handleErr(err, w)
			return
		}
		response = f(&s.state, requestStruct)
		writeJson(w, response)
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
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(respStr))
}

// true if there was an error that we handled
func handleErr(err error, w http.ResponseWriter) bool {
	if err != nil {
		log.Printf("Error: %v", err.Error())
		w.WriteHeader(500)
		writeJson(w, struct {
			Status string `json:"status"`
			Reason string `json:"reason"`
		}{
			Status: "error",
			Reason: err.Error(),
		})
		return true
	}
	return false
}
