package main

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
	session, err := game.LockingAddPlayer(req.PlayerName)
	if err != nil {
		return NewJoinGameResponseError(err.Error())
	}
	state.addSession(session, game)
	return &JoinGameResponse{
		Status:  "ok",
		Session: session,
	}
}
