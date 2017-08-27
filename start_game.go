package main

import "math/rand"

type StartGameRequest struct {
	NumPlayers int    `json:"num_players"`
	Name       string `json:"name"`
}

type StartGameResponse struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

func StartGame(state *ServerState, req StartGameRequest) (StartGameResponse, error) {
	if _, ok := state.Games[req.Name]; ok {
		return StartGameResponse{"error", "game with the same name exists"}, nil
	}
	if req.NumPlayers < 2 || req.NumPlayers > 5 {
		return StartGameResponse{"error", "must specify 2-5 players"}, nil
	}
	deck, cardsById := newDeck()
	newGame := Game{
		Name:       req.Name,
		players:    nil,
		sessions:   make(map[string]SessionToken),
		NumPlayers: req.NumPlayers,
		turns:      make([]Turn, 0),
		deck:       deck,
		hands:      make(map[string][]Card, req.NumPlayers),
		board: map[Color]int{
			White:  0,
			Blue:   0,
			Red:    0,
			Green:  0,
			Yellow: 0},
		discard:   make([]Card, 0),
		cardsById: cardsById,
		whoseTurn: 0,
	}
	state.Games[req.Name] = newGame
	return StartGameResponse{"ok", ""}, nil
}

func newDeck() (Deck, map[int]Card) {
	numCards := 5 * (3 + 2 + 2 + 2 + 1)
	cards := make([]Card, numCards)
	cardsById := make(map[int]Card, numCards)
	order := rand.Perm(numCards)
	p_i := 0
	for _, color := range Colors {
		for n_i, number := range Numbers {
			dupe := []int{3, 2, 2, 2, 1}[n_i]
			for d_i := 0; d_i < dupe; d_i++ {
				cards[order[p_i]] = Card{
					Id:     order[p_i],
					Color:  color,
					Number: number,
				}
				cardsById[order[p_i]] = cards[order[p_i]]
				p_i++
			}
		}
	}
	return Deck(cards), cardsById
}
