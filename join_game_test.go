package main

import "testing"

func serverStateWithGame() (ServerState, Game) {
	s := NewServer().state
	StartGame(&s, &StartGameRequest{2, "test_game"})
	return s, *s.Games["test_game"]
}

func TestJoinGame_Basic(t *testing.T) {
	serverState, game := serverStateWithGame()
	request := JoinGameRequest{"test_game", "player1"}
	response := JoinGame(&serverState, &request).(*JoinGameResponse)
	if response.Status == "error" {
		t.Errorf("Expected status ok but was error: %v", response.Reason)
	}
	if response.Session == "" {
		t.Error("Expected a session token")
	}
	if len(game.players) != 1 {
		t.Errorf("The game should have 1 player but has %v", len(game.players))
	}
	request.PlayerName = "player2"
	response = JoinGame(&serverState, &request).(*JoinGameResponse)
	if response.Status == "error" {
		t.Errorf("Expected status ok but was error: %v", response.Reason)
	}
	if len(game.players) != 2 {
		t.Errorf("The game should have 2 players but has %v", len(game.players))
	}
	request.PlayerName = "player3"
	if response.Status == "ok" {
		t.Errorf("expected to error when adding an extra player")
	}

}

func TestJoinGame_WrongTypeRequest(t *testing.T) {
	serverState, game := serverStateWithGame()
	request := StartGameRequest{2, "test_game"}
	response := JoinGame(&serverState, &request).(*JoinGameResponse)
	if response.Status != "error" {
		t.Errorf("Expected State=error for wrong request type but was %v", response.Status)
	}
	if len(game.players) > 0 {
		t.Errorf("Expected that there are still no players in the game")
	}
}

func TestJoinGame_BadParams(t *testing.T) {
	serverState, game := serverStateWithGame()
	request := JoinGameRequest{"test_game", ""}
	response := JoinGame(&serverState, &request).(*JoinGameResponse)
	if response.Status != "error" {
		t.Errorf("Expected State=error when player has no name")
	}
	if len(game.players) > 0 {
		t.Errorf("Expected that there are still no players in the game")
	}

	request = JoinGameRequest{"not_test_game", "player"}
	response = JoinGame(&serverState, &request).(*JoinGameResponse)
	if response.Status != "error" {
		t.Errorf("Expected State=error when the game doesn't exist")
	}
	if len(game.players) > 0 {
		t.Errorf("Expected that there are still no players in the game")
	}
	request.GameName = "test_game"
	_ = JoinGame(&serverState, &request)
	response = JoinGame(&serverState, &request).(*JoinGameResponse)
	if response.Status != "error" {
		t.Errorf("Expected error for joining the same player twice")
	}
	if len(game.players) != 1 {
		t.Errorf("expected the game has 1 player but has %v", len(game.players))
	}
}
