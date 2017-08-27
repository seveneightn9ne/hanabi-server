package main

import "fmt"

type JoinGameRequest struct {
	GameName   string `json:"game_name"`
	PlayerName string `json:"player_name"`
}

type JoinGameResponse struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

func JoinGame(req JoinGameRequest) JoinGameResponse {
	GamesLock.Lock()
	defer GamesLock.Unlock()
	if g, ok := Games[req.GameName]; ok {
		if len(g.players) >= g.NumPlayers {
			return JoinGameResponse{"error", fmt.Sprintf("this game is full (%v/%v players)", len(g.players), g.NumPlayers)}
		}
		for _, p := range g.players {
			if p == req.PlayerName {
				return JoinGameResponse{"error", "player with that name is already in the game"}
			}
		}
		g.players = append(g.players, req.PlayerName)
		return JoinGameResponse{"ok", ""}
	} else {
		return JoinGameResponse{"error", "no game found with that name"}
	}
}
