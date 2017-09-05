package main

import (
	"encoding/hex"
	"sync"
)

type MoveType string

const (
	Hint    MoveType = "hint"
	Play             = "play"
	Discard          = "discard"
)

type Color string

const (
	Red    Color = "red"
	Yellow Color = "yellow"
	Green  Color = "green"
	Blue   Color = "blue"
	Black  Color = "black"
	White  Color = "white"
)

type GameState string

const (
	NotStarted     GameState = "not-started"
	WaitingForTurn GameState = "waiting-for-turn"
	YourTurn       GameState = "your-turn"
	Finished       GameState = "finished"
)

var Colors = [...]Color{Red, Yellow, Green, Blue, White}
var Numbers = [...]int{1, 2, 3, 4, 5}

type Move struct {
	Type MoveType `json:"type"`
	// for Hint:
	ToPlayer *string `json:"to_player,omitempty"`
	Color    *Color  `json:"color,omitempty"`
	Number   *int    `json:"number,omitempty"`
	CardIds  []int   `json:"card_ids,omitempty"`
	// for Play/Discard:
	CardId *int `json:"card_id,omitempty"`
}

type Card struct {
	Id     int   `json:"id"`
	Color  Color `json:"color"`
	Number int   `json:"number"`
}

// Card implements Cardy
var _ Cardy = (*Card)(nil)

func (f *Card) GetId() int {
	return f.Id
}

func (f *Card) Hide() HiddenCard {
	return HiddenCard{
		Id: f.Id,
	}
}

type HiddenCard struct {
	Id int `json:"id"`
}

// HiddenCard implements Cardy
var _ Cardy = (*HiddenCard)(nil)

func (h *HiddenCard) GetId() int {
	return h.Id
}

type Cardy interface {
	GetId() int
}
type Deck []Card

func (g *Game) DrawCard() *Card {
	l := len(g.deck)
	if l == 0 {
		return nil
	}
	c := g.deck[l-1]
	g.deck = g.deck[l-2:]
	return &c
}

type Turn struct {
	Id     int    `json:"id"`     // turn number starting at 0
	Player string `json:"player"` // player that made the move
	Move   Move   `json:"move"`
	// no NewCard for Hint move
	NewCard Cardy `json:"new_card"`
}

type GameStateSummary struct {
	State      GameState         `json:"state"`
	Players    []string          `json:"players"`
	Hand       []HiddenCard      `json:"hand"`        // the focused player's hand
	OtherHands map[string][]Card `json:"other_hands"` // the other player's hands
	Board      map[Color][]Card  `json:"board"`
	Discard    []Card            `json:"discard"`
	Turns      []Turn            `json:"turns"`
	TurnCursor int               `json:"turn_cursor"`
}

// 64-bit hex
type SessionToken string

func RandomSessionToken() (res SessionToken, err error) {
	bs, err := RandBytes(8)
	if err != nil {
		return res, err
	}
	return SessionToken(hex.EncodeToString(bs)), nil
}

// Calling only methods that start with Locking is threadsafe.
type Game struct {
	sync.Mutex

	// Immutable Fields
	Name       string
	NumPlayers int
	cardsById  map[int]Card

	// Mutable, private fields
	players     []SessionToken
	playerNames map[SessionToken]string
	turns       []Turn
	deck        Deck
	hands       map[SessionToken][]Card
	board       map[Color][]Card // Stack for each color
	bombs       int
	hints       int
	discard     []Card
	whoseTurn   int // Index into players. Use -1 when game is over
}

func (g *Game) cardsInHand() int {
	if g.NumPlayers <= 3 {
		return 5
	}
	return 4
}

// Removes the card from the hand!
// Requires game is locked!
func (g *Game) getCardFromHand(cardId int, player SessionToken) *Card {
	hand := g.hands[player]
	for i, card := range hand {
		if card.Id == cardId {
			g.hands[player] = append(hand[:i], hand[i+1:]...)
			return &card
		}
	}
	return nil
}
