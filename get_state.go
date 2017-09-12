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

	// Blocks iff req.Wait
	gameState := getStateLoop(game, req.Session, req.Wait)

	return &GetStateResponse{
		Status: "ok",
		State:  gameState,
	}
}

func getStateLoop(g *Game, session SessionToken, wait bool) GameStateSummary {
	for {
		res := g.getState(session, 0)

		if !wait || res.State == YourTurn {
			return res
		}
		// TODO use a channel to make this faster
		<-time.NewTimer(time.Millisecond * 500).C
	}
}

func (g *Game) getState(session SessionToken, turnCursor int) GameStateSummary {
	g.Lock()
	defer g.Unlock()

	var resp GameStateSummary
	// fill these no matter what
	resp.Players = nil
	for _, s := range g.players {
		resp.Players = append(resp.Players, g.playerNames[s])
	}
	resp.Board = g.exportBoard()
	resp.Discard = g.discard
	resp.Hand = g.hiddenPlayerHand(session)
	resp.OtherHands = g.otherHands(session)

	if len(g.players) < g.NumPlayers {
		// Game has not started yet
		resp.State = NotStarted
		if len(resp.Turns) == 0 {
			resp.Turns = []Turn{}
		}
		return resp
	} else if g.whoseTurn == -1 {
		resp.State = Finished
	} else if g.players[g.whoseTurn] == session {
		resp.State = YourTurn
	} else {
		resp.State = WaitingForTurn
	}
	resp.Turns = g.turns[turnCursor:]
	if len(resp.Turns) == 0 {
		resp.Turns = []Turn{}
	}
	resp.TurnCursor = len(g.turns)
	return resp
}

func (g *Game) exportBoard() map[Color][]Card {
	res := make(map[Color][]Card)
	for k, v := range g.board {
		if len(v) == 0 {
			res[k] = []Card{}
		} else {
			res[k] = v
		}
	}
	return res
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
