package main

import (
	"log"
	"net/http"
)

func main() {
	dispatcher := NewEvents()
	go dispatcher.receiveEvents()
	go dispatcher.distributeEvents()

	controller := Controller{dispatcher}
	http.HandleFunc("/", controller.handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
