package main

type MoveRequest struct {
	GameName   string       `json:"game_name"`
	PlayerName string       `json:"player_name"`
	Session    SessionToken `json:"session"`
	Move       Move         `json:"move"`
}

type MoveResponse struct {
	Status string            `json:"status"`
	Reason string            `json:"reason"`
	State  *GameStateSummary `json:"state,omitempty"`
}

func NewMoveResponseError(reason string) *MoveResponse {
	return &MoveResponse{
		Status: "error",
		Reason: reason,
	}
}

func DoMove(state *ServerState, req_ interface{}) interface{} {
	req, ok := req_.(*MoveRequest)
	if !ok {
		return NewMoveResponseError("cannot interpret the request as a MoveRequest")
	}
	game := state.lookupGame(req.GameName)
	if game == nil {
		return NewMoveResponseError("no game found with that name")
	}

	gameState, err := game.LockingMove(req.PlayerName, req.Session, req.Move)
	if err != nil {
		return NewMoveResponseError(err.Error())
	}
	return &MoveResponse{
		Status: "ok",
		State:  gameState,
	}
}
