package main

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
		if len(g.Players) >= g.NumPlayers {
			return JoinGameResponse{"error", "this game is full"}
		}
		for _, p := range g.Players {
			if p == req.PlayerName {
				return JoinGameResponse{"error", "player with that name is already in the game"}
			}
		}
		g.Players = append(g.Players, req.PlayerName)
		return JoinGameResponse{"ok", ""}
	} else {
		return JoinGameResponse{"error", "no game found with that name"}
	}
}
