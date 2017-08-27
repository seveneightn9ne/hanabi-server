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

type DoTurnRequest struct {
	Turn Turn
	Resp chan DoTurnResponse
}

type DoTurnResponse struct {
	Response DoTurnResponseType
	Info     InfoResponse
	// Maybe we want more info, like if a bomb went off
}
type DoTurnResponseType int

const (
	Ok DoTurnResponseType = iota
	NotYourTurn
)

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
			g.players = append(g.players, p)
		case r := <-g.RequestInfo:
			r.Resp <- g.InfoResponse(r.Player, r.TurnCursor)
		}
	}
	// TODO: now play the game
	for g.whoseTurn != -1 {
		select {
		case r := <-g.RequestInfo:
			r.Resp <- g.InfoResponse(r.Player, r.TurnCursor)
		case t := <-g.DoTurn:
			g.doTurn(t)
		}
	}
}

func (g *Game) doTurn(turn Turn) {
	return
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
