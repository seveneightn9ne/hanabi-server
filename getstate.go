package main

type GetStateRequest struct {
	GameName   string       `json:"game_name"`
	PlayerName string       `json:"player_name"`
	Session    SessionToken `json:"session"`
	Wait       bool         `json:"wait"`
}

type GetStateResponse struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

func NewGetStateResponseError(reason string) GetStateResponse {
	return GetStateResponse{
		Status: "error",
		Reason: reason,
	}
}

func GetState(state *ServerState, req GetStateRequest) GetStateResponse {
	game := state.lookupGame(req.GameName)
	if game == nil {
		return NewGetStateResponseError("no game found with that name")
	}

	err := game.LockingGetState(req.PlayerName, req.Session, req.Wait)
	if err != nil {
		return NewGetStateResponseError(err.Error())
	}
	return GetStateResponse{
		Status: "ok",
	}
}
