package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"strings"
)

type eventData []map[string]interface{}

func main() {
	url := flag.String("url", "https://prod-coco-up-read.ft.com/content/notifications-push", "notifications push endpoint url")
	auth := flag.String("auth", "***", "basic authentication header")
	flag.Parse()

	req, err := http.NewRequest("GET", *url, nil)
	if err != nil {
		log.Fatalf("Creating request: [%v]", err)
	}
	req.Header.Add("Authorization", "Basic "+*auth)

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		log.Fatalf("Sending request: [%v]", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatalf("Received invalid statusCode: [%v]", resp.StatusCode)
	}

	br := bufio.NewReader(resp.Body)
	for {
		event, err := br.ReadString('\n')
		if err != nil {
			log.Printf("Error: [%v]", err)
			continue
		}
		trimmed := strings.TrimSpace(event)
		if trimmed == "" {
			continue
		}
		data := strings.TrimPrefix(trimmed, "data: ")
		var notification eventData
		err = json.Unmarshal([]byte(data), &notification)
		if err != nil {
			log.Printf("Error: [%v]. \n", err)
			continue
		}
		if len(notification) == 0 {
			log.Println("Received 'heartbeat' event")
			continue
		}
		log.Printf("Received notification: [%v]", notification)
	}
}
