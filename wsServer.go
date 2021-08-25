package main

import (
	"flag"
	"log"
	"net/http"
    "time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "192.168.1.107:8080", "http service address")

var upgrader = websocket.Upgrader{} // use default options

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
        for true {
            time.Sleep(2 * time.Second)
		    log.Printf("recv: %s", message)
		    err = c.WriteMessage(mt, message)
		    if err != nil {
		    	log.Println("write:", err)
		    	break
		    }
        }  
	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/ws", echo)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
