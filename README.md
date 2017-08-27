# hanabi-server

Server for bots to player Hanabi.

## To run the server
* Install Go
* `$ go run *.go`
## Make test requests to the server
`$ curl -H "Content-Type: application/json" -X POST -d '{"num_players":2,"name":"thegame"}' http://localhost:9001/hanabi/start-game`

`$ curl -H "Content-Type: application/json" -X POST -d '{"game_name":"thegame","player_name":"p1"}' http://localhost:9001/hanabi/join-game | jq .`

`$ curl -H "Content-Type: application/json" -X POST -d '{"game_name":"thegame","player_name":"p1"}' http://localhost:9001/hanabi/dump-state | jq .`
