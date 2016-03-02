package main

import (
	"log"
	"net/http"
	"github.com/jawher/mow.cli"
	"os"
"io"
)

const logPattern = log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile | log.LUTC

var infoLogger *log.Logger
var warnLogger *log.Logger
var errorLogger *log.Logger

func main() {
	app := cli.App("notifications-push", "Proactively notifies subscribers about new content publishes/modifications.")
	socksProxy := app.String(cli.StringOpt{
		Name:   "socks-proxy",
		Value:  "",
		Desc:   "Use specified SOCKS proxy (e.g. localhost:2323)",
		EnvVar: "SOCKS_PROXY",
	})
	println(socksProxy)
	app.Action = func() {
		initLogs(os.Stdout, os.Stdout, os.Stderr)

		dispatcher := NewEvents()
		go dispatcher.receiveEvents()
		go dispatcher.distributeEvents()

		controller := Controller{dispatcher}
		http.HandleFunc("/", controller.handler)
		log.Fatal(http.ListenAndServe(":8080", nil))
	}
	app.Run(os.Args)
}

func initLogs(infoHandle io.Writer, warnHandle io.Writer, errorHandle io.Writer) {
	infoLogger = log.New(infoHandle, "INFO  - ", logPattern)
	warnLogger = log.New(warnHandle, "WARN  - ", logPattern)
	errorLogger = log.New(errorHandle, "ERROR - ", logPattern)
}
