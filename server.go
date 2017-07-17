package hanabi

import (
	"encoding/json"
	//"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
)


func okResponse() []byte {
	return []byte("{status: \"ok\"}")
}
func writeJson(w http.ResponseWriter, obj interface{}) {
	respStr, err := json.Marshal(obj)
	if err != nil {
		log.Printf("Error during JSON marshal: %v\n", err)
		return
	}
	w.Write([]byte(respStr))
}

// true if there was an error that we handled
func handleErr(err error, w http.ResponseWriter) bool {
	if err != nil {
		log.Printf("Error: %v", err.Error())
		w.WriteHeader(500)
		writeJson(w, struct {
			status  string
			message string
		}{
			status:  "error",
			message: err.Error(),
		})
		return true
	}
	return false
}

const pfx = "/hanabi/"

func Handle(w http.ResponseWriter, req *http.Request) {
	dec := json.NewDecoder(req.Body)
	var response interface{}
	var err error
	if strings.HasPrefix(req.URL.Path, pfx+"start-game") {
		var v StartGameRequest
		if err = dec.Decode(&v); handleErr(err, w) {
			return
		}
		response, err = StartGame(v)
	/*} else if strings.HasPrefix(pfx+"join-game") {
		var v JoinGameRequest
		if err = dec.Decode(&v); handleErr(err) {
			return
		}
		resp, err = JoinGame(JoinGameRequest)*/
		// TODO(jessk) handle more endpoints here
	} else {
		w.WriteHeader(404)
		return
	}
	writeJson(w, response)
}

func main() {
	Games = make([]Game, 0)
	port := flag.Int("port", 9001, "port to listen on")
	flag.Parse()
	serveStr := fmt.Sprintf(":%v", *port)
	log.Printf("Serving at localhost%v", serveStr)
	http.HandleFunc(pfx, Handle)
	log.Fatal(http.ListenAndServe(serveStr, nil))
}
