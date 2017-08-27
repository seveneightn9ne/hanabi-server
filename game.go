package main

import (
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
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
	ToPlayer string `json:"to_player"`
	Color    Color  `json:"color"`
	Number   int    `json:"number"`
	CardIds  []int  `json:"card_ids"`
	// for Play/Discard:
	CardId int `json:"card_id"`
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
type Turn struct {
	Id     int    `json:"id"`
	Player string `json:"player"`
	Move   Move   `json:"move"`
	// no NewCard for Hint move
	NewCard Card `json:"new_card"`
}

type GameStateSummary struct {
	State      GameState         `json:"state"`
	Players    []string          `json:"players"`
	Hand       []HiddenCard      `json:"hand"`        // the focused player's hand
	OtherHands map[string][]Card `json:"other_hands"` // the other player's hands
	Board      map[Color]int     `json:"board"`
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
	players   []string
	sessions  map[string]SessionToken // player -> session token
	turns     []Turn
	deck      Deck
	hands     map[string][]Card
	board     map[Color]int
	discard   []Card
	whoseTurn int // Index into players. Use -1 when game is over
}

// func (g *Game) move(move Move, player string) (GetStateResponse, error) {
// 	lastTurn := len(g.turns)
// 	if player != g.players[g.whoseTurn] {
// 		return MoveResponse{NotYourTurn, g.InfoResponse(player, lastTurn)}
// 	}
// 	//type Move struct {
// 	//Type MoveType `json:"type"`
// 	//// for Hint:
// 	//ToPlayer string `json:"to_player"`
// 	//Color    Color  `json:"color"`
// 	//Number   int    `json:"number"`
// 	//CardIds  []int  `json:"card_ids"`
// 	//for Play/Discard:
// 	//CardId int `json:"card_id"`

// 	//type Turn struct {
// 	//Id     int    `json:"id"`
// 	//Player string `json:"player"`
// 	//Move   Move   `json:"move"`
// 	// no NewCard for Hint move
// 	//NewCard Card `json:"new_card"`
// 	var turn Turn
// 	turn.Id = g.turns[len(g.turns)-1].Id + 1
// 	turn.Player = player
// 	switch move.Type {
// 	case Hint:
// 		// TODO Make sure the hint is legal
// 		//for _, card := range g.hands[move.ToPlayer] {
// 		//}
// 	case Play:
// 		// TODO Bomb or edit board
// 	case Discard:
// 		// TODO Add to Discard
// 	}
// 	return MoveResponse{Ok, g.InfoResponse(player, lastTurn)}
// }

func (g *Game) LockingAddPlayer(playerName string) (session SessionToken, err error) {
	g.Lock()
	defer g.Unlock()

	if len(g.players) >= g.NumPlayers {
		return session, fmt.Errorf("the game is full (%v/%v players)", len(g.players), g.NumPlayers)
	}
	for _, p := range g.players {
		if p == playerName {
			return session, fmt.Errorf("player with that name is already in the game")
		}
	}
	session, err = RandomSessionToken()
	if err != nil {
		return session, fmt.Errorf("error generating session token")
	}
	g.players = append(g.players, playerName)
	g.sessions[playerName] = session
	return session, nil
}

func (g *Game) LockingGetState(playerName string, session SessionToken, wait bool) (res GameStateSummary, err error) {
	err = g.checkPlayerSession(playerName, session)
	if err != nil {
		return res, err
	}
	for {
		res, err = g.lockingGetStateSummary(playerName, 0)
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

// uses the game lock
func (g *Game) checkPlayerSession(playerName string, session SessionToken) error {
	g.Lock()
	defer g.Unlock()
	if !g.hasPlayer(playerName) {
		return fmt.Errorf("no player '%v'", playerName)
	}
	s2, ok := g.sessions[playerName]
	if !ok {
		// this should never happen
		return fmt.Errorf("missing session for '%v'", playerName)
	}
	if session == "" {
		return fmt.Errorf("session is required")
	}
	if subtle.ConstantTimeCompare([]byte(s2), []byte(session)) != 1 {
		return fmt.Errorf("invalid session for '%v'", playerName)
	}
	return nil
}

func (g *Game) hasPlayer(playerName string) bool {
	for _, p2 := range g.players {
		if playerName == p2 {
			return true
		}
	}
	return false
}

func (g *Game) lockingGetStateSummary(player string, turnCursor int) (GameStateSummary, error) {
	g.Lock()
	defer g.Unlock()
	return g.getStateSummary(player, turnCursor)
}

func (g *Game) getStateSummary(player string, turnCursor int) (GameStateSummary, error) {
	var resp GameStateSummary
	// fill these no matter what
	resp.Players = g.players
	resp.Board = g.board
	resp.Discard = g.discard
	resp.Hand = g.hiddenPlayerHand(player)
	resp.OtherHands = g.otherHands(player)

	if len(g.players) < g.NumPlayers {
		// Game has not started yet
		resp.State = NotStarted
		return resp, nil
	} else if g.whoseTurn == -1 {
		resp.State = Finished
	} else if g.players[g.whoseTurn] == player {
		resp.State = YourTurn
	} else {
		resp.State = WaitingForTurn
	}
	resp.Turns = g.turns[turnCursor:]
	resp.TurnCursor = len(g.turns)
	return resp, nil
}

// The hand of a player, as hidden cards
func (g *Game) hiddenPlayerHand(player string) (res []HiddenCard) {
	hand := g.hands[player]
	for _, card := range hand {
		res = append(res, card.Hide())
	}
	return res
}

// The hands of the players _except_ the specified player.
func (g *Game) otherHands(exceptPlayer string) map[string][]Card {
	res := make(map[string][]Card)
	for p2, hand := range g.hands {
		if p2 == exceptPlayer {
			continue
		}
		res[p2] = hand
	}
	return res
}

// // Take hands and hide the hand of the player.
// func hideCards(hands hands, player string) map[string][]Card {
// 	ret := make(map[string][]Card)
// 	for p, hand := range hands {
// 		if p == player {
// 			// Hide the cards for the player.
// 			ret[p] = nil
// 			for _, card := range hand {
// 				ret[p] = append(ret[p], card.Hide())
// 			}
// 		} else {
// 			// Pass through the other hands unmodified.
// 			ret[p] = hand
// 		}
// 	}
// }
