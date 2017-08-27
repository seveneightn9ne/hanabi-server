package main

import "sync"

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

func (f *Card) Id() int {
	return f.Id
}

type HiddenCard struct {
	Id int `json:"id"`
}

func (h *HiddenCard) Id() int {
	return h.Id
}

type Cardy interface {
	Id() int
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
	Hands      map[string][]Card `json:"hands"`
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

type Game struct {
	// Immutable Fields
	Name        string
	NumPlayers  int
	AddPlayer   chan string
	DoTurn      chan Turn
	RequestInfo chan InfoRequest

	// Mutable, private fields
	players   []string
	turns     []Turn
	deck      Deck
	hands     map[string][]Card
	board     map[Color]int
	discard   []Card
	whoseTurn int // Index into players. Use -1 when game is over
}

var Games map[string]Game
var GamesLock sync.Mutex

func (g *Game) DoGame() {
	for len(g.players) < g.NumPlayers {
		select {
		case p := <-g.AddPlayer:
			g.players = append(g.Players, p)
		case r := <-g.RequestInfo:
			r.Resp <- g.InfoResponse(r.Player, r.TurnCursor)
		}
	}
	// TODO: now play the game
}

func (g *Game) InfoResponse(player string, turnCursor int) InfoResponse {
	var resp InfoResponse
	// fill these no matter what
	resp.Players = g.players
	resp.Board = g.board
	resp.Discard = g.discard
	resp.Hands = sanitize(g.hands, player)

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
