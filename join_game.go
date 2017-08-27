package main

type JoinGameRequest struct {
	GameName   string `json:"game_name"`
	PlayerName string `json:"player_name"`
}

type JoinGameResponse struct {
	Status string       `json:"status"`
	Reason string       `json:"reason"`
	Token  SessionToken `json:"token"`
}

func NewJoinGameResponseError(reason string) JoinGameResponse {
	return JoinGameResponse{
		Status: "error",
		Reason: reason,
	}
}

func JoinGame(state *ServerState, req JoinGameRequest) JoinGameResponse {
	if _, ok := state.Games[req.GameName]; !ok {
		return NewJoinGameResponseError("no game found with that name")
	}

	resCh := make(chan AddPlayerCmdRes)
	state.Games[req.GameName].AddPlayer <- AddPlayerCmd{
		playerName: req.PlayerName,
		resCh:      resCh,
	}
	cmdRes := <-resCh
	if cmdRes.err != nil {
		NewJoinGameResponseError(cmdRes.err.Error())
	}
	return JoinGameResponse{
		Status: "ok",
		Token:  cmdRes.session,
	}
}
