package main

import "fmt"

type MoveRequest struct {
	Session SessionToken `json:"session"`
	Move    Move         `json:"move"`
}

type MoveResponse struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

func NewMoveResponseError(reason string) *MoveResponse {
	return &MoveResponse{
		Status: "error",
		Reason: reason,
	}
}

func MoveHandler(state *ServerState, req_ interface{}) interface{} {
	req, ok := req_.(*MoveRequest)
	if !ok {
		return NewMoveResponseError("cannot interpret the request as a MoveRequest")
	}
	game := state.gameForSession(req.Session)
	if game == nil {
		return NewMoveResponseError("Session token not found")
	}

	err := game.lockingMove(req.Session, req.Move)
	if err != nil {
		return NewMoveResponseError(err.Error())
	}
	return &MoveResponse{
		Status: "ok",
	}
}

func (g *Game) lockingMove(session SessionToken, move Move) (err error) {
	playerIndex, err := g.playerIndex(session)
	if err != nil {
		return err
	}
	if g.whoseTurn != playerIndex {
		return fmt.Errorf("not player %v's turn", playerIndex)
	}

	switch move.Type {
	case Play:
		return fmt.Errorf("TO BE ACCOMPLISHED")
	case Discard:
		return fmt.Errorf("TO BE ACCOMPLISHED")
	case Hint:
		return fmt.Errorf("TO BE ACCOMPLISHED")
	default:
		return fmt.Errorf("unrecognized move type: %v", move.Type)
	}
}

func (g *Game) playerIndex(session SessionToken) (int, error) {
	for i, s := range g.players {
		if s == session {
			return i, nil
		}
	}
	return 0, fmt.Errorf("session not in game")
}
