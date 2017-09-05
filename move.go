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
	g.Lock()
	defer g.Unlock()

	playerIndex, err := g.playerIndex(session)
	if err != nil {
		return err
	}
	if g.whoseTurn != playerIndex {
		return fmt.Errorf("not player %v's turn", playerIndex)
	}

	switch move.Type {
	case Play:
		card := g.getCardFromHand(move.CardId, session)
		if card == nil {
			return fmt.Errorf("Card #%v is not in your hand", move.CardId)
		}
		pile := g.board[card.Color]
		topCard := pile[len(pile)-1]
		if topCard.Number+1 == card.Number {
			// Hooray, well done!
			if card.Number == 5 && g.hints < 8 {
				// Grant a hint
				g.hints += 1
			}
			g.board[card.Color] = append(pile, *card)
		} else {
			// Oof, wrong card
			if g.bombs == 1 {
				// Last chance -- game over
				// TODO(jessk) -- GAME OVER
			}

			g.bombs -= 1

			// Card goes in discard pile
			g.discard = append(g.discard, *card)
		}
		// You get a new card!
		newCard := g.DrawCard()
		g.hands[session] = append(g.hands[session], *newCard)

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
