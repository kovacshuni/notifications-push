package main

import (
        "bufio"
        "encoding/json"
        "log"
        "net/http"
        "time"
)

func main() {
        go fetchEvents()

        go distributeEvents()

        http.HandleFunc("/", handler)
        log.Fatal(http.ListenAndServe(":8080", nil))
}

var (
        incoming = make(chan string)

        addSubscriber    = make(chan chan string)
        removeSubscriber = make(chan chan string)
)

func fetchEvents() {
        // this simulates some incoming events that we wish to distribute to consumers
        go func() {
                for {
                        str, err := json.Marshal(struct {
                                Date time.Time `json:"date"`
                        }{time.Now()})
                        if err != nil {
                                panic(err)
                        }
                        incoming <- string(str)
                        time.Sleep(2 * time.Second)
                }
        }()
}

func distributeEvents() {
        subscribers := make(map[chan string]struct{})
        for {
                select {
                case msg := <-incoming:
                        for sub, _ := range subscribers {
                                select {
                                case sub <- msg:
                                default:
                                        log.Printf("listener too far behind - message dropped")
                                }
                        }
                case sub := <-addSubscriber:
                        subscribers[sub] = struct{}{}
                case sub := <-removeSubscriber:
                        delete(subscribers, sub)
                }
        }
}

func handler(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
        w.Header().Set("Pragma", "no-cache")
        w.Header().Set("Expires", "0")

        bw := bufio.NewWriter(w)

        events := make(chan string)

        addSubscriber <- events
        defer func() { removeSubscriber <- events }()

        for {
                bw.WriteString(<-events)
                bw.WriteByte('\n')
                bw.Flush()
                flusher := w.(http.Flusher)
                flusher.Flush()
        }
}
