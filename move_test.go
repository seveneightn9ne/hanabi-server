package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T, numPlayers int) (server *testServer, players []*testPlayer) {
	server = newTestServer(t)
	server.StartGame()
	for i := 0; i < numPlayers; i++ {
		p := server.newTestPlayer()
		p.JoinGame()
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
		CardIDs:  []int{},
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

type testServer struct {
	T       *testing.T
	Server  *Server
	Players []*testPlayer
}

func newTestServer(t *testing.T) *testServer {
	return &testServer{
		T:      t,
		Server: NewServer(),
	}
}

func (s *testServer) StartGame() {
	req := StartGameRequest{
		NumPlayers: 2,
		Name:       "test-game",
	}
	res := StartGame(&s.Server.state, &req).(*StartGameResponse)
	require.NotNil(s.T, res)
	require.Equal(s.T, "ok", res.Status, "%v", res.Reason)
}

type testPlayer struct {
	T       *testing.T
	Server  *testServer
	Name    string
	Session SessionToken
}

func (s *testServer) newTestPlayer() *testPlayer {
	p := &testPlayer{
		T:      s.T,
		Server: s,
		Name:   fmt.Sprintf("test-player-%v", len(s.Players)),
	}
	s.Players = append(s.Players, p)
	return p
}

func (p *testPlayer) JoinGame() {
	req := JoinGameRequest{
		GameName:   "test-game",
		PlayerName: p.Name,
	}
	res := JoinGame(&p.Server.Server.state, &req).(*JoinGameResponse)
	require.NotNil(p.T, res)
	require.Equal(p.T, "ok", res.Status, "%v", res.Reason)
	p.Session = res.Session
}

func (p *testPlayer) Move(move Move) error {
	req := MoveRequest{
		Session: p.Session,
		Move:    move,
	}
	res := MoveHandler(&p.Server.Server.state, &req).(*MoveResponse)
	require.NotNil(p.T, res)
	if res.Status == "ok" {
		return nil
	}
	return fmt.Errorf("%v", res.Reason)
}
