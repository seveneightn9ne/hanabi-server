package main

import (
	"time"
)

type GetStateRequest struct {
	Session SessionToken `json:"session"`
	Wait    bool         `json:"wait"`
}

type GetStateResponse struct {
	Status string           `json:"status"`
	Reason string           `json:"reason,omitempty"`
	State  GameStateSummary `json:"state,omitempty"`
}

func NewGetStateResponseError(reason string) *GetStateResponse {
	return &GetStateResponse{
		Status: "error",
		Reason: reason,
	}
}

func GetState(state *ServerState, req_ interface{}) interface{} {
	req, ok := req_.(*GetStateRequest)
	if !ok {
		return NewGetStateResponseError("cannot interpret the request as a StartGameRequest")
	}
	game := state.gameForSession(req.Session)
	if game == nil {
		return NewGetStateResponseError("Session token not found")
	}

	gameState, err := game.lockingGetState(req.Session, req.Wait)
	if err != nil {
		return NewGetStateResponseError(err.Error())
	}
	return &GetStateResponse{
		Status: "ok",
		State:  gameState,
	}
}

func (g *Game) lockingGetState(session SessionToken, wait bool) (res GameStateSummary, err error) {
	for {
		res, err = g.lockingGetStateSummary(session, 0)
		if err != nil {
			return res, err
		}
		if !wait || res.State == YourTurn {
			return res, err
		}
		// TODO use a channel to make this faster
		<-time.NewTimer(time.Millisecond * 500).C
	}
}

func (g *Game) lockingGetStateSummary(session SessionToken, turnCursor int) (GameStateSummary, error) {
	g.Lock()
	defer g.Unlock()
	return g.getStateSummary(session, turnCursor)
}

func (g *Game) getStateSummary(session SessionToken, turnCursor int) (GameStateSummary, error) {
	var resp GameStateSummary
	// fill these no matter what
	resp.Players = nil
	for _, s := range g.players {
		resp.Players = append(resp.Players, g.playerNames[s])
	}
	resp.Board = g.board
	resp.Discard = g.discard
	resp.Hand = g.hiddenPlayerHand(session)
	resp.OtherHands = g.otherHands(session)

	if len(g.players) < g.NumPlayers {
		// Game has not started yet
		resp.State = NotStarted
		return resp, nil
	} else if g.whoseTurn == -1 {
		resp.State = Finished
	} else if g.players[g.whoseTurn] == session {
		resp.State = YourTurn
	} else {
		resp.State = WaitingForTurn
	}
	resp.Turns = g.turns[turnCursor:]
	resp.TurnCursor = len(g.turns)
	return resp, nil
}

// The hands of the players _except_ the specified player.
func (g *Game) otherHands(exceptPlayer SessionToken) map[string][]Card {
	res := make(map[string][]Card)
	for p2, hand := range g.hands {
		if p2 == exceptPlayer {
			continue
		}
		res[g.playerNames[p2]] = hand
	}
	return res
}

// The hand of a player, as hidden cards
func (g *Game) hiddenPlayerHand(player SessionToken) (res []HiddenCard) {
	hand := g.hands[player]
	for _, card := range hand {
		res = append(res, card.Hide())
	}
	return res
}
