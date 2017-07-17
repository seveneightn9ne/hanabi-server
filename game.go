package hanabi

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
var Colors = [...]Color{Red, Yellow, Green, Blue, Black, White}
var Numbers = [...]int{1, 2, 3, 4, 5}

type Move struct {
	Type MoveType `json:"type"`
	// for Hint:
	ToPlayer string `json:"to_player"`
	Color    Color  `json:"color"`
	Number   int    `json:"number"`
	// for Play/Discard:
	CardId int `json:"card_id"`
}
type Card struct {
	Id     int   `json:"id"`
	Color  Color `json:"color"`
	Number int   `json:"number"`
}
type Deck []Card
type Turn struct {
	Id     int    `json:"id"`
	Player string `json:"player"`
	Move   Move   `json:"move"`
	// no NewCard for Hint move
	NewCard Card  `json:"new_card"`
}

type Game struct {
	Name    string
	Players map[int]string
	Turns   []Turn
	Deck    Deck
	Hands   map[string][]Card
}

var Games []Game
var GamesLock sync.Mutex
