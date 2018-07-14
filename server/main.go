package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"agones.dev/agones/sdks/go"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		indexFile, err := os.Open("/home/server/site/index.html")
		if err != nil {
			fmt.Println(err)
		}
		index, err := ioutil.ReadAll(indexFile)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Fprintf(w, string(index))
	})

	http.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		counter := 1
		for {
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				fmt.Println(err)
				return
			}
			if string(msg) == "PING" {
				fmt.Println(fmt.Sprintf("ping %d", counter))
				err = conn.WriteMessage(msgType, []byte(fmt.Sprintf("pong %d", counter)))
				if err != nil {
					fmt.Println(err)
					return
				}
				counter = counter + 1
			}
			if string(msg) == "STOP" {
				err := s.Shutdown()
				if err != nil {
					log.Printf("Error shutting down: %v", err)
				}
				conn.Close()
				close(stop)
				os.Exit(0)
				fmt.Println(string(msg))
				return
			}
		}
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
