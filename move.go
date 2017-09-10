package main

import (
	"fmt"
	"sort"
)

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

	playerName, playerIndex, err := g.playerInfo(session)
	if err != nil {
		return err
	}
	if g.whoseTurn != playerIndex {
		return fmt.Errorf("not your turn it's player %v's turn", g.whoseTurn)
	}

	switch move.Type {
	case Play:
		var gameOver bool
		if move.CardID == nil {
			return fmt.Errorf("missing required field card_id for move type PLAY")
		}
		card := g.getCardFromHand(*move.CardID, session)
		if card == nil {
			return fmt.Errorf("Card #%v is not in your hand", *move.CardID)
		}

		pile := g.board[card.Color]
		var topCard int
		if len(pile) == 0 {
			topCard = 0
		} else {
			topCard = pile[len(pile)-1].Number
		}
		if topCard+1 == card.Number {
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
				gameOver = true
			}

			g.bombs -= 1

			// Card goes in discard pile
			g.discard = append(g.discard, *card)
		}
		// You get a new card!
		newCard := g.DrawCard(session)
		g.commitTurn(Turn{
			ID:     len(g.turns),
			Player: playerName,
			Move: Move{
				Type:   Play,
				CardID: move.CardID,
			},
			NewCard: newCard,
		}, gameOver)

		return nil
	case Discard:
		if move.CardID == nil {
			return fmt.Errorf("missing required field card_id for move type DISCARD")
		}
		card := g.getCardFromHand(*move.CardID, session)
		if card == nil {
			return fmt.Errorf("Card #%v is not in your hand", *move.CardID)
		}

		g.discard = append(g.discard, *card)
		// You get a new card!
		newCard := g.DrawCard(session)
		g.commitTurn(Turn{
			ID:     len(g.turns),
			Player: playerName,
			Move: Move{
				Type:   Discard,
				CardID: move.CardID,
			},
			NewCard: newCard,
		}, false /* gameOver */)

		return nil
	case Hint:
		if g.hints < 1 {
			return fmt.Errorf("no hint credits available")
		}
		if move.CardID != nil {
			return fmt.Errorf("unexpected CardID in HINT move")
		}
		turn := Turn{
			ID:     len(g.turns),
			Player: playerName,
			Move: Move{
				Type:     Hint,
				ToPlayer: move.ToPlayer,
				Color:    move.Color,
				Number:   move.Number,
				CardIDs:  move.CardIDs,
			},
			NewCard: nil,
		}

		// Check that the hint is valid
		if turn.Move.ToPlayer == nil {
			return fmt.Errorf("hint missing required field 'to_player'")
		}
		err = g.checkHint(*turn.Move.ToPlayer, turn.Move.Color, turn.Move.Number, turn.Move.CardIDs)
		if err != nil {
			return err
		}

		// commit
		g.hints--
		g.commitTurn(turn, false /* gameOver */)
		return nil
	default:
		return fmt.Errorf("unrecognized move type: %v", move.Type)
	}
}

func (g *Game) playerInfo(session SessionToken) (name string, index int, err error) {
	index = -1
	for i, s := range g.players {
		if s == session {
			index = i
		}
	}
	for s, n := range g.playerNames {
		if s == session {
			name = n
		}
	}
	if index == -1 {
		return name, index, fmt.Errorf("player not found")
	}
	if len(name) == 0 {
		return name, index, fmt.Errorf("fault: empty player name")
	}
	return name, index, nil
}

func (g *Game) lookupPlayerByName(name string) (s SessionToken, err error) {
	for s, n := range g.playerNames {
		if name == n {
			return s, nil
		}
	}
	return s, fmt.Errorf("player not found: %v", name)
}

// Check that the hint and cardIDs line up with the truth.
func (g *Game) checkHint(toPlayer string, color *Color, number *int, cardIDs []int) error {
	if color == nil && number == nil {
		return fmt.Errorf("hint must have color or number but found neither")
	}
	if color != nil && number != nil {
		return fmt.Errorf("hint must have color or number but found both")
	}
	if color != nil {
		err := g.checkCardColor(*color)
		if err != nil {
			return err
		}
	}
	if number != nil {
		err := g.checkCardNumber(*number)
		if err != nil {
			return err
		}
	}

	hintedSession, err := g.lookupPlayerByName(toPlayer)
	if err != nil {
		return err
	}

	var cardIDs2 []int
	for _, cardID := range cardIDs {
		cardIDs2 = append(cardIDs2, cardID)
	}
	sort.Ints(cardIDs2)

	var cardIDsRef []int
	for _, card := range g.hands[hintedSession] {
		if color != nil && card.Color == *color {
			cardIDsRef = append(cardIDsRef, card.ID)
		}
		if number != nil && card.Number == *number {
			cardIDsRef = append(cardIDsRef, card.ID)
		}
	}
	sort.Ints(cardIDsRef)

	if len(cardIDs2) != len(cardIDsRef) {
		return fmt.Errorf("invalid hint: got %v expected %v (length)", cardIDs, cardIDsRef)
	}
	for i, v := range cardIDsRef {
		if cardIDs2[i] != v {
			return fmt.Errorf("invalid hint: got %v expected %v (at index %v)", cardIDs, cardIDsRef, i)
		}
	}
	return nil
}
