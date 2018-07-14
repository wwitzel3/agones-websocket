package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"agones.dev/agones/sdks/go"
)

type State struct {
	Counter int
}

func main() {
	port := flag.String("port", "7654", "The port to listen to udp traffic on")
	flag.Parse()
	if ep := os.Getenv("PORT"); ep != "" {
		port = &ep
	}

	s, err := sdk.NewSDK()
	if err != nil {
		log.Fatalf("Could not connect to sdk: %v", err)
	}

	stop := make(chan struct{})
	go doHealth(s, stop)

	state := &State{Counter: 0}

	hub := newHub()
	go hub.run()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL)
		if r.URL.Path != "/" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		http.ServeFile(w, r, "/home/server/site/index.html")
	})

	http.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
		serveWebsocket(state, s, hub, w, r)
	})

	err = s.Ready()
	if err != nil {
		log.Fatalf("Could not send ready message: %v", err)
	}

	if err := http.ListenAndServe(fmt.Sprintf(":%s", *port), nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// doHealth sends the regular Health Pings
func doHealth(sdk *sdk.SDK, stop <-chan struct{}) {
	tick := time.Tick(2 * time.Second)
	for {
		err := sdk.Health()
		if err != nil {
			log.Fatalf("Could not send health ping, %v", err)
		}
		select {
		case <-stop:
			log.Print("Stopped health pings")
			return
		case <-tick:
		}
	}
}
