package main

import (
	"encoding/hex"
	"fmt"
)

type MoveType int

const (
	Hint MoveType = iota
	Play
	Discard
)

type Color int

const (
	Red Color = iota
	Yellow
	Green
	Blue
	Black
	White
)

type GameState int

const (
	NotStarted GameState = iota
	WaitingForTurn
	YourTurn
	Finished
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

type InfoResponse struct {
	State      GameState         `json:"state"`
	Players    []string          `json:"players"`
	Hand       []HiddenCard      `json:"hand"`        // the focused player's hand
	OtherHands map[string][]Card `json:"other_hands"` // the other player's hands
	Board      map[Color]int     `json:"board"`
	Discard    []Card            `json:"discard"`
	Turns      []Turn            `json:"turns"`
	TurnCursor int               `json:"turn_cursor"`
}

type InfoRequest struct {
	Player     string
	TurnCursor int
	Resp       chan InfoResponse
}

type MoveRequest struct {
	Move   Move
	Player string
	Resp   chan MoveResponse
}

type MoveResponse struct {
	Response MoveResponseType
	Info     InfoResponse
	// Maybe we want more info, like if a bomb went off
}
type MoveResponseType int

const (
	Ok MoveResponseType = iota
	NotYourTurn
)

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

type AddPlayerCmd struct {
	playerName string
	resCh      chan<- AddPlayerCmdRes
}

type AddPlayerCmdRes struct {
	err     error
	session SessionToken
}

func (g *Game) move(move Move, player string) MoveResponse {
	lastTurn := len(g.turns)
	if player != g.players[g.whoseTurn] {
		return MoveResponse{NotYourTurn, g.InfoResponse(player, lastTurn)}
	}
	//type Move struct {
	//Type MoveType `json:"type"`
	//// for Hint:
	//ToPlayer string `json:"to_player"`
	//Color    Color  `json:"color"`
	//Number   int    `json:"number"`
	//CardIds  []int  `json:"card_ids"`
	//for Play/Discard:
	//CardId int `json:"card_id"`

	//type Turn struct {
	//Id     int    `json:"id"`
	//Player string `json:"player"`
	//Move   Move   `json:"move"`
	// no NewCard for Hint move
	//NewCard Card `json:"new_card"`
	var turn Turn
	turn.Id = g.turns[len(g.turns)-1].Id + 1
	turn.Player = player
	switch move.Type {
	case Hint:
		// TODO Make sure the hint is legal
		//for _, card := range g.hands[move.ToPlayer] {
		//}
	case Play:
		// TODO Bomb or edit board
	case Discard:
		// TODO Add to Discard
	}
	return MoveResponse{Ok, g.InfoResponse(player, lastTurn)}
}

func (g *Game) LockingAddPlayer(playerName string) (session SessionToken, err error) {
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

func (g *Game) InfoResponse(player string, turnCursor int) InfoResponse {
	var resp InfoResponse
	// fill these no matter what
	resp.Players = g.players
	resp.Board = g.board
	resp.Discard = g.discard
	resp.Hand = g.hiddenPlayerHand(player)
	resp.OtherHands = g.otherHands(player)

	if len(g.players) < g.NumPlayers {
		// Game has not started yet
		resp.State = NotStarted
		return resp
	} else if g.whoseTurn == -1 {
		resp.State = Finished
	} else if g.players[g.whoseTurn] == player {
		resp.State = YourTurn
	} else {
		resp.State = WaitingForTurn
	}
	resp.Turns = g.turns[turnCursor:]
	resp.TurnCursor = len(g.turns)
	return resp
}

// The hand of a player, as hidden cards
func (g *Game) hiddenPlayerHand(player string) []HiddenCard {
	panic("todo")
}

// The hands of the players _except_ the specified player.
func (g *Game) otherHands(exceptPlayer string) map[string][]Card {
	panic("todo")
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
