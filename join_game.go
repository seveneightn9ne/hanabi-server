package main

import (
	"fmt"
	"log"
)

type JoinGameRequest struct {
	GameName   string `json:"game_name"`
	PlayerName string `json:"player_name"`
}

type JoinGameResponse struct {
	Status  string       `json:"status"`
	Reason  string       `json:"reason,omitempty"`
	Session SessionToken `json:"session,omitempty"`
}

func NewJoinGameResponseError(reason string) *JoinGameResponse {
	return &JoinGameResponse{
		Status: "error",
		Reason: reason,
	}
}

func JoinGame(state *ServerState, req_ interface{}) interface{} {
	req, ok := req_.(*JoinGameRequest)
	if !ok {
		return NewJoinGameResponseError("cannot interpret the request as a StartGameRequest")
	}
	if req.GameName == "" {
		return NewJoinGameResponseError("missing required field \"game_name\"")
	}
	game := state.lookupGame(req.GameName)
	if game == nil {
		return NewJoinGameResponseError("no game found with that name")
	}
	if req.PlayerName == "" {
		return NewJoinGameResponseError("missing required field \"player_name\"")
	}
	session, err := game.lockingJoinGame(req.PlayerName)
	if err != nil {
		return NewJoinGameResponseError(err.Error())
	}
	state.addSession(session, game)
	log.Printf("Player joined %v -> %v", req.PlayerName, req.GameName)
	return &JoinGameResponse{
		Status:  "ok",
		Session: session,
	}
}

func (g *Game) lockingJoinGame(playerName string) (session SessionToken, err error) {
	g.Lock()
	defer g.Unlock()

	if len(g.players) >= g.NumPlayers {
		return session, fmt.Errorf("the game is full (%v/%v players)", len(g.players), g.NumPlayers)
	}
	for _, p := range g.playerNames {
		if p == playerName {
			return session, fmt.Errorf("player with that name is already in the game")
		}
	}
	session, err = RandomSessionToken()
	if err != nil {
		return session, fmt.Errorf("error generating session token")
	}
	g.players = append(g.players, session)
	g.playerNames[session] = playerName

	// Deal cards to player
	c := g.cardsInHand()
	hand := g.deck[:c]
	g.deck = g.deck[c:]
	g.hands[session] = hand

	return session, nil
}
