package main

import "testing"

func serverGamePlayer() (ServerState, *Game, SessionToken) {
	s := NewServer().state
	StartGame(&s, &StartGameRequest{2, "test_game"})
	r := JoinGame(&s, &JoinGameRequest{"test_game", "player1"}).(*JoinGameResponse)
	return s, s.Games["test_game"], r.Session
}

func TestGetState_NotStarted(t *testing.T) {
	serverState, _, session := serverGamePlayer()
	request := GetStateRequest{session, false}
	response := GetState(&serverState, &request).(*GetStateResponse)
	if response.Status == "error" {
		t.Errorf("Expected status ok but was error: %v", response.Reason)
	}
	if s := response.State.State; s != "not-started" {
		t.Errorf("Expected game is 'not-started' but is %v", s)
	}
	if l := len(response.State.Players); l != 1 {
		t.Errorf("There should be 1 player but there's %v", l)
	}
	if p := response.State.Players[0]; p != "player1" {
		t.Errorf("expected the player is me, player1, but it's %v", p)
	}
	if l := len(response.State.Board); l != 5 {
		t.Errorf("expected the board length is 5 but is %v", l)
	}
	for k, v := range response.State.Board {
		if len(v) != 0 {
			t.Errorf("Expected all piles to be empty but pile %v was %v", k, v)
		}
	}
	if l := len(response.State.Hand); l != 5 {
		t.Errorf("Expect I have 5 cards but I have %v", l)
	}
	if l := len(response.State.Discard); l != 0 {
		t.Errorf("Expect discard pile is empty")
	}
	if l := len(response.State.Turns); l != 0 {
		t.Errorf("Expect there are no turns")
	}
}

func TestGetState_YourTurn(t *testing.T) {
	serverState, _, session := serverGamePlayer()
	JoinGame(&serverState, &JoinGameRequest{"test_game", "player2"})
	request := GetStateRequest{session, false}
	response := GetState(&serverState, &request).(*GetStateResponse)
	if response.Status == "error" {
		t.Errorf("Expected status ok but was error: %v", response.Reason)
	}
	if s := response.State.State; s != "your-turn" {
		t.Errorf("Expected game state is 'your-turn' but is %v", s)
	}
	if l := len(response.State.Players); l != 2 {
		t.Errorf("There should be 2 players but there's %v", l)
	}
	if p := response.State.Players[0]; p != "player1" {
		t.Errorf("expected the first player is me, player1, but it's %v", p)
	}
	if l := len(response.State.Hand); l != 5 {
		t.Errorf("Expect I have 5 cards but I have %v", l)
	}
	if h := response.State.OtherHands["player2"]; len(h) != 5 {
		t.Errorf("Expected the opponent has 5 cards but has %v", len(h))
	} else {
		for _, c := range h {
			if c.Color == "" {
				t.Errorf("Expected opponent's cards have a color")
			}
			if c.Number == 0 {
				t.Errorf("Expected opponent's cards have a number")
			}
		}
	}
	if h := response.State.OtherHands["player1"]; h != nil {
		t.Errorf("I shouldn't see my own hand but I see %v", h)
	}
	if l := len(response.State.Discard); l != 0 {
		t.Errorf("Expect discard pile is empty")
	}
	if l := len(response.State.Turns); l != 0 {
		t.Errorf("Expect there are no turns")
	}
}

func TestGetState_WaitingForTurn(t *testing.T) {
	serverState, _, session := serverGamePlayer()
	r := JoinGame(&serverState, &JoinGameRequest{"test_game", "player2"})
	session = r.(*JoinGameResponse).Session
	request := GetStateRequest{session, false}
	response := GetState(&serverState, &request).(*GetStateResponse)
	if response.Status == "error" {
		t.Errorf("Expected status ok but was error: %v", response.Reason)
	}
	if s := response.State.State; s != "waiting-for-turn" {
		t.Errorf("Expected game state is 'waiting-for-turn' but is %v", s)
	}
	if l := len(response.State.Players); l != 2 {
		t.Errorf("There should be 2 players but there's %v", l)
	}
	if p := response.State.Players[0]; p != "player1" {
		t.Errorf("expected the first player is player1, but it's %v", p)
	}
	if l := len(response.State.Hand); l != 5 {
		t.Errorf("Expect I have 5 cards but I have %v", l)
	}
	if h := response.State.OtherHands["player1"]; len(h) != 5 {
		t.Errorf("Expected the opponent has 5 cards but has %v", len(h))
	} else {
		for _, c := range h {
			if c.Color == "" {
				t.Errorf("Expected opponent's cards have a color")
			}
			if c.Number == 0 {
				t.Errorf("Expected opponent's cards have a number")
			}
		}
	}
	if h := response.State.OtherHands["player2"]; h != nil {
		t.Errorf("I shouldn't see my own hand but I see %v", h)
	}
	if l := len(response.State.Discard); l != 0 {
		t.Errorf("Expect discard pile is empty")
	}
	if l := len(response.State.Turns); l != 0 {
		t.Errorf("Expect there are no turns")
	}
}
