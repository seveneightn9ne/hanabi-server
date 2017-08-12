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

func StartGame(req StartGameRequest) (StartGameResponse, error) {
	GamesLock.Lock()
	defer GamesLock.Unlock()
	if _, ok := Games[req.Name]; ok {
		return StartGameResponse{"error", "game with the same name exists"}, nil
	}
	if req.NumPlayers < 2 || req.NumPlayers > 5 {
		return StartGameResponse{"error", "must specify 2-5 players"}, nil
	}
	newGame := Game{
		Name:       req.Name,
		Players:    make([]string, req.NumPlayers),
		NumPlayers: req.NumPlayers,
		Turns:      make([]Turn, 0),
		Deck:       newDeck(),
		Hands:      make(map[string][]Card, req.NumPlayers),
		Board: map[Color]int{
			White:  0,
			Blue:   0,
			Red:    0,
			Green:  0,
			Yellow: 0},
		Discard: make([]Card, 0),
	}
	Games[req.Name] = newGame
	return StartGameResponse{"ok", ""}, nil
}

func newDeck() Deck {
	numCards := 5 * (3 + 2 + 2 + 2 + 1)
	cards := make([]Card, numCards)
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
				p_i++
			}
		}
	}
	return Deck(cards)
}
