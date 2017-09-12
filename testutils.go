package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

//
// Server.
//
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

//
// Player.
//
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

	// Join Game
	req := JoinGameRequest{
		GameName:   "test-game",
		PlayerName: p.Name,
	}
	res := JoinGame(&p.Server.Server.state, &req).(*JoinGameResponse)
	require.NotNil(p.T, res)
	require.Equal(p.T, "ok", res.Status, "%v", res.Reason)
	p.Session = res.Session

	return p
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
