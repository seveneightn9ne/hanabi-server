package main

type DumpStateResponse struct {
	Games map[string]Game `json:"games"`
}

func DumpState() DumpStateResponse {
	GamesLock.Lock()
	defer GamesLock.Unlock()
	return DumpStateResponse{
		Games: Games,
	}
}
