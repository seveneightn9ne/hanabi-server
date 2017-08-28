package main

import "testing"

func TestBasic(t *testing.T) {
	serverState := NewServer().state
	request := StartGameRequest{2, "test_game"}
	response := StartGame(&serverState, &request).(*StartGameResponse)
	if response.Status == "error" {
		t.Errorf("Expected status ok but was error: %v", response.Reason)
	}
	game, ok := serverState.Games["test_game"]
	if !ok {
		t.Errorf("The game was not added to the server's state")
	}
	if game.Name != "test_game" {
		t.Errorf("The game's name should be \"test_game\" but is %v", game.Name)
	}
	if game.NumPlayers != 2 {
		t.Errorf("The game should have 2 players but has %v", game.NumPlayers)
	}
	correctNumCards := 5 * (3 + 2 + 2 + 2 + 1)
	if len(game.deck) != correctNumCards {
		t.Errorf("The deck should have %v cards but has %v", correctNumCards, len(game.deck))
	}
	if len(game.cardsById) != correctNumCards {
		t.Errorf("cardsById should have %v cards but has %v", correctNumCards, len(game.cardsById))
	}
	if game.whoseTurn != 0 {
		t.Errorf("The turn should start at 0 but is %v", game.whoseTurn)
	}
}

func TestWrongTypeRequest(t *testing.T) {
	serverState := NewServer().state
	request := JoinGameRequest{"test_game", "test_player"}
	response := StartGame(&serverState, &request).(*StartGameResponse)
	if response.Status != "error" {
		t.Errorf("Expected State=error for wrong request type but was %v", response.Status)
	}
	if len(serverState.Games) > 0 {
		t.Errorf("Expected that there are still no games in the server state")
	}
}

func TestBadParams(t *testing.T) {
	serverState := NewServer().state
	request := StartGameRequest{0, "test_game"}
	response := StartGame(&serverState, &request).(*StartGameResponse)
	if response.Status != "error" {
		t.Errorf("Expected State=error when NumPlayers is 0")
	}
	if len(serverState.Games) > 0 {
		t.Errorf("Expected that there are still no games in the server state")
	}

	request = StartGameRequest{1, "test_game"}
	response = StartGame(&serverState, &request).(*StartGameResponse)
	if response.Status != "error" {
		t.Errorf("Expected State=error when NumPlayers is 1")
	}
	if len(serverState.Games) > 0 {
		t.Errorf("Expected that there are still no games in the server state")
	}

	request = StartGameRequest{6, "test_game"}
	response = StartGame(&serverState, &request).(*StartGameResponse)
	if response.Status != "error" {
		t.Errorf("Expected State=error when NumPlayers is 6")
	}
	if len(serverState.Games) > 0 {
		t.Errorf("Expected that there are still no games in the server state")
	}

	request = StartGameRequest{5, ""}
	response = StartGame(&serverState, &request).(*StartGameResponse)
	if response.Status != "error" {
		t.Errorf("Expected State=error when Name is empty")
	}
	if len(serverState.Games) > 0 {
		t.Errorf("Expected that there are still no games in the server state")
	}

	// Valid game
	request = StartGameRequest{5, "test_game"}
	response = StartGame(&serverState, &request).(*StartGameResponse)
	if response.Status == "error" {
		t.Errorf("Expected %v to succeed", request)
	}
	if len(serverState.Games) != 1 {
		t.Errorf("Expected that there is now one game in the server state")
	}

	request = StartGameRequest{3, "test_game"}
	response = StartGame(&serverState, &request).(*StartGameResponse)
	if response.Status != "error" {
		t.Errorf("Expected State=error when adding a duplicate game")
	}
	if len(serverState.Games) > 1 {
		t.Errorf("Expected that there is still only one game in the server state")
	}
}
