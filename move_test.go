package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T, numPlayers int) (server *testServer, players []*testPlayer) {
	server = newTestServer(t)
	server.StartGame()
	for i := 0; i < numPlayers; i++ {
		p := server.newTestPlayer()
		players = append(players, p)
	}
	return server, players
}

func TestMove_Hint(t *testing.T) {
	_, players := setupTest(t, 2)
	red := Red

	err := players[0].Move(Move{
		Type:     Hint,
		ToPlayer: nil,
		Color:    &red,
		CardIDs:  []int{5, 8},
	})
	require.Error(t, err, "missing player name")

	err = players[0].Move(Move{
		Type:     Hint,
		ToPlayer: &players[1].Name,
		Color:    &red,
		CardIDs:  []int{5},
	})
	require.Error(t, err, "wrong cards hint")

	err = players[0].Move(Move{
		Type:     Hint,
		ToPlayer: &players[1].Name,
		Color:    &red,
		CardIDs:  []int{5, 8},
	})
	require.NoError(t, err)

	err = players[0].Move(Move{
		Type:     Hint,
		ToPlayer: &players[1].Name,
		Color:    &red,
		CardIDs:  []int{5, 8},
	})
	require.Error(t, err, "not your turn")
}

func TestMove_Play(t *testing.T) {
	_, players := setupTest(t, 2)

	one := 1
	err := players[0].Move(Move{
		Type:   Play,
		CardID: &one,
	})
	require.NoError(t, err)

	eight := 8
	err = players[1].Move(Move{
		Type:   Play,
		CardID: &eight,
	})
	require.NoError(t, err)
}

func TestMove_Discard(t *testing.T) {
	_, players := setupTest(t, 2)

	one := 1
	err := players[0].Move(Move{
		Type:   Discard,
		CardID: &one,
	})
	require.NoError(t, err)

	eight := 8
	err = players[1].Move(Move{
		Type:   Discard,
		CardID: &eight,
	})
	require.NoError(t, err)
}
