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
		return fmt.Errorf("not player %v's turn", playerIndex)
	}

	switch move.Type {
	case Play:
		if move.CardId == nil {
			return fmt.Errorf("missing required field card_id for move type PLAY")
		}
		card := g.getCardFromHand(*move.CardId, session)
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
		if g.hints < 1 {
			return fmt.Errorf("no hint credits available")
		}
		if move.CardId != nil {
			return fmt.Errorf("unexpected CardId in HINT move")
		}
		turn := Turn{
			Id:      len(g.turns),
			Player:  playerName,
			Move:    move,
			NewCard: nil,
		}

		// Check that the hint is valid
		if turn.Move.ToPlayer == nil {
			return fmt.Errorf("hint missing required field 'to_player'")
		}
		err = g.checkHint(*turn.Move.ToPlayer, turn.Move.Color, turn.Move.Number, turn.Move.CardIds)
		if err != nil {
			return err
		}

		// commit
		g.hints--
		g.turns = append(g.turns, turn)
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
	return name, index, fmt.Errorf("session not in game")
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
			cardIDsRef = append(cardIDsRef, card.Id)
		}
		if number != nil && card.Number == *number {
			cardIDsRef = append(cardIDsRef, card.Id)
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

func (g *Game) checkCardColor(color Color) error {
	switch color {
	case Red, Yellow, Green, Blue, Black, White:
	default:
		return fmt.Errorf("invalid color: %v", color)
	}
	return nil
}

func (g *Game) checkCardNumber(number int) error {
	switch number {
	case 1, 2, 3, 4, 5:
	default:
		return fmt.Errorf("invalid number: %v", number)
	}
	return nil
}
